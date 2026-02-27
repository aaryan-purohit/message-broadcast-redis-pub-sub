package redisclient

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestNew(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	// create client
	client, err := New(mr.Addr(), 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("expected redis client, got nil")
	}

}
