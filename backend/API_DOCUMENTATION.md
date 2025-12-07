# API Documentation

## Overview

This document provides detailed information about all available API endpoints, request/response formats, and authentication requirements.

**Base URL**: `http://localhost:8080/api/v1`

## Authentication

Most endpoints require JWT authentication. Include the access token in the Authorization header:

```
Authorization: Bearer <access_token>
```

## Endpoints

---

### 1. Register User

Create a new user account.

**Endpoint**: `POST /auth/register`

**Authentication**: Not required

**Request Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Validation Rules**:
- `email`: Required, must be valid email format
- `password`: Required, minimum 6 characters

**Success Response** (201 Created):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com"
  }
}
```

**Error Responses**:

400 Bad Request - Invalid input:
```json
{
  "error": "Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
```

400 Bad Request - Email already exists:
```json
{
  "error": "email already registered"
}
```

---

### 2. Login User

Authenticate an existing user.

**Endpoint**: `POST /auth/login`

**Authentication**: Not required

**Request Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Validation Rules**:
- `email`: Required, must be valid email format
- `password`: Required

**Success Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com"
  }
}
```

**Error Responses**:

401 Unauthorized - Invalid credentials:
```json
{
  "error": "invalid email or password"
}
```

---

### 3. Refresh Access Token

Generate a new access token using a refresh token.

**Endpoint**: `POST /auth/refresh`

**Authentication**: Not required (uses refresh token)

**Request Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Validation Rules**:
- `refresh_token`: Required

**Success Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Responses**:

401 Unauthorized - Invalid or expired refresh token:
```json
{
  "error": "invalid or expired refresh token"
}
```

401 Unauthorized - Refresh token not found:
```json
{
  "error": "refresh token not found or expired"
}
```

---

### 4. Get Current User Profile

Retrieve the authenticated user's profile information.

**Endpoint**: `GET /auth/me`

**Authentication**: Required (JWT Access Token)

**Request Headers**:
```
Authorization: Bearer <access_token>
```

**Request Body**: None

**Success Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com"
}
```

**Error Responses**:

401 Unauthorized - Missing token:
```json
{
  "error": "Authorization header required"
}
```

401 Unauthorized - Invalid token format:
```json
{
  "error": "Invalid authorization header format"
}
```

401 Unauthorized - Expired or invalid token:
```json
{
  "error": "Invalid or expired token"
}
```

404 Not Found - User not found:
```json
{
  "error": "user not found"
}
```

---

## WebSocket Connection

### Endpoint: `ws://localhost:8080/ws`

**Authentication**: Required (JWT Access Token)

**Connection Methods**:

1. **Query Parameter** (Recommended):
   ```
   ws://localhost:8080/ws?token=<access_token>
   ```

2. **Authorization Header**:
   ```
   Authorization: Bearer <access_token>
   ```

**Connection Flow**:

1. Client sends connection request with JWT token
2. Server validates the token
3. If valid, WebSocket connection is established
4. Client is registered in the hub
5. Client can send/receive messages

**Message Format**:

Messages are sent as plain text strings. The server broadcasts all messages to connected clients.

**Example Messages**:
```
Hello, everyone!
This is a test message
```

**Connection Events**:

- `onopen`: Connection established successfully
- `onmessage`: Message received from server
- `onerror`: Error occurred
- `onclose`: Connection closed

**Error Responses**:

401 Unauthorized - Missing token:
```json
{
  "error": "Authentication token required"
}
```

401 Unauthorized - Invalid token:
```json
{
  "error": "Invalid or expired token"
}
```

---

## Health Check

### Endpoint: `GET /health`

**Authentication**: Not required

**Success Response** (200 OK):
```json
{
  "status": "ok",
  "active_clients": 5
}
```

---

## Token Information

### Access Token
- **Expiration**: 15 minutes
- **Purpose**: Authenticate API requests
- **Storage**: Should be stored in memory (not localStorage)

### Refresh Token
- **Expiration**: 7 days
- **Purpose**: Generate new access tokens
- **Storage**: Can be stored in httpOnly cookies or secure storage

---

## Error Codes Summary

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Created (successful registration) |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (authentication failed) |
| 404 | Not Found (resource doesn't exist) |
| 500 | Internal Server Error |

---

## Rate Limiting

Currently, there is no rate limiting implemented. Consider adding rate limiting in production.

---

## CORS

CORS is not configured by default. Add CORS middleware if accessing from a web application.

Example using Gin CORS middleware:
```go
import "github.com/gin-contrib/cors"

router.Use(cors.Default())
```

