import React, { useState } from 'react';

const CreateGroupModal = ({ onClose, onCreate, availableUsers }) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [selectedMembers, setSelectedMembers] = useState([]);
  const [isCreating, setIsCreating] = useState(false);

  const handleToggleMember = (userId) => {
    setSelectedMembers(prev => 
      prev.includes(userId) 
        ? prev.filter(id => id !== userId)
        : [...prev, userId]
    );
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name.trim()) return;

    setIsCreating(true);
    try {
      await onCreate(name, description, selectedMembers);
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Create New Group</h2>
          <button className="close-btn" onClick={onClose}>✕</button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="group-name">Group Name *</label>
            <input
              id="group-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter group name"
              required
              minLength={2}
              maxLength={100}
            />
          </div>

          <div className="form-group">
            <label htmlFor="group-description">Description</label>
            <textarea
              id="group-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter group description (optional)"
              rows={3}
            />
          </div>

          <div className="form-group">
            <label>Add Members ({selectedMembers.length} selected)</label>
            <div className="member-selector">
              {availableUsers.length === 0 ? (
                <p className="no-users-msg">No users available to add</p>
              ) : (
                availableUsers.map(user => (
                  <div 
                    key={user.id} 
                    className={`member-option ${selectedMembers.includes(user.id) ? 'selected' : ''}`}
                    onClick={() => handleToggleMember(user.id)}
                  >
                    <span className="member-checkbox">
                      {selectedMembers.includes(user.id) ? '☑' : '☐'}
                    </span>
                    <span className="member-name">{user.username || user.email}</span>
                  </div>
                ))
              )}
            </div>
          </div>

          <div className="modal-actions">
            <button type="button" className="cancel-btn" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="create-btn" disabled={!name.trim() || isCreating}>
              {isCreating ? 'Creating...' : 'Create Group'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateGroupModal;

