package internal

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// CommandFunc represents a command handler function
type CommandFunc func(s *Server, c *Client, args []string) error

// Server represents the chat server
type Server struct {
	clients    map[net.Conn]*Client
	mutex      sync.Mutex
	messages   []Message
	maxClients int
	Logfile    *os.File
	rooms      map[string]*ChatRoom
	commands   map[string]CommandFunc
	port       string
}

// Logo constant
const Logo = `Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    '.       | '\ \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     '-'       '--'
[ENTER YOUR NAME]:`

func NewServer() *Server {
	Logfile, err := os.OpenFile("chat.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
	}

	s := &Server{
		clients:    make(map[net.Conn]*Client),
		messages:   []Message{},
		maxClients: 10,
		Logfile:    Logfile,
		rooms:      make(map[string]*ChatRoom),
		commands:   make(map[string]CommandFunc),
	}

	// Create default room
	s.rooms["general"] = &ChatRoom{
		name:     "general",
		clients:  make(map[net.Conn]*Client),
		messages: []Message{},
	}

	// Register commands
	s.registerCommands()
	return s
}

func (s *Server) registerCommands() {
	s.commands = map[string]CommandFunc{
		"help": func(s *Server, c *Client, args []string) error {
			help := `Available commands:
/help           - Show this help
/list           - List online users
/nick <name>    - Change your nickname
/msg <user> <message> - Send private message
/who            - Show users in current room
`
			c.conn.Write([]byte(help))
			return nil
		},

		"list": func(s *Server, c *Client, args []string) error {
			s.mutex.Lock()
			var users []string
			for _, client := range s.clients {
				users = append(users, fmt.Sprintf("%s (in %s)", client.name, client.room))
			}
			s.mutex.Unlock()
			response := fmt.Sprintf("Online users (%d):\n%s\n",
				len(users), strings.Join(users, "\n"))
			c.conn.Write([]byte(response))
			return nil
		},

		"nick": func(s *Server, c *Client, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: /nick <new_name>")
			}
			newName := args[0]
			if err := s.ValidateName(newName); err != nil {
				return err
			}
			oldName := c.name
			c.name = newName
			s.broadcast(Message{
				Type:      MessageTypeSystem,
				Content:   fmt.Sprintf("%s changed name to %s", oldName, newName),
				Timestamp: time.Now(),
			}, nil)
			return nil
		},

		"join": func(s *Server, c *Client, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: /join <room>")
			}
			return s.joinRoom(c, args[0])
		},

		"create": func(s *Server, c *Client, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: /create <room>")
			}
			return s.createRoom(c, args[0])
		},

		"rooms": func(s *Server, c *Client, args []string) error {
			return s.listRooms(c)
		},

		"msg": func(s *Server, c *Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: /msg <user> <message>")
			}
			return s.sendPrivateMessage(c, args[0], strings.Join(args[1:], " "))
		},

		"who": func(s *Server, c *Client, args []string) error {
			if c.room == "" {
				return fmt.Errorf("you are not in any room")
			}
			room := s.rooms[c.room]
			var users []string
			for _, client := range room.clients {
				users = append(users, client.name)
			}
			response := fmt.Sprintf("Users in room %s (%d):\n%s\n",
				c.room, len(users), strings.Join(users, ", "))
			c.conn.Write([]byte(response))
			return nil
		},
	}
}

func (s *Server) logActivity(message string) {
	if s.Logfile != nil {
		fmt.Fprintf(s.Logfile, "[%s] %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			message)
	}
}

func (s *Server) broadcast(msg Message, exclude net.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.messages = append(s.messages, msg)
	formatted := formatMessage(msg)

	for conn := range s.clients {
		if conn != exclude {
			conn.Write([]byte(formatted + "\n"))
		}
	}
}

func (s *Server) handleCommand(client *Client, message string) bool {
	if !strings.HasPrefix(message, "/") {
		return false
	}

	parts := strings.Fields(message)
	command := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	handler, exists := s.commands[command]
	if !exists {
		client.sendMessage(Message{
			Type:      MessageTypeError,
			Content:   "Unknown command. Type /help for available commands.",
			Timestamp: time.Now(),
		})
		return true
	}

	if err := handler(s, client, args); err != nil {
		client.sendMessage(Message{
			Type:      MessageTypeError,
			Content:   err.Error(),
			Timestamp: time.Now(),
		})
	}
	return true
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Send welcome message
	_, err := conn.Write([]byte(Logo))
	if err != nil {
		log.Printf("Error sending logo: %v", err)
		return
	}

	reader := bufio.NewReader(conn)

	// Get and validate client name
	var name string
	for {
		nameBytes, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading name: %v", err)
			return
		}

		name = strings.TrimSpace(nameBytes)
		if err := s.ValidateName(name); err != nil {
			conn.Write([]byte(fmt.Sprintf("Invalid name: %s\nPlease enter another name: ", err)))
			continue
		}
		break
	}

	client := &Client{
		conn:     conn,
		name:     name,
		joinTime: time.Now(),
	}

	// Add client to server and default room
	s.mutex.Lock()
	s.clients[conn] = client
	s.mutex.Unlock()

	// Join default room
	s.joinRoom(client, "general")

	// Message handling loop
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}

		// Handle commands
		if s.handleCommand(client, message) {
			continue
		}

		// Regular message handling
		if client.room != "" {
			room := s.rooms[client.room]
			s.broadcastToRoom(room, Message{
				Type:      MessageTypeChat,
				From:      client.name,
				Content:   message,
				Timestamp: time.Now(),
			}, nil)
		}
	}

	// Handle disconnection
	s.mutex.Lock()
	delete(s.clients, conn)
	if client.room != "" {
		if room, exists := s.rooms[client.room]; exists {
			delete(room.clients, conn)
		}
	}
	s.mutex.Unlock()

	s.broadcast(Message{
		Type:      MessageTypeSystem,
		Content:   fmt.Sprintf("%s has left our chat...", client.name),
		Timestamp: time.Now(),
	}, nil)
	s.logActivity(fmt.Sprintf("User left: %s", client.name))
}

func (s *Server) Start(port string) error {
	s.port = port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Listening on the port :%s\n", port)
	s.logActivity("Server started on port " + port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		s.mutex.Lock()
		if len(s.clients) >= s.maxClients {
			s.mutex.Unlock()
			conn.Write([]byte("Chat is full. Please try again later.\n"))
			conn.Close()
			continue
		}
		s.mutex.Unlock()

		go s.handleConnection(conn)
	}
}

func (s *Server) sendPrivateMessage(from *Client, toName, content string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var to *Client
	for _, c := range s.clients {
		if c.name == toName {
			to = c
			break
		}
	}

	if to == nil {
		return fmt.Errorf("user %s not found", toName)
	}

	msg := Message{
		Type:      MessageTypePrivate,
		From:      from.name,
		To:        to.name,
		Content:   content,
		Timestamp: time.Now(),
	}

	to.sendMessage(msg)
	from.sendMessage(msg)
	s.logActivity(fmt.Sprintf("Private message: %s -> %s: %s",
		from.name, to.name, content))
	return nil
}
