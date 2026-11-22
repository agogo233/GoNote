package handlers

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"

	"gonote/internal/models/logger"
)

// wsSendBuffer is the per-connection outgoing queue size (in messages).
const wsSendBuffer = 64

// wsWriteTimeout bounds a single socket write so a stalled client only
// affects its own delivery, never the fan-out to other clients.
const wsWriteTimeout = 5 * time.Second

// WSMessage represents a WebSocket message sent to clients
type WSMessage struct {
	Type    string      `json:"type"`    // "notes_updated", "folders_updated", etc.
	Payload interface{} `json:"payload"` // Optional payload data
}

// wsClient wraps a WebSocket connection with its own outbound queue and
// writer goroutine. Broadcasts only enqueue onto the queue (non-blocking),
// so one slow client cannot stall the manager or the other clients.
type wsClient struct {
	conn      *websocket.Conn
	send      chan []byte
	done      chan struct{} // closed when the client is removed
	closeOnce sync.Once
}

// close idempotently signals the writer goroutine to stop and closes the
// underlying connection. The send channel is deliberately never closed so
// concurrent broadcasters can never panic on a send-to-closed-channel.
func (c *wsClient) close() {
	c.closeOnce.Do(func() {
		if c.done != nil {
			close(c.done)
		}
		if c.conn != nil {
			c.conn.Close()
		}
	})
}

// WSManager manages WebSocket connections and broadcasts
type WSManager struct {
	clients        map[*websocket.Conn]*wsClient
	mu             sync.RWMutex
	broadcast      chan WSMessage
	register       chan *websocket.Conn
	unregister     chan *websocket.Conn
	stop           chan struct{}
	stopOnce       sync.Once // Ensures Stop() is only called once
	stopped        bool      // Flag to indicate manager has stopped
	maxConnections int       // Maximum concurrent connections (0 = unlimited)
}

// InitWSManager initializes and returns a new WebSocket manager
func InitWSManager(maxConnections int) *WSManager {
	wm := &WSManager{
		clients:        make(map[*websocket.Conn]*wsClient),
		broadcast:      make(chan WSMessage, 100),
		register:       make(chan *websocket.Conn, 10),
		unregister:     make(chan *websocket.Conn, 10),
		stop:           make(chan struct{}),
		maxConnections: maxConnections,
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("WebSocket manager goroutine panic recovered: %v", r)
			}
		}()
		wm.run()
	}()
	return wm
}

// Stop gracefully shuts down the WebSocket manager.
// It closes all connections and stops the run goroutine.
// Safe to call multiple times.
func (m *WSManager) Stop() {
	m.stopOnce.Do(func() {
		// Mark as stopped first to prevent new registrations
		m.mu.Lock()
		m.stopped = true
		// Close all remaining connections
		for conn, client := range m.clients {
			if client != nil {
				client.close()
			} else if conn != nil {
				conn.Close()
			}
			delete(m.clients, conn)
		}
		m.mu.Unlock()

		close(m.stop) // Signal run() to stop

		logger.Println("WebSocket manager stopped")
	})
}

// run handles WebSocket connection lifecycle and broadcasts
func (m *WSManager) run() {
	for {
		select {
		case <-m.stop:
			return // Exit the goroutine

		case conn := <-m.register:
			m.mu.Lock()
			// Defensive: a re-registered connection replaces any stale entry
			// so the old writer goroutine cannot leak.
			if old, ok := m.clients[conn]; ok && old != nil {
				old.close()
			}
			client := &wsClient{
				conn: conn,
				send: make(chan []byte, wsSendBuffer),
				done: make(chan struct{}),
			}
			m.clients[conn] = client
			total := len(m.clients)
			m.mu.Unlock()
			go m.writePump(client)
			logger.Printf("WebSocket client connected. Total: %d", total)

		case conn := <-m.unregister:
			m.mu.Lock()
			client, ok := m.clients[conn]
			if ok {
				delete(m.clients, conn)
			}
			total := len(m.clients)
			m.mu.Unlock()
			if ok {
				if client != nil {
					client.close()
				} else if conn != nil {
					conn.Close()
				}
			}
			logger.Printf("WebSocket client disconnected. Total: %d", total)

		case msg := <-m.broadcast:
			m.broadcastMessage(msg)
		}
	}
}

