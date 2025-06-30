package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type ReactionRepository interface {
	CreateReaction(reaction *models.Reaction) error
	GetReactionByID(id uint) (*models.Reaction, error)
	UpdateReaction(reaction *models.Reaction) error
	DeleteReaction(id uint) error
	ListReactions(filters map[string]interface{}) ([]models.Reaction, int, error)
}

type reactionRepository struct {
	db *gorm.DB
}

func NewReactionRepository(db *gorm.DB) ReactionRepository {
	return &reactionRepository{db: db}
}

func (r *reactionRepository) CreateReaction(reaction *models.Reaction) error {
	return r.db.Create(reaction).Error
}

func (r *reactionRepository) GetReactionByID(id uint) (*models.Reaction, error) {
	var reaction models.Reaction
	if err := r.db.Preload("User").First(&reaction, id).Error; err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *reactionRepository) UpdateReaction(reaction *models.Reaction) error {
	return r.db.Save(reaction).Error
}

func (r *reactionRepository) DeleteReaction(id uint) error {
	return r.db.Delete(&models.Reaction{}, id).Error
}

func (r *reactionRepository) ListReactions(filters map[string]interface{}) ([]models.Reaction, int, error) {
	var reactions []models.Reaction
	query := r.db.Model(&models.Reaction{})

	// Filter by UserID
	if userID, ok := filters["user_id"].(uint); ok {
		query = query.Where("user_id = ?", userID)
	}
	// Filter by ReactableID
	if reactableID, ok := filters["reactable_id"].(uint); ok {
		query = query.Where("reactable_id = ?", reactableID)
	}
	// Filter by ReactableType
	if reactableType, ok := filters["reactable_type"].(string); ok {
		query = query.Where("reactable_type = ?", reactableType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").Find(&reactions).Error; err != nil {
		return nil, 0, err
	}

	return reactions, int(total), nil
}
