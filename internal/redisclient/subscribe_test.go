package redisclient

import (
	"testing"

	"github.com/aaryan-purohit/message-broadcast-redis-pub-sub/internal/processor"
	"github.com/redis/go-redis/v9"
)

func TestNewSubscriber(t *testing.T) {

	mockSubscriber := &Subscriber{
		client:    &redis.Client{},
		channel:   "test-channel",
		processor: &processor.Processor{},
		logger:    nil,
	}

	if mockSubscriber.channel != "test-channel" {
		t.Errorf("Expected channel to be 'test-channel', got '%s'", mockSubscriber.channel)
	}

	if mockSubscriber.processor == nil {
		t.Error("Expected processor to be initialized, got nil")
	}

	if mockSubscriber.client == nil {
		t.Error("Expected Redis client to be initialized, got nil")
	}

	if mockSubscriber.logger != nil {
		t.Error("Expected logger to be nil, got non-nil")
	}

}
