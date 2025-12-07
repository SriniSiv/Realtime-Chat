package db

import (
	"backend/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(username, email, passwordHash string) (*models.User, error) {
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// CreateRefreshToken creates a new refresh token in the database
func (r *UserRepository) CreateRefreshToken(userID uuid.UUID, token string, expiresAt time.Time) error {
	refreshToken := &models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	return r.db.Create(refreshToken).Error
}

// GetRefreshToken retrieves a refresh token from the database
func (r *UserRepository) GetRefreshToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	if err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found or expired")
		}
		return nil, err
	}
	return &refreshToken, nil
}

// DeleteRefreshToken deletes a refresh token from the database
func (r *UserRepository) DeleteRefreshToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.RefreshToken{}).Error
}

// DeleteUserRefreshTokens deletes all refresh tokens for a user
func (r *UserRepository) DeleteUserRefreshTokens(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UsernameExists checks if a username already exists
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAllUsers retrieves all users from the database
func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetUsersWithConversation retrieves users that the current user has had conversations with
func (r *UserRepository) GetUsersWithConversation(currentUserID uuid.UUID) ([]models.User, error) {
	var users []models.User

	// Find all users that the current user has exchanged messages with (excluding broadcasts)
	err := r.db.Raw(`
		SELECT DISTINCT u.* FROM user_details u
		WHERE u.id IN (
			SELECT DISTINCT sender_id FROM chat_messages
			WHERE receiver_id = ? AND type = 'direct'
			UNION
			SELECT DISTINCT receiver_id FROM chat_messages
			WHERE sender_id = ? AND type = 'direct'
		) AND u.id != ? AND u.deleted_at IS NULL
	`, currentUserID, currentUserID, currentUserID).Scan(&users).Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

// SearchUsers searches users by username or email
func (r *UserRepository) SearchUsers(query string) ([]models.User, error) {
	var users []models.User
	searchPattern := "%" + query + "%"
	if err := r.db.Where("username ILIKE ? OR email ILIKE ?", searchPattern, searchPattern).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FilterUsers searches and filters users with pagination
func (r *UserRepository) FilterUsers(filter *models.UserFilterRequest, excludeUserID uuid.UUID) ([]models.User, int64, error) {
	var users []models.User
	var totalCount int64

	query := r.db.Model(&models.User{}).Where("id != ?", excludeUserID)

	// Apply search filter on username or email
	if filter.SearchString != "" {
		searchPattern := "%" + filter.SearchString + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}

	// Get total count before pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if filter.Sorting != nil && filter.Sorting.Field != "" {
		order := "ASC"
		if filter.Sorting.Order == "desc" {
			order = "DESC"
		}
		// Whitelist allowed sort fields
		allowedFields := map[string]bool{"username": true, "email": true, "created_at": true}
		if allowedFields[filter.Sorting.Field] {
			query = query.Order(filter.Sorting.Field + " " + order)
		}
	} else {
		query = query.Order("username ASC")
	}

	// Apply pagination
	if filter.PageInfo != nil {
		page := filter.PageInfo.Page
		if page < 1 {
			page = 1
		}
		pageSize := filter.PageInfo.PageSize
		if pageSize < 1 {
			pageSize = 50
		}
		if pageSize > 100 {
			pageSize = 100
		}
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}

// UpdateUsername updates a user's username
func (r *UserRepository) UpdateUsername(userID uuid.UUID, username string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("username", username).Error
}

// UsernameExistsExcludingUser checks if a username exists for a user other than the given userID
func (r *UserRepository) UsernameExistsExcludingUser(username string, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("username = ? AND id != ?", username, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
