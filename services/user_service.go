package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"fmt"
	"log"
)

type UserService interface {
	CreateUser(username, email, password string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(id uint, username, email, password, role string) (*models.User, error)
	DeleteUser(id uint) error
	ListUsers() ([]models.User, error)
	ListUsersPaginated(limit, offset int) ([]models.User, int64, error)
	ModifyUserStatus(id uint, status string) (*models.User, error)
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(uRepo repositories.UserRepository) UserService {
	return &userService{userRepo: uRepo}
}

func (s *userService) CreateUser(username, email, password string) (*models.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("username, email, and password are required")
	}

	existingUser, err := s.userRepo.GetUserByUsername(username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	existingUser, err = s.userRepo.GetUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("Failed to hash password for %s: %v", email, err)
		return nil, err
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     "user",
		IsActive: true,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		log.Printf("Failed to create user %s: %v", email, err)
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		log.Printf("Failed to get user %d: %v", id, err)
	}
	return user, err
}

func (s *userService) GetUserByUsername(username string) (*models.User, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		log.Printf("Failed to get user by username %s: %v", username, err)
	}
	return user, err
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		log.Printf("Failed to get user by email %s: %v", email, err)
	}
	return user, err
}

func (s *userService) UpdateUser(id uint, username, email, password, role string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	if username != "" {
		existingUser, err := s.userRepo.GetUserByUsername(username)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("username already exists")
		}
		user.Username = username
	}

	if email != "" {
		existingUser, err := s.userRepo.GetUserByEmail(email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = email
	}

	if password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			log.Printf("Failed to hash password for user %d: %v", id, err)
			return nil, err
		}
		user.Password = hashedPassword
	}

	if role != "" {
		switch role {
		case string(models.RoleRoot), string(models.RoleAdmin), string(models.RoleEmployee), string(models.RoleUser):
			user.Role = models.Role(role)
		default:
			return nil, fmt.Errorf("invalid role")
		}
	}

	if err := s.userRepo.UpdateUser(user); err != nil {
		log.Printf("Failed to update user %d: %v", id, err)
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(id uint) error {
	err := s.userRepo.DeleteUser(id)
	if err != nil {
		log.Printf("Failed to delete user %d: %v", id, err)
	}
	return err
}

func (s *userService) ListUsers() ([]models.User, error) {
	users, err := s.userRepo.ListUsers()
	if err != nil {
		log.Printf("Failed to list users: %v", err)
	}
	return users, err
}

func (s *userService) ListUsersPaginated(limit, offset int) ([]models.User, int64, error) {
	users, total, err := s.userRepo.ListUsersPaginated(limit, offset)
	if err != nil {
		log.Printf("Failed to list users: %v", err)
	}
	return users, total, err
}

func (s *userService) ModifyUserStatus(id uint, status string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		log.Printf("Failed to get user %d: %v", id, err)
		return nil, err
	}

	switch status {
	case "ban":
		user.IsBanned = true
	case "unban":
		user.IsBanned = false
	default:
		return nil, fmt.Errorf("invalid status")
	}

	if err := s.userRepo.UpdateUser(user); err != nil {
		log.Printf("Failed to update user status %d: %v", id, err)
		return nil, err
	}

	return user, nil
}
