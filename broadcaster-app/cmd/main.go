package main

import (
	"fmt"
	"os"

	"broadcaster-app/pkg/client"
	"broadcaster-app/pkg/db"
	"broadcaster-app/pkg/server"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start":
		port := "8080"
		if len(os.Args) >= 3 {
			port = os.Args[2]
		}

		database, err := db.Connect()
		if err != nil {
			fmt.Printf("Database connection failed: %v\n", err)
			os.Exit(1)
		}

		srv := server.NewServer(port, database)
		if err := srv.Start(); err != nil {
			fmt.Printf("Server failure: %v\n", err)
			os.Exit(1)
		}
	case "connect":
		addr := "127.0.0.1:8080"
		if len(os.Args) >= 3 {
			addr = os.Args[2]
		}
		if err := client.Connect(addr); err != nil {
			fmt.Printf("Client error: %v\n", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  broadcast-server start [port]        # Default port: 8080")
	fmt.Println("  broadcast-server connect [host:port] # Default host: 127.0.0.1:8080")
}
