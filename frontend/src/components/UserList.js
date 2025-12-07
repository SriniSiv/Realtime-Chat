import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { chatAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';

const UserList = ({ users, selectedUser, onSelectUser, onRefresh }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState(null);
  const [isSearching, setIsSearching] = useState(false);
  const { getValidAccessToken } = useAuth();

  // Debounced search function
  const searchUsers = useCallback(async (query) => {
    if (!query.trim()) {
      setSearchResults(null);
      return;
    }

    setIsSearching(true);
    try {
      const token = await getValidAccessToken();
      const response = await chatAPI.searchUsers(token, query);
      if (response.users) {
        setSearchResults(response.users);
      }
    } catch (error) {
      console.error('Search failed:', error);
    } finally {
      setIsSearching(false);
    }
  }, [getValidAccessToken]);

  // Debounce search input
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      searchUsers(searchQuery);
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [searchQuery, searchUsers]);

  // Use search results if available, otherwise use all users
  const displayUsers = searchResults !== null ? searchResults : users;

  // Sort users: online first, then offline
  const sortedUsers = useMemo(() => {
    return [...displayUsers].sort((a, b) => {
      if (a.is_online === b.is_online) return 0;
      return a.is_online ? -1 : 1;
    });
  }, [displayUsers]);

  const onlineCount = users.filter(u => u.is_online).length;

  return (
    <div className="user-list">
      <div className="user-list-header">
        <span>Users ({onlineCount} online)</span>
        <button onClick={onRefresh} className="refresh-btn" title="Refresh">
          🔄
        </button>
      </div>

      <div className="user-search">
        <input
          type="text"
          placeholder="🔍 Search users..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="search-input"
        />
        {isSearching && <span className="search-loading">...</span>}
      </div>

      <div className="user-list-content">
        {sortedUsers.length === 0 ? (
          <div className="no-users">
            <p>{searchQuery ? 'No users found' : 'No other users'}</p>
            <small>{searchQuery ? 'Try a different search' : 'Invite others to join...'}</small>
          </div>
        ) : (
          sortedUsers.map(user => (
            <div
              key={user.id}
              className={`user-item ${selectedUser?.id === user.id ? 'selected' : ''} ${!user.is_online ? 'offline' : ''}`}
              onClick={() => onSelectUser(user)}
            >
              <div className="user-avatar">
                {(user.username || user.email).charAt(0).toUpperCase()}
              </div>
              <div className="user-info">
                <span className="user-name">{user.username || user.email}</span>
                <span className={`user-status ${user.is_online ? 'online' : 'offline'}`}>
                  ● {user.is_online ? 'Online' : 'Offline'}
                </span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default UserList;

