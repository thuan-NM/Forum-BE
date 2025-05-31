package repositories

import (
	"Forum_BE/models"
	"fmt"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

type CommentRepository interface {
	CreateComment(comment *models.Comment) error
	GetCommentByID(id uint) (*models.Comment, error)
	UpdateComment(comment *models.Comment) error
	DeleteComment(id uint) error
	ListComments(filters map[string]interface{}) ([]models.Comment, int64, error)
	ListReplies(parentID uint, filters map[string]interface{}) ([]models.Comment, int64, error)
	GetAllComments(filters map[string]interface{}) ([]models.Comment, int64, error) // Thêm method mới
	UpdateCommentStatus(id uint, status string) error
	GetAllChildCommentIDs(parentID uint) ([]uint, error)
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) CreateComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) GetCommentByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.Preload("User").
		Where("deleted_at IS NULL").
		First(&comment, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %v", err)
	}
	return &comment, nil
}

func (r *commentRepository) UpdateComment(comment *models.Comment) error {
	return r.db.Save(comment).Error
}
func (r *commentRepository) DeleteComment(id uint) error {
	childIDs, err := r.GetAllChildCommentIDs(id)
	if err != nil {
		return fmt.Errorf("failed to get child comments: %v", err)
	}

	allIDs := append(childIDs, id)

	return r.db.Model(&models.Comment{}).
		Where("id IN ?", allIDs).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *commentRepository) GetAllChildCommentIDs(parentID uint) ([]uint, error) {
	var childIDs []uint
	var queue []uint
	queue = append(queue, parentID)

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		var directChildren []uint
		err := r.db.Model(&models.Comment{}).
			Where("parent_id = ? AND deleted_at IS NULL", currentID).
			Pluck("id", &directChildren).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get direct children for comment %d: %v", currentID, err)
		}

		// Thêm các comment con vào childIDs và queue
		childIDs = append(childIDs, directChildren...)
		queue = append(queue, directChildren...)
	}

	return childIDs, nil
}
func (r *commentRepository) ListComments(filters map[string]interface{}) ([]models.Comment, int64, error) {
	var comments []models.Comment

	query := r.db.Preload("User").
		Where("parent_id IS NULL").
		Where("deleted_at IS NULL")
	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)

	// Default pagination values
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	allowedFilters := map[string]bool{
		"post_id":   true,
		"answer_id": true,
		"user_id":   true,
	}
	if filters != nil {
		for key, value := range filters {
			if key == "content" {
				query = query.Where("content LIKE ?", "%"+value.(string)+"%")
			} else if allowedFilters[key] {
				query = query.Where(fmt.Sprintf("%s = ?", key), value)
			}
		}
	}

	var total int64
	if err := query.Model(&models.Comment{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %v", err)
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).
		Select("comments.*, EXISTS (SELECT 1 FROM comments c WHERE c.parent_id = comments.id AND c.deleted_at IS NULL) AS has_replies").
		Find(&comments).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch comments: %v", err)
	}

	for i := range comments {
		var hasReplies bool
		err := r.db.Raw("SELECT EXISTS (SELECT 1 FROM comments WHERE parent_id = ? AND deleted_at IS NULL)", comments[i].ID).Scan(&hasReplies).Error
		if err == nil {
			comments[i].Metadata = []byte(fmt.Sprintf(`{"has_replies": %t}`, hasReplies))
		}
	}

	return comments, total, nil
}

func (r *commentRepository) ListReplies(parentID uint, filters map[string]interface{}) ([]models.Comment, int64, error) {
	var comments []models.Comment
	query := r.db.Preload("User").
		Where("parent_id = ? AND deleted_at IS NULL", parentID)
	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)

	// Default pagination values
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	var total int64
	if err := query.Model(&models.Comment{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count replies: %v", err)
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Find(&comments).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list replies: %v", err)
	}

	for i := range comments {
		var hasReplies bool
		err := r.db.Raw("SELECT EXISTS (SELECT 1 FROM comments WHERE parent_id = ? AND deleted_at IS NULL)", comments[i].ID).Scan(&hasReplies).Error
		if err == nil {
			comments[i].Metadata = []byte(fmt.Sprintf(`{"has_replies": %t}`, hasReplies))
		}
	}

	return comments, total, nil
}
func (r *commentRepository) GetAllComments(filters map[string]interface{}) ([]models.Comment, int64, error) {
	var comments []models.Comment
	query := r.db.Model(&models.Comment{})
	typefilter, okType := filters["typefilter"].(string)
	status, okStatus := filters["status"].(string)
	search, ok := filters["search"].(string)
	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)

	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	if okStatus && status != "" {
		query = query.Where("status = ?", status)
	}
	if ok && search != "" {
		search = strings.ToLower(search)
		query = query.Where("content LIKE ?", "%"+search+"%")
	}
	if okType && typefilter != "" && typefilter != "all" {
		switch typefilter {
		case "post_id":
			query = query.Where("post_id IS NOT NULL")
		case "answer_id":
			query = query.Where("answer_id IS NOT NULL")
		case "parent_id":
			query = query.Where("parent_id IS NOT NULL")
		}
	}
	var total int64
	if err := query.Model(&models.Comment{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count all comments: %v", err)
	}

	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit).Preload("User").Preload("Post").Preload("Answer").Preload("Parent")
	if err := query.Find(&comments).Error; err != nil {
		log.Printf("Error fetching comment: %v", err)
		return nil, 0, err
	}

	for i := range comments {
		var hasReplies bool
		err := r.db.Raw("SELECT EXISTS (SELECT 1 FROM comments WHERE parent_id = ? AND deleted_at IS NULL)", comments[i].ID).Scan(&hasReplies).Error
		if err == nil {
			comments[i].Metadata = []byte(fmt.Sprintf(`{"has_replies": %t}`, hasReplies))
		}
	}
	return comments, total, nil
}

func (r *commentRepository) UpdateCommentStatus(id uint, status string) error {
	return r.db.Model(&models.Comment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}
