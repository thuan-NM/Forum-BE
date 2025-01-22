package services

import (
	"Forum_BE/models"
	"Forum_BE/utils"
	"errors"
)

type AuthService interface {
	Register(username, email, password string) (*models.User, error)
	Login(email, password string) (string, *models.User, error) // Trả về JWT token và user
	ResetToken(userID uint) (string, error)                     // Reset token mới
}

type authService struct {
	userService UserService
	jwtSecret   string
}

func NewAuthService(u UserService, secret string) AuthService {
	return &authService{userService: u, jwtSecret: secret}
}

func (s *authService) Register(username, email, password string) (*models.User, error) {
	return s.userService.CreateUser(username, email, password)
}

func (s *authService) Login(email, password string) (string, *models.User, error) {
	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		return "", nil, errors.New("Invalid email or password")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", nil, errors.New("Invalid email or password")
	}

	token, err := utils.GenerateJWT(user.ID, s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}
func (s *authService) ResetToken(userID uint) (string, error) {
	token, err := utils.GenerateJWT(userID, s.jwtSecret)
	if err != nil {
		return "", err
	}
	return token, nil
}
