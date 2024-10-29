package internal

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// ChatRoom represents a separate chat room
type ChatRoom struct {
	name     string
	clients  map[net.Conn]*Client
	messages []Message
}

func (s *Server) broadcastToRoom(room *ChatRoom, msg Message, exclude net.Conn) {
	room.messages = append(room.messages, msg)
	formatted := formatMessage(msg)

	for conn := range room.clients {
		if conn != exclude {
			conn.Write([]byte(formatted + "\n"))
		}
	}
}

func (s *Server) createRoom(c *Client, roomName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.rooms[roomName]; exists {
		return fmt.Errorf("room already exists")
	}

	s.rooms[roomName] = &ChatRoom{
		name:     roomName,
		clients:  make(map[net.Conn]*Client),
		messages: []Message{},
	}

	s.logActivity(fmt.Sprintf("Room created: %s by %s", roomName, c.name))
	return s.joinRoom(c, roomName)
}

func (s *Server) joinRoom(c *Client, roomName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, exists := s.rooms[roomName]
	if !exists {
		return fmt.Errorf("room does not exist")
	}

	// Remove from current room if any
	if c.room != "" {
		if oldRoom, exists := s.rooms[c.room]; exists {
			delete(oldRoom.clients, c.conn)
		}
	}

	// Add to new room
	room.clients[c.conn] = c
	c.room = roomName

	// Send room history
	for _, msg := range room.messages {
		c.sendMessage(msg)
	}

	s.broadcastToRoom(room, Message{
		Type:      MessageTypeSystem,
		Content:   fmt.Sprintf("%s joined the room", c.name),
		Timestamp: time.Now(),
	}, nil)

	return nil
}

func (s *Server) listRooms(c *Client) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var rooms []string
	for name, room := range s.rooms {
		rooms = append(rooms, fmt.Sprintf("%s (%d users)",
			name, len(room.clients)))
	}

	response := fmt.Sprintf("Available rooms:\n%s\n",
		strings.Join(rooms, "\n"))
	c.conn.Write([]byte(response))
	return nil
}
