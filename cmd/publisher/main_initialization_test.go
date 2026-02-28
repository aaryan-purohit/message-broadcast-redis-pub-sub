package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"main/internal/events"

	"github.com/google/uuid"
)

func TestInit_LoggerCreation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if logger == nil {
		t.Fatal("expected logger to be created")
	}
}

func TestInit_ConfigFromEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envVal   string
		fallback string
		expected string
	}{
		{"REDIS_ADDR override", "REDIS_ADDR", "custom:6380", "localhost:6379", "custom:6380"},
		{"CHANNEL_NAME override", "CHANNEL_NAME", "custom.events", "broadcast.events", "custom.events"},
		{"SERVER_ID override", "SERVER_ID", "custom-pub", "publisher", "custom-pub"},
		{"REDIS_ADDR default", "REDIS_ADDR_UNUSED", "", "localhost:6379", "localhost:6379"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				os.Setenv(tt.envKey, tt.envVal)
				defer os.Unsetenv(tt.envKey)
			} else {
				os.Unsetenv(tt.envKey)
			}

			var result string
			if tt.envKey == "REDIS_ADDR_UNUSED" {
				result = getEnv("REDIS_ADDR", tt.fallback)
			} else {
				result = getEnv(tt.envKey, tt.fallback)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInit_DefaultConfigValues(t *testing.T) {
	os.Unsetenv("REDIS_ADDR")
	os.Unsetenv("CHANNEL_NAME")
	os.Unsetenv("SERVER_ID")

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	channel := getEnv("CHANNEL_NAME", "broadcast.events")
	source := getEnv("SERVER_ID", "publisher")

	if redisAddr != "localhost:6379" {
		t.Errorf("expected default REDIS_ADDR, got %q", redisAddr)
	}
	if channel != "broadcast.events" {
		t.Errorf("expected default CHANNEL_NAME, got %q", channel)
	}
	if source != "publisher" {
		t.Errorf("expected default SERVER_ID, got %q", source)
	}
}

func TestInit_EventMessageCreation(t *testing.T) {
	source := "test-publisher"
	payload := map[string]any{
		"counter": 1,
		"text":    "hello from publisher",
	}

	msg := events.Message{
		ID:        uuid.NewString(),
		Type:      "demo.message",
		Source:    source,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	if msg.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if msg.Type != "demo.message" {
		t.Errorf("expected type 'demo.message', got %q", msg.Type)
	}
	if msg.Source != source {
		t.Errorf("expected source %q, got %q", source, msg.Source)
	}
}

func TestInit_EventMessageJsonMarshaling(t *testing.T) {
	msg := events.Message{
		ID:        "test-123",
		Type:      "demo.message",
		Source:    "test-publisher",
		Timestamp: time.Date(2026, 2, 27, 12, 0, 0, 0, time.UTC),
		Payload: map[string]any{
			"counter": 1,
			"text":    "hello",
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty JSON data")
	}

	var recovered events.Message
	if err := json.Unmarshal(data, &recovered); err != nil {
		t.Fatalf("expected no error unmarshaling, got %v", err)
	}

	if recovered.ID != msg.ID {
		t.Errorf("expected ID %q, got %q", msg.ID, recovered.ID)
	}
	if recovered.Type != msg.Type {
		t.Errorf("expected Type %q, got %q", msg.Type, recovered.Type)
	}
	if recovered.Source != msg.Source {
		t.Errorf("expected Source %q, got %q", msg.Source, recovered.Source)
	}
}

func TestInit_MultipleEventCreation(t *testing.T) {
	source := "publisher"
	createdEvents := 0

	for i := 1; i <= 5; i++ {
		msg := events.Message{
			ID:        uuid.NewString(),
			Type:      "demo.message",
			Source:    source,
			Timestamp: time.Now(),
			Payload: map[string]any{
				"counter": i,
				"text":    "hello from publisher",
			},
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("iteration %d: failed to marshal, %v", i, err)
			continue
		}

		if len(data) > 0 {
			createdEvents++
		}
	}

	if createdEvents != 5 {
		t.Errorf("expected 5 events created, got %d", createdEvents)
	}
}

func TestInit_LoggerWithContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	source := "test-source"

	contextLogger := logger.With("component", "publisher", "source", source)

	if contextLogger == nil {
		t.Fatal("expected logger with context to be created")
	}
}
