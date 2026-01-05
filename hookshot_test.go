package hookshot

import (
	"testing"
)

func TestRegisterAndClearHandlers(t *testing.T) {
	// Clear any existing handlers
	ClearHandlers()

	// Register a test handler
	called := false
	Register("test-handler", func() {
		called = true
	})

	// Verify handler was registered
	if _, exists := handlers["test-handler"]; !exists {
		t.Error("Handler was not registered")
	}

	// Call the handler
	handlers["test-handler"]()
	if !called {
		t.Error("Handler was not called")
	}

	// Clear handlers
	ClearHandlers()
	if len(handlers) != 0 {
		t.Errorf("Handlers not cleared, got %d handlers", len(handlers))
	}
}

func TestMustRegister(t *testing.T) {
	ClearHandlers()

	// First registration should succeed
	MustRegister("unique-handler", func() {})

	// Second registration with same name should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustRegister should panic on duplicate registration")
		}
	}()
	MustRegister("unique-handler", func() {})
}

func TestMultipleHandlers(t *testing.T) {
	ClearHandlers()

	count := 0
	Register("handler-a", func() { count += 1 })
	Register("handler-b", func() { count += 10 })
	Register("handler-c", func() { count += 100 })

	if len(handlers) != 3 {
		t.Errorf("Expected 3 handlers, got %d", len(handlers))
	}

	handlers["handler-a"]()
	handlers["handler-b"]()
	handlers["handler-c"]()

	if count != 111 {
		t.Errorf("Expected count=111, got %d", count)
	}

	ClearHandlers()
}
