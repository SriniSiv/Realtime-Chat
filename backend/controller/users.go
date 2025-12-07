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

// FilterUsers handles POST /chat/users
// @Summary Filter and search users
// @Description Filter, search and paginate users with online status
// @Router /chat/users [post]
func (ctrl *UserController) FilterUsers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "user not authenticated",
		})
		return
	}

	var filter models.UserFilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {
		// If no body provided, use empty filter
		filter = models.UserFilterRequest{}
	}

	result, err := ctrl.service.FilterUsers(&filter, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to filter users",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateUsername handles PUT /auth/user
// @Summary Update username
// @Description Update the current user's username
// @Router /auth/user [put]
func (ctrl *UserController) UpdateUsername(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "user not authenticated",
		})
		return
	}

	var req models.UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	user, err := ctrl.service.UpdateUsername(userID.(uuid.UUID), req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "username updated successfully",
		"user":    user,
	})
}
