# Message Broadcast System (Star Topology using Redis Pub/Sub)

## Project Overview

This project implements a **message broadcasting system across multiple application servers** using **Redis Pub/Sub** in a **Star Topology** architecture.

In this design:
- **Redis acts as a central hub**
- All application servers publish and subscribe to Redis channels
- Messages sent by any server are **broadcast to all other servers**

The infrastructure for Redis is managed using **Terraform (Infrastructure as Code)** to ensure consistent and reproducible deployments.

---

## 🏗️ Architecture Overview

- **Topology:** Star Topology  
- **Central Node:** Redis  
- **Outer Nodes:** Application servers (publishers & subscribers)

### High-Level Flow
1. Any server publishes a message to Redis
2. Redis broadcasts the message to all subscribed servers
3. Each server independently consumes the message

This architecture provides:
- Loose coupling between services
- Horizontal scalability
- Real-time message propagation

---


## ⚙️ Requirements (Current Scope)

### Infrastructure
- **Terraform** >= 1.3
- **Docker** (for running Redis as a container)
- Access to a local or remote Docker runtime

### Tools
- Git
- Bash / Shell environment

> Application-level publisher/subscriber services will be added in later phases.

---


## 🚀 Applying Terraform Changes

All infrastructure changes are managed from the `iac/` directory.


```bash
cd iac
terraform init
terraform plan
terraform apply
```
