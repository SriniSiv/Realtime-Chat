package models

import "github.com/google/uuid"

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents the refresh token request payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         UserDTO   `json:"user"`
}

// UserDTO represents the user data transfer object
type UserDTO struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// UserWithStatus represents a user with online/offline status (like Slack)
type UserWithStatus struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	IsOnline bool      `json:"is_online"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// MessageResponse represents a generic message response
type MessageResponse struct {
	Message string `json:"message"`
}

// UpdateUsernameRequest represents the update username request payload
type UpdateUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
}

// CreateGroupRequest represents the create group request payload
type CreateGroupRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=100"`
	Description string   `json:"description"`
	MemberIDs   []string `json:"member_ids"` // Optional initial members
}

// UpdateGroupRequest represents the update group request payload
type UpdateGroupRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description"`
}

// AddMembersRequest represents the add members request payload
type AddMembersRequest struct {
	MemberIDs []string `json:"member_ids" binding:"required,min=1"`
}

// GroupDTO represents the group data transfer object
type GroupDTO struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedBy   uuid.UUID         `json:"created_by"`
	CreatorName string            `json:"creator_name"`
	MemberCount int               `json:"member_count"`
	Members     []GroupMemberDTO  `json:"members,omitempty"`
	CreatedAt   string            `json:"created_at"`
}

// GroupMemberDTO represents a group member
type GroupMemberDTO struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	IsOnline bool      `json:"is_online"`
}

