import React, { useState } from 'react';

const MessageInput = ({ onSend, selectedUser, selectedGroup, disabled }) => {
  const [message, setMessage] = useState('');

  const handleSubmit = (e, type = 'direct') => {
    e.preventDefault();
    if (!message.trim()) return;

    if (type === 'direct' && !selectedUser && !selectedGroup) {
      alert('Please select a user or group to send a message');
      return;
    }

    onSend(message.trim(), selectedGroup ? 'group' : type);
    setMessage('');
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e, 'direct');
    }
  };

  const getPlaceholder = () => {
    if (disabled) return 'Connecting...';
    if (selectedGroup) return `Message #${selectedGroup.name}...`;
    if (selectedUser) return `Message ${selectedUser.username || selectedUser.email}...`;
    return 'Select a conversation...';
  };

  const canSend = selectedUser || selectedGroup;

  return (
    <div className="message-input">
      <form onSubmit={(e) => handleSubmit(e, 'direct')} className="input-form">
        <input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          onKeyPress={handleKeyPress}
          placeholder={getPlaceholder()}
          disabled={disabled}
        />
        <button
          type="submit"
          className="send-btn"
          disabled={disabled || !message.trim() || !canSend}
          title={selectedGroup ? "Send to group" : "Send direct message"}
        >
          Send
        </button>
        {!selectedGroup && (
          <button
            type="button"
            className="broadcast-btn"
            onClick={(e) => handleSubmit(e, 'broadcast')}
            disabled={disabled || !message.trim()}
            title="Broadcast to all users"
          >
            📢 All
          </button>
        )}
      </form>
    </div>
  );
};

export default MessageInput;

