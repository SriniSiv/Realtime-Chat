# Realtime Chat - Architecture Documentation


## 1. High Level Design (HLD)

### System Architecture
```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│    Clients   │ ───► │Load Balancer │ ───► │   Backend    │
│  (React App) │      │   (Nginx)    │      │  (Go + Gin)  │
└──────────────┘      └──────────────┘      └──────┬───────┘
                                                   │
                 ┌─────────────────────────────────┼─────────────────────────────────┐
                 │                                 │                                 │
                 ▼                                 ▼                                 ▼
          ┌─────────────┐                  ┌─────────────┐                   ┌─────────────┐
          │    Redis    │                  │    Kafka    │                   │ PostgreSQL  │
          │  (Pub/Sub)  │                  │  (Queue)    │                   │    (DB)     │
          └─────────────┘                  └─────────────┘                   └─────────────┘
```

### Component Responsibilities

React Frontend

. UI rendering and state management
. WebSocket client for real-time updates

Nginx

. Load balancing across backend instances
. SSL termination
. Sticky sessions for WebSocket

Go Backend

. Exposes REST API endpoints
. Manages a WebSocket hub for real-time messaging
. Implements the core business logic

PostgreSQL

. Persists users, messages, and groups

Redis

. Provides cross-instance Pub/Sub for message synchronization
. Tracks user presence (online/offline)

Kafka

. Acts as a message queue for asynchronous database persistence

### Key Design Decisions

Dual Messaging (Redis + Kafka)

. Redis for real-time delivery (fast, ephemeral)
. Kafka for durable persistence (reliable, ordered)

UUID Primary Keys

. Distributed ID generation
. No coordination needed between servers

Stateless Backend

. Easy horizontal scaling
. Any server can handle any request

Consumer Groups

. Exactly-once message persistence
. No duplicate writes to database

---

## 2. Low Level Design (LLD)

### Backend Layered Architecture
```
┌─────────────────────────────────────────────┐
│              ROUTES LAYER                    │  ← Endpoint definitions
├─────────────────────────────────────────────┤
│            CONTROLLER LAYER                  │  ← Request validation, response
├─────────────────────────────────────────────┤
│             SERVICE LAYER                    │  ← Business logic
├─────────────────────────────────────────────┤
│               DB LAYER                       │  ← GORM operations
└─────────────────────────────────────────────┘
```

### Database Schema (ERD)
```
┌─────────────┐       ┌─────────────────┐       ┌─────────────┐
│   USERS     │       │   MESSAGES      │       │   GROUPS    │
├─────────────┤       ├─────────────────┤       ├─────────────┤
│ id (UUID)   │◄──────│ sender_id (FK)  │       │ id (UUID)   │
│ email       │       │ receiver_id (FK)│───────│ name        │
│ username    │       │ group_id (FK)   │──────►│ created_by  │
│ password    │       │ content         │       │ created_at  │
│ created_at  │       │ type            │       └──────┬──────┘
└─────────────┘       │ created_at      │              │
                      └─────────────────┘              │
                                                       ▼
                                            ┌─────────────────┐
                                            │ GROUP_MEMBERS   │
                                            ├─────────────────┤
                                            │ group_id (FK)   │
                                            │ user_id (FK)    │
                                            │ joined_at       │
                                            └─────────────────┘
```

### WebSocket Hub Design
```
                    ┌─────────────────────────────────┐
                    │           HUB                   │
                    │  ┌───────────────────────────┐  │
  register ────────►│  │ clients map[id]*Client   │  │
  unregister ──────►│  │                           │  │
  broadcast ───────►│  │ Channels:                 │  │
  direct ──────────►│  │  - register               │  │
                    │  │  - unregister             │  │
                    │  │  - broadcast              │  │
                    │  │  - direct                 │  │
                    │  └───────────────────────────┘  │
                    └─────────────────────────────────┘
                                   │
                    ┌──────────────┼──────────────┐
                    ▼              ▼              ▼
              ┌──────────┐  ┌──────────┐  ┌──────────┐
              │ Client 1 │  │ Client 2 │  │ Client N │
              │ ReadPump │  │ ReadPump │  │ ReadPump │
              │WritePump │  │WritePump │  │WritePump │
              └──────────┘  └──────────┘  └──────────┘
```

### API Endpoints

Authentication

. POST /auth/register - Register new user
. POST /auth/login - Login and get JWT tokens
. POST /auth/refresh - Refresh access token

Chat

. POST /chat/users - Filter/search users
. GET /chat/history - Get chat history between users

