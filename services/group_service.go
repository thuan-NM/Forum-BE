package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type GroupService interface {
	CreateGroup(name, description string) (*models.Group, error)
	GetGroupByID(id uint) (*models.Group, error)
	UpdateGroup(id uint, name, description string) (*models.Group, error)
	DeleteGroup(id uint) error
	ListGroups() ([]models.Group, error)
}

type groupService struct {
	groupRepo repositories.GroupRepository
}

func NewGroupService(gRepo repositories.GroupRepository) GroupService {
	return &groupService{groupRepo: gRepo}
}

func (s *groupService) CreateGroup(name, description string) (*models.Group, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	// Kiểm tra xem nhóm đã tồn tại chưa
	existingGroup, err := s.groupRepo.GetGroupByName(name)
	if err == nil && existingGroup != nil {
		return nil, errors.New("group already exists")
	}

	group := &models.Group{
		Name:        name,
		Description: description,
	}

	if err := s.groupRepo.CreateGroup(group); err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) GetGroupByID(id uint) (*models.Group, error) {
	return s.groupRepo.GetGroupByID(id)
}

func (s *groupService) UpdateGroup(id uint, name, description string) (*models.Group, error) {
	group, err := s.groupRepo.GetGroupByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		existingGroup, err := s.groupRepo.GetGroupByName(name)
		if err == nil && existingGroup != nil && existingGroup.ID != id {
			return nil, errors.New("group name already exists")
		}
		group.Name = name
	}

	if description != "" {
		group.Description = description
	}

	if err := s.groupRepo.UpdateGroup(group); err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) DeleteGroup(id uint) error {
	return s.groupRepo.DeleteGroup(id)
}

func (s *groupService) ListGroups() ([]models.Group, error) {
	return s.groupRepo.ListGroups()
}
