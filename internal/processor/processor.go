package processor

import (
	"context"
	"errors"
	"log/slog"
	"main/internal/dispatcher"
	"main/internal/events"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrQueueFull = errors.New("processor queue is full")
)

type Processor struct {
	queue      chan events.Message
	dispatcher *dispatcher.Dispatcher
	logger     *slog.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	processed  atomic.Int64
	dropped    atomic.Int64
	maxRetries int
	retryDelay time.Duration
}

func New(dispatcher *dispatcher.Dispatcher, logger *slog.Logger, workers int, buffer int) *Processor {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Processor{
		queue:      make(chan events.Message, buffer),
		dispatcher: dispatcher,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
	}

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	return p
}

func (p *Processor) Submit(event events.Message) error {
	select {
	case p.queue <- event:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
		p.dropped.Add(1)
		p.logger.Warn("message dropped, queue full", "event_id", event.ID, "total_dropped", p.dropped.Load())
		return ErrQueueFull
	}
}

func (p *Processor) Stop() {
	p.logger.Info("processor stopping, draining queue", "processed", p.processed.Load(), "dropped", p.dropped.Load())
	p.cancel()
	close(p.queue)
	p.wg.Wait()
	p.logger.Info("processor stopped", "total_processed", p.processed.Load(), "total_dropped", p.dropped.Load())
}

func (p *Processor) GetMetrics() map[string]int64 {
	return map[string]int64{
		"processed": p.processed.Load(),
		"dropped":   p.dropped.Load(),
		"queued":    int64(len(p.queue)),
	}
}

func (p *Processor) worker(id int) {
	defer p.wg.Done()
	p.logger.Info("worker started", "worker_id", id)
	for {
		select {
		case <-p.ctx.Done():
			p.logger.Info("worker stopping", "worker_id", id)
			return
		case event, ok := <-p.queue:
			if !ok {
				p.logger.Info("worker queue closed", "worker_id", id)
				return
			}
			p.processWithRetry(event)
			p.processed.Add(1)
		}
	}
}

func (p *Processor) processWithRetry(event events.Message) {
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-p.ctx.Done():
				return
			case <-time.After(p.retryDelay * time.Duration(attempt)):
			}
		}

		p.dispatcher.Dispatch(p.ctx, event)
		if ctx := p.ctx; ctx.Err() == nil {
			return
		}
		if attempt < p.maxRetries {
			p.logger.Warn("dispatch failed, retrying", "event_id", event.ID, "attempt", attempt+1, "max_retries", p.maxRetries)
		}
	}
	p.logger.Error("dispatch failed after retries", "event_id", event.ID, "max_retries", p.maxRetries)
}
