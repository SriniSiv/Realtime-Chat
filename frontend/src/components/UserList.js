import React from 'react';

const UserList = ({ users, selectedUser, onSelectUser, onRefresh }) => {
  return (
    <div className="user-list">
      <div className="user-list-header">
        <span>Online Users ({users.length})</span>
        <button onClick={onRefresh} className="refresh-btn" title="Refresh">
          🔄
        </button>
      </div>
      <div className="user-list-content">
        {users.length === 0 ? (
          <div className="no-users">
            <p>No other users online</p>
            <small>Waiting for others to join...</small>
          </div>
        ) : (
          users.map(user => (
            <div
              key={user.id}
              className={`user-item ${selectedUser?.id === user.id ? 'selected' : ''}`}
              onClick={() => onSelectUser(user)}
            >
              <div className="user-avatar">
                {user.email.charAt(0).toUpperCase()}
              </div>
              <div className="user-info">
                <span className="user-name">{user.email}</span>
                <span className="user-status">● Online</span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default UserList;

