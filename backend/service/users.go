package service

import (
	"backend/db"
	"backend/models"
	"backend/utils"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// OnlineUserInfo represents a user that is currently online
type OnlineUserInfo struct {
	ID    uuid.UUID
	Email string
}

// OnlineUsersProvider interface for getting online users (implemented by websocket.Hub)
type OnlineUsersProvider interface {
	GetOnlineUserIDs() []uuid.UUID
}

// UserService handles business logic for user operations
type UserService struct {
	repo                *db.UserRepository
	onlineUsersProvider OnlineUsersProvider
}

// NewUserService creates a new user service
func NewUserService(repo *db.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// SetOnlineUsersProvider sets the provider for online users (e.g., websocket hub)
func (s *UserService) SetOnlineUsersProvider(provider OnlineUsersProvider) {
	s.onlineUsersProvider = provider
}

// GetOnlineUserIDs returns the IDs of all online users
func (s *UserService) GetOnlineUserIDs() []uuid.UUID {
	if s.onlineUsersProvider == nil {
		return []uuid.UUID{}
	}
	return s.onlineUsersProvider.GetOnlineUserIDs()
}

// buildOnlineMap creates a map of online user IDs for quick lookup
func (s *UserService) buildOnlineMap() map[uuid.UUID]bool {
	onlineMap := make(map[uuid.UUID]bool)
	onlineIDs := s.GetOnlineUserIDs()
	for _, id := range onlineIDs {
		onlineMap[id] = true
	}
	return onlineMap
}

// Register creates a new user account
func (s *UserService) Register(username, email, password string) (*models.AuthResponse, error) {
	// Check if username already exists
	usernameExists, err := s.repo.UsernameExists(username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, errors.New("username already taken")
	}

	// Check if email already exists
	emailExists, err := s.repo.EmailExists(email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user, err := s.repo.CreateUser(username, email, string(hashedPassword))
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Store refresh token
	expiresAt := time.Now().Add(utils.RefreshTokenExpiry)
	if err := s.repo.CreateRefreshToken(user.ID, refreshToken, expiresAt); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

// Login authenticates a user and returns tokens
func (s *UserService) Login(email, password string) (*models.AuthResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Store refresh token
	expiresAt := time.Now().Add(utils.RefreshTokenExpiry)
	if err := s.repo.CreateRefreshToken(user.ID, refreshToken, expiresAt); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

// RefreshAccessToken generates a new access token using refresh token
func (s *UserService) RefreshAccessToken(refreshTokenString string) (string, error) {
	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}

	// Check if refresh token exists in database
	_, err = s.repo.GetRefreshToken(refreshTokenString)
	if err != nil {
		return "", errors.New("refresh token not found or expired")
	}

	// Generate new access token
	accessToken, err := utils.GenerateAccessToken(claims.UserID, claims.Email)
	if err != nil {
		return "", errors.New("failed to generate access token")
	}

	return accessToken, nil
}

// GetUserProfile retrieves user profile by ID
func (s *UserService) GetUserProfile(userID uuid.UUID) (*models.UserDTO, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &models.UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// GetAllUsers retrieves all users from the database
func (s *UserService) GetAllUsers() ([]models.UserDTO, error) {
	users, err := s.repo.GetAllUsers()
	if err != nil {
		return nil, errors.New("failed to fetch users")
	}

	userDTOs := make([]models.UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = models.UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}
	}

	return userDTOs, nil
}

// GetUsersWithConversation retrieves users that the current user has had conversations with
func (s *UserService) GetUsersWithConversation(currentUserID uuid.UUID) ([]models.UserDTO, error) {
	users, err := s.repo.GetUsersWithConversation(currentUserID)
	if err != nil {
		return nil, errors.New("failed to fetch conversation users")
	}

	userDTOs := make([]models.UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = models.UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}
	}

	return userDTOs, nil
}

// SearchUsers searches users by username or email
func (s *UserService) SearchUsers(query string) ([]models.UserDTO, error) {
	users, err := s.repo.SearchUsers(query)
	if err != nil {
		return nil, errors.New("failed to search users")
	}

	userDTOs := make([]models.UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = models.UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}
	}

	return userDTOs, nil
}

// GetUsersWithConversationAndStatus retrieves users with conversation history including online status
func (s *UserService) GetUsersWithConversationAndStatus(currentUserID uuid.UUID) ([]models.UserWithStatus, error) {
	users, err := s.GetUsersWithConversation(currentUserID)
	if err != nil {
		return nil, err
	}

	onlineMap := s.buildOnlineMap()

	usersWithStatus := make([]models.UserWithStatus, len(users))
	for i, user := range users {
		usersWithStatus[i] = models.UserWithStatus{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsOnline: onlineMap[user.ID],
		}
	}

	return usersWithStatus, nil
}

// SearchUsersWithStatus searches users and returns with online status
func (s *UserService) SearchUsersWithStatus(query string) ([]models.UserWithStatus, error) {
	var users []models.UserDTO
	var err error

	if query == "" {
		users, err = s.GetAllUsers()
	} else {
		users, err = s.SearchUsers(query)
	}
	if err != nil {
		return nil, err
	}

	onlineMap := s.buildOnlineMap()

	usersWithStatus := make([]models.UserWithStatus, len(users))
	for i, user := range users {
		usersWithStatus[i] = models.UserWithStatus{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsOnline: onlineMap[user.ID],
		}
	}

	return usersWithStatus, nil
}
