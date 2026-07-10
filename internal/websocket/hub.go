package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"ai_assistant/internal/model"
	"ai_assistant/internal/repository"
)

// Hub maintains WebSocket connections and chat rooms.
type Hub struct {
	rooms    map[string]*Room
	register chan *Client
	unregister chan *Client
	broadcast chan *Message
	mutex    sync.RWMutex
	logger   *slog.Logger
	msgRepo  *repository.MessageRepository
}

// NewHub creates a new WebSocket hub.
func NewHub(logger *slog.Logger, msgRepo *repository.MessageRepository) *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
		logger:     logger,
		msgRepo:    msgRepo,
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			room := h.rooms[client.RoomID]
			if room == nil {
				room = NewRoom(client.RoomID, h.logger)
				h.rooms[client.RoomID] = room
				go room.Run()
			}
			room.Join(client)
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if room, ok := h.rooms[client.RoomID]; ok {
				room.Leave(client)
				if room.Empty() {
					delete(h.rooms, client.RoomID)
				}
			}
			h.mutex.Unlock()

		case msg := <-h.broadcast:
			h.mutex.RLock()
			if room, ok := h.rooms[msg.RoomID]; ok {
				room.Broadcast(msg)
			}
			h.mutex.RUnlock()
		}
	}
}

// Broadcast sends a message to a room.
func (h *Hub) Broadcast(roomID, sender, content, msgType string) {
	msg := &Message{
		RoomID:   roomID,
		Sender:   sender,
		Content:  content,
		Type:     msgType,
		SentAt:   time.Now(),
	}
	h.broadcast <- msg

	// Persist message
	if msgType == model.MessageTypeMessage {
		m := model.NewMessage(roomID, sender, content, msgType)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.msgRepo.Create(ctx, m)
	}
}

// Register registers a client.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Message represents a chat message.
type Message struct {
	RoomID  string    `json:"room_id"`
	Sender  string    `json:"sender"`
	Content string    `json:"content"`
	Type    string    `json:"type"`
	SentAt  time.Time `json:"sent_at"`
}

// ToJSON converts message to JSON.
func (m *Message) ToJSON() []byte {
	data, _ := json.Marshal(m)
	return data
}
