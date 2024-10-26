// main_test.go
package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	dialTimeout      = 2 * time.Second
	messageTimeout   = 500 * time.Millisecond
	serverStartDelay = 100 * time.Millisecond
	defaultTestPort  = "8989"
)

type TestClient struct {
	conn   net.Conn
	reader *bufio.Reader
}

func newTestClient(t *testing.T, address string) (*TestClient, error) {
	dialer := net.Dialer{Timeout: dialTimeout}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("could not connect to server: %v", err)
	}

	return &TestClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *TestClient) expectMessage(t *testing.T, expected string) error {
	deadline := time.Now().Add(messageTimeout)
	c.conn.SetReadDeadline(deadline)

	for {
		msg, err := c.reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read message: %v", err)
		}
		msg = strings.TrimSpace(msg)
		if strings.Contains(msg, expected) {
			return nil
		}
	}
}

func (c *TestClient) sendMessage(message string) error {
	_, err := c.conn.Write([]byte(message + "\n"))
	return err
}

func (c *TestClient) close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func setupTestServer(t *testing.T, port string) error {
	// Validate port number before attempting to start server
	if portNum, err := strconv.Atoi(port); err != nil || portNum < 0 || portNum > 65535 {
		return fmt.Errorf("invalid port number: %s", port)
	}

	errChan := make(chan error, 1)
	readyChan := make(chan bool, 1)

	go func() {
		s := NewServer()
		defer s.logFile.Close()

		// Send ready signal before starting the server
		readyChan <- true

		err := s.Start(port)
		if err != nil {
			errChan <- err
			return
		}
	}()

	// Wait for either ready signal or error
	select {
	case <-readyChan:
		time.Sleep(serverStartDelay)
		// Try to connect to verify server is actually running
		conn, err := net.DialTimeout("tcp", "localhost:"+port, dialTimeout)
		if err != nil {
			return fmt.Errorf("server started but connection failed: %v", err)
		}
		conn.Close()
		return nil
	case err := <-errChan:
		return err
	case <-time.After(dialTimeout):
		return fmt.Errorf("server startup timeout")
	}
}

func TestServerStartup(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"DefaultPort", defaultTestPort, false},
		{"CustomPort", "9090", false},
		{"InvalidPort", "99999", true},
		{"NegativePort", "-1", true},
		{"NonNumericPort", "abc", true},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			err := setupTestServer(t, tt.port)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for port %s, got nil", tt.port)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for port %s: %v", tt.port, err)
			}

			client, err := newTestClient(t, "localhost:"+tt.port)
			if err != nil {
				t.Fatalf("Client connection failed: %v", err)
			}
			defer client.close()

			if err := client.expectMessage(t, "Welcome"); err != nil {
				t.Errorf("Welcome message test failed: %v", err)
			}
		})
	}
}

func TestClientConnection(t *testing.T) {
	err := setupTestServer(t, "8991")
	if err != nil {
		t.Fatalf("Server setup failed: %v", err)
	}

	t.Run("SingleClient", func(t *testing.T) {
		client, err := newTestClient(t, "localhost:8991")
		if err != nil {
			t.Fatalf("Client connection failed: %v", err)
		}
		defer client.close()

		if err := client.expectMessage(t, "Welcome"); err != nil {
			t.Fatalf("Welcome message failed: %v", err)
		}
		if err := client.sendMessage("TestUser1"); err != nil {
			t.Fatalf("Send message failed: %v", err)
		}
		if err := client.expectMessage(t, "joined"); err != nil {
			t.Fatalf("Join message failed: %v", err)
		}
	})

	t.Run("MultipleClients", func(t *testing.T) {
		clients := make([]*TestClient, 0)
		defer func() {
			for _, c := range clients {
				c.close()
			}
		}()

		for i := 0; i < 3; i++ {
			client, err := newTestClient(t, "localhost:8991")
			if err != nil {
				t.Fatalf("Client %d connection failed: %v", i, err)
			}
			clients = append(clients, client)

			if err := client.expectMessage(t, "Welcome"); err != nil {
				t.Fatalf("Welcome message failed for client %d: %v", i, err)
			}
			if err := client.sendMessage(fmt.Sprintf("User%d", i)); err != nil {
				t.Fatalf("Send message failed for client %d: %v", i, err)
			}
			if err := client.expectMessage(t, "joined"); err != nil {
				t.Fatalf("Join message failed for client %d: %v", i, err)
			}
		}
	})
}

func TestMessageBroadcast(t *testing.T) {
	err := setupTestServer(t, "8992")
	if err != nil {
		t.Fatalf("Server setup failed: %v", err)
	}

	client1, err := newTestClient(t, "localhost:8992")
	if err != nil {
		t.Fatalf("Client1 connection failed: %v", err)
	}
	defer client1.close()

	client2, err := newTestClient(t, "localhost:8992")
	if err != nil {
		t.Fatalf("Client2 connection failed: %v", err)
	}
	defer client2.close()

	// Setup clients
	setupClients := func() error {
		if err := client1.expectMessage(t, "Welcome"); err != nil {
			return fmt.Errorf("client1 welcome failed: %v", err)
		}
		if err := client1.sendMessage("User1"); err != nil {
			return fmt.Errorf("client1 name failed: %v", err)
		}
		if err := client2.expectMessage(t, "Welcome"); err != nil {
			return fmt.Errorf("client2 welcome failed: %v", err)
		}
		if err := client2.sendMessage("User2"); err != nil {
			return fmt.Errorf("client2 name failed: %v", err)
		}
		return nil
	}

	if err := setupClients(); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	testMessage := "Hello everyone!"
	if err := client1.sendMessage(testMessage); err != nil {
		t.Fatalf("Send message failed: %v", err)
	}
	if err := client2.expectMessage(t, testMessage); err != nil {
		t.Fatalf("Broadcast failed: %v", err)
	}
}

func TestDisconnect(t *testing.T) {
	err := setupTestServer(t, "8993")
	if err != nil {
		t.Fatalf("Server setup failed: %v", err)
	}

	client1, err := newTestClient(t, "localhost:8993")
	if err != nil {
		t.Fatalf("Client1 connection failed: %v", err)
	}

	client2, err := newTestClient(t, "localhost:8993")
	if err != nil {
		t.Fatalf("Client2 connection failed: %v", err)
	}
	defer client2.close()

	// Setup clients
	if err := client1.expectMessage(t, "Welcome"); err != nil {
		t.Fatalf("Client1 welcome failed: %v", err)
	}
	if err := client1.sendMessage("User1"); err != nil {
		t.Fatalf("Client1 name failed: %v", err)
	}
	if err := client2.expectMessage(t, "Welcome"); err != nil {
		t.Fatalf("Client2 welcome failed: %v", err)
	}
	if err := client2.sendMessage("User2"); err != nil {
		t.Fatalf("Client2 name failed: %v", err)
	}

	// Test disconnection
	client1.close()
	if err := client2.expectMessage(t, "has left"); err != nil {
		t.Fatalf("Disconnect message failed: %v", err)
	}
}
