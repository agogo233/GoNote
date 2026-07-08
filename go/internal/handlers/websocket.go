package handlers

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"

	"gonote/internal/models/logger"
)

// WSMessage represents a WebSocket message sent to clients
type WSMessage struct {
	Type    string      `json:"type"`    // "notes_updated", "folders_updated", etc.
	Payload interface{} `json:"payload"` // Optional payload data
}

// WSManager manages WebSocket connections and broadcasts
type WSManager struct {
	connections    map[*websocket.Conn]bool
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
		connections:    make(map[*websocket.Conn]bool),
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
		for conn := range m.connections {
			if conn != nil {
				conn.Close()
			}
			delete(m.connections, conn)
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
			m.connections[conn] = true
			m.mu.Unlock()
			logger.Printf("WebSocket client connected. Total: %d", len(m.connections))

		case conn := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.connections[conn]; ok {
				delete(m.connections, conn)
				if conn != nil {
					conn.Close()
				}
			}
			m.mu.Unlock()
			logger.Printf("WebSocket client disconnected. Total: %d", len(m.connections))

		case msg := <-m.broadcast:
			m.broadcastMessage(msg)
		}
	}
}

// broadcastMessage sends a message to all connected clients.
// JSON serialization is done outside the lock to minimize lock hold time.
func (m *WSManager) broadcastMessage(msg WSMessage) {
	// Serialize message outside of lock
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Printf("WebSocket: failed to marshal message: %v", err)
		return
	}

	// Get a snapshot of connections under read lock
	m.mu.RLock()
	// Early exit if no connections
	if len(m.connections) == 0 {
		m.mu.RUnlock()
		return
	}
	
	// Copy connections to slice for iteration outside lock
	conns := make([]*websocket.Conn, 0, len(m.connections))
	for conn := range m.connections {
		conns = append(conns, conn)
	}
	m.mu.RUnlock()

	// Send messages without holding the lock
	var failedConns []*websocket.Conn
	for _, conn := range conns {
		// Skip nil connections (defensive)
		if conn == nil {
			continue
		}
		// Set write deadline to prevent blocking on slow clients
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			failedConns = append(failedConns, conn)
		}
		conn.SetWriteDeadline(time.Time{}) // Reset deadline
	}

	// Remove failed connections with write lock
	if len(failedConns) > 0 {
		m.mu.Lock()
		for _, conn := range failedConns {
			if conn != nil {
				conn.Close()
			}
			delete(m.connections, conn)
		}
		m.mu.Unlock()
		logger.Printf("WebSocket: removed %d failed connections", len(failedConns))
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
	if m.maxConnections > 0 && len(m.connections) >= m.maxConnections {
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
		// Channel full, close connection directly
		m.mu.Lock()
		if _, ok := m.connections[conn]; ok {
			delete(m.connections, conn)
			conn.Close()
		}
		m.mu.Unlock()
	}
}

// BroadcastNotesUpdated notifies clients that notes have been updated
func BroadcastNotesUpdated(wm *WSManager) {
	if wm != nil {
		wm.Broadcast("notes_updated", nil)
	}
}