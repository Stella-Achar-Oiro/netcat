package internal

import (
	"fmt"
	"log"
	"strings"
)

func formatMessage(msg Message) string {
	timestamp := msg.Timestamp.Format("2006-01-02 15:04:05")
	switch msg.Type {
	case MessageTypePrivate:
		return fmt.Sprintf("[%s][PM from %s]: %s", timestamp, msg.From, msg.Content)
	case MessageTypeSystem:
		return fmt.Sprintf("[%s] %s", timestamp, msg.Content)
	case MessageTypeError:
		return fmt.Sprintf("[%s][ERROR] %s", timestamp, msg.Content)
	default:
		return fmt.Sprintf("[%s][%s]: %s", timestamp, msg.From, msg.Content)
	}
}

func RunWithUI(server *Server) error {
	ui, err := NewChatUI(server)
	if err != nil {
		return err
	}
	defer ui.Close()

	// Start server in goroutine
	go func() {
		if err := server.Start("8989"); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Run UI
	return ui.Run()
}

func (c *Client) sendMessage(msg Message) {
	formatted := formatMessage(msg)
	c.conn.Write([]byte(formatted + "\n"))
}

func (s *Server) isNameTaken(name string) bool {
	for _, client := range s.clients {
		if strings.EqualFold(client.name, name) {
			return true
		}
	}
	return false
}

func (s *Server) ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) < 2 {
		return fmt.Errorf("name too short (minimum 2 characters)")
	}
	if len(name) > 20 {
		return fmt.Errorf("name too long (maximum 20 characters)")
	}
	if s.isNameTaken(name) {
		return fmt.Errorf("name already taken")
	}
	return nil
}
