# 🏗️ Architecture Documentation

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                             │
│  (Web Browser, Mobile App, Postman, cURL, WebSocket Client)     │
└─────────────────────────────────────────────────────────────────┘
                              ↓ HTTP/WebSocket
┌─────────────────────────────────────────────────────────────────┐
│                      GIN WEB FRAMEWORK                           │
│                    (Port 8080 - main.go)                         │
└─────────────────────────────────────────────────────────────────┘
                              ↓
        ┌─────────────────────┴─────────────────────┐
        ↓                                           ↓
┌──────────────────┐                    ┌──────────────────────┐
│   HTTP ROUTES    │                    │  WEBSOCKET HANDLER   │
│  (routes/api.go) │                    │ (websocket/handler)  │
└──────────────────┘                    └──────────────────────┘
        ↓                                           ↓
┌──────────────────┐                    ┌──────────────────────┐
│  AUTH MIDDLEWARE │                    │   WEBSOCKET HUB      │
│(middleware/auth) │                    │  (websocket/hub.go)  │
└──────────────────┘                    └──────────────────────┘
        ↓                                           ↓
┌──────────────────┐                    ┌──────────────────────┐
│   CONTROLLER     │                    │   CLIENT MANAGER     │
│(controller/users)│                    │  (Active Clients)    │
└──────────────────┘                    └──────────────────────┘
        ↓
┌──────────────────┐
│     SERVICE      │
│ (service/users)  │
│  Business Logic  │
└──────────────────┘
        ↓
┌──────────────────┐
│   REPOSITORY     │
│   (db/users)     │
│  DB Operations   │
└──────────────────┘
        ↓
┌──────────────────┐
│   GORM ORM       │
└──────────────────┘
        ↓
┌──────────────────┐
│   PostgreSQL     │
│    Database      │
└──────────────────┘
```

## Request Flow Diagrams

### 1. Registration Flow

```
Client                Controller           Service              Repository          Database
  │                       │                   │                     │                  │
  │──POST /register──────>│                   │                     │                  │
  │  {email, password}    │                   │                     │                  │
  │                       │                   │                     │                  │
  │                       │──Validate Input──>│                     │                  │
  │                       │                   │                     │                  │
  │                       │                   │──Check Email────────>│                  │
  │                       │                   │   Exists            │                  │
  │                       │                   │                     │──SELECT──────────>│
  │                       │                   │                     │<─Result──────────│
  │                       │                   │<─Email Available────│                  │
  │                       │                   │                     │                  │
  │                       │                   │──Hash Password      │                  │
  │                       │                   │  (bcrypt)           │                  │
  │                       │                   │                     │                  │
  │                       │                   │──Create User────────>│                  │
  │                       │                   │                     │──INSERT──────────>│
  │                       │                   │                     │<─User Created────│
  │                       │                   │<─User Object────────│                  │
  │                       │                   │                     │                  │
  │                       │                   │──Generate JWT       │                  │
  │                       │                   │  (Access+Refresh)   │                  │
  │                       │                   │                     │                  │
  │                       │                   │──Store Refresh──────>│                  │
  │                       │                   │   Token             │──INSERT──────────>│
  │                       │                   │                     │<─Token Stored────│
  │                       │<─Auth Response────│                     │                  │
  │<─201 Created──────────│                   │                     │                  │
  │  {tokens, user}       │                   │                     │                  │
```

### 2. Login Flow

```
Client                Controller           Service              Repository          Database
  │                       │                   │                     │                  │
  │──POST /login─────────>│                   │                     │                  │
  │  {email, password}    │                   │                     │                  │
  │                       │                   │                     │                  │
  │                       │──Validate Input──>│                     │                  │
  │                       │                   │                     │                  │
  │                       │                   │──Get User By────────>│                  │
  │                       │                   │   Email             │                  │
  │                       │                   │                     │──SELECT──────────>│
  │                       │                   │                     │<─User Data───────│
  │                       │                   │<─User Object────────│                  │
  │                       │                   │                     │                  │
  │                       │                   │──Verify Password    │                  │
  │                       │                   │  (bcrypt.Compare)   │                  │
  │                       │                   │                     │                  │
  │                       │                   │──Generate JWT       │                  │
  │                       │                   │  (Access+Refresh)   │                  │
  │                       │                   │                     │                  │
  │                       │                   │──Store Refresh──────>│                  │
  │                       │                   │   Token             │──INSERT──────────>│
  │                       │<─Auth Response────│                     │                  │
  │<─200 OK───────────────│                   │                     │                  │
  │  {tokens, user}       │                   │                     │                  │
```

### 3. Protected Route Flow (GET /me)

```
Client                Middleware           Controller           Service              Repository
  │                       │                   │                   │                     │
  │──GET /me─────────────>│                   │                   │                     │
  │  Authorization:       │                   │                   │                     │
  │  Bearer <token>       │                   │                   │                     │
  │                       │                   │                   │                     │
  │                       │──Extract Token    │                   │                     │
  │                       │                   │                   │                     │
  │                       │──Validate JWT     │                   │                     │
  │                       │  (utils.Validate) │                   │                     │
  │                       │                   │                   │                     │
  │                       │──Set User ID──────>│                   │                     │
  │                       │   in Context      │                   │                     │
  │                       │                   │                   │                     │
  │                       │                   │──Get User ID      │                     │
  │                       │                   │   from Context    │                     │
  │                       │                   │                   │                     │
  │                       │                   │──Get Profile──────>│                     │
  │                       │                   │                   │                     │
  │                       │                   │                   │──Get User By ID────>│
  │                       │                   │                   │<─User Data─────────│
  │                       │                   │<─User DTO─────────│                     │
  │<─200 OK───────────────│<──────────────────│                   │                     │
  │  {id, email}          │                   │                   │                     │
