# 📡 Message Broadcast System  
## Star Topology using Redis Pub/Sub (Go + Terraform)

A production-ready **message broadcasting system** using Redis Pub/Sub in a **Star Topology** with thread-safe event dispatch, worker pools, graceful shutdown, and comprehensive error handling.

---

## 🏗️ Architecture

### Star Topology Design
```
        ┌─────────────┐
        │   Redis     │
        │  Pub/Sub    │
        │   (Hub)     │
        └─────────────┘
              │
    ┌─────────┼─────────┐
    │         │         │
┌───▼────┐ ┌─▼─────┐ ┌─▼────┐
│Server-1│ │Server-2│ │Server-N│
│(Sub)   │ │(Sub)   │ │(Sub)  │
└────────┘ └─┬─────┘ └──────┘
    │        │
    └────────┼─────────┐
             │         │
        ┌────▼───┐ ┌───▼─────┐
        │Publisher│ │Publisher│
        └────────┘ └────────┘
```

### Component Architecture

**Transport Layer** (`internal/redisclient/`)
- Publisher: Publishes events to Redis channel
- Subscriber: Listens to Redis channel, deserializes JSON

**Event Processing** (`internal/processor/`)
- Worker Pool: Async processing with configurable workers
- Queue: Buffered channel for backpressure handling
- Retry Logic: Exponential backoff on failures
- Metrics: Processed/dropped event counters

**Event Routing** (`internal/dispatcher/`)
- Handler Registry: Wait-free concurrent registration (RWMutex)
- Dispatch Engine: Route events by type to appropriate handlers
- Thread-Safe: Multiple concurrent dispatches supported

**Event Handlers** (`internal/handlers/`)
- Extensible pattern for different event types
- Clean interface for adding business logic

---

## ⚙️ Requirements

### Infrastructure
- **Terraform** >= 1.3
- **Docker** (for Redis container)
- Docker daemon running

### Application
- **Go** >= 1.21
- Redis server reachable over network

---

## 🚀 Quick Start

### 1. Setup Redis Infrastructure

```bash
cd iac
terraform init
terraform plan
terraform apply
```

Redis will be accessible at: `localhost:6379`

To verify Redis is running:
```bash
redis-cli -p 6379 ping
# Output: PONG
```

### 2. Run Subscribers

Open **multiple terminals** and start subscribers (they listen for events):

```bash
# Terminal 1 - Server 1
SERVER_ID=server-1 go run ./cmd/subscriber/main.go

# Terminal 2 - Server 2
SERVER_ID=server-2 go run ./cmd/subscriber/main.go

# Terminal 3 - Server N
SERVER_ID=server-n go run ./cmd/subscriber/main.go
```

Expected output:
```json
{"time":"2026-02-22T10:00:00Z","level":"INFO","msg":"subscribed to redis","channel":"broadcast.events","server_id":"server-1","component":"subscriber"}
```

### 3. Run Publisher

Open another terminal and publish events:

```bash
go run ./cmd/publisher/main.go
```

Publisher sends 5 demo events with 2-second intervals. Subscribers receive and process them.

---

## 🔧 Configuration

Environment variables for both publisher and subscriber:

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ADDR` | `localhost:6379` | Redis server address |
| `CHANNEL_NAME` | `broadcast.events` | Redis pub/sub channel |
| `SERVER_ID` | `unknown-server` | Subscriber/Publisher identifier |

Example:
```bash
REDIS_ADDR=redis.prod.example.com:6379 \
CHANNEL_NAME=events.prod \
SERVER_ID=api-server-01 \
go run ./cmd/subscriber/main.go
```

---

## 📊 Event Structure

Events are structured JSON with the following schema:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "demo.message",
  "source": "publisher",
  "timestamp": "2026-02-22T10:00:00Z",
  "payload": {
    "counter": 1,
    "text": "hello from publisher"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string (UUID) | Unique event identifier |
| `type` | string | Event type for routing to handlers |
| `source` | string | Publisher/source identifier |
| `timestamp` | ISO8601 | Event creation time |
| `payload` | object | Custom event data |

---

## 🎯 Key Features

### Thread-Safe Dispatch
- RWMutex protects handler registry
- Multiple concurrent dispatches allowed
- Single writer for handler registration

### Worker Pool Processing
- Configurable number of workers (default: 4)
- Buffered queue (default: 100 items)
- Backpressure handling with error returns

### Graceful Shutdown
- Signal handling (SIGINT, SIGTERM)
- Queue draining before exit
- WaitGroup ensures all workers complete

### Error Handling & Resilience
- Queue full detection with `ErrQueueFull`
- Exponential backoff retry (3 attempts)
- Dead-letter logging for visibility
- Atomic metrics counters

### Production-Ready Logging
- Structured JSON logging (slog)
- Context-aware log fields
- Error tracking and metrics

---

## 🔌 Extending: Adding Event Type Handlers

1. **Create a handler** in `internal/handlers/my_event.go`:

```go
package handlers

