package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"errors"
)

type UserService interface {
	CreateUser(username, email, password string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(id uint, username, email, password, role string) (*models.User, error)
	DeleteUser(id uint) error
	ListUsers() ([]models.User, error)
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(uRepo repositories.UserRepository) UserService {
	return &userService{uRepo}
}

func (s *userService) CreateUser(username, email, password string) (*models.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email, and password are required")
	}

	// Kiểm tra xem username hoặc email đã tồn tại chưa
	existingUser, err := s.userRepo.GetUserByUsername(username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	existingUser, err = s.userRepo.GetUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Mã hóa mật khẩu
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     models.RoleUser, // Mặc định là User
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetUserByID(id)
}

func (s *userService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepo.GetUserByUsername(username)
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.GetUserByEmail(email)
}

func (s *userService) UpdateUser(id uint, username, email, password, role string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	if username != "" {
		// Kiểm tra xem username đã tồn tại chưa
		existingUser, err := s.userRepo.GetUserByUsername(username)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, errors.New("username already exists")
		}
		user.Username = username
	}

	if email != "" {
		// Kiểm tra xem email đã tồn tại chưa
		existingUser, err := s.userRepo.GetUserByEmail(email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, errors.New("email already exists")
		}
		user.Email = email
	}

	if password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}
		user.Password = hashedPassword
	}

	if role != "" {
		// Kiểm tra role hợp lệ
		switch role {
		case string(models.RoleRoot), string(models.RoleAdmin), string(models.RoleEmployee), string(models.RoleUser):
			user.Role = models.Role(role)
		default:
			return nil, errors.New("invalid role")
		}
	}

	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(id uint) error {
	return s.userRepo.DeleteUser(id)
}

func (s *userService) ListUsers() ([]models.User, error) {
	return s.userRepo.ListUsers()
}
