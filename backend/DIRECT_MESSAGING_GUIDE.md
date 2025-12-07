# 💬 Direct Messaging Guide - User-to-User Communication

## Overview

This guide explains how User 1 can send messages to User 2 and establish direct communication in the Realtime Chat application.

---

## 🔄 Communication Flow

### Step-by-Step: How User 1 Sees User 2's Message

```
┌─────────────┐                  ┌─────────────┐                  ┌─────────────┐
│   User 1    │                  │   Server    │                  │   User 2    │
│  (Browser)  │                  │  (WebSocket │                  │  (Browser)  │
│             │                  │     Hub)    │                  │             │
└─────────────┘                  └─────────────┘                  └─────────────┘
       │                                │                                │
       │  1. Login & Get JWT            │                                │
       ├───────────────────────────────>│                                │
       │<───────────────────────────────┤                                │
       │                                │                                │
       │  2. Connect WebSocket          │                                │
       │     with JWT token             │                                │
       ├───────────────────────────────>│                                │
       │<─── Connected ─────────────────┤                                │
       │                                │                                │
       │                                │  3. Login & Get JWT            │
       │                                │<───────────────────────────────┤
       │                                ├───────────────────────────────>│
       │                                │                                │
       │                                │  4. Connect WebSocket          │
       │                                │<───────────────────────────────┤
       │                                ├─── Connected ─────────────────>│
       │                                │                                │
       │<─── "User2 is online" ─────────┤                                │
       │                                │                                │
       │  5. Get Online Users           │                                │
       ├───────────────────────────────>│                                │
       │<─── [User2 info] ──────────────┤                                │
       │                                │                                │
       │                                │  6. Send Direct Message        │
       │                                │     to User 1                  │
       │                                │<───────────────────────────────┤
       │                                │                                │
       │<─── Message from User2 ────────┤                                │
       │     (Direct delivery)          │                                │
       │                                │                                │
       │  7. Reply to User 2            │                                │
       ├───────────────────────────────>│                                │
       │                                ├───────────────────────────────>│
       │                                │     (Direct delivery)          │
```

---

## 📋 Implementation Details

### 1. **User Registration & Login**

Both users must first register and login to get JWT tokens.

**API Endpoints:**
- `POST /api/realtime-chat/auth/register` - Create account
- `POST /api/realtime-chat/auth/login` - Get JWT token

**Example:**
```bash
# User 1 Login
curl -X POST http://localhost:8080/api/realtime-chat/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user1@test.com", "password": "password123"}'

# Response:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user1@test.com"
  }
}
```

### 2. **WebSocket Connection**

Each user connects to the WebSocket endpoint with their JWT token.

**WebSocket Endpoint:**
```
ws://localhost:8080/api/realtime-chat/ws?token=<JWT_ACCESS_TOKEN>
```

**JavaScript Example:**
```javascript
const token = "eyJhbGc..."; // From login response
const ws = new WebSocket(`ws://localhost:8080/api/realtime-chat/ws?token=${token}`);

ws.onopen = () => {
    console.log('Connected to WebSocket');
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
};
```

### 3. **Get Online Users**

User 1 can fetch the list of currently online users to see who they can message.

**API Endpoint:**
```
GET /api/realtime-chat/chat/online-users
Authorization: Bearer <JWT_ACCESS_TOKEN>
```

**Response:**
```json
{
  "online_users": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "email": "user2@test.com"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "email": "user3@test.com"
    }
  ],
  "count": 2
}
```

### 4. **Send Direct Message**

User 2 sends a direct message to User 1.

**Message Format (JSON):**
```json
{
  "to": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Hello User 1! How are you?",
  "type": "direct"
}
```

**JavaScript Example:**
```javascript
function sendDirectMessage(recipientId, content) {
    const message = {
        to: recipientId,
        content: content,
        type: 'direct'
    };
    ws.send(JSON.stringify(message));
}

