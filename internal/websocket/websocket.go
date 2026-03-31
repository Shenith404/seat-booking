package websocket

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shenith404/seat-booking/internal/pubsub"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev; restrict in production
	},
}

// Client represents a WebSocket client connection
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	showID string
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[string]map[*Client]bool // showID -> clients
	broadcast  chan broadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	pubsub     *pubsub.PubSub
}

type broadcastMessage struct {
	showID string
	data   []byte
}

// NewHub creates a new Hub
func NewHub(ps *pubsub.PubSub) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan broadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		pubsub:     ps,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.showID] == nil {
				h.clients[client.showID] = make(map[*Client]bool)
				// Start subscribing to this show's Redis channel
				go h.subscribeToShow(ctx, client.showID)
			}
			h.clients[client.showID][client] = true
			h.mu.Unlock()
			log.Printf("Client connected to show %s", client.showID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.showID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.showID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client disconnected from show %s", client.showID)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.showID]
			for client := range clients {
				select {
				case client.send <- message.data:
				default:
					close(client.send)
					delete(clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// subscribeToShow subscribes to a show's Redis pub/sub channel
func (h *Hub) subscribeToShow(ctx context.Context, showID string) {
	eventChan, cleanup := h.pubsub.Subscribe(ctx, showID)
	defer cleanup()

	for event := range eventChan {
		data, err := encodeEvent(event)
		if err != nil {
			log.Printf("Failed to encode event: %v", err)
			continue
		}

		h.broadcast <- broadcastMessage{
			showID: showID,
			data:   data,
		}
	}
}

// BroadcastToShow sends a message to all clients connected to a show
func (h *Hub) BroadcastToShow(showID string, data []byte) {
	h.broadcast <- broadcastMessage{
		showID: showID,
		data:   data,
	}
}

// Handler handles WebSocket connections
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeWS handles WebSocket requests
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	showID := r.URL.Query().Get("show_id")
	if showID == "" {
		http.Error(w, "show_id query parameter required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:    h.hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		showID: showID,
	}

	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func encodeEvent(event pubsub.Event) ([]byte, error) {
	return []byte(`{"type":"` + string(event.Type) + `","show_id":"` + event.ShowID + `","seat_id":"` + event.SeatID + `"}`), nil
}
