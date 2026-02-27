package redisclient

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"main/internal/dispatcher"
	"main/internal/events"
	"main/internal/processor"
)

// fakeHandler captures submitted events
type fakeHandler struct {
	mu     sync.Mutex
	events []events.Message
}

func (f *fakeHandler) Handle(ctx context.Context, event events.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, event)
	return nil
}

// newFakeProcessor creates a processor with a fake handler that captures events
func newFakeProcessor(handler *fakeHandler) *processor.Processor {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := dispatcher.New(logger)
	dispatcher.Register("demo.message", handler)
	return processor.New(dispatcher, logger, 1, 10)
}

func TestSubscriber_Start_ReceivesMessage(t *testing.T) {
	// --- Setup Redis ---
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// --- Setup Processor with fake handler ---
	handler := &fakeHandler{}
	proc := newFakeProcessor(handler)
	defer proc.Stop()

	// --- Subscriber ---
	channel := "test.channel"
	sub := NewSubscriber(rdb, channel, proc, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Start subscriber ---
	go func() {
		_ = sub.Start(ctx)
	}()

	// Give subscriber time to subscribe
	time.Sleep(100 * time.Millisecond)

	// --- Publish message ---
	event := events.Message{
		ID:     "test-id",
		Type:   "demo.message",
		Source: "test-publisher",
		Payload: map[string]any{
			"text": "hello",
		},
	}

	data, _ := json.Marshal(event)
	err = rdb.Publish(ctx, channel, data).Err()
	if err != nil {
		t.Fatalf("failed to publish message: %v", err)
	}

	// --- Assert ---
	time.Sleep(200 * time.Millisecond)

	handler.mu.Lock()
	count := len(handler.events)
	handler.mu.Unlock()

	if count != 1 {
		t.Fatalf("expected 1 event, got %d", count)
	}
}

func TestSubscriber_Start_ContextCancel(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	handler := &fakeHandler{}
	proc := newFakeProcessor(handler)
	defer proc.Stop()

	sub := NewSubscriber(rdb, "test.channel", proc, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	done := make(chan struct{})

	go func() {
		_ = sub.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
		// success - context timeout or cancellation was handled
	case <-time.After(2 * time.Second):
		t.Fatal("subscriber did not stop after context timeout")
	}
}
