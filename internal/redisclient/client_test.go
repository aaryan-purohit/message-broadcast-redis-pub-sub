package redisclient

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestNew(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// create client
	client, err := New(mr.Addr(), 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("expected redis client, got nil")
	}

	// Verify client is usable
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to ping redis: %v", err)
	}
}

func TestNew_InvalidAddress(t *testing.T) {
	// Use an invalid address that won't have a server listening
	client, err := New("localhost:9999", 0)
	if err == nil {
		t.Fatal("expected error for invalid address, got nil")
	}
	if client != nil {
		t.Fatal("expected nil client for failed connection")
	}
}

func TestNew_WithDifferentDB(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Test creating client with different DB numbers
	for db := 0; db < 3; db++ {
		client, err := New(mr.Addr(), db)
		if err != nil {
			t.Fatalf("expected no error for db %d, got %v", db, err)
		}

		if client == nil {
			t.Fatalf("expected redis client for db %d, got nil", db)
		}

		// Verify client is usable
		ctx := context.Background()
		if err := client.Ping(ctx).Err(); err != nil {
			t.Fatalf("failed to ping redis on db %d: %v", db, err)
		}
	}
}

func TestNew_ConnectionFunctional(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client, err := New(mr.Addr(), 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx := context.Background()

	// Test basic SET/GET operations
	key := "test-key"
	value := "test-value"

	if err := client.Set(ctx, key, value, 0).Err(); err != nil {
		t.Fatalf("failed to set key: %v", err)
	}

	retrieved, err := client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("failed to get key: %v", err)
	}

	if retrieved != value {
		t.Fatalf("expected %s, got %s", value, retrieved)
	}
}
