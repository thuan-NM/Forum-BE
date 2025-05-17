package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type PostService interface {
	CreatePost(content string, userID uint, status models.PostStatus) (*models.Post, error)
	GetPostByID(id uint) (*models.Post, error)
	DeletePost(id uint) error
	UpdatePost(id uint, content string, status models.PostStatus) (*models.Post, error)
	ListPosts(filters map[string]interface{}) ([]models.Post, error)
}

type postService struct {
	postRepo    repositories.PostRepository
	redisClient *redis.Client
}

func NewPostService(postRepo repositories.PostRepository, redisClient *redis.Client) PostService {
	return &postService{postRepo: postRepo, redisClient: redisClient}
}

func (s *postService) CreatePost(content string, userID uint, status models.PostStatus) (*models.Post, error) {
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	post := &models.Post{
		Content: content,
		UserID:  userID,
		Status:  status,
	}

	if err := s.postRepo.CreatePost(post); err != nil {
		return nil, err
	}

	s.updateCacheAfterCreate(post)
	
	return post, nil
}

func (s *postService) GetPostByID(id uint) (*models.Post, error) {
	cacheKey := fmt.Sprintf("post:%d", id)

	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var post models.Post
		if err := json.Unmarshal([]byte(cached), &post); err == nil {
			log.Printf("Cache hit for post:%d", id)
			return &post, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for post:%d: %v", id, err)
	}

	post, err := s.postRepo.GetPostByID(id)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(post)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for post:%d: %v", id, err)
		}
	}

	return post, nil
}

func (s *postService) DeletePost(id uint) error {
	_, err := s.postRepo.GetPostByID(id)
	if err != nil {
		return err
	}

	err = s.postRepo.DeletePost(id)
	if err != nil {
		return err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("post:%d", id))
	s.invalidateCache("posts:*")

	return nil
}

func (s *postService) UpdatePost(id uint, content string, status models.PostStatus) (*models.Post, error) {
	post, err := s.postRepo.GetPostByID(id)
	if err != nil {
		return nil, err
	}

	if content != "" {
		post.Content = content
	}
	if status != "" {
		post.Status = status
	}

	if err := s.postRepo.UpdatePost(post); err != nil {
		return nil, err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("post:%d", id))
	s.invalidateCache("posts:*")

	return post, nil
}

func (s *postService) ListPosts(filters map[string]interface{}) ([]models.Post, error) {
	cacheKey := utils.GenerateCacheKey("posts", 0, filters)

	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var posts []models.Post
		if err := json.Unmarshal([]byte(cached), &posts); err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return posts, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	posts, err := s.postRepo.List(filters)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(posts)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}

	return posts, nil
}

func (s *postService) updateCacheAfterCreate(post *models.Post) {
	cachePattern := "posts:*"
	ctx := context.Background()

	// Sử dụng pipeline để lấy và cập nhật cache
	pipe := s.redisClient.Pipeline()
	keys, err := s.redisClient.Keys(ctx, cachePattern).Result()
	if err != nil {
		log.Printf("Failed to get cache keys for pattern %s: %v", cachePattern, err)
		return
	}

	for _, key := range keys {
		cached, err := s.redisClient.Get(ctx, key).Result()
		if err == nil {
			var posts []models.Post
			if err := json.Unmarshal([]byte(cached), &posts); err == nil {
				posts = append([]models.Post{*post}, posts...)
				data, err := json.Marshal(posts)
				if err == nil {
					pipe.Set(ctx, key, data, 2*time.Minute)
				}
			}
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("Failed to update cache for pattern %s: %v", cachePattern, err)
	}
}

func (s *postService) invalidateCache(pattern string) {
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Failed to invalidate cache for pattern %s: %v", pattern, err)
		return
	}

	if len(keys) > 0 {
		if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Failed to delete cache keys %v: %v", keys, err)
		}
	}
}
