// main.go
package main

import (
	"fmt"
	"log"
	"os"

	"netcat/internal"
)

// Message types for different kinds of messages
const (
	MessageTypeChat = iota
	MessageTypeSystem
	MessageTypePrivate
	MessageTypeError
)

func main() {
	// Parse command line arguments
	port := "8989" // default port
	useUI := false

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-ui":
			useUI = true
		default:
			if len(os.Args) > 2 {
				fmt.Println("[USAGE]: ./TCPChat $port")
				return
			}
			port = os.Args[i]
		}
	}

	// Create and start server
	server := internal.NewServer()
	defer server.Logfile.Close()

	if useUI {
		if err := internal.RunWithUI(server); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := server.Start(port); err != nil {
			log.Fatal(err)
		}
	}
}
