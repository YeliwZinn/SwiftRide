package websockets

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket Upgrader (allows cross-origin requests)
var Upgrader = websocket.Upgrader{
	
	CheckOrigin: func(r *http.Request) bool {
		// Allow any origin (CORS for development purposes)
		
		return true // Allow any origin
	},
}

// Client represents a WebSocket connection
type Client struct {
	Conn   *websocket.Conn
	UserID string
	Role   string // "rider" or "driver"
}

// Hub manages all WebSocket connections
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan Notification
	Register   chan *Client
	Unregister chan *Client
	Mu         sync.Mutex
}

// Notification struct for messages
type Notification struct {
    Type    string      `json:"type"` // ride_request, ride_response, payment_request
    UserID  string      `json:"user_id"`
    Payload interface{} `json:"payload"`
}

// Global WebSocket hub instance
var WS_HUB *Hub

// Initialize WebSocket Hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Notification),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run the WebSocket Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			h.Clients[client] = true
			h.Mu.Unlock()
			log.Printf("Client %s connected", client.UserID)

		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Conn.Close()
				log.Printf("Client %s disconnected", client.UserID)
			}
			h.Mu.Unlock()

		case notification := <-h.Broadcast:
			h.Mu.Lock()
			for client := range h.Clients {
				if client.UserID == notification.UserID {
					err := client.Conn.WriteJSON(gin.H{
						"type":    notification.Type,
						"payload": notification.Payload,
					})
					
					if err != nil {
						log.Println("Error sending message:", err)
						client.Conn.Close()
						delete(h.Clients, client)
					}
				}
			}
			h.Mu.Unlock()
		}
	}
}



