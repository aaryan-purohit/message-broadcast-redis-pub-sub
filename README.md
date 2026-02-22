# Message Broadcast System  
## Star Topology using Redis Pub/Sub (Go + Terraform)

---

## 📌 Project Overview

This project implements a **message broadcasting system across multiple application servers** using **Redis Pub/Sub** in a **Star Topology** architecture.

In this design:
- **Redis acts as a central hub**
- Multiple independent application servers connect to Redis
- Any server can publish an event
- Redis broadcasts the event to **all subscribed servers in real time**

The system is implemented in **Go**, with:
- Clean separation between transport, routing, and business logic
- Structured event messages
- Asynchronous processing with worker pools

Redis infrastructure is provisioned using **Terraform (IaC)** for consistent setup.

---


## ⚙️ Requirements (Current Scope)

### Infrastructure
- **Terraform** >= 1.3
- **Docker** (for running Redis)
- Local or remote Docker runtime

### Application
- **Go** >= 1.21
- Redis server reachable over network

### Tools
- Git
- Bash / Shell environment


---

## 🚀 Applying Terraform Changes (Redis Setup)

All infrastructure provisioning is managed from the `iac/` directory.

```bash
cd iac
terraform init
terraform plan
terraform apply
```

After apply completes, Redis will be accessible on: `localhost:6379` (or the configured Docker host)

## How to run

1. For subscribers, open multiple termianl as per your wish and run the following command 
    `SERVER_ID=server-1 run ./cmd/subscriber/main.go`
2. For publishers, open another terminal and run `go run ./cmd/publisher/main.go`