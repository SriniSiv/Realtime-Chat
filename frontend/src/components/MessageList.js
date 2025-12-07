import React, { useEffect, useRef, useMemo } from 'react';

const MessageList = ({ messages, currentUserId, selectedUserId }) => {
  const messagesEndRef = useRef(null);

  // Filter messages for the selected conversation
  const filteredMessages = useMemo(() => {
    if (!selectedUserId) return [];

    return messages.filter(msg => {
      // Show broadcast messages
      if (msg.type === 'broadcast') return true;

      // Show messages between current user and selected user
      const isFromSelected = msg.from === selectedUserId;
      const isToSelected = msg.to === selectedUserId;
      const isFromMe = msg.from === currentUserId;
      const isToMe = msg.to === currentUserId;

      return (isFromSelected && isToMe) || (isFromMe && isToSelected);
    });
  }, [messages, currentUserId, selectedUserId]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [filteredMessages]);

  const formatTime = (timestamp) => {
    return new Date(timestamp).toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  return (
    <div className="message-list">
      {filteredMessages.length === 0 ? (
        <div className="no-messages">
          <div className="empty-icon">💬</div>
          <p>No messages yet</p>
          <small>Start a conversation!</small>
        </div>
      ) : (
        filteredMessages.map((msg, index) => {
          const isSent = msg.isSent || msg.from === currentUserId;
          const isSystem = msg.type === 'system';

          if (isSystem) {
            return (
              <div key={index} className="message system">
                <span className="system-icon">ℹ️</span>
                <span>{msg.content}</span>
              </div>
            );
          }

          return (
            <div 
              key={index} 
              className={`message ${isSent ? 'sent' : 'received'}`}
            >
              <div className="message-bubble">
                {!isSent && (
                  <div className="message-sender">{msg.from_email}</div>
                )}
                <div className="message-content">{msg.content}</div>
                <div className="message-time">
                  {msg.type === 'broadcast' && <span className="broadcast-tag">📢 Broadcast</span>}
                  {formatTime(msg.timestamp)}
                </div>
              </div>
            </div>
          );
        })
      )}
      <div ref={messagesEndRef} />
    </div>
  );
};

export default MessageList;

