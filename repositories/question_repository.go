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
	GetPassedQuestionIDs(userID uint) ([]uint, error)
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
		Preload("Answers").
		Preload("Comments").
		Preload("Tags").
		Preload("Follows").
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
	query := r.db.Preload("User").Preload("Tags").Preload("Answers").Preload("Comments").Preload("Follows")
	if sort, ok := filters["sort"]; ok && sort == "follow_count" {
		query = query.Order("(SELECT COUNT(*) FROM follows WHERE follows.question_id = questions.id) DESC")
	}

	err := query.Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}
func (r *questionRepository) GetPassedQuestionIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Table("passed_questions").Where("user_id = ?", userID).Pluck("question_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}
