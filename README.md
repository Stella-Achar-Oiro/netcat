# NetCat Chat Server

![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

A TCP-based chat server implementation in Go that recreates NetCat functionality with enhanced features. This project implements a Server-Client Architecture that supports multiple clients, chat rooms, and real-time message broadcasting.

## 🌟 Features

- 🔄 TCP connection between server and multiple clients
- 👥 Support for up to 10 concurrent clients
- 💬 Real-time message broadcasting
- 🏷️ Username validation and management
- 📝 Message history for new clients
- 🚪 Join/leave notifications
- 💌 Private messaging
- 🏠 Multiple chat rooms
- 📜 Activity logging

## 🚀 Getting Started

### Prerequisites

- Go 1.16 or higher
- netcat (nc) for client connections

### Installation

1. Clone the repository:
```bash
git clone https://github.com/Stella-Achar-Oiro/netcat.git
cd netcat
```

2. Build the project:
```bash
# Using build script
chmod +x build.sh
./build.sh --build

# Or using Go directly
go build -o TCPChat
```

### Running the Server

```bash
# Default port (8989)
./TCPChat

# Custom port
./TCPChat 2525
```

### Connecting as a Client

```bash
# Connect to default port
nc localhost 8989

# Connect to custom port
nc localhost 2525
```

## 🎮 Usage

### Available Commands

```
/help           - Show available commands
/list           - Show online users
/nick <name>    - Change your nickname
/msg <user> <message> - Send private message
/join <room>    - Join a chat room
/rooms          - List available rooms
/create <room>  - Create a new room
/quit           - Leave chat
```

### Example Session

1. Start the server:
```bash
./TCPChat
# Server starts listening on port 8989
```

2. Connect as a client:
```bash
nc localhost 8989
# You'll see the welcome logo
# Enter your name when prompted
```

3. Chat commands:
```
/list           # See who's online
/msg Alice Hi!  # Send private message to Alice
/create room1   # Create a new chat room
/join room1     # Join a chat room
```

## 🏗️ Project Structure

```
├── main.go         # Main server implementation
├── ui.go          # Terminal UI implementation
├── main_test.go   # Test suite
└── build.sh       # Build script
```

## 🧪 Testing

Run the test suite:
```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestServerStartup

# Run with race detector
go test -v -race
```

## 🛠️ Build Options

The build script provides several options:

```bash
./build.sh --help          # Show help
./build.sh --clean        # Clean project
./build.sh --test         # Run tests
./build.sh --build        # Build for current platform
./build.sh --all         # Build for all platforms
./build.sh --run         # Build and run server
```

## 📝 Message Format

Messages are formatted as:
```
[2024-01-20 15:48:41][username]: message
```

System messages:
```
[2024-01-20 15:48:41] username has joined our chat...
[2024-01-20 15:48:41] username has left our chat...
```

## ⚡ Features in Detail

### Message Broadcasting
- All messages are broadcast to all clients in the same room
- Messages include timestamps and sender information
- Empty messages are not broadcast

### Chat History
- New clients receive all previous messages upon joining
- History is maintained per room
- System messages are included in history

### User Management
- Usernames must be unique
- Name changes are broadcast to all users
- User list is maintained and available via `/list`

### Room Management
- Multiple chat rooms supported
- Room creation restricted to existing users
- Room membership tracking
- Room-specific message broadcasting

## 🔍 Logging

The server maintains a log file (`chat.log`) containing:
- Server start/stop events
- Client connections/disconnections
- Username changes
- Room creation/deletion
- Error events

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Authors

* **Stella Achar-Oiro** - *Initial work* - [Stella-Achar-Oiro](https://github.com/Stella-Achar-Oiro)

## 🙏 Acknowledgments

* Inspired by the NetCat utility
* Thanks to all contributors and testers