import (
    "context"
    "log/slog"
    "main/internal/events"
)

type MyEventHandler struct {
    logger *slog.Logger
}

func NewMyEventHandler(logger *slog.Logger) *MyEventHandler {
    return &MyEventHandler{logger: logger}
}

func (h *MyEventHandler) Handle(ctx context.Context, event events.Message) error {
    h.logger.Info("my.event handled", "payload", event.Payload)
    // Custom logic here
    return nil
}
```

2. **Register the handler** in `cmd/subscriber/main.go`:

```go
d := dispatcher.New(logger)
d.Register("demo.message", handlers.NewDemoMessageHandler(logger))
d.Register("my.event", handlers.NewMyEventHandler(logger))  // Add this
```

3. **Adapt publisher** to send your event type:

```go
event := events.Message{
    ID:     uuid.NewString(),
    Type:   "my.event",  // Your type
    Source: source,
    Timestamp: time.Now(),
    Payload: map[string]any{
        "custom": "data",
    },
}
```

---

## 📈 Performance & Scaling

### Worker Pool Tuning
```go
// Increase workers for CPU-bound handlers
p := processor.New(d, logger, 8, 100)  // 8 workers

// Increase buffer for bursty traffic
p := processor.New(d, logger, 4, 500)  // 500 item queue
```

### Scaling Recommendations
| Scenario | Workers | Buffer | Notes |
|----------|---------|--------|-------|
| Low throughput | 2-4 | 50-100 | Dev/testing |
| Medium throughput | 4-8 | 100-300 | Standard production |
| High throughput | 8-16 | 500-1000 | High-volume systems |
| Bursty traffic | 4-8 | 1000+ | Large spike handling |

### Monitoring Metrics
The processor exposes metrics via `GetMetrics()`:
```go
metrics := p.GetMetrics()
// {
//   "processed": 1250,
//   "dropped": 3,
//   "queued": 42
// }
```

Monitor these for:
- **High `dropped`**: Increase buffer size or add workers
- **High `queued`**: Subscribers can't keep up, scale horizontally
- **Low throughput**: Check handler performance, enable profiling

---

## 🛠️ Troubleshooting

### Redis Connection Fails
```
Error: failed to connect to redis
```
**Solution:** Verify Redis is running and accessible:
```bash
redis-cli -h localhost -p 6379 ping
```

### No Messages Received
- Ensure subscribers started before publishers
- Check `CHANNEL_NAME` matches on both sides
- Verify Redis connectivity with `redis-cli PING`

### High Message Drop Rate
1. Check `dropped` metric: `p.GetMetrics()["dropped"]`
2. Increase buffer: `processor.New(d, logger, 4, 500)`
3. Add more workers: `processor.New(d, logger, 8, 100)`

### Subscribers Don't Gracefully Shutdown
- Ensure SIGINT/SIGTERM handling enabled
- Check `p.Stop()` is called on shutdown signal
- Monitor goroutines for leaks


---

## 📝 Development


### Build Binary
```bash
go build -o subscriber ./cmd/subscriber/main.go
go build -o publisher ./cmd/publisher/main.go
```

### Run with Different Redis Host
```bash
REDIS_ADDR=redis.example.com:6380 go run ./cmd/subscriber/main.go
```
---

## 🤝 Contributing

Improvements welcome! Consider implementing:
- Unit/integration tests
- Prometheus metrics endpoint
- Dead-letter queue system
- Event validation framework
