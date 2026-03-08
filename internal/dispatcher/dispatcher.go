package dispatcher

import (
	"context"
	"log/slog"
	"sync"

	"github.com/aaryan-purohit/message-broadcast-redis-pub-sub/internal/events"
)

type Handler interface {
	Handle(ctx context.Context, event events.Message) error
}

type Dispatcher struct {
	mu       sync.RWMutex
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
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventType] = handler
	d.logger.Info("handler registered", "event_type", eventType)
}

func (d *Dispatcher) Dispatch(ctx context.Context, event events.Message) error {
	d.mu.RLock()
	handler, ok := d.handlers[event.Type]
	d.mu.RUnlock()

	if !ok {
		d.logger.Warn("no handler found", "event_type", event.Type)
		return nil // nothing to do
	}

	if err := handler.Handle(ctx, event); err != nil {
		d.logger.Error("handler failed", "event_type", event.Type, "error", err)
		return err
	}
	return nil
}
