package repositories

import (
	"Forum_BE/models"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
)

type ReactionRepository interface {
	CreateReaction(reaction *models.Reaction) error
	GetReactionByID(id uint) (*models.Reaction, error)
	UpdateReaction(reaction *models.Reaction) error
	DeleteReaction(id uint) error
	ListReactions(filters map[string]interface{}) ([]models.Reaction, int, error)
	GetReactionCount(postID, commentID, answerID *uint) (int64, error)
	ValidateReactionID(postID, commentID, answerID *uint) error
}

type reactionRepository struct {
	db *gorm.DB
}

func NewReactionRepository(db *gorm.DB) ReactionRepository {
	if db == nil {
		log.Fatal("database connection is nil")
	}
	log.Printf("Initialized ReactionRepository with db: %v", db != nil)
	return &reactionRepository{db: db}
}

func (r *reactionRepository) CreateReaction(reaction *models.Reaction) error {
	if r.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	log.Printf("Creating reaction with db: %v", r.db != nil)
	return r.db.Create(reaction).Error
}

func (r *reactionRepository) GetReactionByID(id uint) (*models.Reaction, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	log.Printf("Getting reaction %d with db: %v", id, r.db != nil)
	var reaction models.Reaction
	query := r.db.Preload("User").Where("deleted_at IS NULL")
	if err := query.Preload("Post").Preload("Comment").Preload("Answer").First(&reaction, id).Error; err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *reactionRepository) UpdateReaction(reaction *models.Reaction) error {
	if r.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	log.Printf("Updating reaction with db: %v", r.db != nil)
	return r.db.Save(reaction).Error
}

func (r *reactionRepository) DeleteReaction(id uint) error {
	if r.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	log.Printf("Deleting reaction %d with db: %v, hardDelete: %v", id, r.db != nil)

	return r.db.Unscoped().Where("id = ?", id).Delete(&models.Reaction{}).Error

}

func (r *reactionRepository) ListReactions(filters map[string]interface{}) ([]models.Reaction, int, error) {
	if r.db == nil {
		return nil, 0, fmt.Errorf("database connection is not initialized")
	}
	log.Printf("Listing reactions with db: %v", r.db != nil)
	var reactions []models.Reaction
	query := r.db.Model(&models.Reaction{}).Preload("User").Preload("Post").Preload("Comment").Preload("Answer").Where("deleted_at IS NULL")

	if userID, ok := filters["user_id"].(uint); ok {
		query = query.Where("user_id = ?", userID)
	}
	if postID, ok := filters["post_id"].(uint); ok {
		query = query.Where("post_id = ?", postID)
	}
	if commentID, ok := filters["comment_id"].(uint); ok {
		query = query.Where("comment_id = ?", commentID)
	}
	if answerID, ok := filters["answer_id"].(uint); ok {
		query = query.Where("answer_id = ?", answerID)
	}

	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}

	var total int64
	if err := query.Model(&models.Reaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&reactions).Error; err != nil {
		return nil, 0, err
	}

	return reactions, int(total), nil
}

func (r *reactionRepository) GetReactionCount(postID, commentID, answerID *uint) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection is not initialized")
	}
	log.Printf("Getting reaction count with db: %v", r.db != nil)
	var count int64
	query := r.db.Model(&models.Reaction{}).Where("deleted_at IS NULL")
	if postID != nil {
		query = query.Where("post_id = ?", *postID)
	} else if commentID != nil {
		query = query.Where("comment_id = ?", *commentID)
	} else if answerID != nil {
		query = query.Where("answer_id = ?", *answerID)
	} else {
		return 0, errors.New("at least one of post_id, comment_id, or answer_id must be provided")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reactionRepository) ValidateReactionID(postID, commentID, answerID *uint) error {
	if r.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	count := 0
	if postID != nil {
		count++
	}
	if commentID != nil {
		count++
	}
	if answerID != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("exactly one of post_id, comment_id, or answer_id must be provided")
	}

	var dbCount int64
	if postID != nil {
		log.Printf("Validating post_id %d with db: %v", *postID, r.db != nil)
		if err := r.db.Model(&models.Post{}).Where("id = ? AND deleted_at IS NULL", *postID).Count(&dbCount).Error; err != nil {
			return err
		}
		if dbCount == 0 {
			return fmt.Errorf("post with ID %d does not exist", *postID)
		}
	} else if commentID != nil {
		log.Printf("Validating comment_id %d with db: %v", *commentID, r.db != nil)
		if err := r.db.Model(&models.Comment{}).Where("id = ? AND deleted_at IS NULL", *commentID).Count(&dbCount).Error; err != nil {
			return err
		}
		if dbCount == 0 {
			return fmt.Errorf("comment with ID %d does not exist", *commentID)
		}
	} else if answerID != nil {
		log.Printf("Validating answer_id %d with db: %v", *answerID, r.db != nil)
		if err := r.db.Model(&models.Answer{}).Where("id = ? AND deleted_at IS NULL", *answerID).Count(&dbCount).Error; err != nil {
			return err
		}
		if dbCount == 0 {
			return fmt.Errorf("answer with ID %d does not exist", *answerID)
		}
	}
	return nil
}
