import React, { useState, useEffect, useCallback } from 'react';
import { groupAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';

const GroupList = ({ groups, selectedGroup, onSelectGroup, onCreateGroup, onRefresh, onAddMembers, onJoinGroup }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState(null);
  const [availableGroups, setAvailableGroups] = useState([]);
  const [isSearching, setIsSearching] = useState(false);
  const [showAvailable, setShowAvailable] = useState(false);
  const [joiningGroupId, setJoiningGroupId] = useState(null);
  const { getValidAccessToken } = useAuth();

  // Search groups
  const searchGroups = useCallback(async (query) => {
    if (!query.trim()) {
      setSearchResults(null);
      setAvailableGroups([]);
      return;
    }

    setIsSearching(true);
    try {
      const token = await getValidAccessToken();

      // Search user's groups and available groups in parallel
      const [userGroupsRes, availableRes] = await Promise.all([
        groupAPI.filterGroups(token, { search_text: query, available_only: false }),
        groupAPI.filterGroups(token, { search_text: query, available_only: true }),
      ]);

      if (userGroupsRes.groups) {
        setSearchResults(userGroupsRes.groups);
      }
      if (availableRes.groups) {
        setAvailableGroups(availableRes.groups);
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
      searchGroups(searchQuery);
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [searchQuery, searchGroups]);

  // Handle join group
  const handleJoinGroup = async (group) => {
    if (joiningGroupId || !onJoinGroup) return;
    setJoiningGroupId(group.id);
    try {
      await onJoinGroup(group);
      // Remove from available groups after joining
      setAvailableGroups(prev => prev.filter(g => g.id !== group.id));
      // Clear search to show updated group list
      setSearchQuery('');
      setSearchResults(null);
      setShowAvailable(false);
    } catch (error) {
      console.error('Failed to join group:', error);
    } finally {
      setJoiningGroupId(null);
    }
  };

  // Use search results if available, otherwise use all groups
  const displayGroups = searchResults !== null ? searchResults : groups;

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

      <div className="group-search">
        <input
          type="text"
          placeholder="🔍 Search groups..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="search-input"
        />
        {isSearching && <span className="search-loading">...</span>}
      </div>

      {/* Toggle to show available groups */}
      {searchQuery && availableGroups.length > 0 && (
        <div className="group-toggle">
          <button
            className={`toggle-btn ${!showAvailable ? 'active' : ''}`}
            onClick={() => setShowAvailable(false)}
          >
            My Groups ({displayGroups.length})
          </button>
          <button
            className={`toggle-btn ${showAvailable ? 'active' : ''}`}
            onClick={() => setShowAvailable(true)}
          >
            Available ({availableGroups.length})
          </button>
        </div>
      )}

      <div className="group-list-content">
        {showAvailable && searchQuery ? (
          // Show available groups to join
          availableGroups.length === 0 ? (
            <div className="no-groups">
              <p>No available groups found</p>
              <small>Try a different search</small>
            </div>
          ) : (
            availableGroups.map(group => (
              <div
                key={group.id}
                className="group-item available"
              >
                <div className="group-avatar">
                  #
                </div>
                <div className="group-info">
                  <span className="group-name">{group.name}</span>
                  <span className="group-members">
                    {group.member_count} member{group.member_count !== 1 ? 's' : ''} • by {group.creator_name}
                  </span>
                </div>
                <button
                  className="join-group-btn"
                  onClick={() => handleJoinGroup(group)}
                  disabled={joiningGroupId === group.id}
                  title="Join Group"
                >
                  {joiningGroupId === group.id ? '...' : 'Join'}
                </button>
              </div>
            ))
          )
        ) : (
          // Show user's groups
          displayGroups.length === 0 ? (
            <div className="no-groups">
              <p>{searchQuery ? 'No groups found' : 'No groups yet'}</p>
              <small>{searchQuery ? 'Try a different search' : 'Click ➕ to create a group'}</small>
            </div>
          ) : (
            displayGroups.map(group => (
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
                {selectedGroup?.id === group.id && onAddMembers && (
                  <button
                    className="add-member-btn"
                    onClick={(e) => {
                      e.stopPropagation();
                      onAddMembers(group);
                    }}
                    title="Add Members"
                  >
                    👤+
                  </button>
                )}
              </div>
            ))
          )
        )}
      </div>
    </div>
  );
};

export default GroupList;

