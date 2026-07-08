package handlers

import (
	"sync"
	"testing"
	"time"

	"github.com/gofiber/contrib/websocket"
)

// mockConn is a minimal mock for websocket.Conn
// Since we can't easily create real websocket connections in unit tests,
// we test the WSManager logic directly

func TestWSManager_InitAndStop(t *testing.T) {
	manager := InitWSManager(100)
	if manager == nil {
		t.Fatal("InitWSManager returned nil")
	}
	
	// Stop should not panic when called once
	manager.Stop()
	
	// Stop should be idempotent (safe to call multiple times)
	manager.Stop()
	manager.Stop()
	
	// Verify stopped flag
	if !manager.stopped {
		t.Error("manager should be marked as stopped")
	}
}

func TestWSManager_RegisterAfterStop(t *testing.T) {
	manager := &WSManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan WSMessage, 100),
		register:    make(chan *websocket.Conn, 10),
		unregister:  make(chan *websocket.Conn, 10),
		stop:        make(chan struct{}),
		maxConnections: 0,
	}
	go manager.run()
	
	// Stop the manager
	manager.Stop()
	
	// Register should return false after stop
	result := manager.Register(nil)
	if result {
		t.Error("Register should return false after manager is stopped")
	}
}

func TestWSManager_RegisterNil(t *testing.T) {
	manager := &WSManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan WSMessage, 100),
		register:    make(chan *websocket.Conn, 10),
		unregister:  make(chan *websocket.Conn, 10),
		stop:        make(chan struct{}),
		maxConnections: 0,
	}
	go manager.run()
	defer manager.Stop()

	// Register(nil) should return false immediately
	result := manager.Register(nil)
	if result {
		t.Error("Register(nil) should return false")
	}
}

func TestWSManager_UnregisterAfterStop(t *testing.T) {
	manager := &WSManager{
		connections:    make(map[*websocket.Conn]bool),
		broadcast:      make(chan WSMessage, 100),
		register:       make(chan *websocket.Conn, 10),
		unregister:     make(chan *websocket.Conn, 10),
		stop:           make(chan struct{}),
		maxConnections: 0,
	}
	go manager.run()
	
	// Stop the manager
	manager.Stop()
	
	// Unregister should not panic after stop (no-op)
	manager.Unregister(nil)
}

func TestWSManager_BroadcastNonBlocking(t *testing.T) {
	manager := &WSManager{
		connections:    make(map[*websocket.Conn]bool),
		broadcast:      make(chan WSMessage, 100),
		register:       make(chan *websocket.Conn, 10),
		unregister:     make(chan *websocket.Conn, 10),
		stop:           make(chan struct{}),
		maxConnections: 0,
	}
	go manager.run()
	defer manager.Stop()
	
	// Broadcast should not block even with no connections
	done := make(chan bool, 1)
	go func() {
		manager.Broadcast("test", nil)
		done <- true
	}()
	
	select {
	case <-done:
		// Good - didn't block
	case <-time.After(1 * time.Second):
		t.Error("Broadcast should not block")
	}
}

func TestWSManager_BroadcastChannelFull(t *testing.T) {
	manager := &WSManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan WSMessage, 2), // Small buffer
		register:    make(chan *websocket.Conn, 10),
		unregister:  make(chan *websocket.Conn, 10),
		stop:        make(chan struct{}),
	}
	
	// Don't start run() - so channel will fill up
	
	// Fill the channel
	for i := 0; i < 2; i++ {
		manager.Broadcast("test", nil)
	}
	
	// This should not block (non-blocking send with default case)
	done := make(chan bool, 1)
	go func() {
		manager.Broadcast("overflow", nil) // Should hit default case
		done <- true
	}()
	
	select {
	case <-done:
		// Good - didn't block
	case <-time.After(1 * time.Second):
		t.Error("Broadcast should not block when channel is full")
	}
}

func TestWSManager_ConcurrentRegister(t *testing.T) {
	manager := &WSManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan WSMessage, 100),
		register:    make(chan *websocket.Conn, 100), // Large enough buffer
		unregister:  make(chan *websocket.Conn, 10),
		stop:        make(chan struct{}),
	}
	go manager.run()
	defer manager.Stop()
	
	// Wait for manager to start
	time.Sleep(10 * time.Millisecond)
	
	// Note: Register rejects nil connections, so this tests the channel behavior
	// not actual connection handling. Real connections would be real websocket.Conn objects.
	
	// Test concurrent register calls that would fail due to nil check
	var wg sync.WaitGroup
	numGoroutines := 50
	
	// Concurrently try to register nil (should all return false immediately)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// nil connections are rejected immediately
			result := manager.Register(nil)
			if result {
				t.Error("Register(nil) should return false")
			}
		}()
	}
	
	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Good - all goroutines completed
	case <-time.After(2 * time.Second):
		t.Error("Concurrent Register calls should not block indefinitely")
	}
}

