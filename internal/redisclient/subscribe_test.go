package redisclient

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"main/internal/dispatcher"
	"main/internal/events"
	"main/internal/processor"
	"main/internal/testutils"
)

// fakeHandler captures submitted events deterministically
type fakeHandler struct {
	mu     sync.Mutex
	events []events.Message
	seen   chan struct{}
}

func newFakeHandler() *fakeHandler {
	return &fakeHandler{
		seen: make(chan struct{}, 1),
	}
}

func (f *fakeHandler) Handle(ctx context.Context, event events.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.events = append(f.events, event)

	// Non-blocking signal
	select {
	case f.seen <- struct{}{}:
	default:
	}

	return nil
}

func newFakeProcessor(t *testing.T, handler *fakeHandler) *processor.Processor {
	t.Helper()

	logger := testutils.TestLogger()

	d := dispatcher.New(logger)
	d.Register("demo.message", handler)

	return processor.New(d, logger, 1, 10)
}

// waitForSubscription blocks until Redis confirms a subscriber
func waitForSubscription(
	t *testing.T,
	rdb *redis.Client,
	channel string,
) {
	t.Helper()

	ctx := context.Background()
	timeout := time.After(1 * time.Second)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("timeout waiting for subscription to channel %q", channel)
		case <-tick.C:
			res := rdb.PubSubNumSub(ctx, channel)
			if n := res.Val()[channel]; n > 0 {
				return
			}
		}
	}
}

func TestSubscriber_Start_ReceivesMessage(t *testing.T) {
	t.Parallel()

	mr := testutils.SetupRedis(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	handler := newFakeHandler()
	proc := newFakeProcessor(t, handler)
	defer proc.Stop()

	channel := "test.channel"
	sub := NewSubscriber(rdb, channel, proc, testutils.TestLogger())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = sub.Start(ctx)
	}()

	// 🔑 IMPORTANT: wait until subscription is active
	waitForSubscription(t, rdb, channel)

	event := events.Message{
		ID:     "test-id",
		Type:   "demo.message",
		Source: "test-publisher",
		Payload: map[string]any{
			"text": "hello",
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	if err := rdb.Publish(ctx, channel, data).Err(); err != nil {
		t.Fatalf("failed to publish message: %v", err)
	}

	select {
	case <-handler.seen:
		// success
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}

	handler.mu.Lock()
	defer handler.mu.Unlock()

	if len(handler.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(handler.events))
	}

	received := handler.events[0]

	if received.ID != "test-id" {
		t.Fatalf("unexpected ID: %s", received.ID)
	}
	if received.Type != "demo.message" {
		t.Fatalf("unexpected type: %s", received.Type)
	}

	payload, ok := received.Payload.(map[string]any)
	if !ok {
		t.Fatalf("unexpected payload type: %T", received.Payload)
	}

	if payload["text"] != "hello" {
		t.Fatalf("unexpected payload value: %#v", payload)
	}
}

func TestSubscriber_Start_ContextCancel(t *testing.T) {
	t.Parallel()

	mr := testutils.SetupRedis(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	handler := newFakeHandler()
	proc := newFakeProcessor(t, handler)
	defer proc.Stop()

	sub := NewSubscriber(rdb, "test.channel", proc, testutils.TestLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	done := make(chan struct{})

	go func() {
		_ = sub.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("subscriber did not stop after context cancellation")
	}
}

func TestSubscriber_IgnoresInvalidJSON(t *testing.T) {
	t.Parallel()

	mr := testutils.SetupRedis(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	handler := newFakeHandler()
	proc := newFakeProcessor(t, handler)
	defer proc.Stop()

	channel := "test.channel"
	sub := NewSubscriber(rdb, channel, proc, testutils.TestLogger())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = sub.Start(ctx)
	}()

	waitForSubscription(t, rdb, channel)

	_ = rdb.Publish(ctx, channel, []byte("invalid-json"))

	select {
	case <-handler.seen:
		t.Fatal("handler should not receive invalid JSON")
	case <-time.After(300 * time.Millisecond):
		// success
	}
}

func TestSubscriber_IgnoresOtherChannels(t *testing.T) {
	t.Parallel()

	mr := testutils.SetupRedis(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	handler := newFakeHandler()
	proc := newFakeProcessor(t, handler)
	defer proc.Stop()

	sub := NewSubscriber(rdb, "allowed.channel", proc, testutils.TestLogger())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = sub.Start(ctx)
	}()

	waitForSubscription(t, rdb, "allowed.channel")

	event := events.Message{
		ID:   "id",
		Type: "demo.message",
	}

	data, _ := json.Marshal(event)

	_ = rdb.Publish(ctx, "other.channel", data)

	select {
	case <-handler.seen:
		t.Fatal("handler should not receive events from other channels")
	case <-time.After(300 * time.Millisecond):
		// success
	}
}
