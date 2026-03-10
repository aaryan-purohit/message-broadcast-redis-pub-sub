package dispatcher

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/aaryan-purohit/message-broadcast-redis-pub-sub/internal/events"
)

type mockHandler struct {
}

func (m *mockHandler) Handle(ctx context.Context, event events.Message) error {
	return nil
}

type mockHandlerWithError struct {
}

func (m *mockHandlerWithError) Handle(ctx context.Context, event events.Message) error {
	return os.ErrInvalid
}

func TestNewDispatcher(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := New(logger)

	if dispatcher == nil {
		t.Fatal("expected dispatcher to be created, got nil")
	}
	if dispatcher.logger != logger {
		t.Fatal("expected logger to be set correctly")
	}
	if len(dispatcher.handlers) != 0 {
		t.Fatal("expected handlers map to be initialized empty")
	}
}

func TestRegisterHandler(t *testing.T) {

	eventType := "test_event"
	handler := &mockHandler{}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := New(logger)

	dispatcher.Register(eventType, handler)

	if dispatcher.handlers[eventType] != handler {
		t.Fatalf("expected handler to be registered for event type %s", eventType)
	}

}

func TestDispatch(t *testing.T) {

	eventType := "test_event"
	handler := &mockHandler{}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := New(logger)

	dispatcher.Register(eventType, handler)

	event := events.Message{
		Type:    eventType,
		Payload: []byte("test data"),
	}

	if err := dispatcher.Dispatch(context.Background(), event); err != nil {
		t.Fatalf("expected dispatch to succeed, got error: %v", err)
	}

}

func TestDispatchNoHandler(t *testing.T) {
	eventType := "unregistered_event"

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := New(logger)
	event := events.Message{
		Type:    eventType,
		Payload: []byte("test data"),
	}

	if err := dispatcher.Dispatch(context.Background(), event); err != nil {
		t.Fatalf("expected dispatch to succeed even with no handler, got error: %v", err)
	}
}

func TestDispatchHandlerError(t *testing.T) {
	eventType := "error_event"

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := New(logger)

	errorHandler := &mockHandlerWithError{}
	dispatcher.Register(eventType, errorHandler)

	event := events.Message{
		Type:    eventType,
		Payload: []byte("test data"),
	}

	if err := dispatcher.Dispatch(context.Background(), event); err == nil {
		t.Fatal("expected dispatch to return error from handler, got nil")
	}
}
