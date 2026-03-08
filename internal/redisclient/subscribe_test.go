package redisclient

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"main/internal/events"
	"main/internal/processor"
)

func TestNewSubscriber(t *testing.T) {
	mr := SetupRedis(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	channel := "test.channel"
	logger := TestLogger()

	proc := &processor.Processor{}

	sub := NewSubscriber(rdb, channel, proc, logger)

	if sub == nil {
		t.Fatal("expected non-nil subscriber")
	}
	if sub.client != rdb {
		t.Fatal("subscriber has incorrect Redis client")
	}
	if sub.channel != channel {
		t.Fatalf("expected channel %q, got %q", channel, sub.channel)
	}
	if sub.logger != logger {
		t.Fatal("subscriber has incorrect logger")
	}
	if sub.processor != proc {
		t.Fatal("subscriber has incorrect processor")
	}
}

func TestSubscriber_Start_ProcessMessage(t *testing.T) {
	mr := SetupRedis(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()
	logger := TestLogger()

	eventProcessed := make(chan bool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := NewSubscriber(rdb, "test.channel", &processor.Processor{}, logger)

	go func() {
		if err := sub.Start(ctx); err != nil {
			t.Errorf("subscriber Start error: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	// publish message
	event := events.Message{
		ID: "123",
	}

	data, _ := json.Marshal(event)

	rdb.Publish(ctx, "test-channel", data)

	select {
	case <-eventProcessed:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("event was not processed")
	}

	cancel()
}
