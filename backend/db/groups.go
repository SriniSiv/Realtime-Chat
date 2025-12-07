package db

import (
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GroupRepository handles database operations for groups
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// CreateGroup creates a new group
func (r *GroupRepository) CreateGroup(group *models.Group) error {
	return r.db.Create(group).Error
}

// GetGroupByID retrieves a group by ID
func (r *GroupRepository) GetGroupByID(groupID uuid.UUID) (*models.Group, error) {
	var group models.Group
	if err := r.db.Preload("Creator").First(&group, "id = ?", groupID).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// GetUserGroups retrieves all groups a user is a member of
func (r *GroupRepository) GetUserGroups(userID uuid.UUID) ([]models.Group, error) {
	var groups []models.Group
	err := r.db.Raw(`
		SELECT g.* FROM chat_groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ? AND g.deleted_at IS NULL
		ORDER BY g.updated_at DESC
	`, userID).Scan(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// AddMember adds a user to a group
func (r *GroupRepository) AddMember(member *models.GroupMember) error {
	return r.db.Create(member).Error
}

// RemoveMember removes a user from a group
func (r *GroupRepository) RemoveMember(groupID, userID uuid.UUID) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{}).Error
}

// GetGroupMembers retrieves all members of a group
func (r *GroupRepository) GetGroupMembers(groupID uuid.UUID) ([]models.GroupMember, error) {
	var members []models.GroupMember
	err := r.db.Preload("User").Where("group_id = ?", groupID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetMemberCount returns the number of members in a group
func (r *GroupRepository) GetMemberCount(groupID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.GroupMember{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}

// IsMember checks if a user is a member of a group
func (r *GroupRepository) IsMember(groupID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count).Error
	return count > 0, err
}

// GetMemberRole gets a user's role in a group
func (r *GroupRepository) GetMemberRole(groupID, userID uuid.UUID) (string, error) {
	var member models.GroupMember
	err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

// UpdateGroup updates a group's details
func (r *GroupRepository) UpdateGroup(group *models.Group) error {
	return r.db.Save(group).Error
}

// DeleteGroup soft deletes a group
func (r *GroupRepository) DeleteGroup(groupID uuid.UUID) error {
	return r.db.Delete(&models.Group{}, "id = ?", groupID).Error
}

// GetGroupMemberIDs returns all member user IDs for a group
func (r *GroupRepository) GetGroupMemberIDs(groupID uuid.UUID) ([]uuid.UUID, error) {
	var memberIDs []uuid.UUID
	err := r.db.Model(&models.GroupMember{}).
		Where("group_id = ?", groupID).
		Pluck("user_id", &memberIDs).Error
	return memberIDs, err
}

// GetGroupByName checks if a group with the given name exists
func (r *GroupRepository) GetGroupByName(name string) (*models.Group, error) {
	var group models.Group
	if err := r.db.Where("LOWER(name) = LOWER(?) AND deleted_at IS NULL", name).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// SearchAvailableGroups searches for groups the user is NOT a member of
func (r *GroupRepository) SearchAvailableGroups(userID uuid.UUID, query string) ([]models.Group, error) {
	var groups []models.Group
	searchPattern := "%" + query + "%"
	err := r.db.Raw(`
		SELECT g.* FROM chat_groups g
		WHERE g.id NOT IN (
			SELECT gm.group_id FROM group_members gm WHERE gm.user_id = ?
		)
		AND g.deleted_at IS NULL
		AND (LOWER(g.name) LIKE LOWER(?) OR LOWER(g.description) LIKE LOWER(?))
		ORDER BY g.name ASC
		LIMIT 20
	`, userID, searchPattern, searchPattern).Scan(&groups).Error
	return groups, err
}

// SearchUserGroups searches for groups the user IS a member of
func (r *GroupRepository) SearchUserGroups(userID uuid.UUID, query string) ([]models.Group, error) {
	var groups []models.Group
	searchPattern := "%" + query + "%"
	err := r.db.Raw(`
		SELECT g.* FROM chat_groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ? AND g.deleted_at IS NULL
		AND (LOWER(g.name) LIKE LOWER(?) OR LOWER(g.description) LIKE LOWER(?))
		ORDER BY g.updated_at DESC
	`, userID, searchPattern, searchPattern).Scan(&groups).Error
	return groups, err
}

