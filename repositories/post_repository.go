package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type PostRepository interface {
	CreatePost(post *models.Post) error
	GetPostByID(id uint) (*models.Post, error)
	UpdatePost(post *models.Post) error
	DeletePost(id uint) error
	List(filters map[string]interface{}) ([]models.Post, error)
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) CreatePost(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *postRepository) GetPostByID(id uint) (*models.Post, error) {
	var post models.Post
	if err := r.db.
		Preload("User").
		Preload("Comments").
		Preload("Group").
		Preload("Tags").
		First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) UpdatePost(post *models.Post) error {
	return r.db.Save(post).Error
}

func (r *postRepository) DeletePost(id uint) error {
	return r.db.Delete(&models.Post{}, id).Error
}

func (r *postRepository) List(filters map[string]interface{}) ([]models.Post, error) {
	var posts []models.Post
	dbQuery := r.db.Model(&models.Post{})
	for key, value := range filters {
		if key == "content LIKE ?" {
			dbQuery = dbQuery.Where(key, value)
		} else {
			dbQuery = dbQuery.Where(key, value)
		}
	}

	if err := dbQuery.
		Preload("User").
		Preload("Comments").
		Preload("Group").
		Preload("Tags").
		Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}
