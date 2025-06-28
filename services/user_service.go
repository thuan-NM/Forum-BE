package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log/slog"
	"regexp"
	"strconv"
	"time"
)

var (
	ErrInvalidUsername = errors.New("invalid username format")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be at least 6 characters")
	ErrInvalidRole     = errors.New("invalid role")
	ErrInvalidStatus   = errors.New("invalid status")
)

type UserService interface {
	CreateUser(username, email, password, fullname string, isVerify bool) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(id uint, username, email, password, role, status string, emailVerified bool) (*models.User, error)
	DeleteUser(id uint) error
	GetAllUsers(filters map[string]interface{}) ([]models.User, int64, error)
	ModifyUserStatus(id uint, status string) (*models.User, error)
}

type userService struct {
	userRepo    repositories.UserRepository
	redisClient *redis.Client
}

func NewUserService(uRepo repositories.UserRepository, redisClient *redis.Client) UserService {
	return &userService{userRepo: uRepo, redisClient: redisClient}
}

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func (s *userService) CreateUser(username, email, password, fullname string, isVerify bool) (*models.User, error) {

	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	if len(password) < 6 {
		return nil, ErrInvalidPassword
	}

	existingUser, err := s.userRepo.GetUserByUsername(username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("Username already exists")
	}
	if !errors.Is(err, repositories.ErrNotFound) && err != nil {
		slog.Error("Failed to check username", "username", username, "error", err)
		return nil, err
	}

	existingUser, err = s.userRepo.GetUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("Email already exists")
	}
	if !errors.Is(err, repositories.ErrNotFound) && err != nil {
		slog.Error("Failed to check email", "email", email, "error", err)
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		slog.Error("Failed to hash password", "email", email, "error", err)
		return nil, err
	}

	user := &models.User{
		Username:      username,
		Email:         email,
		Password:      hashedPassword,
		FullName:      fullname,
		Role:          models.RoleUser,
		Status:        models.StatusInactive,
		EmailVerified: isVerify,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		slog.Error("Failed to create user", "email", email, "error", err)
		return nil, err
	}

	if s.redisClient != nil {
		userJSON, err := json.Marshal(user)
		if err != nil {
			slog.Error("Failed to marshal user for caching", "id", user.ID, "error", err)
		} else {
			cacheKey := fmt.Sprintf("user:%d", user.ID)
			if err := s.redisClient.Set(context.Background(), cacheKey, userJSON, 24*time.Hour).Err(); err != nil {
				slog.Error("Failed to cache user", "cache_key", cacheKey, "error", err)
			}
		}
	}

	if s.redisClient != nil {
		s.invalidateUsersCache(context.Background())
	}

	return user, nil
}

func (s *userService) GetUserByID(id uint) (*models.User, error) {
	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("user:%d", id)
		cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
		if err == nil {
			var user models.User
			if err := json.Unmarshal([]byte(cached), &user); err == nil {
				slog.Info("Cache hit for user", "id", id)
				return &user, nil
			}
		} else if err != redis.Nil {
			slog.Error("Redis error", "id", id, "error", err)
		}
	}

	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		slog.Error("Failed to get user", "id", id, "error", err)
		return nil, err
	}

	// Cache the result
	if s.redisClient != nil && user != nil {
		userJSON, err := json.Marshal(user)
		if err != nil {
			slog.Error("Failed to marshal user for caching", "id", user.ID, "error", err)
		} else {
			cacheKey := fmt.Sprintf("user:%d", id)
			if err := s.redisClient.Set(context.Background(), cacheKey, userJSON, 24*time.Hour).Err(); err != nil {
				slog.Error("Failed to cache user", "cache_key", cacheKey, "error", err)
			}
		}
	}

	return user, nil
}

func (s *userService) GetUserByUsername(username string) (*models.User, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		slog.Error("Failed to get user by username", "username", username, "error", err)
	}
	return user, err
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		slog.Error("Failed to get user by email", "email", email, "error", err)
	}
	return user, err
}

func (s *userService) UpdateUser(id uint, username, email, password, role, status string, emailVerified bool) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		slog.Error("Failed to get user", "id", id, "error", err)
		return nil, err
	}

	if username != "" {
		existingUser, err := s.userRepo.GetUserByUsername(username)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("username already exists")
		}
		if !errors.Is(err, repositories.ErrNotFound) && err != nil {
			slog.Error("Failed to check username", "username", username, "error", err)
			return nil, err
		}
		user.Username = username
	}

	if email != "" {
		if !emailRegex.MatchString(email) {
			return nil, ErrInvalidEmail
		}
		existingUser, err := s.userRepo.GetUserByEmail(email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("email already exists")
		}
		if !errors.Is(err, repositories.ErrNotFound) && err != nil {
			slog.Error("Failed to check email", "email", email, "error", err)
			return nil, err
		}
		user.Email = email
	}

	if password != "" {
		if len(password) < 6 {
			return nil, ErrInvalidPassword
		}
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			slog.Error("Failed to hash password", "id", id, "error", err)
			return nil, err
		}
		user.Password = hashedPassword
	}

	if role != "" {
		switch models.Role(role) {
		case models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser:
			user.Role = models.Role(role)
		default:
			return nil, ErrInvalidRole
		}
	}

	if status != "" {
		switch models.Status(status) {
		case models.StatusActive, models.StatusInactive, models.StatusBanned:
			user.Status = models.Status(status)
		default:
			return nil, ErrInvalidStatus
		}
	}

	user.EmailVerified = emailVerified

	if err := s.userRepo.UpdateUser(user); err != nil {
		slog.Error("Failed to update user", "id", id, "error", err)
		return nil, err
	}

	if s.redisClient != nil {
		userJSON, err := json.Marshal(user)
		if err != nil {
			slog.Error("Failed to marshal user for caching", "id", user.ID, "error", err)
		} else {
			cacheKey := fmt.Sprintf("user:%d", id)
			if err := s.redisClient.Set(context.Background(), cacheKey, userJSON, 24*time.Hour).Err(); err != nil {
				slog.Error("Failed to cache user", "cache_key", cacheKey, "error", err)
			}
			cacheStatusKey := fmt.Sprintf("user:status:%d", id)
			if err := s.redisClient.Set(context.Background(), cacheStatusKey, string(user.Status), 1*time.Hour).Err(); err != nil {
				slog.Error("Failed to cache user status", "cache_key", cacheStatusKey, "error", err)
			}
		}

		s.invalidateUsersCache(context.Background())
	}

	return user, nil
}