// writePump drains the per-connection outbound queue. It exits when the
// client is removed (done closed) or a write fails.
func (m *WSManager) writePump(client *wsClient) {
	for {
		select {
		case <-client.done:
			return
		case data := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
			if err := client.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				m.removeClient(client.conn)
				return
			}
		}
	}
}

// removeClient idempotently removes and closes a client by its connection.
func (m *WSManager) removeClient(conn *websocket.Conn) {
	m.mu.Lock()
	client, ok := m.clients[conn]
	if ok {
		delete(m.clients, conn)
	}
	m.mu.Unlock()
	if ok && client != nil {
		client.close()
	}
}

// broadcastMessage enqueues a message to every connected client without
// blocking. Each client has its own buffered queue; a client whose queue is
// full (slow consumer) simply misses this message instead of stalling the
// fan-out for everyone else. JSON serialization happens once, outside the
// lock.
func (m *WSManager) broadcastMessage(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Printf("WebSocket: failed to marshal message: %v", err)
		return
	}

	// Snapshot the clients under read lock
	m.mu.RLock()
	if len(m.clients) == 0 {
		m.mu.RUnlock()
		return
	}
	clients := make([]*wsClient, 0, len(m.clients))
	for _, client := range m.clients {
		if client != nil {
			clients = append(clients, client)
		}
	}
	m.mu.RUnlock()

	// Enqueue without holding the lock; never block on a slow client
	dropped := 0
	for _, client := range clients {
		select {
		case client.send <- data:
		default:
			dropped++
		}
	}
	if dropped > 0 {
		logger.Printf("WebSocket: dropped message for %d slow client(s)", dropped)
	}
}

// Broadcast sends a message to all connected WebSocket clients
func (m *WSManager) Broadcast(msgType string, payload interface{}) {
	select {
	case m.broadcast <- WSMessage{Type: msgType, Payload: payload}:
	default:
		// Channel full, skip broadcast (non-blocking)
		logger.Printf("WebSocket broadcast channel full, message '%s' dropped", msgType)
	}
}

// Register registers a new WebSocket connection.
// Returns false if the manager has been stopped, connection is nil, or at capacity.
func (m *WSManager) Register(conn *websocket.Conn) bool {
	// Reject nil connections
	if conn == nil {
		return false
	}

	m.mu.RLock()
	if m.stopped {
		m.mu.RUnlock()
		return false
	}
	// Check connection limit (0 = unlimited)
	if m.maxConnections > 0 && len(m.clients) >= m.maxConnections {
		m.mu.RUnlock()
		return false
	}
	m.mu.RUnlock()

	// Non-blocking send to prevent deadlock during shutdown
	select {
	case m.register <- conn:
		return true
	case <-m.stop:
		return false
	default:
		// Channel full, log and return false
		logger.Printf("WebSocket: register channel full, connection rejected")
		return false
	}
}

// Unregister unregisters a WebSocket connection.
// No-op if the manager has been stopped or connection is nil.
func (m *WSManager) Unregister(conn *websocket.Conn) {
	// No-op for nil connections
	if conn == nil {
		return
	}

	m.mu.RLock()
	if m.stopped {
		m.mu.RUnlock()
		return
	}
	m.mu.RUnlock()

	// Non-blocking send to prevent deadlock during shutdown
	select {
	case m.unregister <- conn:
	case <-m.stop:
		// Manager stopping, connection will be closed by Stop()
	default:
		// Channel full, remove the client directly
		m.removeClient(conn)
	}
}

// BroadcastNotesUpdated notifies clients that notes have been updated
func BroadcastNotesUpdated(wm *WSManager) {
	if wm != nil {
		wm.Broadcast("notes_updated", nil)
	}
}
