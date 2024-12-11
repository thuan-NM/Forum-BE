package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type QuestionRepository interface {
	CreateQuestion(question *models.Question) error
	GetQuestionByID(id uint) (*models.Question, error)
	UpdateQuestion(question *models.Question) error
	DeleteQuestion(id uint) error
	ListQuestions(filters map[string]interface{}) ([]models.Question, error)
}

type questionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) QuestionRepository {
	return &questionRepository{db: db}
}

func (r *questionRepository) CreateQuestion(question *models.Question) error {
	return r.db.Create(question).Error
}

func (r *questionRepository) GetQuestionByID(id uint) (*models.Question, error) {
	var question models.Question
	err := r.db.Preload("User").
		Preload("Group").
		Preload("Answers").
		Preload("Comments").
		Preload("Tags").
		First(&question, id).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *questionRepository) UpdateQuestion(question *models.Question) error {
	return r.db.Save(question).Error
}

func (r *questionRepository) DeleteQuestion(id uint) error {
	return r.db.Delete(&models.Question{}, id).Error
}

func (r *questionRepository) ListQuestions(filters map[string]interface{}) ([]models.Question, error) {
	var questions []models.Question
	query := r.db.Preload("User").Preload("Group").Preload("Tags").Preload("Answers").Preload("Comments")

	// Áp dụng các bộ lọc nếu có
	if filters != nil {
		for key, value := range filters {
			if key == "title_search" {
				query = query.Where("title LIKE ?", "%"+value.(string)+"%")
			} else {
				query = query.Where(key+" = ?", value)
			}
		}
	}

	err := query.Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}
