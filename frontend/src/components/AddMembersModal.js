import React, { useState, useEffect } from 'react';
import { groupAPI, chatAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';

const AddMembersModal = ({ group, onClose, onMembersAdded }) => {
  const [allUsers, setAllUsers] = useState([]);
  const [currentMembers, setCurrentMembers] = useState([]);
  const [selectedMembers, setSelectedMembers] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isAdding, setIsAdding] = useState(false);
  const [error, setError] = useState('');
  const { getValidAccessToken } = useAuth();

  // Fetch all users and current members
  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      try {
        const token = await getValidAccessToken();
        
        // Fetch all users and current group members in parallel
        const [usersRes, membersRes] = await Promise.all([
          chatAPI.searchUsers(token, ''), // Get all users
          groupAPI.getGroupMembers(token, group.id),
        ]);

        if (usersRes.users) {
          setAllUsers(usersRes.users);
        }
        if (membersRes.members) {
          setCurrentMembers(membersRes.members.map(m => m.user_id));
        }
      } catch (err) {
        console.error('Failed to fetch data:', err);
        setError('Failed to load users');
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [getValidAccessToken, group.id]);

  // Filter users: exclude current members and apply search
  const availableUsers = allUsers.filter(user => {
    const isNotMember = !currentMembers.includes(user.id);
    const matchesSearch = !searchQuery || 
      (user.username?.toLowerCase().includes(searchQuery.toLowerCase()) ||
       user.email?.toLowerCase().includes(searchQuery.toLowerCase()));
    return isNotMember && matchesSearch;
  });

  const handleToggleMember = (userId) => {
    setSelectedMembers(prev => 
      prev.includes(userId) 
        ? prev.filter(id => id !== userId)
        : [...prev, userId]
    );
  };

  const handleAddMembers = async () => {
    if (selectedMembers.length === 0) return;

    setIsAdding(true);
    setError('');
    try {
      const token = await getValidAccessToken();
      const response = await groupAPI.addMembers(token, group.id, selectedMembers);
      
      if (response.error) {
        setError(response.error);
      } else {
        onMembersAdded(response.group);
        onClose();
      }
    } catch (err) {
      console.error('Failed to add members:', err);
      setError('Failed to add members');
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Add Members to {group.name}</h2>
          <button className="close-btn" onClick={onClose}>✕</button>
        </div>

        <div className="modal-body">
          {error && <div className="error-msg">{error}</div>}

          <div className="form-group">
            <input
              type="text"
              placeholder="🔍 Search users..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="search-input"
            />
          </div>

          <div className="form-group">
            <label>Select Users ({selectedMembers.length} selected)</label>
            <div className="member-selector">
              {isLoading ? (
                <p className="loading-msg">Loading users...</p>
              ) : availableUsers.length === 0 ? (
                <p className="no-users-msg">
                  {searchQuery ? 'No users found' : 'All users are already members'}
                </p>
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
                    <span className={`member-status ${user.is_online ? 'online' : 'offline'}`}>
                      ● {user.is_online ? 'Online' : 'Offline'}
                    </span>
                  </div>
                ))
              )}
            </div>
          </div>

          <div className="modal-actions">
            <button type="button" className="cancel-btn" onClick={onClose}>
              Cancel
            </button>
            <button 
              type="button" 
              className="create-btn" 
              disabled={selectedMembers.length === 0 || isAdding}
              onClick={handleAddMembers}
            >
              {isAdding ? 'Adding...' : `Add ${selectedMembers.length} Member${selectedMembers.length !== 1 ? 's' : ''}`}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AddMembersModal;

