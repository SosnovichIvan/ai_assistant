package handler

import (
	"log/slog"
	"net/http"

	ws "ai_assistant/internal/websocket"

	"github.com/go-chi/chi/v5"
)

// WebSocketHandler handles WebSocket connections.
type WebSocketHandler struct {
	hub    *ws.Hub
	logger *slog.Logger
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(hub *ws.Hub, logger *slog.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

// Handle handles WebSocket upgrade and connection.
func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "room_id")
	userID := r.URL.Query().Get("user_id")

	if roomID == "" {
		http.Error(w, "room_id required", http.StatusBadRequest)
		return
	}
	if userID == "" {
		userID = "anonymous"
	}

	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", slog.Any("error", err))
		return
	}

	client := ws.NewClient(h.hub, conn, roomID, userID)
	h.hub.Register(client)

	go client.WritePump(h.logger)
	go client.ReadPump(h.logger)
}
