package service

import (
	"backend/db"
	"backend/models"
	"errors"
	"time"

	"github.com/google/uuid"
)

// GroupService handles business logic for groups
type GroupService struct {
	groupRepo *db.GroupRepository
	userRepo  *db.UserRepository
}

// NewGroupService creates a new group service
func NewGroupService(groupRepo *db.GroupRepository, userRepo *db.UserRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// CreateGroup creates a new group and adds the creator as admin
func (s *GroupService) CreateGroup(creatorID uuid.UUID, name, description string, memberIDs []uuid.UUID) (*models.GroupDTO, error) {
	// Check for duplicate group name
	existingGroup, _ := s.groupRepo.GetGroupByName(name)
	if existingGroup != nil {
		return nil, errors.New("a group with this name already exists")
	}

	// Create the group
	group := &models.Group{
		Name:        name,
		Description: description,
		CreatedBy:   creatorID,
	}

	if err := s.groupRepo.CreateGroup(group); err != nil {
		return nil, errors.New("failed to create group")
	}

	// Add creator as admin
	creatorMember := &models.GroupMember{
		GroupID: group.ID,
		UserID:  creatorID,
		Role:    "admin",
	}
	if err := s.groupRepo.AddMember(creatorMember); err != nil {
		return nil, errors.New("failed to add creator to group")
	}

	// Add other members
	for _, memberID := range memberIDs {
		if memberID == creatorID {
			continue // Skip creator, already added
		}
		member := &models.GroupMember{
			GroupID: group.ID,
			UserID:  memberID,
			Role:    "member",
		}
		s.groupRepo.AddMember(member) // Ignore errors for individual members
	}

	return s.GetGroupDTO(group.ID)
}

// GetGroupDTO returns a GroupDTO for a group
func (s *GroupService) GetGroupDTO(groupID uuid.UUID) (*models.GroupDTO, error) {
	group, err := s.groupRepo.GetGroupByID(groupID)
	if err != nil {
		return nil, errors.New("group not found")
	}

	memberCount, _ := s.groupRepo.GetMemberCount(groupID)
	creator, _ := s.userRepo.GetUserByID(group.CreatedBy)
	creatorName := ""
	if creator != nil {
		creatorName = creator.Username
	}

	return &models.GroupDTO{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedBy:   group.CreatedBy,
		CreatorName: creatorName,
		MemberCount: int(memberCount),
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetUserGroups returns all groups a user is a member of
func (s *GroupService) GetUserGroups(userID uuid.UUID) ([]models.GroupDTO, error) {
	groups, err := s.groupRepo.GetUserGroups(userID)
	if err != nil {
		return nil, errors.New("failed to fetch groups")
	}

	groupDTOs := make([]models.GroupDTO, len(groups))
	for i, group := range groups {
		memberCount, _ := s.groupRepo.GetMemberCount(group.ID)
		creator, _ := s.userRepo.GetUserByID(group.CreatedBy)
		creatorName := ""
		if creator != nil {
			creatorName = creator.Username
		}

		groupDTOs[i] = models.GroupDTO{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			CreatedBy:   group.CreatedBy,
			CreatorName: creatorName,
			MemberCount: int(memberCount),
			CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		}
	}

	return groupDTOs, nil
}

// AddMembers adds members to a group (public groups - anyone can add members)
func (s *GroupService) AddMembers(groupID uuid.UUID, requestingUserID uuid.UUID, memberIDs []uuid.UUID) error {
	// For public groups, anyone can add members (including self-join)
	// Just verify the requesting user exists (optional check)
	_ = requestingUserID // Acknowledge parameter (no admin check for public groups)

	for _, memberID := range memberIDs {
		// Check if already a member
		isMember, _ := s.groupRepo.IsMember(groupID, memberID)
		if isMember {
			continue
		}

		member := &models.GroupMember{
			GroupID: groupID,
			UserID:  memberID,
			Role:    "member",
		}
		s.groupRepo.AddMember(member)
	}

	return nil
}

// GetGroupMembers returns all members of a group with their details
func (s *GroupService) GetGroupMembers(groupID uuid.UUID) ([]models.GroupMemberDTO, error) {
	members, err := s.groupRepo.GetGroupMembers(groupID)
	if err != nil {
		return nil, errors.New("failed to fetch members")
	}

	memberDTOs := make([]models.GroupMemberDTO, len(members))
	for i, member := range members {
		memberDTOs[i] = models.GroupMemberDTO{
			ID:       member.ID,
			UserID:   member.UserID,
			Username: member.User.Username,
			Email:    member.User.Email,
			Role:     member.Role,
		}
	}

	return memberDTOs, nil
}

// RemoveMember removes a member from a group
func (s *GroupService) RemoveMember(groupID, requestingUserID, memberID uuid.UUID) error {
	// Check if requesting user is admin or is removing themselves
	role, err := s.groupRepo.GetMemberRole(groupID, requestingUserID)
	if err != nil {
		return errors.New("you are not a member of this group")
	}

	if requestingUserID != memberID && role != "admin" {
		return errors.New("only admins can remove other members")
	}

	return s.groupRepo.RemoveMember(groupID, memberID)
}

// IsMember checks if a user is a member of a group
func (s *GroupService) IsMember(groupID, userID uuid.UUID) (bool, error) {
	return s.groupRepo.IsMember(groupID, userID)
}

// GetGroupMemberIDs returns all member IDs for a group
func (s *GroupService) GetGroupMemberIDs(groupID uuid.UUID) ([]uuid.UUID, error) {
	return s.groupRepo.GetGroupMemberIDs(groupID)
}

// SearchAvailableGroups searches for groups the user is NOT a member of
func (s *GroupService) SearchAvailableGroups(userID uuid.UUID, query string) ([]models.GroupDTO, error) {
	groups, err := s.groupRepo.SearchAvailableGroups(userID, query)
	if err != nil {
		return nil, errors.New("failed to search groups")
	}

	groupDTOs := make([]models.GroupDTO, len(groups))
	for i, group := range groups {
		memberCount, _ := s.groupRepo.GetMemberCount(group.ID)
		creator, _ := s.userRepo.GetUserByID(group.CreatedBy)
		creatorName := ""
		if creator != nil {
			creatorName = creator.Username
		}

		groupDTOs[i] = models.GroupDTO{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			CreatedBy:   group.CreatedBy,
			CreatorName: creatorName,
			MemberCount: int(memberCount),
			CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		}
	}

	return groupDTOs, nil
}

// SearchUserGroups searches for groups the user IS a member of
func (s *GroupService) SearchUserGroups(userID uuid.UUID, query string) ([]models.GroupDTO, error) {
	groups, err := s.groupRepo.SearchUserGroups(userID, query)
	if err != nil {
		return nil, errors.New("failed to search groups")
	}

	groupDTOs := make([]models.GroupDTO, len(groups))
	for i, group := range groups {
		memberCount, _ := s.groupRepo.GetMemberCount(group.ID)
		creator, _ := s.userRepo.GetUserByID(group.CreatedBy)
		creatorName := ""
		if creator != nil {
			creatorName = creator.Username
		}

		groupDTOs[i] = models.GroupDTO{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			CreatedBy:   group.CreatedBy,
			CreatorName: creatorName,
			MemberCount: int(memberCount),
			CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		}
	}

	return groupDTOs, nil
}

