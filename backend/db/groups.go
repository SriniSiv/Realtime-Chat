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

// GroupNameExistsExcluding checks if a group name exists for a group other than the given groupID
func (r *GroupRepository) GroupNameExistsExcluding(name string, groupID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Group{}).Where("LOWER(name) = LOWER(?) AND id != ? AND deleted_at IS NULL", name, groupID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateGroupNameAndDescription updates a group's name and description
func (r *GroupRepository) UpdateGroupNameAndDescription(groupID uuid.UUID, name, description string) error {
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&models.Group{}).Where("id = ?", groupID).Updates(updates).Error
}

// FilterGroups filters and searches groups with pagination
func (r *GroupRepository) FilterGroups(filter *models.GroupFilterRequest, userID uuid.UUID) ([]models.Group, int64, error) {
	var groups []models.Group
	var totalCount int64

	// Build base query based on available_only flag
	var baseQuery string
	var countQuery string
	args := []interface{}{}

	if filter.AvailableOnly {
		// Groups user is NOT a member of
		baseQuery = `
			SELECT g.* FROM chat_groups g
			WHERE g.id NOT IN (
				SELECT gm.group_id FROM group_members gm WHERE gm.user_id = ?
			)
			AND g.deleted_at IS NULL`
		countQuery = `
			SELECT COUNT(*) FROM chat_groups g
			WHERE g.id NOT IN (
				SELECT gm.group_id FROM group_members gm WHERE gm.user_id = ?
			)
			AND g.deleted_at IS NULL`
		args = append(args, userID)
	} else {
		// Groups user IS a member of
		baseQuery = `
			SELECT g.* FROM chat_groups g
			INNER JOIN group_members gm ON g.id = gm.group_id
			WHERE gm.user_id = ? AND g.deleted_at IS NULL`
		countQuery = `
			SELECT COUNT(*) FROM chat_groups g
			INNER JOIN group_members gm ON g.id = gm.group_id
			WHERE gm.user_id = ? AND g.deleted_at IS NULL`
		args = append(args, userID)
	}

	// Apply search filter
	if filter.SearchString != "" {
		searchPattern := "%" + filter.SearchString + "%"
		baseQuery += ` AND (g.name ILIKE ? OR g.description ILIKE ?)`
		countQuery += ` AND (g.name ILIKE ? OR g.description ILIKE ?)`
		args = append(args, searchPattern, searchPattern)
	}

	// Get total count
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := r.db.Raw(countQuery, countArgs...).Scan(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := "g.updated_at DESC" // default
	if filter.Sorting != nil && filter.Sorting.Field != "" {
		validFields := map[string]string{
			"name":       "g.name",
			"created_at": "g.created_at",
			"updated_at": "g.updated_at",
		}
		if dbField, ok := validFields[filter.Sorting.Field]; ok {
			order := "ASC"
			if filter.Sorting.Order == "desc" {
				order = "DESC"
			}
			orderBy = dbField + " " + order
		}
	}
	baseQuery += " ORDER BY " + orderBy

	// Apply pagination
	page := 1
	pageSize := 50
	if filter.PageInfo != nil {
		if filter.PageInfo.Page > 0 {
			page = filter.PageInfo.Page
		}
		if filter.PageInfo.PageSize > 0 {
			pageSize = filter.PageInfo.PageSize
		}
	}
	offset := (page - 1) * pageSize
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	// Execute query
	if err := r.db.Raw(baseQuery, args...).Scan(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, totalCount, nil
}

