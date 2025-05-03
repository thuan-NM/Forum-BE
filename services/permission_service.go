package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"fmt"
	"log"
)

type PermissionService interface {
	CreatePermission(role, resource, action string, allowed bool) (*models.Permission, error)
	GetPermission(role, resource, action string) (*models.Permission, error)
	UpdatePermission(role, resource, action string, allowed bool) (*models.Permission, error)
	DeletePermission(id uint) error
	ListPermissions() ([]models.Permission, error)
	GetUserRole(userID uint) (models.Role, error)
}

type permissionService struct {
	permissionRepo repositories.PermissionRepository
	userRepo       repositories.UserRepository
}

func NewPermissionService(pRepo repositories.PermissionRepository, uRepo repositories.UserRepository) PermissionService {
	return &permissionService{permissionRepo: pRepo, userRepo: uRepo}
}

func (s *permissionService) CreatePermission(role, resource, action string, allowed bool) (*models.Permission, error) {
	if role == "" || resource == "" || action == "" {
		return nil, fmt.Errorf("role, resource, and action are required")
	}

	existingPermission, err := s.permissionRepo.GetPermission(role, resource, action)
	if err == nil && existingPermission != nil {
		return nil, fmt.Errorf("permission already exists")
	}

	permission := &models.Permission{
		Role:     models.Role(role),
		Resource: resource,
		Action:   action,
		Allowed:  allowed,
	}

	if err := s.permissionRepo.CreatePermission(permission); err != nil {
		log.Printf("Failed to create permission for %s:%s:%s: %v", role, resource, action, err)
		return nil, err
	}

	return permission, nil
}

func (s *permissionService) GetPermission(role, resource, action string) (*models.Permission, error) {
	permission, err := s.permissionRepo.GetPermission(role, resource, action)
	if err != nil {
		log.Printf("Failed to get permission for %s:%s:%s: %v", role, resource, action, err)
	}
	return permission, err
}

func (s *permissionService) UpdatePermission(role, resource, action string, allowed bool) (*models.Permission, error) {
	permission, err := s.permissionRepo.GetPermission(role, resource, action)
	if err != nil {
		log.Printf("Failed to find permission for %s:%s:%s: %v", role, resource, action, err)
		return nil, err
	}

	permission.Allowed = allowed

	if err := s.permissionRepo.UpdatePermission(permission); err != nil {
		log.Printf("Failed to update permission for %s:%s:%s: %v", role, resource, action, err)
		return nil, err
	}

	return permission, nil
}

func (s *permissionService) DeletePermission(id uint) error {
	err := s.permissionRepo.DeletePermission(id)
	if err != nil {
		log.Printf("Failed to delete permission %d: %v", id, err)
	}
	return err
}

func (s *permissionService) ListPermissions() ([]models.Permission, error) {
	permissions, err := s.permissionRepo.ListPermissions()
	if err != nil {
		log.Printf("Failed to list permissions: %v", err)
	}
	return permissions, err
}

func (s *permissionService) GetUserRole(userID uint) (models.Role, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to get user %d: %v", userID, err)
		return "", err
	}
	return user.Role, nil
}
