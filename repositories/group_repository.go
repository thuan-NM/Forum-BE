package repositories

import (
	"Forum_BE/models"
	"errors"
	"gorm.io/gorm"
)

type GroupRepository interface {
	CreateGroup(group *models.Group) error
	GetGroupByID(id uint) (*models.Group, error)
	UpdateGroup(group *models.Group) error
	DeleteGroup(id uint) error
	ListGroups() ([]models.Group, error)
	GetGroupByName(name string) (*models.Group, error) // <-- Add this method
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) CreateGroup(group *models.Group) error {
	return r.db.Create(group).Error
}

func (r *groupRepository) GetGroupByID(id uint) (*models.Group, error) {
	var group models.Group
	err := r.db.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) UpdateGroup(group *models.Group) error {
	return r.db.Save(group).Error
}

func (r *groupRepository) DeleteGroup(id uint) error {
	return r.db.Delete(&models.Group{}, id).Error
}

func (r *groupRepository) ListGroups() ([]models.Group, error) {
	var groups []models.Group
	err := r.db.Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *groupRepository) GetGroupByName(name string) (*models.Group, error) {
	var group models.Group
	err := r.db.Where("name = ?", name).First(&group).Error
	if err != nil {
		// If the group doesn't exist, return nil and the error (which will be gorm.ErrRecordNotFound)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}
