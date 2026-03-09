package redisclient

import (
	"testing"

	"github.com/alicebob/miniredis"
)

func TestNewClient(t *testing.T) {
	// Test with valid Redis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client, err := New(mr.Addr(), 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("Expected a valid Redis client, got nil")
	}

}

func TestNewClientInvalidAddress(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	_, err = New("invalid:address", 0)
	if err == nil {
		t.Fatal("Expected an error for invalid address, got nil")
	}
}
