# Realtime Chat Backend - Authentication & WebSocket API

Complete backend implementation with user authentication and WebSocket support.

## 🏗️ Architecture

```
backend/
├── config/          # Database configuration
├── controller/      # Request validation & HTTP handlers
├── db/             # Database operations (Repository pattern)
├── middleware/     # JWT authentication middleware
├── migrations/     # SQL migration files
├── models/         # Data models & DTOs
├── routes/         # API route definitions
├── service/        # Business logic layer
├── utils/          # Utility functions (JWT)
├── websocket/      # WebSocket hub & handlers
└── main.go         # Application entry point
```

## 🚀 Tech Stack

- **Language**: Go 1.23+
- **Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT (Access + Refresh tokens)
- **WebSockets**: Gorilla WebSocket
- **Password Hashing**: bcrypt

## 📦 Installation

### 1. Install Dependencies

```bash
cd backend
go mod download
```

### 2. Setup PostgreSQL Database

```bash
# Create database
createdb realtime_chat

# Run migrations (optional - GORM auto-migrates)
psql -d realtime_chat -f migrations/script.sql
```

### 3. Configure Database Connection

Edit `config/db.go` and update the DSN:

```go
dsn := "host=localhost user=YOUR_USER password=YOUR_PASSWORD dbname=realtime_chat port=5432 sslmode=disable"
```

### 4. Run the Server

```bash
go run main.go
```

Server will start on `http://localhost:8080`

## 📡 API Endpoints

### Base URL: `http://localhost:8080/api/v1`

### 1. Register User

**POST** `/auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (201):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com"
  }
}
```

### 2. Login User

**POST** `/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com"
  }
}
```

### 3. Refresh Access Token

**POST** `/auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 4. Get Current User Profile

**GET** `/auth/me`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com"
}
```

## 🔌 WebSocket Connection

### Endpoint: `ws://localhost:8080/ws`

### Authentication

Send JWT token via query parameter or header:

**Option 1: Query Parameter**
```
ws://localhost:8080/ws?token=<access_token>
```

**Option 2: Authorization Header**
```
Authorization: Bearer <access_token>
```

### Example WebSocket Client (JavaScript)

```javascript
const token = "your_access_token_here";
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = () => {
  console.log("Connected to WebSocket");
  ws.send("Hello, Server!");
};

ws.onmessage = (event) => {
  console.log("Message from server:", event.data);
};

ws.onerror = (error) => {
  console.error("WebSocket error:", error);
};

ws.onclose = () => {
  console.log("Disconnected from WebSocket");
};
```

## 🧪 Testing with cURL

### 1. Register a new user

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 3. Get user profile (protected route)

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 4. Refresh token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### 5. Health check

```bash
curl http://localhost:8080/health
```

## 🔐 Security Features

- ✅ Password hashing with bcrypt
- ✅ JWT-based authentication (Access + Refresh tokens)
- ✅ Token expiration (15 min for access, 7 days for refresh)
- ✅ Protected routes with middleware
- ✅ WebSocket authentication
- ✅ Input validation with Gin binding
- ✅ SQL injection prevention with GORM

## 📊 Database Schema

### Users Table
```sql
id            UUID PRIMARY KEY
email         VARCHAR(255) UNIQUE NOT NULL
password_hash TEXT NOT NULL
created_at    TIMESTAMP
updated_at    TIMESTAMP
deleted_at    TIMESTAMP (soft delete)
```

### Refresh Tokens Table
```sql
id         UUID PRIMARY KEY
user_id    UUID REFERENCES users(id)
token      TEXT NOT NULL
expires_at TIMESTAMP NOT NULL
created_at TIMESTAMP
```

## 🎯 Token Expiration

- **Access Token**: 15 minutes
- **Refresh Token**: 7 days

To change these values, edit `utils/jwt.go`:

```go
AccessTokenExpiry  = 15 * time.Minute
RefreshTokenExpiry = 7 * 24 * time.Hour
```

## 🔧 Configuration

### JWT Secrets

⚠️ **IMPORTANT**: Change the JWT secrets in production!

Edit `utils/jwt.go`:

```go
AccessTokenSecret  = []byte("your-access-token-secret-key-change-this")
RefreshTokenSecret = []byte("your-refresh-token-secret-key-change-this")
```

### Database Connection

Edit `config/db.go` to update database credentials.

## 📝 Project Structure Details

### Layered Architecture

1. **Routes Layer** (`routes/api.go`)
   - Defines API endpoints
   - Groups related routes
   - Applies middleware

2. **Controller Layer** (`controller/users.go`)
   - Handles HTTP requests
   - Validates request payloads
   - Returns HTTP responses

3. **Service Layer** (`service/users.go`)
   - Contains business logic
   - Handles authentication logic
   - Manages token generation

4. **Repository Layer** (`db/users.go`)
   - Database operations
   - CRUD operations
   - Query execution

## 🌐 WebSocket Features

- ✅ JWT authentication before connection
- ✅ Active client tracking
- ✅ Broadcast messages to all clients
- ✅ Direct messaging support
- ✅ Automatic ping/pong for connection health
- ✅ Graceful connection handling

## 🚦 Error Handling

All endpoints return consistent error responses:

```json
{
  "error": "Error message here"
}
```

HTTP Status Codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (invalid/expired token)
- `404` - Not Found
- `500` - Internal Server Error

## 📚 Next Steps

1. Add environment variables for configuration
2. Implement rate limiting
3. Add email verification
4. Implement password reset
5. Add user roles and permissions
6. Implement message persistence
7. Add unit and integration tests
8. Set up Docker containerization
9. Add API documentation with Swagger
10. Implement logging with structured logger

## 🤝 Contributing

Feel free to submit issues and enhancement requests!

## 📄 License

MIT License

