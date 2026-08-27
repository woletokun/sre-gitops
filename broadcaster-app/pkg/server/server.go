package server

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"broadcaster-app/pkg/auth"
	"broadcaster-app/pkg/db"

	"github.com/gorilla/websocket"
)

//go:embed index.html
var content embed.FS

type Server struct {
	port       string
	db         *db.DB
	tcpClients map[net.Conn]string
	wsClients  map[*websocket.Conn]string
	broadcast  chan string
	mu         sync.Mutex
}

type AuthRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewServer(port string, database *db.DB) *Server {
	return &Server{
		port:       port,
		db:         database,
		tcpClients: make(map[net.Conn]string),
		wsClients:  make(map[*websocket.Conn]string),
		broadcast:  make(chan string),
	}
}

func (s *Server) Start() error {
	go s.startHTTPServer()

	listener, err := net.Listen("tcp", "0.0.0.0:"+s.port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Broadcast TCP engine listening on 0.0.0.0:%s\n", s.port)
	go s.handleMessages()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go s.handleTCPClient(conn)
	}
}

func (s *Server) startHTTPServer() {
	// Serve embedded Frontend UI at / and /broadcaster
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimSuffix(r.URL.Path, "/")
		if p == "" || p == "/broadcaster" {
			data, err := content.ReadFile("index.html")
			if err != nil {
				http.Error(w, "Could not load UI", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(data)
			return
		}
		http.NotFound(w, r)
	})

	http.HandleFunc("/api/register", s.handleRegister)
	http.HandleFunc("/api/login", s.handleLogin)
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK"}`))
	})

	fmt.Println("HTTP/WebSocket API & Web UI listening on port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Printf("HTTP server error: %v\n", err)
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	id, err := s.db.RegisterUser(req.Username, req.Email, hash)
	if err != nil {
		http.Error(w, "Registration failed: "+err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id, "message": "User registered successfully"})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	id, _, hash, err := s.db.GetUserByUsername(req.Username)
	if err != nil || !auth.CheckPasswordHash(req.Password, hash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(id, req.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token, "username": req.Username})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}

	s.mu.Lock()
	s.wsClients[ws] = username
	s.mu.Unlock()

	s.broadcast <- fmt.Sprintf("--> %s joined via Web UI", username)

	for {
		_, msgBytes, err := ws.ReadMessage()
		if err != nil {
			s.mu.Lock()
			delete(s.wsClients, ws)
			s.mu.Unlock()
			break
		}
		text := string(msgBytes)
		msg := fmt.Sprintf("[%s]: %s", username, text)

		go s.db.SaveMessage("", username, text)
		s.broadcast <- msg
	}
}

func (s *Server) handleTCPClient(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.tcpClients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	var username string

	fmt.Fprintln(conn, "Enter Username:")
	if scanner.Scan() {
		username = scanner.Text()
	} else {
		return
	}

	s.mu.Lock()
	s.tcpClients[conn] = username
	s.mu.Unlock()

	s.broadcast <- fmt.Sprintf("--> %s joined via CLI", username)

	for scanner.Scan() {
		text := scanner.Text()
		msg := fmt.Sprintf("[%s]: %s", username, text)

		go s.db.SaveMessage("", username, text)
		s.broadcast <- msg
	}
}

func (s *Server) handleMessages() {
	for msg := range s.broadcast {
		s.mu.Lock()
		for conn := range s.tcpClients {
			fmt.Fprintln(conn, msg)
		}
		for ws := range s.wsClients {
			ws.WriteMessage(websocket.TextMessage, []byte(msg))
		}
		s.mu.Unlock()
	}
}
