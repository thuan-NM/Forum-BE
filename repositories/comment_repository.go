package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type CommentRepository interface {
	CreateComment(comment *models.Comment) error
	GetCommentByID(id uint) (*models.Comment, error)
	UpdateComment(comment *models.Comment) error
	DeleteComment(id uint) error
	ListComments(filters map[string]interface{}) ([]models.Comment, error)
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
		Preload("Question").
		Preload("Answer").
		Preload("Votes").
		First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) UpdateComment(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

func (r *commentRepository) DeleteComment(id uint) error {
	return r.db.Delete(&models.Comment{}, id).Error
}

func (r *commentRepository) ListComments(filters map[string]interface{}) ([]models.Comment, error) {
	var comments []models.Comment
	query := r.db.Preload("User").Preload("Question").Preload("Answer").Preload("Votes")

	// Áp dụng các bộ lọc nếu có
	if filters != nil {
		for key, value := range filters {
			query = query.Where(key+" = ?", value)
		}
	}

	err := query.Find(&comments).Error
	if err != nil {
		return nil, err
	}

	return comments, nil
}
