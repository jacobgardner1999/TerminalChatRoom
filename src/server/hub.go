package main

import (
    "log"
	"sync"
)

type Hub struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
}

func (h *Hub) RegisterClient(client *Client, roomName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

    if client == nil {
        log.Println("RegisterClient: Client is nil")
        return
    }

    room, ok := h.rooms[roomName]
    if !ok {
        room = NewRoom(roomName)
        h.rooms[roomName] = room
    }
    client.room = room
    room.RegisterClient(client)
}

func (h *Hub) UnregisterClient(client *Client, roomName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[roomName]; ok {
		room.UnregisterClient(client)
		if len(room.clients) == 0 {
			delete(h.rooms, roomName)
		}
	}
}

func (h *Hub) GetRoom(roomName string) *Room {
    h.mu.Lock()
    defer h.mu.Unlock()

    if roomName == "" {
        log.Println("GetRoom: Empty room name")
        return nil
    }

    room, ok := h.rooms[roomName]
    if !ok {
        room = NewRoom(roomName)
        h.rooms[roomName] = room
    }

    return room
}

func (h *Hub) BroadcastToRoom(roomName string, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[roomName]; ok {
		room.Broadcast(message)
	}
}
