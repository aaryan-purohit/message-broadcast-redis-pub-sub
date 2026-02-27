package handlers

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"main/internal/events"
)

func TestNewDemoMessageHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewDemoMessageHandler(logger)

	if handler == nil {
		t.Fatal("expected handler to be non-nil")
	}
}

func TestDemoMessageHandler_Handle_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewDemoMessageHandler(logger)

	msg := events.Message{
		ID:        "test-123",
		Type:      "demo.message",
		Source:    "test-publisher",
		Timestamp: time.Now(),
		Payload:   map[string]any{"text": "hello"},
	}

	err := handler.Handle(context.Background(), msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDemoMessageHandler_Handle_WithEmptyMessage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewDemoMessageHandler(logger)

	msg := events.Message{}

	err := handler.Handle(context.Background(), msg)
	if err != nil {
		t.Fatalf("expected no error for empty message, got %v", err)
	}
}

func TestDemoMessageHandler_Handle_WithCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewDemoMessageHandler(logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg := events.Message{
		ID:     "test-123",
		Type:   "demo.message",
		Source: "test",
	}

	err := handler.Handle(ctx, msg)
	if err != nil {
		t.Fatalf("expected no error with cancelled context, got %v", err)
	}
}
