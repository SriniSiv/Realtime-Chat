import React, { useEffect, useRef } from 'react';

const MessageList = ({ messages, currentUserId }) => {
  const messagesEndRef = useRef(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const formatTime = (timestamp) => {
    return new Date(timestamp).toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  return (
    <div className="message-list">
      {messages.length === 0 ? (
        <div className="no-messages">
          <div className="empty-icon">💬</div>
          <p>No messages yet</p>
          <small>Start a conversation!</small>
        </div>
      ) : (
        messages.map((msg, index) => {
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

