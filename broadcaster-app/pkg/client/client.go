package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("could not connect to server: %w", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to broadcast server at %s\nType messages below:\n", address)

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		fmt.Fprintln(conn, text)
	}

	return nil
}
