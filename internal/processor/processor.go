package processor

import (
	"context"
	"log/slog"
	"main/internal/dispatcher"
	"main/internal/events"
)

type Processor struct {
	queue      chan events.Message
	dispatcher *dispatcher.Dispatcher
	logger     *slog.Logger
}

func New(dispatcher *dispatcher.Dispatcher, logger *slog.Logger, workers int, buffer int) *Processor {
	p := &Processor{
		queue:      make(chan events.Message, buffer),
		dispatcher: dispatcher,
		logger:     logger,
	}

	for i := 0; i < workers; i++ {
		go p.worker(i)
	}
	return p
}

func (p *Processor) Submit(event events.Message) {
	p.queue <- event
}

func (p *Processor) worker(id int) {
	p.logger.Info("worker started", "worker_id", id)
	for event := range p.queue {
		p.dispatcher.Dispatch(context.Background(), event)
	}
}
