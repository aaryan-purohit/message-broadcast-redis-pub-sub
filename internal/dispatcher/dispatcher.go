package dispatcher

import (
	"context"
	"log/slog"
	"main/internal/events"
)

type Handler interface {
	Handle(ctx context.Context, event events.Message) error
}

type Dispatcher struct {
	handlers map[string]Handler
	logger   *slog.Logger
}

func New(logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]Handler),
		logger:   logger,
	}
}

func (d *Dispatcher) Register(eventType string, handler Handler) {
	d.handlers[eventType] = handler
	d.logger.Info("handler register", "event_type", eventType)
}

func (d *Dispatcher) Dispatch(ctx context.Context, event events.Message) {
	handler, ok := d.handlers[event.Type]
	if !ok {
		d.logger.Warn("no handler found", "event_type", event.Type)
		return
	}

	if err := handler.Handle(ctx, event); err != nil {
		d.logger.Error("handler failed", "event_type", event.Type, "error", err)
	}
}