```

### 4. WebSocket Connection Flow

```
Client              WS Handler           Hub                 Connection          Database
  │                     │                 │                      │                  │
  │──WS Connect────────>│                 │                      │                  │
  │  ?token=<jwt>       │                 │                      │                  │
  │                     │                 │                      │                  │
  │                     │──Extract Token  │                      │                  │
  │                     │                 │                      │                  │
  │                     │──Validate JWT   │                      │                  │
  │                     │  (utils.Validate)                      │                  │
  │                     │                 │                      │                  │
  │                     │──Upgrade HTTP   │                      │                  │
  │                     │   to WebSocket  │                      │                  │
  │<────Upgraded────────│                 │                      │                  │
  │                     │                 │                      │                  │
  │                     │──Create Client──>│                      │                  │
  │                     │                 │                      │                  │
  │                     │                 │──Register Client     │                  │
  │                     │                 │   (Add to Map)       │                  │
  │                     │                 │                      │                  │
  │                     │                 │──Start Read Pump─────>│                  │
  │                     │                 │                      │                  │
  │                     │                 │──Start Write Pump────>│                  │
  │                     │                 │                      │                  │
  │<────Connected───────│                 │                      │                  │
  │                     │                 │                      │                  │
  │──Send Message──────────────────────────────────────────────>│                  │
  │                     │                 │                      │                  │
  │                     │                 │<─Broadcast Message───│                  │
  │                     │                 │                      │                  │
  │                     │                 │──Send to All Clients─>                  │
  │<────Message─────────│<────────────────│                      │                  │
```

## Component Responsibilities

### Routes Layer (`routes/api.go`)
**Responsibility**: Define API endpoints and apply middleware
- Maps HTTP methods to controller functions
- Groups related endpoints
- Applies authentication middleware to protected routes

### Controller Layer (`controller/users.go`)
**Responsibility**: Handle HTTP requests and validate input
- Parse request body
- Validate request data using Gin binding
- Call service layer functions
- Return HTTP responses with appropriate status codes

### Service Layer (`service/users.go`)
**Responsibility**: Implement business logic
- User registration logic
- Authentication logic
- Token generation and validation
- Password hashing and verification
- Coordinate between controller and repository

### Repository Layer (`db/users.go`)
**Responsibility**: Database operations
- CRUD operations for users
- CRUD operations for refresh tokens
- Query execution
- Error handling for database operations

### Middleware Layer (`middleware/auth.go`)
**Responsibility**: Request authentication
- Extract JWT token from headers
- Validate token
- Set user context
- Block unauthorized requests

### WebSocket Layer (`websocket/`)
**Responsibility**: Real-time communication
- **Hub**: Manage active clients, broadcast messages
- **Connection**: Handle read/write operations, ping/pong
- **Handler**: Authenticate and upgrade HTTP to WebSocket

### Utils Layer (`utils/jwt.go`)
**Responsibility**: JWT token operations
- Generate access tokens
- Generate refresh tokens
- Validate tokens
- Extract claims from tokens

### Models Layer (`models/`)
**Responsibility**: Data structures
- **user.go**: Database models (User, RefreshToken)
- **dto.go**: Request/Response DTOs

### Config Layer (`config/db.go`)
**Responsibility**: Application configuration
- Database connection
- GORM setup
- Auto-migration

## Data Flow

### Authentication Data Flow
```
Request → Controller → Service → Repository → Database
                ↓
            JWT Utils
                ↓
            Response
```

### WebSocket Data Flow
```
Client → WebSocket Handler → Hub → All Connected Clients
            ↑
        JWT Validation
```

## Security Layers

```
┌─────────────────────────────────────┐
│   Input Validation (Gin Binding)    │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   JWT Authentication (Middleware)   │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   Password Hashing (bcrypt)         │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   SQL Injection Prevention (GORM)   │
└─────────────────────────────────────┘
```

## Scalability Considerations

1. **Horizontal Scaling**: Stateless design allows multiple instances
2. **Database Connection Pooling**: GORM manages connection pool
3. **WebSocket Hub**: Can be extended to use Redis for multi-instance support
4. **Repository Pattern**: Easy to swap database implementations
5. **Layered Architecture**: Each layer can be scaled independently

## Technology Stack

```
┌─────────────────────────────────────┐
│         Application Layer           │
│    Go 1.23+ with Gin Framework      │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│         ORM Layer                   │
│         GORM v1.31.1                │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│         Database Layer              │
│         PostgreSQL                  │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│      Authentication Layer           │
│   JWT (golang-jwt/jwt/v5)           │
│   bcrypt (golang.org/x/crypto)      │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│      WebSocket Layer                │
│   Gorilla WebSocket v1.5.3          │
└─────────────────────────────────────┘
```

---

This architecture provides:
- ✅ Clear separation of concerns
- ✅ Easy to test and maintain
- ✅ Scalable and extensible
- ✅ Follows Go best practices
- ✅ Production-ready design

