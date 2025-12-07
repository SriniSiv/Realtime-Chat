package db

import (
	"backend/models"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageRepository handles database operations for messages
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// SaveMessage saves a chat message to the database
func (r *MessageRepository) SaveMessage(senderID, receiverID uuid.UUID, content, msgType string) (*models.ChatMessage, error) {
	log.Printf("Saving message: sender=%s, receiver=%s, type=%s, content=%s", senderID, receiverID, msgType, content)

	message := &models.ChatMessage{
		SenderID:   senderID,
		ReceiverID: &receiverID, // Use pointer for nullable foreign key
		Content:    content,
		Type:       msgType,
	}

	if err := r.db.Create(message).Error; err != nil {
		log.Printf("Error saving message to DB: %v", err)
		return nil, err
	}

	log.Printf("Message saved successfully with ID: %s", message.ID)
	return message, nil
}

// GetChatHistory retrieves chat history between two users
func (r *MessageRepository) GetChatHistory(userID1, userID2 uuid.UUID, limit, offset int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	err := r.db.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID1, userID2, userID2, userID1,
	).Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return messages, nil
}

// GetBroadcastMessages retrieves broadcast messages
func (r *MessageRepository) GetBroadcastMessages(limit, offset int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	err := r.db.Where("type = ?", "broadcast").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return messages, nil
}

// GetUserMessages retrieves all messages for a user (sent or received)
func (r *MessageRepository) GetUserMessages(userID uuid.UUID, limit, offset int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	err := r.db.Where(
		"sender_id = ? OR receiver_id = ? OR type = ?",
		userID, userID, "broadcast",
	).Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return messages, nil
}

// SaveGroupMessage saves a group chat message to the database
func (r *MessageRepository) SaveGroupMessage(senderID, groupID uuid.UUID, content string) (*models.ChatMessage, error) {
	log.Printf("Saving group message: sender=%s, group=%s, content=%s", senderID, groupID, content)

	message := &models.ChatMessage{
		SenderID: senderID,
		GroupID:  &groupID, // Use pointer for nullable foreign key
		Content:  content,
		Type:     "group",
	}

	if err := r.db.Create(message).Error; err != nil {
		log.Printf("Error saving group message to DB: %v", err)
		return nil, err
	}

	log.Printf("Group message saved successfully with ID: %s", message.ID)
	return message, nil
}

// GetGroupMessages retrieves messages for a group
func (r *MessageRepository) GetGroupMessages(groupID uuid.UUID, limit, offset int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	err := r.db.Where("group_id = ? AND type = ?", groupID, "group").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return messages, nil
}
