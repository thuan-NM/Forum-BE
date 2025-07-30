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
	UpdateUser(id uint, updateDTO UpdateUserDTO) (*models.User, error)
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

type UpdateUserDTO struct {
	Username      *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Email         *string `json:"email,omitempty" binding:"omitempty,email"`
	Password      *string `json:"password,omitempty" binding:"omitempty,min=6"`
	Role          *string `json:"role,omitempty" binding:"omitempty,oneof=root admin employee user"`
	Status        *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive banned"`
	FullName      *string `json:"full_name,omitempty" binding:"omitempty,min=1,max=100"`
	Avatar        *string `json:"avatar,omitempty" binding:"omitempty"`
	Bio           *string `json:"bio,omitempty" binding:"omitempty"`
	Location      *string `json:"location,omitempty" binding:"omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty" binding:"omitempty"`
}

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

func (s *userService) UpdateUser(id uint, updateDTO UpdateUserDTO) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		slog.Error("Không thể lấy thông tin người dùng", "id", id, "error", err)
		return nil, err
	}

	// Kiểm tra và cập nhật username
	if updateDTO.Username != nil {
		existingUser, err := s.userRepo.GetUserByUsername(*updateDTO.Username)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("tên người dùng đã tồn tại")
		}
		if !errors.Is(err, repositories.ErrNotFound) && err != nil {
			slog.Error("Không thể kiểm tra tên người dùng", "username", *updateDTO.Username, "error", err)
			return nil, err
		}
		user.Username = *updateDTO.Username
	}

	// Kiểm tra và cập nhật email
	if updateDTO.Email != nil {
		if !emailRegex.MatchString(*updateDTO.Email) {
			return nil, ErrInvalidEmail
		}
		existingUser, err := s.userRepo.GetUserByEmail(*updateDTO.Email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("email đã tồn tại")
		}
		if !errors.Is(err, repositories.ErrNotFound) && err != nil {
			slog.Error("Không thể kiểm tra email", "email", *updateDTO.Email, "error", err)
			return nil, err
		}
		user.Email = *updateDTO.Email
	}

	// Kiểm tra và cập nhật password
	if updateDTO.Password != nil {
		if len(*updateDTO.Password) < 6 {
			return nil, ErrInvalidPassword
		}
		hashedPassword, err := utils.HashPassword(*updateDTO.Password)
		if err != nil {
			slog.Error("Không thể mã hóa mật khẩu", "id", id, "error", err)
			return nil, err
		}
		user.Password = hashedPassword
	}

	// Kiểm tra và cập nhật role
	if updateDTO.Role != nil {
		switch models.Role(*updateDTO.Role) {
		case models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser:
			user.Role = models.Role(*updateDTO.Role)
		default:
			return nil, ErrInvalidRole
		}
	}

	// Kiểm tra và cập nhật status
	if updateDTO.Status != nil {
		switch models.Status(*updateDTO.Status) {
		case models.StatusActive, models.StatusInactive, models.StatusBanned:
			user.Status = models.Status(*updateDTO.Status)
		default:
			return nil, ErrInvalidStatus
		}
	}

	// Cập nhật các trường khác nếu được cung cấp
	if updateDTO.FullName != nil {
		user.FullName = *updateDTO.FullName
	}
	if updateDTO.Avatar != nil {
		user.Avatar = updateDTO.Avatar
	}
	if updateDTO.Bio != nil {
		user.Bio = updateDTO.Bio
	}
	if updateDTO.Location != nil {
		user.Location = updateDTO.Location
	}
	if updateDTO.EmailVerified != nil {
		user.EmailVerified = *updateDTO.EmailVerified
	}

	// Lưu thông tin người dùng vào cơ sở dữ liệu
	if err := s.userRepo.UpdateUser(user); err != nil {
		slog.Error("Không thể cập nhật người dùng", "id", id, "error", err)
		return nil, err
	}

	// Cập nhật cache
	if s.redisClient != nil {
		userJSON, err := json.Marshal(user)
		if err != nil {
			slog.Error("Không thể mã hóa người dùng để lưu cache", "id", user.ID, "error", err)
		} else {
			cacheKey := fmt.Sprintf("user:%d", id)
			if err := s.redisClient.Set(context.Background(), cacheKey, userJSON, 24*time.Hour).Err(); err != nil {
				slog.Error("Không thể lưu người dùng vào cache", "cache_key", cacheKey, "error", err)
			}
			cacheStatusKey := fmt.Sprintf("user:status:%d", id)
			if err := s.redisClient.Set(context.Background(), cacheStatusKey, string(user.Status), 1*time.Hour).Err(); err != nil {
				slog.Error("Không thể lưu trạng thái người dùng vào cache", "cache_key", cacheStatusKey, "error", err)
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