func (s *userService) DeleteUser(id uint) error {
	err := s.userRepo.DeleteUser(id)
	if err != nil {
		slog.Error("Failed to delete user", "id", id, "error", err)
		return err
	}
	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("user:%d", id)
		cacheStatusKey := fmt.Sprintf("user:status:%d", id)
		if err := s.redisClient.Del(context.Background(), cacheKey, cacheStatusKey).Err(); err != nil {
			slog.Error("Failed to delete from Redis", "id", id, "error", err)
		}
		s.invalidateUsersCache(context.Background())
	}
	return err
}

func (s *userService) GetAllUsers(filters map[string]interface{}) ([]models.User, int64, error) {
	cacheKey := utils.GenerateCacheKey("users:all", 0, filters)
	ctx := context.Background()

	// Check Redis cache
	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var users []models.User
			if err := json.Unmarshal([]byte(cached), &users); err == nil {
				var total int64
				totalKey := cacheKey + ":total"
				totalStr, err := s.redisClient.Get(ctx, totalKey).Result()
				if err == nil {
					total, _ = strconv.ParseInt(totalStr, 10, 64)
				} else if err != redis.Nil {
					slog.Error("Redis error when getting total", "cache_key", totalKey, "error", err)
				}
				slog.Info("Cache hit for users", "cache_key", cacheKey)
				return users, total, nil
			}
		} else if err != redis.Nil {
			slog.Error("Redis error", "cache_key", cacheKey, "error", err)
		}
	}

	// Query database
	users, total, err := s.userRepo.GetAllUsers(filters)
	if err != nil {
		slog.Error("Failed to get all users", "error", err)
		return nil, 0, err
	}

	// Cache results in Redis
	if s.redisClient != nil && len(users) > 0 {
		usersJSON, err := json.Marshal(users)
		if err != nil {
			slog.Error("Failed to marshal users for caching", "error", err)
		} else {
			if err := s.redisClient.Set(ctx, cacheKey, usersJSON, 5*time.Minute).Err(); err != nil {
				slog.Error("Failed to cache users", "cache_key", cacheKey, "error", err)
			}
			if err := s.redisClient.Set(ctx, cacheKey+":total", total, 5*time.Minute).Err(); err != nil {
				slog.Error("Failed to cache total", "cache_key", cacheKey+":total", "error", err)
			}
		}
	}

	return users, total, nil
}

func (s *userService) ModifyUserStatus(id uint, status string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		slog.Error("Failed to get user", "id", id, "error", err)
		return nil, err
	}

	switch status {
	case "ban":
		user.Status = models.StatusBanned
	case "active":
		user.Status = models.StatusActive
	case "unban", "inactive":
		user.Status = models.StatusInactive
	default:
		return nil, ErrInvalidStatus
	}

	if err := s.userRepo.UpdateUser(user); err != nil {
		slog.Error("Failed to update user status", "id", id, "error", err)
		return nil, err
	}

	if s.redisClient != nil {
		userJSON, err := json.Marshal(user)
		if err != nil {
			slog.Error("Failed to marshal user for caching", "id", user.ID, "error", err)
		} else {
			cacheKey := fmt.Sprintf("user:%d", id)
			if err := s.redisClient.Set(context.Background(), cacheKey, userJSON, 24*time.Hour).Err(); err != nil {
				slog.Error("Failed to cache user", "cache_key", cacheKey, "error", err)
			}
			cacheStatusKey := fmt.Sprintf("user:status:%d", id)
			if err := s.redisClient.Set(context.Background(), cacheStatusKey, string(user.Status), 1*time.Hour).Err(); err != nil {
				slog.Error("Failed to cache user status", "cache_key", cacheStatusKey, "error", err)
			}
		}

		s.invalidateUsersCache(context.Background())
	}

	return user, nil
}

func (s *userService) invalidateUsersCache(ctx context.Context) {
	if s.redisClient == nil {
		return
	}

	keys, err := s.redisClient.Keys(ctx, "users:all:*").Result()
	if err != nil {
		slog.Error("Failed to find users cache keys", "error", err)
		return
	}
	var allKeys []string
	for _, key := range keys {
		allKeys = append(allKeys, key)
		totalKey := key + ":total"
		allKeys = append(allKeys, totalKey)
	}

	if len(allKeys) > 0 {
		if err := s.redisClient.Del(ctx, allKeys...).Err(); err != nil {
			slog.Error("Failed to delete users cache", "error", err)
		} else {
			slog.Info("Invalidated users cache", "keys", allKeys)
		}
	}
}
