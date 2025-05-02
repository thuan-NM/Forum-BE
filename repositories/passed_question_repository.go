package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type PassRepository interface {
	PassQuestion(userID, questionID uint) error
	IsPassed(userID, questionID uint) (bool, error)
	GetPassedQuestionIDs(userID uint) ([]uint, error)
}

type passRepository struct {
	db *gorm.DB
}

func NewPassRepository(db *gorm.DB) PassRepository {
	return &passRepository{db: db}
}

func (r *passRepository) PassQuestion(userID, questionID uint) error {
	pass := &models.PassedQuestion{
		UserID:     userID,
		QuestionID: questionID,
	}
	return r.db.Create(pass).Error
}

func (r *passRepository) IsPassed(userID, questionID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.PassedQuestion{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Count(&count).Error
	return count > 0, err
}

func (r *passRepository) GetPassedQuestionIDs(userID uint) ([]uint, error) {
	var passed []models.PassedQuestion
	err := r.db.Where("user_id = ?", userID).Find(&passed).Error
	if err != nil {
		return nil, err
	}
	var ids []uint
	for _, p := range passed {
		ids = append(ids, p.QuestionID)
	}
	return ids, nil
}
