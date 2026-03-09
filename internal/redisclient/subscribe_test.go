package redisclient

import (
	"log/slog"
	"os"
	"testing"

	"github.com/aaryan-purohit/message-broadcast-redis-pub-sub/internal/processor"
	"github.com/alicebob/miniredis"
)

func TestNewSubscriber(t *testing.T) {

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client, err := New(mr.Addr(), 0)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}

	mockSubscriber := NewSubscriber(client, "test-channel", &processor.Processor{}, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if mockSubscriber.channel != "test-channel" {
		t.Errorf("Expected channel to be 'test-channel', got '%s'", mockSubscriber.channel)
	}

	if mockSubscriber.processor == nil {
		t.Error("Expected processor to be initialized, got nil")
	}

	if mockSubscriber.client == nil {
		t.Error("Expected Redis client to be initialized, got nil")
	}

	if mockSubscriber.logger == nil {
		t.Error("Expected logger to be initialized, got nil")
	}

}
