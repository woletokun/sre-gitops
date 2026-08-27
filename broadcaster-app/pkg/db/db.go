package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func Connect() (*DB, error) {
	host := getEnv("DB_HOST", "postgres-service")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "broadcaster_admin")
	password := getEnv("DB_PASSWORD", "BroadCasterSecurePass2026!")
	dbname := getEnv("DB_NAME", "broadcaster")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var db *sql.DB
	var err error

	// Retry connection up to 10 times to handle pod startup ordering
	for i := 1; i <= 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				fmt.Println("Successfully connected to PostgreSQL database")
				return &DB{db}, nil
			}
		}
		fmt.Printf("Waiting for database (attempt %d/10): %v\n", i, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after 10 attempts: %w", err)
}

func (db *DB) RegisterUser(username, email, passwordHash string) (string, error) {
	var id string
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`
	err := db.QueryRow(query, username, email, passwordHash).Scan(&id)
	return id, err
}

func (db *DB) GetUserByUsername(username string) (string, string, string, error) {
	var id, email, passwordHash string
	query := `SELECT id, email, password_hash FROM users WHERE username = $1`
	err := db.QueryRow(query, username).Scan(&id, &email, &passwordHash)
	return id, email, passwordHash, err
}

func (db *DB) SaveMessage(userID, username, content string) error {
	query := `INSERT INTO messages (user_id, username, content) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, userID, username, content)
	return err
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
