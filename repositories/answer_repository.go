package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type AnswerRepository interface {
	CreateAnswer(answer *models.Answer) error
	GetAnswerByID(id uint) (*models.Answer, error)
	UpdateAnswer(answer *models.Answer) error
	DeleteAnswer(id uint) error
	ListAnswers(filters map[string]interface{}) ([]models.Answer, error)
}

type answerRepository struct {
	db *gorm.DB
}

func NewAnswerRepository(db *gorm.DB) AnswerRepository {
	return &answerRepository{db: db}
}

func (r *answerRepository) CreateAnswer(answer *models.Answer) error {
	return r.db.Create(answer).Error
}

func (r *answerRepository) GetAnswerByID(id uint) (*models.Answer, error) {
	var answer models.Answer
	err := r.db.Preload("User").
		Preload("Question").
		Preload("Comments").
		Preload("Votes").
		First(&answer, id).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func (r *answerRepository) UpdateAnswer(answer *models.Answer) error {
	return r.db.Save(answer).Error
}

func (r *answerRepository) DeleteAnswer(id uint) error {
	return r.db.Delete(&models.Answer{}, id).Error
}

func (r *answerRepository) ListAnswers(filters map[string]interface{}) ([]models.Answer, error) {
	var answers []models.Answer
	query := r.db.Preload("User").Preload("Question").Preload("Comments").Preload("Votes")

	// Áp dụng các bộ lọc nếu có
	if filters != nil {
		for key, value := range filters {
			query = query.Where(key+" = ?", value)
		}
	}

	err := query.Find(&answers).Error
	if err != nil {
		return nil, err
	}

	return answers, nil
}
