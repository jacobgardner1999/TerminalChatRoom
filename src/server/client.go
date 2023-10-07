package main

import (
    "fmt"
    "strings"
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
    writeWait = 10 * time.Second
    pongWait = 60 * time.Second
    pingPeriod = (pongWait * 9) / 10
    maxMessageSize = 512
)

var (
    newline = []byte{'\n'}
    space = []byte{' '}
)

var upgrader = websocket.Upgrader{
    ReadBufferSize: 1024,
    WriteBufferSize: 1024,
}

type Client struct {
    hub *Hub
    conn *websocket.Conn
    send chan []byte
    username string
    room *Room
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		c.room.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		if bytes.HasPrefix(message, []byte("/")) {
            if !c.handleCommand(string(message), hub) {
                log.Printf("Unknown command: %s", message)
		    } 
        } else {
	        if c.room != nil {
                c.room.Broadcast(c.username, message)
            } else {
                log.Println("Client's room is not set")
            }
        }
}
}

func (c *Client) handleCommand(command string, hub *Hub) bool {
	parts := strings.Fields(command)
	if len(parts) < 1 {
		return false
	}

	switch parts[0] {
	case "/join":
		return c.handleJoinCommand(parts, hub)
	case "/name":
		return c.handleNameCommand(parts)
    case "/rooms":
        return c.handleRoomsCommand(hub)
	default:
		return false
	}
}

func (c *Client) handleJoinCommand(parts []string, hub *Hub) bool {
    if len(parts) != 2 {
		log.Println("Invalid /join command format")
		return false
	}

	roomName := parts[1]

    c.hub.RegisterClient(c, roomName)
	c.room = hub.GetRoom(roomName)

	c.room.Broadcast("Server", []byte(fmt.Sprintf("%s joined the room", c.username)))
    message := fmt.Sprintf("/userRoom %s", c.room.name) 
    c.send <- []byte(message)
    time.Sleep(5 * time.Millisecond)
    c.room.BroadcastChatLog(c)

	return true
}

func (c *Client) handleNameCommand(parts []string) bool {
    if len(parts) != 2 {
        log.Println("Invalid /name command format")
        return false
    }

    newName := parts[1]
    oldName := c.username

    c.username = newName
    c.room.Broadcast("Server", []byte(fmt.Sprintf("%s set their name to %s", oldName, newName)))

    return true
}

func (c *Client) handleRoomsCommand(hub *Hub) bool {
    var roomList []string

	for roomName := range hub.rooms {
		roomList = append(roomList, roomName)
	}

	message := "Room List: " + strings.Join(roomList, ", ")

    currentTime := time.Now().Format("15:04")
    c.send <- []byte("Server|" + currentTime + "|" + message)

    return true
}

func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)
            n := len(c.send)
            for i := 0; i < n; i++ {
                w.Write(newline)
                w.Write(<-c.send)
            }

            if err := w.Close(); err != nil {
                return
            }
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    client := &Client{hub: hub, username: "New User", conn: conn, send: make(chan []byte, 256)}
    hub.RegisterClient(client, "waitingRoom")
    
    go client.writePump()
    go client.readPump(hub)
}