// Send message to User 1
sendDirectMessage('550e8400-e29b-41d4-a716-446655440000', 'Hello User 1!');
```

### 5. **Receive Message**

User 1 receives the message through their WebSocket connection.

**Received Message Format:**
```json
{
  "from": "550e8400-e29b-41d4-a716-446655440001",
  "from_email": "user2@test.com",
  "to": "550e8400-e29b-41d4-a716-446655440000",
  "to_email": "user1@test.com",
  "content": "Hello User 1! How are you?",
  "type": "direct",
  "timestamp": "2024-12-05T10:30:00Z"
}
```

**JavaScript Handler:**
```javascript
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    
    if (message.type === 'direct') {
        console.log(`Direct message from ${message.from_email}: ${message.content}`);
        displayMessage(message);
    } else if (message.type === 'system') {
        console.log(`System: ${message.content}`);
    }
};
```

---

## 🎯 Message Types

### 1. **Direct Message** (User-to-User)
```json
{
  "to": "user-uuid",
  "content": "Your message",
  "type": "direct"
}
```
- Sent to **one specific user**
- Only the recipient receives it

### 2. **Broadcast Message** (User-to-All)
```json
{
  "to": "all",
  "content": "Your message",
  "type": "broadcast"
}
```
- Sent to **all connected users** (except sender)
- Everyone online receives it

### 3. **System Message** (Server-to-Users)
```json
{
  "from_email": "user@test.com",
  "content": "user@test.com is now online",
  "type": "system"
}
```
- Automatically sent when users connect/disconnect
- Notifies all users about status changes

---

## 🧪 Testing Direct Messaging

### Option 1: Using the HTML Test Client

1. **Open the test file:**
   ```bash
   open backend/test_direct_messaging.html
   ```

2. **Login as User 1:**
   - Email: `user1@test.com`
   - Password: `password123`
   - Click "Login" then "Connect WebSocket"

3. **Open another browser tab/window:**
   - Open the same HTML file again
   - Login as User 2: `user2@test.com`
   - Click "Login" then "Connect WebSocket"

4. **Send Messages:**
   - In User 1's window, click on User 2 in the online users list
   - Type a message and click "Send"
   - User 2 will receive the message instantly

5. **Reply:**
   - In User 2's window, click on User 1
   - Type a reply and send
   - User 1 receives it immediately

### Option 2: Using cURL and wscat

**Terminal 1 (User 1):**
```bash
# Login
TOKEN1=$(curl -s -X POST http://localhost:8080/api/realtime-chat/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user1@test.com","password":"password123"}' | jq -r '.access_token')

# Connect WebSocket
wscat -c "ws://localhost:8080/api/realtime-chat/ws?token=$TOKEN1"
```

**Terminal 2 (User 2):**
```bash
# Login
TOKEN2=$(curl -s -X POST http://localhost:8080/api/realtime-chat/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user2@test.com","password":"password123"}' | jq -r '.access_token')

# Get User 1's ID
curl -s http://localhost:8080/api/realtime-chat/chat/online-users \
  -H "Authorization: Bearer $TOKEN2" | jq

# Connect WebSocket
wscat -c "ws://localhost:8080/api/realtime-chat/ws?token=$TOKEN2"

# Send message to User 1 (replace USER1_ID)
{"to":"USER1_ID","content":"Hello from User 2!","type":"direct"}
```

---

## 🔧 Backend Implementation

### WebSocket Hub (`websocket/hub.go`)

The hub manages all active WebSocket connections and routes messages:

```go
// When a message is received
case message := <-h.broadcast:
    // Enrich with sender/recipient email
    h.EnrichMessage(message)

    if message.Type == "direct" && message.To != uuid.Nil {
        // Send to specific user
        if client, ok := h.clients[message.To]; ok {
            client.Send <- messageJSON
        }
    } else {
        // Broadcast to all (except sender)
        for id, client := range h.clients {
            if id != message.From {
                client.Send <- messageJSON
            }
        }
    }
```

### Message Parsing (`websocket/hub.go`)

Incoming messages are parsed and routed:

```go
func ParseIncomingMessage(rawMessage []byte, senderID uuid.UUID) (*Message, error) {
    var incoming IncomingMessage
    json.Unmarshal(rawMessage, &incoming)

    message := &Message{
        From:      senderID,
        Content:   incoming.Content,
        Timestamp: time.Now(),
    }

    // Determine if direct or broadcast
    if incoming.To == "" || incoming.To == "all" {
        message.Type = "broadcast"
    } else {
        message.Type = "direct"
        message.To = uuid.Parse(incoming.To)
    }

    return message, nil
}
```

### Online Users API (`routes/api.go`)

```go
chat.GET("/online-users", func(c *gin.Context) {
    users := hub.GetOnlineUsers()
    c.JSON(200, gin.H{
        "online_users": users,
        "count":        len(users),
    })
})
```

---

## 📊 Data Flow Summary

1. **User 1 & User 2** → Login → Get JWT tokens
2. **User 1 & User 2** → Connect to WebSocket with JWT
3. **Server** → Registers both users in the hub
4. **Server** → Broadcasts "User X is online" system messages
5. **User 1** → Fetches online users list → Sees User 2
6. **User 2** → Sends direct message to User 1's UUID
7. **Server Hub** → Routes message only to User 1
8. **User 1** → Receives message in real-time
9. **User 1** → Can reply directly to User 2

---

## 🎨 Features

✅ **Real-time Communication** - Instant message delivery
✅ **Direct Messaging** - Private user-to-user messages
✅ **Broadcast Messaging** - Send to all online users
✅ **Online Status** - See who's currently online
✅ **System Notifications** - User join/leave notifications
✅ **JWT Authentication** - Secure WebSocket connections
✅ **Message History** - Timestamps for all messages
✅ **User Identification** - Email and UUID for each user

---

## 🔐 Security

- ✅ JWT token required for WebSocket connection
- ✅ Token validated before connection upgrade
- ✅ User ID extracted from JWT (can't be spoofed)
- ✅ Messages include verified sender information
- ✅ Only authenticated users can see online users list

---

## 🚀 Quick Start

```bash
# 1. Start the server
cd backend
go run main.go

# 2. Register two users
curl -X POST http://localhost:8080/api/realtime-chat/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user1@test.com","password":"password123"}'

curl -X POST http://localhost:8080/api/realtime-chat/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user2@test.com","password":"password123"}'

# 3. Open test client
open backend/test_direct_messaging.html

# 4. Open another browser window/tab with the same file

# 5. Login as different users in each window and start chatting!
```

---

## 📝 Notes

- Messages are **not persisted** in the database (in-memory only)
- If a user is offline, direct messages to them are **not delivered**
- System messages notify all users when someone connects/disconnects
- The sender does **not** receive their own broadcast messages
- All timestamps are in UTC

---

## 🔮 Future Enhancements

- [ ] Message persistence in database
- [ ] Message history retrieval
- [ ] Offline message queue
- [ ] Read receipts
- [ ] Typing indicators
- [ ] File/image sharing
- [ ] Group chat rooms
- [ ] Message reactions
- [ ] User blocking
- [ ] Message encryption

---

**Happy Chatting! 💬**


