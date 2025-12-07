package controller

import (
	"backend/models"
	"backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserController handles HTTP requests for user operations
type UserController struct {
	service *service.UserService
}

// NewUserController creates a new user controller
func NewUserController(service *service.UserService) *UserController {
	return &UserController{service: service}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Router /auth/register [post]
func (ctrl *UserController) Register(c *gin.Context) {
	var req models.RegisterRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Call service layer
	response, err := ctrl.service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Router /auth/login [post]
func (ctrl *UserController) Login(c *gin.Context) {
	var req models.LoginRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Call service layer
	response, err := ctrl.service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Generate a new access token using refresh token
// @Router /auth/refresh [post]
func (ctrl *UserController) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Call service layer
	accessToken, err := ctrl.service.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

// GetUserDetails handles getting current user profile
// @Summary Get current user profile
// @Description Get logged-in user's profile information
// @Router /auth/me [get]
func (ctrl *UserController) GetUserDetails(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// Call service layer
	user, err := ctrl.service.GetUserProfile(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetOnlineUsers handles GET /chat/online-users
// @Summary Get online users
// @Description Get list of currently online users
// @Router /chat/online-users [get]
func (ctrl *UserController) GetOnlineUsers(c *gin.Context) {
	onlineIDs := ctrl.service.GetOnlineUserIDs()
	c.JSON(http.StatusOK, gin.H{
		"online_users": onlineIDs,
		"count":        len(onlineIDs),
	})
}

// GetUsersWithConversation handles GET /chat/users
// @Summary Get users with conversation history
// @Description Get list of users current user has had conversations with (like Slack DM sidebar)
// @Router /chat/users [get]
func (ctrl *UserController) GetUsersWithConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "user not authenticated",
		})
		return
	}

	users, err := ctrl.service.GetUsersWithConversationAndStatus(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to fetch users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// SearchUsers handles GET /chat/users/search
// @Summary Search users
// @Description Search users by username or email with online status
// @Router /chat/users/search [get]
func (ctrl *UserController) SearchUsers(c *gin.Context) {
	query := c.Query("q")

	users, err := ctrl.service.SearchUsersWithStatus(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to search users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}
