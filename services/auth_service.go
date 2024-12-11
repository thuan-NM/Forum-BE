package services

import (
	"Forum_BE/models"
	"Forum_BE/utils"
	"errors"
)

type AuthService interface {
	Register(username, email, password string) (*models.User, error)
	Login(username, password string) (string, error) // Trả về JWT token
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

func (s *authService) Login(username, password string) (string, error) {
	user, err := s.userService.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid username or password")
	}

	token, err := utils.GenerateJWT(user.ID, s.jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}
