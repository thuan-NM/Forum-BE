package repositories

import (
	"Forum_BE/models"
	"fmt"
	"gorm.io/gorm"
	"log"
	"strconv"
	"time"
)

type QuestionRepository interface {
	CreateQuestion(question *models.Question) error
	GetQuestionByID(id uint) (*models.Question, error)
	UpdateQuestion(question *models.Question) error
	DeleteQuestion(id uint) error
	ListQuestions(filters map[string]interface{}) ([]models.Question, int, error)
	ListQuestionsExcludingPassed(filters map[string]interface{}) ([]models.Question, int, error)
	GetPassedQuestionIDs(userID uint) ([]uint, error)
	UpdateInteractionStatus(id uint, status string) error
	UpdateQuestionStatus(id uint, status string) error
	GetQuestionByIDMinimal(id uint) (*models.Question, error)
	GetAllQuestion(filters map[string]interface{}) ([]models.Question, int, error)
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
		Preload("Follows").
		Preload("Topic").
		First(&question, id).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *questionRepository) GetQuestionByIDMinimal(id uint) (*models.Question, error) {
	var question models.Question
	err := r.db.Model(&models.Question{}).First(&question, id).Error
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

func (r *questionRepository) ListQuestions(filters map[string]interface{}) ([]models.Question, int, error) {
	var questions []models.Question
	// Process pagination parameters
	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Process sort parameter
	sortOrder := "created_at DESC"
	if sort, ok := filters["sort"].(string); ok {
		if sort == "asc" {
			sortOrder = "created_at ASC"
		} else if sort == "desc" {
			sortOrder = "created_at DESC"
		}
	}

	// Query for counting total
	countQuery := r.db.Model(&models.Question{})
	if search, ok := filters["title_search"]; ok {
		countQuery = countQuery.Where("title LIKE ?", "%"+search.(string)+"%")
	}
	if status, ok := filters["status"]; ok {
		countQuery = countQuery.Where("status = ?", status)
	}
	if interstatus, ok := filters["interstatus"]; ok {
		countQuery = countQuery.Where("interaction_status = ?", interstatus)
	}
	if topicIDs, ok := filters["topic_id"].([]string); ok && len(topicIDs) > 0 {
		topicIDList := make([]uint, 0, len(topicIDs))
		for _, id := range topicIDs {
			if topicID, err := strconv.ParseUint(id, 10, 64); err == nil {
				topicIDList = append(topicIDList, uint(topicID))
			}
		}
		if len(topicIDList) > 0 {
			countQuery = countQuery.Where("topic_id IN ?", topicIDList)
		}
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		log.Printf("Error counting questions: %v", err)
		return nil, 0, err
	}

	// Apply filters and pagination
	query := r.db.Model(&models.Question{}).Preload("User").Preload("Topic").Preload("Answers").Preload("Follows")
	if search, ok := filters["title_search"]; ok {
		query = query.Where("title LIKE ?", "%"+search.(string)+"%")
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if interstatus, ok := filters["interstatus"]; ok {
		query = query.Where("interaction_status = ?", interstatus)
	}
	if topicIDs, ok := filters["topic_id"].([]string); ok && len(topicIDs) > 0 {
		topicIDList := make([]uint, 0, len(topicIDs))
		for _, id := range topicIDs {
			if topicID, err := strconv.ParseUint(id, 10, 64); err == nil {
				topicIDList = append(topicIDList, uint(topicID))
			}
		}
		if len(topicIDList) > 0 {
			query = query.Where("topic_id IN ?", topicIDList)
		}
	}
	query = query.Offset(offset).Limit(limit).Order(sortOrder)
	err := query.Find(&questions).Error
	if err != nil {
		log.Printf("Error fetching questions: %v", err)
		return nil, 0, err
	}
	log.Printf("Found %d questions with total %d", len(questions), total)
	return questions, int(total), nil
}

func (r *questionRepository) GetAllQuestion(filters map[string]interface{}) ([]models.Question, int, error) {
	var questions []models.Question
	// Process pagination parameters
	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Process sort parameter
	sortOrder := "created_at DESC"
	if sort, ok := filters["sort"].(string); ok {
		if sort == "asc" {
			sortOrder = "created_at ASC"
		} else if sort == "desc" {
			sortOrder = "created_at DESC"
		}
	}

	// Query for counting total
	countQuery := r.db.Model(&models.Question{})
	if search, ok := filters["title_search"]; ok {
		countQuery = countQuery.Where("title LIKE ?", "%"+search.(string)+"%")
	}
	if status, ok := filters["status"]; ok {
		countQuery = countQuery.Where("status = ?", status)
	}
	if interstatus, ok := filters["interstatus"]; ok {
		countQuery = countQuery.Where("interaction_status = ?", interstatus)
	}
	if topicIDs, ok := filters["topic_id"].([]string); ok && len(topicIDs) > 0 {
		topicIDList := make([]uint, 0, len(topicIDs))
		for _, id := range topicIDs {
			if topicID, err := strconv.ParseUint(id, 10, 64); err == nil {
				topicIDList = append(topicIDList, uint(topicID))
			}
		}
		if len(topicIDList) > 0 {
			countQuery = countQuery.Where("topic_id IN ?", topicIDList)
		}
	}
	if user_id, okUserId := filters["user_id"]; okUserId {
		countQuery = countQuery.Where("user_id = ?", user_id)
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		log.Printf("Error counting questions: %v", err)
		return nil, 0, err
	}

	// Apply filters and pagination
	query := r.db.Model(&models.Question{}).Preload("User").Preload("Topic").Preload("Answers").Preload("Follows")
	if search, ok := filters["title_search"]; ok {
		query = query.Where("title LIKE ?", "%"+search.(string)+"%")
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if interstatus, ok := filters["interstatus"]; ok {
		query = query.Where("interaction_status = ?", interstatus)
	}
	if topicIDs, ok := filters["topic_id"].([]string); ok && len(topicIDs) > 0 {
		topicIDList := make([]uint, 0, len(topicIDs))
		for _, id := range topicIDs {
			if topicID, err := strconv.ParseUint(id, 10, 64); err == nil {
				topicIDList = append(topicIDList, uint(topicID))
			}
		}
		if len(topicIDList) > 0 {
			query = query.Where("topic_id IN ?", topicIDList)
		}
	}
	if user_id, okUserId := filters["user_id"]; okUserId {
		query = query.Where("user_id = ?", user_id)
	}
	query = query.Offset(offset).Limit(limit).Order(sortOrder)
	err := query.Find(&questions).Error
	if err != nil {
		log.Printf("Error fetching questions: %v", err)
		return nil, 0, err
	}
	log.Printf("Found %d questions with total %d", len(questions), total)
	return questions, int(total), nil
}

func (r *questionRepository) ListQuestionsExcludingPassed(filters map[string]interface{}) ([]models.Question, int, error) {
	var questions []models.Question
	// Process pagination parameters
	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Process sort parameter
	sortOrder := "created_at DESC"
	if sort, ok := filters["sort"].(string); ok {
		if sort == "asc" {
			sortOrder = "created_at ASC"
		} else if sort == "desc" {
			sortOrder = "created_at DESC"
		}
	}

	// Extract user_id for passed questions filtering
	userID, ok := filters["user_id"].(uint)
	if !ok || userID == 0 {
		return nil, 0, fmt.Errorf("user_id is required for excluding passed questions")
	}

	// Query for counting total
	countQuery := r.db.Model(&models.Question{})
	countQuery = countQuery.Where("questions.id NOT IN (?)",
		r.db.Table("passed_questions").Select("question_id").Where("user_id = ?", userID))
	if search, ok := filters["title_search"]; ok {
		countQuery = countQuery.Where("title LIKE ?", "%"+search.(string)+"%")
	}
	if status, ok := filters["status"]; ok {
		countQuery = countQuery.Where("status = ?", status)
	}
	if interstatus, ok := filters["interstatus"]; ok {
		countQuery = countQuery.Where("interaction_status = ?", interstatus)
	}
	if topicIDs, ok := filters["topic_id"].([]string); ok && len(topicIDs) > 0 {
		topicIDList := make([]uint, 0, len(topicIDs))
		for _, id := range topicIDs {
			if topicID, err := strconv.ParseUint(id, 10, 64); err == nil {
				topicIDList = append(topicIDList, uint(topicID))
			}
		}
		if len(topicIDList) > 0 {
			countQuery = countQuery.Where("topic_id IN ?", topicIDList)
		}
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		log.Printf("Error counting questions excluding passed: %v", err)
		return nil, 0, err
	}

	// Apply filters and pagination
	query := r.db.Model(&models.Question{}).Preload("User").Preload("Topic").Preload("Answers").Preload("Follows")
	query = query.Where("questions.id NOT IN (?)",
		r.db.Table("passed_questions").Select("question_id").Where("user_id = ?", userID))
	if search, ok := filters["title_search"]; ok {
		query = query.Where("title LIKE ?", "%"+search.(string)+"%")
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if interstatus, ok := filters["interstatus"]; ok {
		query = query.Where("interaction_status = ?", interstatus)
	}
	if topicIDs, ok := filters["topic_id"].([]string); ok && len(topicIDs) > 0 {
		topicIDList := make([]uint, 0, len(topicIDs))
		for _, id := range topicIDs {
			if topicID, err := strconv.ParseUint(id, 10, 64); err == nil {
				topicIDList = append(topicIDList, uint(topicID))
			}
		}
		if len(topicIDList) > 0 {
			query = query.Where("topic_id IN ?", topicIDList)
		}
	}
	query = query.Offset(offset).Limit(limit).Order(sortOrder)
	err := query.Find(&questions).Error
	if err != nil {
		log.Printf("Error fetching questions excluding passed: %v", err)
		return nil, 0, err
	}
	log.Printf("Found %d questions excluding passed with total %d", len(questions), total)
	return questions, int(total), nil
}

func (r *questionRepository) GetPassedQuestionIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Table("passed_questions").Where("user_id = ?", userID).Pluck("question_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *questionRepository) UpdateInteractionStatus(id uint, status string) error {
	return r.db.Model(&models.Question{}).Where("id = ?", id).Updates(map[string]interface{}{
		"interaction_status": status,
		"updated_at":         time.Now(),
	}).Error
}

func (r *questionRepository) UpdateQuestionStatus(id uint, status string) error {
	return r.db.Model(&models.Question{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}
