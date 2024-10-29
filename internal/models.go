package internal

import (
	"net"
	"time"
)

// Client represents a connected chat client
type Client struct {
	conn     net.Conn
	name     string
	joinTime time.Time
	room     string // Current room name
}

// Message represents a chat message
type Message struct {
	Type      int
	From      string
	To        string // For private messages
	Content   string
	Timestamp time.Time
}

// Message types for different kinds of messages
const (
	MessageTypeChat = iota
	MessageTypeSystem
	MessageTypePrivate
	MessageTypeError
)