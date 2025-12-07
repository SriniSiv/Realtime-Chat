import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useAuth } from '../context/AuthContext';
import { chatAPI, groupAPI, createWebSocket } from '../services/api';
import UserList from './UserList';
import GroupList from './GroupList';
import MessageList from './MessageList';
import MessageInput from './MessageInput';
import CreateGroupModal from './CreateGroupModal';
import './ChatRoom.css';

const ChatRoom = () => {
  const { user, logout, getValidAccessToken } = useAuth();
  const [messages, setMessages] = useState([]);
  const [allUsers, setAllUsers] = useState([]);
  const [groups, setGroups] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const [connectionStatus, setConnectionStatus] = useState('connecting');
  const [loadingHistory, setLoadingHistory] = useState(false);
  const [showCreateGroup, setShowCreateGroup] = useState(false);
  const [activeTab, setActiveTab] = useState('dms'); // 'dms' or 'groups'
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const reconnectAttemptsRef = useRef(0);

  // Fetch all users with online/offline status (like Slack)
  const fetchUsers = useCallback(async () => {
    try {
      const token = await getValidAccessToken();
      if (!token) return;
      const response = await chatAPI.getAllUsers(token);
      if (response.users) {
        // Filter out current user
        const filteredUsers = response.users.filter(u => u.id !== user.id);
        setAllUsers(filteredUsers);
      }
    } catch (err) {
      console.error('Failed to fetch users:', err);
    }
  }, [getValidAccessToken, user.id]);

  // Fetch user's groups
  const fetchGroups = useCallback(async () => {
    try {
      const token = await getValidAccessToken();
      if (!token) return;
      const response = await groupAPI.getUserGroups(token);
      if (response.groups) {
        setGroups(response.groups);
      }
    } catch (err) {
      console.error('Failed to fetch groups:', err);
    }
  }, [getValidAccessToken]);

  // Fetch chat history when a user is selected
  const fetchChatHistory = useCallback(async (userId) => {
    if (!userId) return;

    setLoadingHistory(true);
    try {
      const token = await getValidAccessToken();
      if (!token) return;
      const response = await chatAPI.getChatHistory(token, userId);
      if (response.messages) {
        // Convert API response to match WebSocket message format
        const historyMessages = response.messages.map(msg => ({
          from: msg.sender_id,
          from_email: msg.sender_email,
          to: msg.receiver_id,
          content: msg.content,
          type: msg.type,
          timestamp: msg.created_at,
          isSent: msg.sender_id === user.id,
        }));
        setMessages(historyMessages);
      }
    } catch (err) {
      console.error('Failed to fetch chat history:', err);
    } finally {
      setLoadingHistory(false);
    }
  }, [getValidAccessToken, user.id]);

  // Fetch group messages
  const fetchGroupMessages = useCallback(async (groupId) => {
    if (!groupId) return;

    setLoadingHistory(true);
    try {
      const token = await getValidAccessToken();
      if (!token) return;
      const response = await groupAPI.getGroupMessages(token, groupId);
      if (response.messages) {
        const historyMessages = response.messages.map(msg => ({
          from: msg.sender_id,
          from_email: msg.sender_email,
          group_id: msg.group_id,
          content: msg.content,
          type: 'group',
          timestamp: msg.created_at,
          isSent: msg.sender_id === user.id,
        }));
        setMessages(historyMessages);
      }
    } catch (err) {
      console.error('Failed to fetch group messages:', err);
    } finally {
      setLoadingHistory(false);
    }
  }, [getValidAccessToken, user.id]);

  // Handle user selection
  const handleSelectUser = useCallback((selectedUserData) => {
    setSelectedGroup(null); // Clear group selection
    setSelectedUser(selectedUserData);
    if (selectedUserData) {
      fetchChatHistory(selectedUserData.id);
    } else {
      setMessages([]);
    }
  }, [fetchChatHistory]);

  // Handle group selection
  const handleSelectGroup = useCallback((selectedGroupData) => {
    setSelectedUser(null); // Clear user selection
    setSelectedGroup(selectedGroupData);
    if (selectedGroupData) {
      fetchGroupMessages(selectedGroupData.id);
    } else {
      setMessages([]);
    }
  }, [fetchGroupMessages]);

  // Fetch all messages for current user on login
  const fetchAllMessages = useCallback(async () => {
    try {
      const token = await getValidAccessToken();
      if (!token) return;
      const response = await chatAPI.getAllMessages(token);
      if (response.messages && response.messages.length > 0) {
        const historyMessages = response.messages.map(msg => ({
          from: msg.sender_id,
          from_email: msg.sender_email,
          to: msg.receiver_id,
          content: msg.content,
          type: msg.type,
          timestamp: msg.created_at,
          isSent: msg.sender_id === user.id,
        }));
        setMessages(historyMessages);
      }
    } catch (err) {
      console.error('Failed to fetch all messages:', err);
    }
  }, [getValidAccessToken, user.id]);

  const connectWebSocket = useCallback(async () => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    // Get a valid (fresh) token before connecting
    const token = await getValidAccessToken();
    if (!token) {
      console.error('No valid token available for WebSocket connection');
      setConnectionStatus('error');
      return;
    }

    console.log('Connecting WebSocket with fresh token...');
    const ws = createWebSocket(token);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
      setConnectionStatus('connected');
      reconnectAttemptsRef.current = 0;
      fetchUsers();
      fetchGroups();
      fetchAllMessages(); // Load all messages on login
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);

        // Don't add system messages (online/offline) to chat - refresh user list instead
        if (message.type === 'system') {
          setTimeout(fetchUsers, 500);
          return;
        }

        setMessages(prev => [...prev, message]);
      } catch (err) {
        console.error('Failed to parse message:', err);
      }
    };

    ws.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);
      setConnectionStatus('disconnected');

      // Exponential backoff for reconnection (max 30 seconds)
      const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
      reconnectAttemptsRef.current++;
      console.log(`Reconnecting in ${delay/1000} seconds...`);
      reconnectTimeoutRef.current = setTimeout(connectWebSocket, delay);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setConnectionStatus('error');
    };
  }, [getValidAccessToken, fetchUsers, fetchAllMessages]);

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
    console.log('sendMessage called:', { content, type, wsState: wsRef.current?.readyState, selectedUser, selectedGroup });

    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected. State:', wsRef.current?.readyState);
      return;
    }

    let message;
    if (selectedGroup) {
      // Group message
      message = {
        group_id: selectedGroup.id,
        content,
        type: 'group',
      };
    } else {
      // Direct or broadcast message
      message = {
        to: type === 'broadcast' ? 'all' : selectedUser?.id,
        content,
        type,
      };
    }

    console.log('Sending message via WebSocket:', message);
    wsRef.current.send(JSON.stringify(message));
    console.log('Message sent successfully');

    // Add sent message to local state
    const sentMessage = {
      from: user.id,
      from_email: user.email,
      to: selectedUser?.id || 'all',
      to_email: selectedUser?.email || 'Everyone',
      group_id: selectedGroup?.id,
      group_name: selectedGroup?.name,
      content,
      type: selectedGroup ? 'group' : type,
      timestamp: new Date().toISOString(),
      isSent: true,
    };
    setMessages(prev => [...prev, sentMessage]);
  };

  // Handle group creation
  const handleCreateGroup = async (name, description, memberIds) => {
    try {
      const token = await getValidAccessToken();
      const response = await groupAPI.createGroup(token, name, description, memberIds);
      if (response.group) {
        setGroups(prev => [...prev, response.group]);
        setShowCreateGroup(false);
        handleSelectGroup(response.group);
      }
    } catch (err) {
      console.error('Failed to create group:', err);
    }
  };

  const getConnectionStatusText = () => {
    if (connectionStatus === 'connected') return '● Connected';
    if (connectionStatus === 'connecting') return '○ Connecting...';
    return '● Disconnected';
  };

  const getChatTitle = () => {
    if (selectedGroup) return `# ${selectedGroup.name}`;
    if (selectedUser) return `Chat with ${selectedUser.username || selectedUser.email}`;
    return 'Select a conversation';
  };

  return (
    <div className="chat-room">
      <header className="chat-header">
        <div className="header-left">
          <h1>💬 Realtime Chat</h1>
          <span className={`status-badge ${connectionStatus}`}>
            {getConnectionStatusText()}
          </span>
        </div>
        <div className="header-right">
          <span className="user-email">{user.email}</span>
          <button onClick={logout} className="logout-btn">Logout</button>
        </div>
      </header>

      <div className="chat-container">
        <div className="sidebar">
          <div className="sidebar-tabs">
            <button
              className={`tab-btn ${activeTab === 'dms' ? 'active' : ''}`}
              onClick={() => setActiveTab('dms')}
            >
              💬 DMs
            </button>
            <button
              className={`tab-btn ${activeTab === 'groups' ? 'active' : ''}`}
              onClick={() => setActiveTab('groups')}
            >
              👥 Groups
            </button>
          </div>

          {activeTab === 'dms' ? (
            <UserList
              users={allUsers}
              selectedUser={selectedUser}
              onSelectUser={handleSelectUser}
              onRefresh={fetchUsers}
            />
          ) : (
            <GroupList
              groups={groups}
              selectedGroup={selectedGroup}
              onSelectGroup={handleSelectGroup}
              onCreateGroup={() => setShowCreateGroup(true)}
              onRefresh={fetchGroups}
            />
          )}
        </div>

        <div className="chat-main">
          <div className="chat-title">
            {getChatTitle()}
            {loadingHistory && <span className="loading-indicator"> Loading...</span>}
          </div>
          <MessageList
            messages={messages}
            currentUserId={user.id}
            selectedUserId={selectedUser?.id}
            selectedGroupId={selectedGroup?.id}
          />
          <MessageInput
            onSend={sendMessage}
            selectedUser={selectedUser}
            selectedGroup={selectedGroup}
            disabled={connectionStatus !== 'connected'}
          />
        </div>
      </div>

      {showCreateGroup && (
        <CreateGroupModal
          onClose={() => setShowCreateGroup(false)}
          onCreate={handleCreateGroup}
          availableUsers={allUsers}
        />
      )}
    </div>
  );
};

export default ChatRoom;

