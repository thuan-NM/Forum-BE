package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type PermissionService interface {
	CreatePermission(role, resource, action string, allowed bool) (*models.Permission, error)
	GetPermission(role, resource, action string) (*models.Permission, error)
	UpdatePermission(role, resource, action string, allowed bool) (*models.Permission, error)
	DeletePermission(id uint) error
	ListPermissions() ([]models.Permission, error)
	GetUserRole(userID uint) (models.Role, error) // Thêm phương thức này để middleware có thể lấy role
}

type permissionService struct {
	permissionRepo repositories.PermissionRepository
	userRepo       repositories.UserRepository // Thêm repository user để lấy role
}

func NewPermissionService(pRepo repositories.PermissionRepository, uRepo repositories.UserRepository) PermissionService {
	return &permissionService{permissionRepo: pRepo, userRepo: uRepo}
}

func (s *permissionService) CreatePermission(role, resource, action string, allowed bool) (*models.Permission, error) {
	if role == "" || resource == "" || action == "" {
		return nil, errors.New("role, resource and action are required")
	}

	// Kiểm tra xem permission đã tồn tại chưa
	existingPermission, err := s.permissionRepo.GetPermission(role, resource, action)
	if err == nil && existingPermission != nil {
		return nil, errors.New("permission already exists")
	}

	permission := &models.Permission{
		Role:     models.Role(role),
		Resource: resource,
		Action:   action,
		Allowed:  allowed,
	}

	if err := s.permissionRepo.CreatePermission(permission); err != nil {
		return nil, err
	}

	return permission, nil
}

func (s *permissionService) GetPermission(role, resource, action string) (*models.Permission, error) {
	return s.permissionRepo.GetPermission(role, resource, action)
}

func (s *permissionService) UpdatePermission(role, resource, action string, allowed bool) (*models.Permission, error) {
	permission, err := s.permissionRepo.GetPermission(role, resource, action)
	if err != nil {
		return nil, err
	}

	permission.Allowed = allowed

	if err := s.permissionRepo.UpdatePermission(permission); err != nil {
		return nil, err
	}

	return permission, nil
}

func (s *permissionService) DeletePermission(id uint) error {
	return s.permissionRepo.DeletePermission(id)
}

func (s *permissionService) ListPermissions() ([]models.Permission, error) {
	return s.permissionRepo.ListPermissions()
}

func (s *permissionService) GetUserRole(userID uint) (models.Role, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return "", err
	}
	return user.Role, nil
}
