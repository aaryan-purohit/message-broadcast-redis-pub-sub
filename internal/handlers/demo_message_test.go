package handlers

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/aaryan-purohit/message-broadcast-redis-pub-sub/internal/events"
)

func TestNewDemoMessageHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewDemoMessageHandler(logger)

	if handler.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

func TestHandle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewDemoMessageHandler(logger)

	event := events.Message{
		ID:      "123",
		Source:  "test-source",
		Payload: "test-payload",
	}

	err := handler.Handle(context.Background(), event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
