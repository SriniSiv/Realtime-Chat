import React, { useState, useMemo } from 'react';

const UserList = ({ users, selectedUser, onSelectUser, onRefresh }) => {
  const [searchQuery, setSearchQuery] = useState('');

  // Filter and sort users
  const filteredAndSortedUsers = useMemo(() => {
    let filtered = users;

    // Filter by search query
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = users.filter(user =>
        user.username?.toLowerCase().includes(query) ||
        user.email.toLowerCase().includes(query)
      );
    }

    // Sort: online first, then offline
    return [...filtered].sort((a, b) => {
      if (a.is_online === b.is_online) return 0;
      return a.is_online ? -1 : 1;
    });
  }, [users, searchQuery]);

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
      </div>

      <div className="user-list-content">
        {filteredAndSortedUsers.length === 0 ? (
          <div className="no-users">
            <p>{searchQuery ? 'No users found' : 'No other users'}</p>
            <small>{searchQuery ? 'Try a different search' : 'Invite others to join...'}</small>
          </div>
        ) : (
          filteredAndSortedUsers.map(user => (
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

