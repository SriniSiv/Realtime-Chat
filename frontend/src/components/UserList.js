import React from 'react';

const UserList = ({ users, selectedUser, onSelectUser, onRefresh }) => {
  // Sort users: online first, then offline
  const sortedUsers = [...users].sort((a, b) => {
    if (a.is_online === b.is_online) return 0;
    return a.is_online ? -1 : 1;
  });

  const onlineCount = users.filter(u => u.is_online).length;

  return (
    <div className="user-list">
      <div className="user-list-header">
        <span>Users ({onlineCount} online)</span>
        <button onClick={onRefresh} className="refresh-btn" title="Refresh">
          🔄
        </button>
      </div>
      <div className="user-list-content">
        {sortedUsers.length === 0 ? (
          <div className="no-users">
            <p>No other users</p>
            <small>Invite others to join...</small>
          </div>
        ) : (
          sortedUsers.map(user => (
            <div
              key={user.id}
              className={`user-item ${selectedUser?.id === user.id ? 'selected' : ''} ${!user.is_online ? 'offline' : ''}`}
              onClick={() => onSelectUser(user)}
            >
              <div className="user-avatar">
                {user.email.charAt(0).toUpperCase()}
              </div>
              <div className="user-info">
                <span className="user-name">{user.email}</span>
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

