# 🚀 Quick Test Guide - Direct Messaging

## Test User-to-User Communication in 5 Minutes

### Prerequisites
- Server running on `localhost:8080`
- Two browser windows/tabs

---

## Step 1: Start the Server (30 seconds)

```bash
cd backend
go run main.go
```

You should see:
```
Successfully connected to the database!
Database migration completed!
Server starting on :8080
```

---

## Step 2: Register Two Test Users (1 minute)

**User 1:**
```bash
curl -X POST http://localhost:8080/api/realtime-chat/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@test.com","password":"password123"}'
```

**User 2:**
```bash
curl -X POST http://localhost:8080/api/realtime-chat/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"bob@test.com","password":"password123"}'
```

---

## Step 3: Open Test Client (30 seconds)

1. **Browser Window 1:**
   ```bash
   open backend/test_direct_messaging.html
   ```

2. **Browser Window 2:**
   - Open another browser window/tab
   - Open the same file: `backend/test_direct_messaging.html`

---

## Step 4: Login as Different Users (1 minute)

**Window 1 (Alice):**
- Email: `alice@test.com`
- Password: `password123`
- Click **"Login"**
- Click **"Connect WebSocket"**

**Window 2 (Bob):**
- Email: `bob@test.com`
- Password: `password123`
- Click **"Login"**
- Click **"Connect WebSocket"**

---

## Step 5: Send Direct Messages (2 minutes)

### Alice sends to Bob:

1. In **Window 1 (Alice)**:
   - You'll see "Bob is now online" notification
   - Click **"Refresh"** button in the Online Users section
   - Click on **"bob@test.com"** in the online users list
   - Type: `"Hi Bob! How are you?"`
   - Click **"Send"**

2. In **Window 2 (Bob)**:
   - You'll instantly see Alice's message appear!

### Bob replies to Alice:

1. In **Window 2 (Bob)**:
   - Click **"Refresh"** to see online users
   - Click on **"alice@test.com"**
   - Type: `"Hi Alice! I'm doing great, thanks!"`
   - Click **"Send"**

2. In **Window 1 (Alice)**:
   - You'll instantly see Bob's reply!

---

## 🎉 Success!

You've successfully established **direct user-to-user communication**!

---

## Additional Tests

### Test Broadcast Messages

1. In either window, type a message
2. Click **"Broadcast to All"** instead of "Send"
3. The message will be sent to **all online users**

### Test Online/Offline Notifications

1. Close one browser window
2. In the other window, you'll see: `"user@test.com is now offline"`
3. Reopen and reconnect
4. You'll see: `"user@test.com is now online"`

### Test Multiple Users

1. Open a third browser window
2. Register and login as `charlie@test.com`
3. All three users can now see each other and chat!

---

## 📊 What You Can See

### In Each Window:

✅ **Connection Status** - Green "Connected" badge  
✅ **Online Users List** - All currently connected users  
✅ **Direct Messages** - Private messages (blue, right-aligned for sent)  
✅ **Received Messages** - Messages from others (white, left-aligned)  
✅ **System Notifications** - Yellow boxes for user status changes  
✅ **Timestamps** - Time each message was sent  

---

## 🔍 Verify Communication Flow

### Check Server Logs:

```
Client registered: <uuid> (alice@test.com)
Client registered: <uuid> (bob@test.com)
Direct message sent from alice@test.com to bob@test.com
Direct message sent from bob@test.com to alice@test.com
```

### Check Browser Console:

Press `F12` to open Developer Tools and check the Console tab:
```
✅ Logged in as alice@test.com
✅ WebSocket connected
Received: {from: "...", from_email: "bob@test.com", content: "Hi Alice!", ...}
```

---

## 🐛 Troubleshooting

### "Login failed"
- Make sure the server is running
- Check if users are already registered
- Verify database connection

### "WebSocket error"
- Ensure you logged in first
- Check if the access token is valid (15 min expiry)
- Verify server is running on port 8080

### "No other users online"
- Make sure both users are connected
- Click the "Refresh" button
- Check if the other user's WebSocket is connected (green badge)

### Messages not appearing
- Verify both users are connected (green "Connected" status)
- Check that you selected a recipient before sending
- Look at browser console for errors

---

## 📝 API Endpoints Used

| Endpoint | Purpose |
|----------|---------|
| `POST /api/realtime-chat/auth/register` | Create user account |
| `POST /api/realtime-chat/auth/login` | Get JWT token |
| `GET /api/realtime-chat/chat/online-users` | Get list of online users |
| `ws://localhost:8080/api/realtime-chat/ws` | WebSocket connection |

---

## 🎯 Message Format

### Sending Direct Message:
```json
{
  "to": "recipient-user-id",
  "content": "Your message here",
  "type": "direct"
}
```

### Receiving Message:
```json
{
  "from": "sender-user-id",
  "from_email": "sender@test.com",
  "to": "your-user-id",
  "to_email": "you@test.com",
  "content": "Message content",
  "type": "direct",
  "timestamp": "2024-12-05T10:30:00Z"
}
```

---

## ✨ Features Demonstrated

✅ Real-time bidirectional communication  
✅ Direct user-to-user messaging  
✅ Online user presence  
✅ System notifications  
✅ JWT authentication  
✅ Multiple concurrent users  
✅ Broadcast messaging  

---

**Enjoy testing! 🎊**

