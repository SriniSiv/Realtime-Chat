import React, { useState } from 'react';

const MessageInput = ({ onSend, selectedUser, disabled }) => {
  const [message, setMessage] = useState('');

  const handleSubmit = (e, type = 'direct') => {
    e.preventDefault();
    if (!message.trim()) return;
    
    if (type === 'direct' && !selectedUser) {
      alert('Please select a user to send a direct message');
      return;
    }

    onSend(message.trim(), type);
    setMessage('');
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e, 'direct');
    }
  };

  return (
    <div className="message-input">
      <form onSubmit={(e) => handleSubmit(e, 'direct')} className="input-form">
        <input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          onKeyPress={handleKeyPress}
          placeholder={
            disabled 
              ? 'Connecting...' 
              : selectedUser 
                ? `Message ${selectedUser.email}...` 
                : 'Select a user to chat...'
          }
          disabled={disabled}
        />
        <button 
          type="submit" 
          className="send-btn"
          disabled={disabled || !message.trim() || !selectedUser}
          title="Send direct message"
        >
          Send
        </button>
        <button 
          type="button"
          className="broadcast-btn"
          onClick={(e) => handleSubmit(e, 'broadcast')}
          disabled={disabled || !message.trim()}
          title="Broadcast to all users"
        >
          📢 All
        </button>
      </form>
    </div>
  );
};

export default MessageInput;

