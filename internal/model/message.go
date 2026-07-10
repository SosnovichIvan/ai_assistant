package model

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message.
type Message struct {
	ID          uuid.UUID `db:"id" json:"id"`
	RoomID      string    `db:"room_id" json:"room_id"`
	Sender      string    `db:"sender" json:"sender"`
	Content     string    `db:"content" json:"content"`
	MessageType string    `db:"message_type" json:"message_type"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// NewMessage creates a new message with generated ID.
func NewMessage(roomID, sender, content, msgType string) *Message {
	return &Message{
		ID:          uuid.New(),
		RoomID:      roomID,
		Sender:      sender,
		Content:     content,
		MessageType: msgType,
		CreatedAt:   time.Now(),
	}
}

// WebSocketMessage represents a message sent via WebSocket.
type WebSocketMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp,omitempty"`
}

// MessageType constants.
const (
	MessageTypeMessage = "message"
	MessageTypeTyping  = "typing"
	MessageTypeLeave   = "leave"
	MessageTypeJoin    = "join"
)
