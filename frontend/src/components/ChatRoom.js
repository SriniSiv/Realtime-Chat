import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useAuth } from '../context/AuthContext';
import { chatAPI, createWebSocket } from '../services/api';
import UserList from './UserList';
import MessageList from './MessageList';
import MessageInput from './MessageInput';
import './ChatRoom.css';

const ChatRoom = () => {
  const { user, accessToken, logout } = useAuth();
  const [messages, setMessages] = useState([]);
  const [onlineUsers, setOnlineUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [connectionStatus, setConnectionStatus] = useState('connecting');
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);

  const fetchOnlineUsers = useCallback(async () => {
    try {
      const response = await chatAPI.getOnlineUsers(accessToken);
      if (response.online_users) {
        const filteredUsers = response.online_users.filter(u => u.id !== user.id);
        setOnlineUsers(filteredUsers);
      }
    } catch (err) {
      console.error('Failed to fetch online users:', err);
    }
  }, [accessToken, user.id]);

  const connectWebSocket = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const ws = createWebSocket(accessToken);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnectionStatus('connected');
      fetchOnlineUsers();
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        setMessages(prev => [...prev, message]);
        
        if (message.type === 'system') {
          setTimeout(fetchOnlineUsers, 500);
        }
      } catch (err) {
        console.error('Failed to parse message:', err);
      }
    };

    ws.onclose = () => {
      setConnectionStatus('disconnected');
      reconnectTimeoutRef.current = setTimeout(connectWebSocket, 3000);
    };

    ws.onerror = () => {
      setConnectionStatus('error');
    };
  }, [accessToken, fetchOnlineUsers]);

  useEffect(() => {
    connectWebSocket();
    
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [connectWebSocket]);

  const sendMessage = (content, type = 'direct') => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      return;
    }

    const message = {
      to: type === 'broadcast' ? 'all' : selectedUser?.id,
      content,
      type,
    };

    wsRef.current.send(JSON.stringify(message));

    // Add sent message to local state
    const sentMessage = {
      from: user.id,
      from_email: user.email,
      to: selectedUser?.id || 'all',
      to_email: selectedUser?.email || 'Everyone',
      content,
      type,
      timestamp: new Date().toISOString(),
      isSent: true,
    };
    setMessages(prev => [...prev, sentMessage]);
  };

  return (
    <div className="chat-room">
      <header className="chat-header">
        <div className="header-left">
          <h1>💬 Realtime Chat</h1>
          <span className={`status-badge ${connectionStatus}`}>
            {connectionStatus === 'connected' ? '● Connected' : 
             connectionStatus === 'connecting' ? '○ Connecting...' : '● Disconnected'}
          </span>
        </div>
        <div className="header-right">
          <span className="user-email">{user.email}</span>
          <button onClick={logout} className="logout-btn">Logout</button>
        </div>
      </header>

      <div className="chat-container">
        <UserList 
          users={onlineUsers} 
          selectedUser={selectedUser} 
          onSelectUser={setSelectedUser}
          onRefresh={fetchOnlineUsers}
        />
        <div className="chat-main">
          <div className="chat-title">
            {selectedUser ? `Chat with ${selectedUser.email}` : 'Select a user to start chatting'}
          </div>
          <MessageList messages={messages} currentUserId={user.id} />
          <MessageInput 
            onSend={sendMessage} 
            selectedUser={selectedUser}
            disabled={connectionStatus !== 'connected'}
          />
        </div>
      </div>
    </div>
  );
};

export default ChatRoom;

