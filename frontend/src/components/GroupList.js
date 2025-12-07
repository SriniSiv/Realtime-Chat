import React from 'react';

const GroupList = ({ groups, selectedGroup, onSelectGroup, onCreateGroup, onRefresh }) => {
  return (
    <div className="group-list">
      <div className="group-list-header">
        <span>Groups ({groups.length})</span>
        <div className="header-actions">
          <button onClick={onCreateGroup} className="create-btn" title="Create Group">
            ➕
          </button>
          <button onClick={onRefresh} className="refresh-btn" title="Refresh">
            🔄
          </button>
        </div>
      </div>

      <div className="group-list-content">
        {groups.length === 0 ? (
          <div className="no-groups">
            <p>No groups yet</p>
            <small>Click ➕ to create a group</small>
          </div>
        ) : (
          groups.map(group => (
            <div
              key={group.id}
              className={`group-item ${selectedGroup?.id === group.id ? 'selected' : ''}`}
              onClick={() => onSelectGroup(group)}
            >
              <div className="group-avatar">
                #
              </div>
              <div className="group-info">
                <span className="group-name">{group.name}</span>
                <span className="group-members">
                  {group.member_count} member{group.member_count !== 1 ? 's' : ''}
                </span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default GroupList;

