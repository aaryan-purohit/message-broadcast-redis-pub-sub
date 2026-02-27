package dispatcher

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"main/internal/events"
)

type mockHandler struct {
	called bool
	event  events.Message
	err    error
}

func (m *mockHandler) Handle(ctx context.Context, event events.Message) error {
	m.called = true
	m.event = event
	return m.err
}

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := New(logger)

	if d == nil {
		t.Fatal("expected dispatcher to be non-nil")
	}
	if d.handlers == nil {
		t.Fatal("expected handlers map to be initialized")
	}
}

func TestDispatcher_Register(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := New(logger)

	handler := &mockHandler{}
	d.Register("test.event", handler)

	if _, ok := d.handlers["test.event"]; !ok {
		t.Fatal("expected handler to be registered")
	}
}

func TestDispatcher_Dispatch_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := New(logger)

	handler := &mockHandler{}
	d.Register("demo.message", handler)

	msg := events.Message{
		ID:        "test-123",
		Type:      "demo.message",
		Source:    "test",
		Timestamp: time.Now(),
		Payload:   map[string]any{"key": "value"},
	}

	d.Dispatch(context.Background(), msg)

	if !handler.called {
		t.Fatal("expected handler to be called")
	}
	if handler.event.ID != "test-123" {
		t.Errorf("expected event ID 'test-123', got %q", handler.event.ID)
	}
}

func TestDispatcher_Dispatch_NoHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := New(logger)

	msg := events.Message{
		ID:   "test-123",
		Type: "unknown.event",
	}

	d.Dispatch(context.Background(), msg)
}

func TestDispatcher_Dispatch_HandlerError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := New(logger)

	handler := &mockHandler{err: errors.New("handler error")}
	d.Register("error.event", handler)

	msg := events.Message{
		ID:   "test-123",
		Type: "error.event",
	}

	d.Dispatch(context.Background(), msg)

	if !handler.called {
		t.Fatal("expected handler to be called")
	}
}

func TestDispatcher_Dispatch_MultipleHandlers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := New(logger)

	handler1 := &mockHandler{}
	handler2 := &mockHandler{}
	handler3 := &mockHandler{}

	d.Register("event.type1", handler1)
	d.Register("event.type2", handler2)
	d.Register("event.type3", handler3)

	msg1 := events.Message{ID: "1", Type: "event.type1"}
	msg2 := events.Message{ID: "2", Type: "event.type2"}
	msg3 := events.Message{ID: "3", Type: "event.type3"}

	d.Dispatch(context.Background(), msg1)
	d.Dispatch(context.Background(), msg2)
	d.Dispatch(context.Background(), msg3)

	if !handler1.called || handler1.event.ID != "1" {
		t.Error("handler1 was not called correctly")
	}
	if !handler2.called || handler2.event.ID != "2" {
		t.Error("handler2 was not called correctly")
	}
	if !handler3.called || handler3.event.ID != "3" {
		t.Error("handler3 was not called correctly")
	}
}

func TestDispatcher_Register_Overwrite(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d := New(logger)

	handler1 := &mockHandler{}
	handler2 := &mockHandler{}

	d.Register("test.event", handler1)
	d.Register("test.event", handler2)

	msg := events.Message{ID: "test", Type: "test.event"}
	d.Dispatch(context.Background(), msg)

	if !handler2.called {
		t.Fatal("expected second handler to be called")
	}
	if handler1.called {
		t.Fatal("expected first handler not to be called")
	}
}
