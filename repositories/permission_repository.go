package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	CreatePermission(permission *models.Permission) error
	GetPermission(role string, resource string, action string) (*models.Permission, error)
	UpdatePermission(permission *models.Permission) error
	DeletePermission(id uint) error
	ListPermissions() ([]models.Permission, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) CreatePermission(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

func (r *permissionRepository) GetPermission(role string, resource string, action string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.Where("role = ? AND resource = ? AND action = ?", role, resource, action).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) UpdatePermission(permission *models.Permission) error {
	return r.db.Save(permission).Error
}

func (r *permissionRepository) DeletePermission(id uint) error {
	return r.db.Delete(&models.Permission{}, id).Error
}

func (r *permissionRepository) ListPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
