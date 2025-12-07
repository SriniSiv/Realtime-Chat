import React, { useState } from 'react';
import { authAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';

const EditProfileModal = ({ onClose, onUpdate, currentUsername }) => {
  const [username, setUsername] = useState(currentUsername || '');
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState('');
  const { getValidAccessToken } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!username.trim() || username.length < 3) {
      setError('Username must be at least 3 characters');
      return;
    }

    setIsUpdating(true);
    setError('');
    try {
      const token = await getValidAccessToken();
      const response = await authAPI.updateUsername(token, username.trim());
      if (response.error) {
        setError(response.error);
      } else if (response.user) {
        onUpdate(response.user);
        onClose();
      }
    } catch (err) {
      setError('Failed to update username');
      console.error('Update username error:', err);
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Edit Profile</h2>
          <button className="close-btn" onClick={onClose}>✕</button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="username">Username</label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter username"
              required
              minLength={3}
              maxLength={50}
              autoFocus
            />
            <small className="field-hint">3-50 characters</small>
          </div>

          {error && <div className="error-message">{error}</div>}

          <div className="modal-actions">
            <button type="button" className="cancel-btn" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="create-btn" disabled={!username.trim() || isUpdating}>
              {isUpdating ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default EditProfileModal;

