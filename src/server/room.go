package main

import (
	"fmt"
	"sync"
	"time"
)

type Room struct {
	name    string
	clients map[*Client]bool
    chatLog []string
	mu      sync.Mutex 
}

func NewRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[*Client]bool),
        chatLog: make([]string, 0),
	}
}

func (r *Room) RegisterClient(client *Client) {
	r.clients[client] = true
}

func (r *Room) BroadcastChatLog(client *Client) {
    for _, message := range r.chatLog {
        client.send <- []byte(message)
        time.Sleep(50 * time.Millisecond)
    }
}

func (r *Room) UnregisterClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, client)
}

func (r *Room) Broadcast(username string, content []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

    currentTime := time.Now()
    formattedTime := currentTime.Format("15:04")

    message := fmt.Sprintf("%s|%s|%s", username, formattedTime, string(content))
    r.chatLog = append(r.chatLog, message)
	for client := range r.clients {
    
		select {
		case client.send <- []byte(message):
		default:
			close(client.send)
			delete(r.clients, client)
		}
    }
}

