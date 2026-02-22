package handlers

import (
	"context"
	"log/slog"
	"main/internal/events"
)

type DemoMessageHandler struct {
	logger *slog.Logger
}

func NewDemoMessageHandler(logger *slog.Logger) *DemoMessageHandler {
	return &DemoMessageHandler{logger: logger}
}

func (h *DemoMessageHandler) Handle(ctx context.Context, event events.Message) error {
	h.logger.Info(
		"demo.message handled",
		"id", event.ID,
		"source", event.Source,
		"payload", event.Payload,
	)
	return nil
}
