package controller

import (
	"backend/models"
	"backend/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GroupController handles group-related HTTP requests
type GroupController struct {
	groupService   *service.GroupService
	messageService *service.MessageService
}

// NewGroupController creates a new group controller
func NewGroupController(groupService *service.GroupService) *GroupController {
	return &GroupController{groupService: groupService}
}

// SetMessageService sets the message service for group message operations
func (c *GroupController) SetMessageService(messageService *service.MessageService) {
	c.messageService = messageService
}

// CreateGroup handles POST /groups
func (c *GroupController) CreateGroup(ctx *gin.Context) {
	// Get current user ID from context
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := currentUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	var req models.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse member IDs
	memberIDs := make([]uuid.UUID, 0)
	for _, idStr := range req.MemberIDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			memberIDs = append(memberIDs, id)
		}
	}

	group, err := c.groupService.CreateGroup(userID, req.Name, req.Description, memberIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"group": group})
}

// GetUserGroups handles GET /groups
func (c *GroupController) GetUserGroups(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := currentUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	groups, err := c.groupService.GetUserGroups(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"groups": groups, "count": len(groups)})
}

// GetGroupMembers handles GET /groups/:id/members
func (c *GroupController) GetGroupMembers(ctx *gin.Context) {
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	members, err := c.groupService.GetGroupMembers(groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"members": members, "count": len(members)})
}

// AddMembers handles POST /groups/:id/members
func (c *GroupController) AddMembers(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := currentUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	var req models.AddMembersRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	memberIDs := make([]uuid.UUID, 0)
	for _, idStr := range req.MemberIDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			memberIDs = append(memberIDs, id)
		}
	}

	if err := c.groupService.AddMembers(groupID, userID, memberIDs); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Fetch and return the updated group
	group, err := c.groupService.GetGroupDTO(groupID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "members added successfully"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "members added successfully", "group": group})
}

// GetGroup handles GET /groups/:id
func (c *GroupController) GetGroup(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := currentUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	// Check if user is a member of the group
	isMember, err := c.groupService.IsMember(groupID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isMember {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this group"})
		return
	}

	group, err := c.groupService.GetGroupDTO(groupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"group": group})
}

// SearchAvailableGroups handles GET /groups/search/available
func (c *GroupController) SearchAvailableGroups(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := currentUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "search query 'q' is required"})
		return
	}

	groups, err := c.groupService.SearchAvailableGroups(userID, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"groups": groups, "count": len(groups)})
}

// SearchUserGroups handles GET /groups/search
func (c *GroupController) SearchUserGroups(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := currentUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "search query 'q' is required"})
		return
	}

	groups, err := c.groupService.SearchUserGroups(userID, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"groups": groups, "count": len(groups)})
}

// GetGroupMessages handles GET /groups/:id/messages
func (c *GroupController) GetGroupMessages(ctx *gin.Context) {
	if c.messageService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "message service not configured"})
		return
	}

	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := currentUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	// Check if user is a member of the group
	isMember, err := c.groupService.IsMember(groupID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isMember {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this group"})
		return
	}

	// Parse pagination params
	limit := 50
	offset := 0
	if l := ctx.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := ctx.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	messages, err := c.messageService.GetGroupMessages(groupID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"messages": messages, "count": len(messages)})
}
