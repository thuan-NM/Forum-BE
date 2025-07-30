package repositories

import (
	"Forum_BE/models"
	"Forum_BE/utils"
	"gorm.io/gorm"
	"log"
	"time"
)

type PostRepository interface {
	CreatePost(post *models.Post, tagIds []uint) error
	GetPostByID(id uint) (*models.Post, error)
	GetPostByIDSimple(id uint) (*models.Post, error)
	UpdatePost(post *models.Post, tagId []uint) error
	UpdatePostStatus(id uint, status string) error
	DeletePost(id uint) error
	List(filters map[string]interface{}) ([]models.Post, int, error)
	GetAllPosts(filters map[string]interface{}) ([]models.Post, int, error)
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) CreatePost(post *models.Post, tagIds []uint) error {
	post.PlainContent = utils.StripHTML(post.Content)
	tx := r.db.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(post).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(tagIds) > 0 {
		var tags []models.Tag
		if err := tx.Where("id IN ?", tagIds).Find(&tags).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(post).Association("Tags").Replace(tags); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *postRepository) GetPostByID(id uint) (*models.Post, error) {
	var post models.Post
	if err := r.db.
		Preload("User").
		Preload("Tags").
		Preload("Comments").
		First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) GetPostByIDSimple(id uint) (*models.Post, error) {
	var post models.Post
	if err := r.db.First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) UpdatePost(post *models.Post, tagId []uint) error {
	post.PlainContent = utils.StripHTML(post.Content)
	tx := r.db.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Save(post).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(tagId) > 0 {
		var tags []models.Tag
		if err := tx.Where("id IN ?", tagId).Find(&tags).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(post).Association("Tags").Replace(tags); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *postRepository) UpdatePostStatus(id uint, status string) error {
	return r.db.Model(&models.Post{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}

func (r *postRepository) DeletePost(id uint) error {
	return r.db.Delete(&models.Post{}, id).Error
}

func (r *postRepository) List(filters map[string]interface{}) ([]models.Post, int, error) {
	var posts []models.Post

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

	// Build count query
	countQuery := r.db.Model(&models.Post{})
	for key, value := range filters {
		if key != "limit" && key != "page" {
			countQuery = countQuery.Where(key, value)
		}
	}
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		log.Printf("Error counting posts: %v", err)
		return nil, 0, err
	}

	// Build data query
	query := r.db.Preload("User").Preload("Tags").Preload("Comments")
	for key, value := range filters {
		if key != "limit" && key != "page" {
			query = query.Where(key, value)
		}
	}

	query = query.Offset(offset).Limit(limit).Order("created_at DESC")
	if err := query.Find(&posts).Error; err != nil {
		log.Printf("Error fetching posts: %v", err)
		return nil, 0, err
	}

	log.Printf("Found %d posts with total %d", len(posts), total)
	return posts, int(total), nil
}

func (r *postRepository) GetAllPosts(filters map[string]interface{}) ([]models.Post, int, error) {
	var posts []models.Post
	query := r.db.Model(&models.Post{})

	// Process filters
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("LOWER(title) LIKE LOWER(?)", "%"+search+"%")
	}
	if user_id, ok := filters["user_id"].(uint); ok {
		query = query.Where("user_id = ?", user_id)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if tagfilter, ok := filters["tagfilter"].(uint); ok {
		query = query.
			Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.id = ?", tagfilter)
	}

	// Process pagination
	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("Error counting posts: %v", err)
		return nil, 0, err
	}

	// Fetch data
	query = query.Offset(offset).Limit(limit).Preload("User").Preload("Tags").Preload("Comments").Order("created_at DESC")
	if err := query.Find(&posts).Error; err != nil {
		log.Printf("Error fetching posts: %v", err)
		return nil, 0, err
	}

	log.Printf("Found %d posts with total %d", len(posts), total)
	return posts, int(total), nil
}
