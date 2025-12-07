package service

import (
	"backend/db"
	"backend/models"

	"github.com/google/uuid"
)

// MessageService handles business logic for messages
type MessageService struct {
	messageRepo *db.MessageRepository
	userRepo    *db.UserRepository
}

// NewMessageService creates a new message service
func NewMessageService(messageRepo *db.MessageRepository, userRepo *db.UserRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

// ChatMessageResponse represents a message with user emails
type ChatMessageResponse struct {
	ID          uuid.UUID `json:"id"`
	SenderID    uuid.UUID `json:"sender_id"`
	SenderEmail string    `json:"sender_email"`
	ReceiverID  uuid.UUID `json:"receiver_id,omitempty"`
	GroupID     uuid.UUID `json:"group_id,omitempty"`
	Content     string    `json:"content"`
	Type        string    `json:"type"`
	CreatedAt   string    `json:"created_at"`
}

// SaveMessage saves a message to the database
func (s *MessageService) SaveMessage(senderID, receiverID uuid.UUID, content, msgType string) (*models.ChatMessage, error) {
	return s.messageRepo.SaveMessage(senderID, receiverID, content, msgType)
}

// GetChatHistory retrieves chat history between two users with email info
func (s *MessageService) GetChatHistory(userID1, userID2 uuid.UUID, limit, offset int) ([]ChatMessageResponse, error) {
	messages, err := s.messageRepo.GetChatHistory(userID1, userID2, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.enrichMessages(messages)
}

// GetUserMessages retrieves all messages for a user with email info
func (s *MessageService) GetUserMessages(userID uuid.UUID, limit, offset int) ([]ChatMessageResponse, error) {
	messages, err := s.messageRepo.GetUserMessages(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.enrichMessages(messages)
}

// SaveGroupMessage saves a group message to the database
func (s *MessageService) SaveGroupMessage(senderID, groupID uuid.UUID, content string) (*models.ChatMessage, error) {
	return s.messageRepo.SaveGroupMessage(senderID, groupID, content)
}

// GetGroupMessages retrieves messages for a group with email info
func (s *MessageService) GetGroupMessages(groupID uuid.UUID, limit, offset int) ([]ChatMessageResponse, error) {
	messages, err := s.messageRepo.GetGroupMessages(groupID, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.enrichMessages(messages)
}

// enrichMessages adds user emails to messages
func (s *MessageService) enrichMessages(messages []models.ChatMessage) ([]ChatMessageResponse, error) {
	// Cache user emails to avoid repeated DB queries
	emailCache := make(map[uuid.UUID]string)

	response := make([]ChatMessageResponse, len(messages))
	for i, msg := range messages {
		// Get sender email
		senderEmail, ok := emailCache[msg.SenderID]
		if !ok {
			user, err := s.userRepo.GetUserByID(msg.SenderID)
			if err == nil {
				senderEmail = user.Email
				emailCache[msg.SenderID] = senderEmail
			}
		}

		response[i] = ChatMessageResponse{
			ID:          msg.ID,
			SenderID:    msg.SenderID,
			SenderEmail: senderEmail,
			ReceiverID:  msg.ReceiverID,
			GroupID:     msg.GroupID,
			Content:     msg.Content,
			Type:        msg.Type,
			CreatedAt:   msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return response, nil
}

