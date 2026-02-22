package redisclient

import (
	"context"
	"encoding/json"
	"log/slog"
	"main/internal/events"
	"main/internal/processor"

	"github.com/redis/go-redis/v9"
)

type Subscriber struct {
	client    *redis.Client
	channel   string
	processor processor.Processor
	logger    *slog.Logger
}

func NewSubscriber(client *redis.Client, channel string, processor *processor.Processor, logger *slog.Logger) *Subscriber {
	return &Subscriber{
		client:    client,
		channel:   channel,
		processor: *processor,
		logger:    logger,
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	sub := s.client.Subscribe(ctx, s.channel)
	defer sub.Close()

	s.logger.Info("subscribed to redis", "channel", s.channel)

	for {
		msg, err := sub.ReceiveMessage(ctx)
		if err != nil {
			return err
		}

		var event events.Message
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			s.logger.Error("invalid message", "error", err)
			continue
		}
		s.processor.Submit(event)
	}

}
