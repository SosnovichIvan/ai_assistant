package websocket

import (
	"log/slog"
	"sync"
)

// Room represents a chat room.
type Room struct {
	ID       string
	clients  map[*Client]bool
	broadcast chan []byte
	join     chan *Client
	leave    chan *Client
	mutex    sync.RWMutex
	logger   *slog.Logger
}

// NewRoom creates a new chat room.
func NewRoom(id string, logger *slog.Logger) *Room {
	return &Room{
		ID:       id,
		clients:  make(map[*Client]bool),
		broadcast: make(chan []byte, 256),
		join:     make(chan *Client),
		leave:    make(chan *Client),
		logger:   logger,
	}
}

// Run starts the room's main loop.
func (r *Room) Run() {
	for {
		select {
		case client := <-r.join:
			r.mutex.Lock()
			r.clients[client] = true
			r.mutex.Unlock()
			r.logger.Info("client joined room",
				slog.String("room", r.ID),
				slog.String("user", client.UserID),
			)

		case client := <-r.leave:
			r.mutex.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.Send)
			}
			r.mutex.Unlock()
			r.logger.Info("client left room",
				slog.String("room", r.ID),
				slog.String("user", client.UserID),
			)

		case message := <-r.broadcast:
			r.mutex.RLock()
			for client := range r.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(r.clients, client)
				}
			}
			r.mutex.RUnlock()
		}
	}
}

// Join adds a client to the room.
func (r *Room) Join(client *Client) {
	r.join <- client
}

// Leave removes a client from the room.
func (r *Room) Leave(client *Client) {
	r.leave <- client
}

// Broadcast sends a message to all clients in the room.
func (r *Room) Broadcast(msg *Message) {
	data := msg.ToJSON()
	r.broadcast <- data
}

// Empty returns true if the room has no clients.
func (r *Room) Empty() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.clients) == 0
}

// ClientCount returns the number of clients in the room.
func (r *Room) ClientCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.clients)
}
