package controller

import (
	"backend/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MessageController handles message-related HTTP requests
type MessageController struct {
	messageService *service.MessageService
}

// NewMessageController creates a new message controller
func NewMessageController(messageService *service.MessageService) *MessageController {
	return &MessageController{messageService: messageService}
}

// GetChatHistory returns chat history between current user and another user
func (c *MessageController) GetChatHistory(ctx *gin.Context) {
	// Get current user ID from context (set by auth middleware)
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	currentUserID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get other user ID from query param
	otherUserIDStr := ctx.Query("user_id")
	if otherUserIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter is required"})
		return
	}

	otherUserID, err := uuid.Parse(otherUserIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	// Get pagination params
	limit := 50
	offset := 0

	if l := ctx.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := ctx.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get chat history
	messages, err := c.messageService.GetChatHistory(currentUserID, otherUserID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chat history"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
		"limit":    limit,
		"offset":   offset,
	})
}

// GetAllMessages returns all messages for the current user
func (c *MessageController) GetAllMessages(ctx *gin.Context) {
	// Get current user ID from context
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	currentUserID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get pagination params
	limit := 50
	offset := 0

	if l := ctx.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := ctx.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get all messages
	messages, err := c.messageService.GetUserMessages(currentUserID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
		"limit":    limit,
		"offset":   offset,
	})
}

