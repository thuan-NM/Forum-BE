package repositories

import (
	"Forum_BE/models"
	"Forum_BE/utils"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

type AnswerRepository interface {
	CreateAnswer(answer *models.Answer) error
	GetAnswerByID(id uint) (*models.Answer, error)
	UpdateAnswer(answer *models.Answer) error
	DeleteAnswer(id uint) error
	ListAnswers(filters map[string]interface{}) ([]models.Answer, error)
	GetAllAnswers(filters map[string]interface{}) ([]models.Answer, int, error)
	UpdateAnswerStatus(id uint, status string) error
	GetAnswerByIDSimple(id uint) (*models.Answer, error)
}

type answerRepository struct {
	db *gorm.DB
}

func NewAnswerRepository(db *gorm.DB) AnswerRepository {
	return &answerRepository{db: db}
}

func (r *answerRepository) GetAllAnswers(filters map[string]interface{}) ([]models.Answer, int, error) {
	var answers []models.Answer
	query := r.db.Model(&models.Answer{})

	// Process filters
	search, ok := filters["search"].(string)
	questiontitle, okQt := filters["questiontitle"].(string)
	status, okStatus := filters["status"].(string)
	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)

	// Default pagination values
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	if okQt && questiontitle != "" {
		query = query.Joins("JOIN questions ON questions.id = answers.question_id").
			Where("LOWER(questions.title) LIKE LOWER(?)", "%"+questiontitle+"%")
	}
	if okStatus && status != "" {
		query = query.Where("status = ?", status)
	}
	if ok && search != "" {
		search = strings.ToLower(search)
		query = query.Where("plain_content LIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("Error counting answers: %v", err)
		return nil, 0, err
	}

	query = query.Offset(offset).Limit(limit).Preload("User").Preload("Question").Preload("Comments").Preload("Votes")
	if err := query.Find(&answers).Error; err != nil {
		log.Printf("Error fetching answers: %v", err)
		return nil, 0, err
	}

	log.Printf("Found %d answers for search '%s'", len(answers), search)
	return answers, int(total), nil
}

func (r *answerRepository) CreateAnswer(answer *models.Answer) error {
	answer.PlainContent = utils.StripHTML(answer.Content)
	return r.db.Create(answer).Error
}

func (r *answerRepository) GetAnswerByID(id uint) (*models.Answer, error) {
	var answer models.Answer
	err := r.db.Preload("User").
		Preload("Question").
		Preload("Comments").
		First(&answer, id).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func (r *answerRepository) GetAnswerByIDSimple(id uint) (*models.Answer, error) {
	var answer models.Answer
	err := r.db.First(&answer, id).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func (r *answerRepository) UpdateAnswer(answer *models.Answer) error {
	answer.PlainContent = utils.StripHTML(answer.Content)
	return r.db.Save(answer).Error
}

func (r *answerRepository) DeleteAnswer(id uint) error {
	return r.db.Delete(&models.Answer{}, id).Error
}

func (r *answerRepository) ListAnswers(filters map[string]interface{}) ([]models.Answer, error) {
	var answers []models.Answer
	query := r.db.Preload("User").Preload("Question").Preload("Comments").Preload("Votes")

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

func (r *answerRepository) UpdateAnswerStatus(id uint, status string) error {
	return r.db.Model(&models.Answer{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}