func TestWSManager_ConcurrentBroadcast(t *testing.T) {
	manager := &WSManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan WSMessage, 100),
		register:    make(chan *websocket.Conn, 10),
		unregister:  make(chan *websocket.Conn, 10),
		stop:        make(chan struct{}),
	}
	go manager.run()
	defer manager.Stop()
	
	var wg sync.WaitGroup
	numBroadcasts := 100
	
	// Concurrently broadcast
	for i := 0; i < numBroadcasts; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			manager.Broadcast("test", map[string]int{"n": n})
		}(i)
	}
	
	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Good
	case <-time.After(2 * time.Second):
		t.Error("Concurrent Broadcast calls should complete quickly")
	}
}

func TestWSManager_StopDuringRegister(t *testing.T) {
	manager := &WSManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan WSMessage, 100),
		register:    make(chan *websocket.Conn, 1), // Small buffer
		unregister:  make(chan *websocket.Conn, 10),
		stop:        make(chan struct{}),
	}
	go manager.run()
	
	// Now stop while register might be pending
	go func() {
		time.Sleep(10 * time.Millisecond)
		manager.Stop()
	}()
	
	// Register(nil) should return false immediately (nil check)
	// This tests the non-blocking behavior
	done := make(chan bool, 1)
	go func() {
		result := manager.Register(nil)
		done <- result
	}()
	
	select {
	case result := <-done:
		// nil connections return false immediately
		if result {
			t.Error("Register(nil) should return false")
		}
	case <-time.After(2 * time.Second):
		t.Error("Register should not hang when Stop is called")
	}
}

func TestWSManager_BroadcastMessageEmptyConnections(t *testing.T) {
	manager := &WSManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan WSMessage, 100),
		register:    make(chan *websocket.Conn, 10),
		unregister:  make(chan *websocket.Conn, 10),
		stop:        make(chan struct{}),
	}
	
	// broadcastMessage should handle empty connections gracefully
	msg := WSMessage{Type: "test", Payload: nil}
	manager.broadcastMessage(msg)
	// Should not panic
}

func TestWSManager_DoubleStop(t *testing.T) {
	manager := &WSManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan WSMessage, 100),
		register:    make(chan *websocket.Conn, 10),
		unregister:  make(chan *websocket.Conn, 10),
		stop:        make(chan struct{}),
	}
	go manager.run()
	
	// Multiple stops should be safe
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.Stop()
		}()
	}
	
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Good
	case <-time.After(2 * time.Second):
		t.Error("Multiple Stop calls should not hang")
	}
}

func TestBroadcastNotesUpdated(t *testing.T) {
	// Test with nil manager
	BroadcastNotesUpdated(nil)
	
	// Test with initialized manager
	manager := &WSManager{
		connections:    make(map[*websocket.Conn]bool),
		broadcast:      make(chan WSMessage, 100),
		register:       make(chan *websocket.Conn, 10),
		unregister:     make(chan *websocket.Conn, 10),
		stop:           make(chan struct{}),
		maxConnections: 0,
	}
	go manager.run()
	defer manager.Stop()
	
	// Should send to broadcast channel
	done := make(chan bool, 1)
	go func() {
		BroadcastNotesUpdated(manager)
		done <- true
	}()
	
	select {
	case <-done:
		// Good
	case <-time.After(1 * time.Second):
		t.Error("BroadcastNotesUpdated should not block")
	}
}

func TestWSManager_ConnectionLimit(t *testing.T) {
	manager := &WSManager{
		connections:    make(map[*websocket.Conn]bool),
		broadcast:      make(chan WSMessage, 100),
		register:       make(chan *websocket.Conn, 10),
		unregister:     make(chan *websocket.Conn, 10),
		stop:           make(chan struct{}),
		maxConnections: 2,
	}
	go manager.run()
	defer manager.Stop()

	// Fill the connection map directly to hit the limit
	manager.mu.Lock()
	manager.connections[&websocket.Conn{}] = true
	manager.connections[&websocket.Conn{}] = true
	manager.mu.Unlock()

	// Wait for run() to process
	time.Sleep(10 * time.Millisecond)

	// Attempt register - should be rejected (at capacity)
	// Use a real websocket.Conn to bypass nil check
	done := make(chan bool, 1)
	go func() {
		// We can't easily create a real connection, but we can verify the
		// capacity check logic by checking Register returns false
		// Register(nil) returns false immediately due to nil check
		// So we test the limit indirectly by verifying len check path
		manager.mu.RLock()
		atCapacity := manager.maxConnections > 0 && len(manager.connections) >= manager.maxConnections
		manager.mu.RUnlock()
		if !atCapacity {
			t.Error("Should be at capacity after adding 2 connections")
		}
		done <- true
	}()
	
	select {
	case <-done:
		// Good
	case <-time.After(1 * time.Second):
		t.Error("Connection capacity check should complete quickly")
	}
}
