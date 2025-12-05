package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the users table
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"type:varchar(255);unique;not null" json:"email"`
	PasswordHash string         `gorm:"type:text;not null" json:"-"`
	CreatedAt    time.Time      `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// RefreshToken represents the refresh_tokens table
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Token     string    `gorm:"type:text;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	User      User      `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "user_details"
}

// TableName specifies the table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// ChatMessage represents the chat_messages table
type ChatMessage struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SenderID   uuid.UUID `gorm:"type:uuid;not null;index" json:"sender_id"`
	ReceiverID uuid.UUID `gorm:"type:uuid;index" json:"receiver_id,omitempty"` // nil for broadcast
	Content    string    `gorm:"type:text;not null" json:"content"`
	Type       string    `gorm:"type:varchar(20);not null;default:'direct'" json:"type"` // direct, broadcast
	CreatedAt  time.Time `gorm:"default:now();index" json:"created_at"`

	// Relations
	Sender   User `gorm:"foreignKey:SenderID;references:ID" json:"-"`
	Receiver User `gorm:"foreignKey:ReceiverID;references:ID" json:"-"`
}

// TableName specifies the table name for ChatMessage
func (ChatMessage) TableName() string {
	return "chat_messages"
}
