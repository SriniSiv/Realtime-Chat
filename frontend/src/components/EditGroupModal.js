import React, { useState } from 'react';
import { groupAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';

const EditGroupModal = ({ group, onClose, onUpdate }) => {
  const [name, setName] = useState(group?.name || '');
  const [description, setDescription] = useState(group?.description || '');
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState('');
  const { getValidAccessToken } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name.trim() || name.length < 2) {
      setError('Group name must be at least 2 characters');
      return;
    }

    setIsUpdating(true);
    setError('');
    try {
      const token = await getValidAccessToken();
      const response = await groupAPI.updateGroup(token, group.id, name.trim(), description.trim());
      if (response.error) {
        setError(response.error);
      } else if (response.group) {
        onUpdate(response.group);
        onClose();
      }
    } catch (err) {
      setError('Failed to update group');
      console.error('Update group error:', err);
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Edit Group</h2>
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
              autoFocus
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

          {error && <div className="error-message">{error}</div>}

          <div className="modal-actions">
            <button type="button" className="cancel-btn" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="create-btn" disabled={!name.trim() || isUpdating}>
              {isUpdating ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default EditGroupModal;

