const API_BASE_URL = 'http://localhost:8080/api/realtime-chat';
const WS_BASE_URL = 'ws://localhost:8080/api/realtime-chat';

// Auth API calls
export const authAPI = {
  register: async (username, email, password) => {
    const response = await fetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, email, password }),
    });
    return response.json();
  },

  login: async (email, password) => {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    return response.json();
  },

  refreshToken: async (refreshToken) => {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    return response.json();
  },

  getUserDetails: async (accessToken) => {
    const response = await fetch(`${API_BASE_URL}/auth/user-details`, {
      headers: { 'Authorization': `Bearer ${accessToken}` },
    });
    return response.json();
  },
};

// Chat API calls
export const chatAPI = {
  getOnlineUsers: async (accessToken) => {
    const response = await fetch(`${API_BASE_URL}/chat/online-users`, {
      headers: { 'Authorization': `Bearer ${accessToken}` },
    });
    return response.json();
  },

  // Get all users with online/offline status (like Slack)
  getAllUsers: async (accessToken) => {
    const response = await fetch(`${API_BASE_URL}/chat/users`, {
      headers: { 'Authorization': `Bearer ${accessToken}` },
    });
    return response.json();
  },

  // Search users by username or email
  searchUsers: async (accessToken, query) => {
    const response = await fetch(`${API_BASE_URL}/chat/users/search?q=${encodeURIComponent(query)}`, {
      headers: { 'Authorization': `Bearer ${accessToken}` },
    });
    return response.json();
  },

  // Get chat history with a specific user
  getChatHistory: async (accessToken, userId, limit = 50, offset = 0) => {
    const response = await fetch(
      `${API_BASE_URL}/chat/history?user_id=${userId}&limit=${limit}&offset=${offset}`,
      {
        headers: { 'Authorization': `Bearer ${accessToken}` },
      }
    );
    return response.json();
  },

  // Get all messages for current user
  getAllMessages: async (accessToken, limit = 50, offset = 0) => {
    const response = await fetch(
      `${API_BASE_URL}/chat/messages?limit=${limit}&offset=${offset}`,
      {
        headers: { 'Authorization': `Bearer ${accessToken}` },
      }
    );
    return response.json();
  },
};

// Group API calls
export const groupAPI = {
  // Create a new group
  createGroup: async (accessToken, name, description, memberIds = []) => {
    const response = await fetch(`${API_BASE_URL}/groups`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`,
      },
      body: JSON.stringify({ name, description, member_ids: memberIds }),
    });
    return response.json();
  },

  // Get all groups for current user
  getUserGroups: async (accessToken) => {
    const response = await fetch(`${API_BASE_URL}/groups`, {
      headers: { 'Authorization': `Bearer ${accessToken}` },
    });
    return response.json();
  },

  // Get group details
  getGroup: async (accessToken, groupId) => {
    const response = await fetch(`${API_BASE_URL}/groups/${groupId}`, {
      headers: { 'Authorization': `Bearer ${accessToken}` },
    });
    return response.json();
  },

  // Get group members
  getGroupMembers: async (accessToken, groupId) => {
    const response = await fetch(`${API_BASE_URL}/groups/${groupId}/members`, {
      headers: { 'Authorization': `Bearer ${accessToken}` },
    });
    return response.json();
  },

  // Add members to a group
  addMembers: async (accessToken, groupId, memberIds) => {
    const response = await fetch(`${API_BASE_URL}/groups/${groupId}/members`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`,
      },
      body: JSON.stringify({ member_ids: memberIds }),
    });
    return response.json();
  },

  // Get group messages
  getGroupMessages: async (accessToken, groupId, limit = 50, offset = 0) => {
    const response = await fetch(
      `${API_BASE_URL}/groups/${groupId}/messages?limit=${limit}&offset=${offset}`,
      {
        headers: { 'Authorization': `Bearer ${accessToken}` },
      }
    );
    return response.json();
  },
};

// WebSocket connection
export const createWebSocket = (accessToken) => {
  return new WebSocket(`${WS_BASE_URL}/ws?token=${accessToken}`);
};

export { API_BASE_URL, WS_BASE_URL };

