package repositories

import (
	"Forum_BE/models"
	"errors"
	"gorm.io/gorm"
	"log/slog"
)

var (
	ErrNotFound = errors.New("user not found")
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uint) error
	GetAllUsers(filters map[string]interface{}) ([]models.User, int64, error)
	GetUserByIDWithPassword(id uint) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Đếm số lượng posts, answers, questions
	type Counts struct {
		PostCount     int64
		AnswerCount   int64
		QuestionCount int64
	}
	var counts Counts
	r.db.Model(&models.Post{}).Where("user_id = ?", id).Count(&counts.PostCount)
	r.db.Model(&models.Answer{}).Where("user_id = ?", id).Count(&counts.AnswerCount)
	r.db.Model(&models.Question{}).Where("user_id = ?", id).Count(&counts.QuestionCount)

	// Gán vào struct User
	user.PostCount = counts.PostCount
	user.AnswerCount = counts.AnswerCount
	user.QuestionCount = counts.QuestionCount

	slog.Info("User loaded", "id", id, "postCount", user.PostCount, "answerCount", user.AnswerCount, "questionCount", user.QuestionCount)
	return &user, nil
}

func (r *userRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) DeleteUser(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) GetAllUsers(filters map[string]interface{}) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{}).Where("deleted_at IS NULL")

	// Apply search filter
	if search, ok := filters["search"].(string); ok && search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username LIKE ? OR email LIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if limit, ok := filters["limit"].(int); ok && limit > 0 {
		query = query.Limit(limit)
	}
	if offset, ok := filters["offset"].(int); ok && offset >= 0 {
		query = query.Offset(offset)
	}

	// Execute query
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) GetUserByIDWithPassword(id uint) (*models.User, error) {
	var user models.User

	// Truy vấn đầy đủ, bao gồm cả Password
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
