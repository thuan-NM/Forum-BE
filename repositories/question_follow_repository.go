package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type QuestionFollowRepository interface {
	CreateFollow(follow *models.QuestionFollow) error
	DeleteFollow(userID, questionID uint) error
	GetFollowsByQuestion(questionID uint) ([]models.QuestionFollow, error)
	ExistsByQuestionAndUser(questionID, userID uint) (bool, error)
}

type questionFollowRepository struct {
	db *gorm.DB
}

func NewQuestionFollowRepository(db *gorm.DB) QuestionFollowRepository {
	return &questionFollowRepository{db: db}
}

func (r *questionFollowRepository) CreateFollow(follow *models.QuestionFollow) error {
	return r.db.Create(follow).Error
}

func (r *questionFollowRepository) DeleteFollow(userID, questionID uint) error {
	return r.db.
		Unscoped(). // Bỏ qua soft delete
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&models.QuestionFollow{}).Error
}

func (r *questionFollowRepository) GetFollowsByQuestion(questionID uint) ([]models.QuestionFollow, error) {
	var follows []models.QuestionFollow
	err := r.db.Where("question_id = ?", questionID).Find(&follows).Error
	return follows, err
}
func (r *questionFollowRepository) ExistsByQuestionAndUser(questionID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.QuestionFollow{}).Where("question_id = ? AND user_id = ?", questionID, userID).Count(&count).Error
	return count > 0, err
}
