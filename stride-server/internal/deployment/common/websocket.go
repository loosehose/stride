package common

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/loosehose/stride/stride-server/logging"
	"github.com/rs/zerolog/log"
)

func init() {
	logging.InitLogger()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust as necessary for production
	},
}

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	connections sync.Map
	broadcast   chan []byte
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
}

// NewWebSocketManager creates a new instance of WebSocketManager
func NewWebSocketManager() *WebSocketManager {
	wsm := &WebSocketManager{
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}

	go wsm.run()

	return wsm
}

// run manages the WebSocket connections
func (wsm *WebSocketManager) run() {
	connections := make(map[*websocket.Conn]bool)

	for {
		select {
		case conn := <-wsm.register:
			connections[conn] = true

		case conn := <-wsm.unregister:
			if _, ok := connections[conn]; ok {
				delete(connections, conn)
				conn.Close()
			}

		case message := <-wsm.broadcast:
			for conn := range connections {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Debug().Msgf("Error broadcasting message: %s", err)
					conn.Close()
					delete(connections, conn)
				}
			}
		}
	}
}

// Handler upgrades HTTP to WebSocket and handles incoming connections
func (wsm *WebSocketManager) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Debug().Msgf("Error upgrading to websocket: %s", err)
		return
	}

	wsm.register <- conn

	go func() {
		defer func() {
			wsm.unregister <- conn
			conn.Close()
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Debug().Msgf("WebSocket connection closed: %s", err)
				} else {
					log.Debug().Msgf("Error reading message: %s", err)
				}
				break
			}

			log.Debug().Msgf("Received: %s", message)
			// Handle messages as needed
		}
	}()
}

// BroadcastMessage sends a message to all connected clients
func (wsm *WebSocketManager) BroadcastMessage(message []byte) {
	wsm.broadcast <- message
}
