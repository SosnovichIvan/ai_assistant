package websocket

import (
	"encoding/json"
	"log/slog"
	"time"

	"ai_assistant/internal/model"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client represents a WebSocket client.
type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	RoomID string
	UserID string
	Send   chan []byte
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *websocket.Conn, roomID, userID string) *Client {
	return &Client{
		Hub:    hub,
		Conn:   conn,
		RoomID: roomID,
		UserID: userID,
		Send:   make(chan []byte, 256),
	}
}

// ReadPump handles incoming WebSocket messages.
func (c *Client) ReadPump(logger *slog.Logger) {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("websocket read error", slog.Any("error", err))
			}
			break
		}

		var msg model.WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Warn("invalid message format", slog.String("message", string(message)))
			continue
		}

		msg.Sender = c.UserID

		// Broadcast to room
		c.Hub.Broadcast(c.RoomID, msg.Sender, msg.Content, msg.Type)
	}
}

// WritePump handles outgoing WebSocket messages.
func (c *Client) WritePump(logger *slog.Logger) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
