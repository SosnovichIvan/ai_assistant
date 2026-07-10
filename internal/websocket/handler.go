package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader converts HTTP connections to WebSocket connections.
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure for production
	},
}
