package events

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessage_Creation(t *testing.T) {
	now := time.Now()
	payload := map[string]any{
		"text":  "hello world",
		"count": 42,
	}

	msg := Message{
		ID:        "test-123",
		Type:      "demo.message",
		Source:    "test-publisher",
		Timestamp: now,
		Payload:   payload,
	}

	if msg.ID != "test-123" {
		t.Errorf("expected ID 'test-123', got %q", msg.ID)
	}
	if msg.Type != "demo.message" {
		t.Errorf("expected Type 'demo.message', got %q", msg.Type)
	}
	if msg.Source != "test-publisher" {
		t.Errorf("expected Source 'test-publisher', got %q", msg.Source)
	}
	if !msg.Timestamp.Equal(now) {
		t.Errorf("expected Timestamp %v, got %v", now, msg.Timestamp)
	}
}

func TestMessage_RoundTrip(t *testing.T) {
	original := Message{
		ID:        "round-trip-123",
		Type:      "test.event",
		Source:    "test-service",
		Timestamp: time.Now().UTC(),
		Payload: map[string]any{
			"key1": "value1",
			"key2": 42,
		},
	}

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var recovered Message
	if err := json.Unmarshal(jsonData, &recovered); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if recovered.ID != original.ID {
		t.Errorf("ID mismatch: expected %q, got %q", original.ID, recovered.ID)
	}
	if recovered.Type != original.Type {
		t.Errorf("Type mismatch: expected %q, got %q", original.Type, recovered.Type)
	}
	if recovered.Source != original.Source {
		t.Errorf("Source mismatch: expected %q, got %q", original.Source, recovered.Source)
	}
}

func TestMessage_ValidPayloadTypes(t *testing.T) {
	tests := []struct {
		name    string
		payload any
	}{
		{"map payload", map[string]any{"key": "value"}},
		{"string payload", "simple string"},
		{"number payload", 123},
		{"nil payload", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				ID:        "valid-test",
				Type:      "test.event",
				Source:    "test",
				Timestamp: time.Now(),
				Payload:   tt.payload,
			}

			data, err := json.Marshal(msg)
			if err != nil {
				t.Fatalf("failed to marshal with %s: %v", tt.name, err)
			}

			var recovered Message
			if err := json.Unmarshal(data, &recovered); err != nil {
				t.Fatalf("failed to unmarshal with %s: %v", tt.name, err)
			}
		})
	}
}

func TestMessage_InvalidJSON(t *testing.T) {
	invalidJSON := `{
		"id": "test",
		"type": "event",
		"source": "test",
		"timestamp": "invalid-date",
		"payload": {}
	}`

	var msg Message
	err := json.Unmarshal([]byte(invalidJSON), &msg)
	if err == nil {
		t.Fatal("expected error for invalid timestamp, got nil")
	}
}
