package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uint) error
	ListUsers() ([]models.User, error)
	ListUsersPaginated(limit, offset int) ([]models.User, int64, error)
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
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
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

func (r *userRepository) ListUsers() ([]models.User, error) {
	var users []models.User
	err := r.db.Where("deleted_at IS NULL").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) ListUsersPaginated(limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	err := r.db.Model(&models.User{}).Where("deleted_at IS NULL").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("deleted_at IS NULL").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
