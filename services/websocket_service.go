package services

import (
	"bulk-email-mailgun/models"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketService struct {
	clients   map[*websocket.Conn]bool
	mu        sync.Mutex
	broadcast chan models.ProgressUpdate
}

func NewWebSocketService() *WebSocketService {
	ws := &WebSocketService{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan models.ProgressUpdate, 100),
	}
	go ws.handleBroadcasts()
	return ws
}

func (ws *WebSocketService) AddClient(conn *websocket.Conn) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.clients[conn] = true
}

func (ws *WebSocketService) RemoveClient(conn *websocket.Conn) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	delete(ws.clients, conn)
	conn.Close()
}

func (ws *WebSocketService) GetBroadcastChannel() chan<- models.ProgressUpdate {
	return ws.broadcast
}

func (ws *WebSocketService) handleBroadcasts() {
	for msg := range ws.broadcast {
		ws.mu.Lock()
		for client := range ws.clients {
			err := client.WriteJSON(map[string]interface{}{
				"type": "progress",
				"data": msg,
			})
			if err != nil {
				delete(ws.clients, client)
				client.Close()
			}
		}
		ws.mu.Unlock()
	}
}
