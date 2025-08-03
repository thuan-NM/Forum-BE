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
	CreateAnswer(answer *models.Answer, tagId []uint) error
	GetAnswerByID(id uint) (*models.Answer, error)
	UpdateAnswer(answer *models.Answer, tagId []uint) error
	DeleteAnswer(id uint) error
	ListAnswers(filters map[string]interface{}) ([]models.Answer, int, error)
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
	user_id, okUserId := filters["user_id"].(uint)

	// Process sort parameter
	sortOrder := "created_at DESC"
	if sort, ok := filters["sort"].(string); ok {
		if sort == "asc" {
			sortOrder = "created_at ASC"
		} else if sort == "desc" {
			sortOrder = "created_at DESC"
		}
		log.Printf("Applying sort order: %s", sortOrder)
	}

	// Process tag_ids filter
	if tagIDs, ok := filters["tagfilter"].([]uint); ok && len(tagIDs) > 0 {
		query = query.
			Joins("JOIN answer_tags ON answer_tags.answer_id = answers.id").
			Where("answer_tags.tag_id IN ?", tagIDs).
			Group("answers.id")
	}

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
	if okUserId && user_id != 0 {
		query = query.Where("user_id = ?", user_id)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("Error counting answers: %v", err)
		return nil, 0, err
	}

	query = query.Offset(offset).Limit(limit).Preload("User").Preload("Question").Preload("Comments").Preload("Tags").Order(sortOrder)
	if err := query.Find(&answers).Error; err != nil {
		log.Printf("Error fetching answers: %v", err)
		return nil, 0, err
	}
	log.Printf("Found %d answers with total %d, sort order: %s", len(answers), total, sortOrder)
	return answers, int(total), nil
}

func (r *answerRepository) CreateAnswer(answer *models.Answer, tagIDs []uint) error {
	answer.PlainContent = utils.StripHTML(answer.Content)
	tx := r.db.Begin()
	if err := tx.Error; err != nil {
		return err
	}
	if err := tx.Create(answer).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(tagIDs) > 0 {
		var tags []models.Tag
		if err := tx.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(answer).Association("Tags").Replace(tags); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *answerRepository) GetAnswerByID(id uint) (*models.Answer, error) {
	var answer models.Answer
	err := r.db.Preload("User").
		Preload("Question").
		Preload("Comments").
		Preload("Tags").
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

func (r *answerRepository) UpdateAnswer(answer *models.Answer, tagId []uint) error {
	answer.PlainContent = utils.StripHTML(answer.Content)
	tx := r.db.Begin()
	if err := tx.Error; err != nil {
		return err
	}
	if err := tx.Save(answer).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(tagId) > 0 {
		var tags []models.Tag
		if err := tx.Where("id IN ?", tagId).Find(&tags).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(answer).Association("Tags").Replace(tags); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *answerRepository) DeleteAnswer(id uint) error {
	return r.db.Delete(&models.Answer{}, id).Error
}

func (r *answerRepository) ListAnswers(filters map[string]interface{}) ([]models.Answer, int, error) {
	var answers []models.Answer

	// Process sort parameter
	sortOrder := "created_at DESC"
	if sort, ok := filters["sort"].(string); ok {
		if sort == "asc" {
			sortOrder = "created_at ASC"
		} else if sort == "desc" {
			sortOrder = "created_at DESC"
		}
		log.Printf("Applying sort order: %s", sortOrder)
	}

	// Build count query
	countQuery := r.db.Model(&models.Answer{})
	if filters != nil {
		if tagIDs, ok := filters["tagfilter"].([]uint); ok && len(tagIDs) > 0 {
			countQuery = countQuery.
				Joins("JOIN answer_tags ON answer_tags.answer_id = answers.id").
				Where("answer_tags.tag_id IN ?", tagIDs).
				Group("answers.id")
		}
		for key, value := range filters {
			if key != "limit" && key != "page" && key != "tag_ids" && key != "sort" {
				countQuery = countQuery.Where(key, value)
			}
		}
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		log.Printf("Error counting answers: %v", err)
		return nil, 0, err
	}

	// Build data query
	query := r.db.Preload("User").Preload("Question.User").Preload("Question").Preload("Comments").Preload("Tags").Order(sortOrder)
	if filters != nil {
		if tagIDs, ok := filters["tagfilter"].([]uint); ok && len(tagIDs) > 0 {
			query = query.
				Joins("JOIN answer_tags ON answer_tags.answer_id = answers.id").
				Where("answer_tags.tag_id IN ?", tagIDs).
				Group("answers.id")
		}
		for key, value := range filters {
			if key != "limit" && key != "page" && key != "tag_ids" && key != "sort" {
				query = query.Where(key, value)
			}
		}
	}

	// Apply pagination
	if limit, ok := filters["limit"].(int); ok && limit > 0 {
		query = query.Limit(limit)
	}
	if page, ok := filters["page"].(int); ok && page > 0 {
		offset := (page - 1) * filters["limit"].(int)
		query = query.Offset(offset)
	}

	err := query.Find(&answers).Error
	if err != nil {
		log.Printf("Error fetching answers: %v", err)
		return nil, 0, err
	}
	log.Printf("Found %d answers with total %d, sort order: %s", len(answers), total, sortOrder)
	return answers, int(total), nil
}

func (r *answerRepository) UpdateAnswerStatus(id uint, status string) error {
	return r.db.Model(&models.Answer{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}