Groups

. POST /groups/create - Create new group
. POST /groups - Filter/search groups
. GET /groups/:id/messages - Get group messages

WebSocket

. WS /ws?token=<jwt> - Real-time connection

---

## 3. Data Flow: Message Sent & Received

### Direct Message Flow
```
Sender ──► WebSocket ──► Hub ──┬──► Redis Pub/Sub ──► Other Server Instances ──► Recipient
                               │
                               └──► Kafka ──► Consumer ──► PostgreSQL (persist)
```

Steps

. Client sends message via WebSocket
. Hub delivers locally if recipient on same server
. Redis Pub/Sub broadcasts to other instances
. Kafka queues message for async DB persistence

### Group Message Flow
```
Sender ──► Hub ──► Lookup Members ──► Fan-out to all members (local + Redis)
```

---

## 4. Technology Choices

### PostgreSQL

Why

. ACID transactions for data integrity
. Relational model fits Users, Groups, Messages with FKs
. Native JSONB and UUID support

Trade-off

. Slower writes (mitigated by Kafka async persistence)
. Schema migrations needed for changes

### Redis

Why

. Pub/Sub for cross-instance message sync
. Sub-millisecond latency for real-time
. User online/offline presence tracking

Trade-off

. Data loss on crash (acceptable for presence data)
. Memory limited storage

### Kafka

Why

. Durable message persistence with ordering
. Exactly-once delivery via consumer groups
. Partition by sender_id for message ordering

Trade-off

. Operational complexity (Zookeeper, brokers)
. Adds latency for persistence path

### WebSocket

Why

. Bidirectional, low latency communication
. Server push without client polling
. Perfect for real-time chat

Trade-off

. Stateful connections (needs sticky sessions)
. Connection limits per server instance

### Golang

Why

. Goroutines handle 10K+ concurrent connections
. Low memory footprint (~50MB per 1K connections)
. Simple deployment (single binary)

Trade-off

. Verbose error handling
. Smaller ecosystem than Node.js/Python

---

## 5. Horizontal Scaling

```
        Load Balancer (Sticky Sessions)
                    │
    ┌───────────────┼───────────────┐
    ▼               ▼               ▼
┌────────┐     ┌────────┐     ┌────────┐
│Server 1│◄───►│Server 2│◄───►│Server 3│  ← Redis Sync
└────────┘     └────────┘     └────────┘
```

Backend

. Add more instances (stateless design)

PostgreSQL

. Add read replicas for scaling reads

Redis

. Use Redis Cluster for horizontal scaling

Kafka

. Add partitions for higher throughput

---

## 6. Fault Tolerance

Server Crash

. Clients auto-reconnect to other available instance

PostgreSQL Down

. Promote read replica to primary

Kafka Broker Down

. Partition rebalance + producer retry

Redis Down

. Sentinel failover to replica

---

## 7. Load Balancing

WebSocket

. Sticky sessions using ip_hash
. Same client always connects to same server

REST API

. Round-robin distribution

---

## 8. Message Consistency

Ordering

. Kafka partition by sender_id ensures order

No Loss

. Kafka durability + acknowledgments

No Duplicates

. Single consumer group handles each message once

---

## 9. Golang SDK

### Features

Initialization

. sdk.NewClient(Config{}) - Connect with APIKey + Endpoint

Send Direct Message

. SendMessage(to, content) - Send to specific user

Send Group Message

. SendGroupMessage(groupID, content) - Send to group

Receive Messages

. OnMessage(handler) - Callback for incoming messages

Graceful Shutdown

. Close() - Disconnect cleanly

### Usage Example
```go
// a. Initialization
client, _ := sdk.NewClient(sdk.Config{
    APIKey:   "jwt-access-token",
    Endpoint: "ws://localhost:8080/api/realtime-chat/ws",
})

// b. Sending Messages
client.SendMessage("user-id", "Hello!")
client.SendGroupMessage("group-id", "Hello group!")

// Receive messages
client.OnMessage(func(msg sdk.Message) {
    fmt.Printf("From %s: %s\n", msg.From, msg.Content)
})

// c. Graceful Shutdown
client.Close()
```

---

## 10. Summary

Real-time Delivery

. WebSocket + Redis Pub/Sub

Persistence

. Kafka → PostgreSQL

Scaling

. Stateless servers + Redis sync

Fault Tolerance

. Auto-reconnect + replicas

Load Balancing

. Sticky sessions for WebSocket

Consistency

. Kafka ordering guarantees

Developer SDK

. Go SDK with init, send, close


