package main

import (
    "fmt"
	"sync"
	"time"
)

type Room struct {
	name    string
	clients map[*Client]bool
	mu      sync.Mutex 
}

func NewRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[*Client]bool),
	}
}

func (r *Room) RegisterClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client] = true
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

	for client := range r.clients {
		select {
		case client.send <- []byte(message):
		default:
			close(client.send)
			delete(r.clients, client)
		}
    }
}

