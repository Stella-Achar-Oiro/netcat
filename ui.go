// ui.go
package main

import (
    "fmt"
    "strings"
    "time"

    "github.com/jroimartin/gocui"
)

type ChatUI struct {
    gui         *gocui.Gui
    server      *Server
    msgView     string
    inputView   string
    statusView  string
    userView    string
    roomView    string
    helpView    string
    activeView  string
    showHelp    bool
    currentRoom string
}

func NewChatUI(server *Server) (*ChatUI, error) {
    g, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        return nil, err
    }

    ui := &ChatUI{
        gui:         g,
        server:      server,
        msgView:     "messages",
        inputView:   "input",
        statusView:  "status",
        userView:    "users",
        roomView:    "rooms",
        helpView:    "help",
        activeView:  "input",
        showHelp:    false,
        currentRoom: "general",
    }

    g.SetManagerFunc(ui.layout)
    return ui, nil
}

func (ui *ChatUI) layout(g *gocui.Gui) error {
    maxX, maxY := g.Size()
    
    sidebarWidth := 20
    msgWidth := maxX - sidebarWidth - 1
    msgHeight := maxY - 5
    roomHeight := 10

    // Messages view
    if v, err := g.SetView(ui.msgView, 0, 0, msgWidth, msgHeight); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Messages"
        v.Wrap = true
        v.Autoscroll = true
    }

    // Rooms view
    if v, err := g.SetView(ui.roomView, msgWidth+1, 0, maxX-1, roomHeight); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Rooms"
        v.Wrap = true
        ui.updateRooms()
    }

    // Users view
    if v, err := g.SetView(ui.userView, msgWidth+1, roomHeight+1, maxX-1, msgHeight); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Online Users"
        v.Wrap = true
        ui.updateUsers()
    }

    // Status bar
    if v, err := g.SetView(ui.statusView, 0, msgHeight+1, maxX-1, msgHeight+3); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Status"
        v.Wrap = true
        ui.updateStatus(fmt.Sprintf("Connected to port %s | Room: %s | Ctrl-H: Help", 
            ui.server.port, ui.currentRoom))
    }

    // Input field
    if v, err := g.SetView(ui.inputView, 0, msgHeight+3, maxX-1, maxY-1); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Input"
        v.Editable = true
        v.Wrap = true
        
        if _, err := g.SetCurrentView(ui.inputView); err != nil {
            return err
        }
    }

    // Help window
    if ui.showHelp {
        helpX1 := maxX/6
        helpY1 := maxY/6
        helpX2 := maxX*5/6
        helpY2 := maxY*5/6
        if v, err := g.SetView(ui.helpView, helpX1, helpY1, helpX2, helpY2); err != nil {
            if err != gocui.ErrUnknownView {
                return err
            }
            v.Title = "Help"
            fmt.Fprintln(v, `Commands:
/help           - Show this help
/list           - List online users
/nick <name>    - Change your nickname
/msg <user> <message> - Send private message
/join <room>    - Join a chat room
/rooms          - List available rooms
/create <room>  - Create a new room
/quit           - Leave chat

Keybindings:
Ctrl-C          - Quit
Ctrl-H          - Toggle help
Tab             - Switch views
Enter           - Send message`)
        }
    }

    return nil
}

func (ui *ChatUI) updateUsers() {
    ui.gui.Update(func(g *gocui.Gui) error {
        v, err := g.View(ui.userView)
        if err != nil {
            return err
        }
        v.Clear()

        ui.server.mutex.Lock()
        for _, client := range ui.server.clients {
            fmt.Fprintf(v, "%s (%s)\n", client.name, client.room)
        }
        ui.server.mutex.Unlock()
        return nil
    })
}

func (ui *ChatUI) updateRooms() {
    ui.gui.Update(func(g *gocui.Gui) error {
        v, err := g.View(ui.roomView)
        if err != nil {
            return err
        }
        v.Clear()

        ui.server.mutex.Lock()
        for name, room := range ui.server.rooms {
            prefix := "  "
            if name == ui.currentRoom {
                prefix = "* "
            }
            fmt.Fprintf(v, "%s%s (%d)\n", prefix, name, len(room.clients))
        }
        ui.server.mutex.Unlock()
        return nil
    })
}

func (ui *ChatUI) updateStatus(status string) {
    ui.gui.Update(func(g *gocui.Gui) error {
        v, err := g.View(ui.statusView)
        if err != nil {
            return err
        }
        v.Clear()
        fmt.Fprint(v, status)
        return nil
    })
}

func (ui *ChatUI) keybindings() error {
    // Quit
    if err := ui.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone,
        func(g *gocui.Gui, _ *gocui.View) error {
            return gocui.ErrQuit
        }); err != nil {
        return err
    }

    // Toggle help
    if err := ui.gui.SetKeybinding("", gocui.KeyCtrlH, gocui.ModNone,
        func(_ *gocui.Gui, _ *gocui.View) error {
            ui.showHelp = !ui.showHelp
            return nil
        }); err != nil {
        return err
    }

    // Send message
    if err := ui.gui.SetKeybinding(ui.inputView, gocui.KeyEnter, gocui.ModNone,
        ui.handleInput); err != nil {
        return err
    }

    // Switch views
    if err := ui.gui.SetKeybinding("", gocui.KeyTab, gocui.ModNone,
        func(g *gocui.Gui, v *gocui.View) error {
            nextView := map[string]string{
                ui.msgView:   ui.roomView,
                ui.roomView:  ui.userView,
                ui.userView:  ui.inputView,
                ui.inputView: ui.msgView,
            }
            if next, ok := nextView[v.Name()]; ok {
                ui.activeView = next
                _, err := g.SetCurrentView(next)
                return err
            }
            return nil
        }); err != nil {
        return err
    }

    return nil
}

func (ui *ChatUI) handleInput(_ *gocui.Gui, v *gocui.View) error {
    input := strings.TrimSpace(v.Buffer())
    if input == "" {
        v.Clear()
        v.SetCursor(0, 0)
        return nil
    }

    v.Clear()
    v.SetCursor(0, 0)

    // Create a mock client for UI commands
    client := &Client{
        name:     "Server",
        joinTime: time.Now(),
        room:     ui.currentRoom,
    }

    if strings.HasPrefix(input, "/") {
        ui.server.handleCommand(client, input)
        ui.updateRooms()
        ui.updateUsers()
    } else {
        msg := Message{
            Type:      MessageTypeChat,
            From:      "Server",
            Content:   input,
            Timestamp: time.Now(),
        }
        ui.server.broadcast(msg, nil)
    }

    return nil
}

func (ui *ChatUI) Run() error {
    if err := ui.keybindings(); err != nil {
        return err
    }

    if err := ui.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
        return err
    }

    return nil
}

func (ui *ChatUI) Close() {
    ui.gui.Close()
}