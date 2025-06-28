package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type PostService interface {
	CreatePost(content string, userID uint, tagId []uint) (*models.Post, error)
	GetPostByID(id uint) (*models.Post, error)
	GetPostByIDSimple(id uint) (*models.Post, error)
	DeletePost(id uint) error
	UpdatePost(id uint, content string, status models.PostStatus, tagNames []string) (*models.Post, error)
	UpdatePostStatus(id uint, status string) (*models.Post, error)
	ListPosts(filters map[string]interface{}) ([]models.Post, int, error)
	GetAllPosts(filters map[string]interface{}) ([]models.Post, int, error)
}

type postService struct {
	postRepo    repositories.PostRepository
	redisClient *redis.Client
}

func NewPostService(postRepo repositories.PostRepository, redisClient *redis.Client) PostService {
	return &postService{postRepo: postRepo, redisClient: redisClient}
}

func (s *postService) CreatePost(content string, userID uint, tagId []uint) (*models.Post, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}

	post := &models.Post{
		Content: content,
		UserID:  userID,
	}

	if err := s.postRepo.CreatePost(post, tagId); err != nil {
		log.Printf("Failed to create post: %v", err)
		return nil, err
	}

	s.invalidateCache("posts:*")
	s.invalidateCache("tags:*") // Thêm invalidation cho tag cache
	log.Printf("Cache invalidated for posts:* and tags:* due to new post %d", post.ID)

	return post, nil
}

func (s *postService) GetPostByID(id uint) (*models.Post, error) {
	cacheKey := fmt.Sprintf("post:%d", id)
	ctx := context.Background()

	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var post models.Post
			if err := json.Unmarshal([]byte(cached), &post); err == nil {
				log.Printf("Cache hit for post:%d", id)
				return &post, nil
			}
			log.Printf("Failed to unmarshal cache for post %d: %v", id, err)
		}
		if err != redis.Nil {
			log.Printf("Redis error for post:%d (attempt %d): %v", id, attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	post, err := s.postRepo.GetPostByID(id)
	if err != nil {
		log.Printf("Failed to get post %d: %v", id, err)
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

func (s *postService) GetPostByIDSimple(id uint) (*models.Post, error) {
	cacheKey := fmt.Sprintf("post:simple:%d", id)
	ctx := context.Background()

	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var post models.Post
			if err := json.Unmarshal([]byte(cached), &post); err == nil {
				log.Printf("Cache hit for post:simple:%d", id)
				return &post, nil
			}
			log.Printf("Failed to unmarshal cache for post:simple:%d: %v", id, err)
		}
		if err != redis.Nil {
			log.Printf("Redis error for post:simple:%d (attempt %d): %v", id, attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	post, err := s.postRepo.GetPostByIDSimple(id)
	if err != nil {
		log.Printf("Failed to get post:simple:%d: %v", id, err)
		return nil, err
	}

	data, err := json.Marshal(post)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for post:simple:%d: %v", id, err)
		}
	}

	return post, nil
}

func (s *postService) DeletePost(id uint) error {
	_, err := s.postRepo.GetPostByIDSimple(id)
	if err != nil {
		log.Printf("Failed to get post %d: %v", id, err)
		return err
	}

	err = s.postRepo.DeletePost(id)
	if err != nil {
		log.Printf("Failed to delete post %d: %v", id, err)
		return err
	}

	s.invalidateCache(fmt.Sprintf("post:%d", id))
	s.invalidateCache("posts:*")
	s.invalidateCache("tags:*") // Thêm invalidation cho tag cache
	log.Printf("Cache invalidated for posts:* and tags:* due to deleted post %d", id)

	return nil
}

func (s *postService) UpdatePost(id uint, content string, status models.PostStatus, tagNames []string) (*models.Post, error) {
	post, err := s.postRepo.GetPostByID(id)
	if err != nil {
		log.Printf("Failed to get post %d: %v", id, err)
		return nil, err
	}

	if content != "" {
		post.Content = content
	}
	if status != "" {
		if !IsValidStatus(string(status)) {
			return nil, errors.New("invalid status")
		}
		post.Status = status
	}

	if err := s.postRepo.UpdatePost(post, tagNames); err != nil {
		log.Printf("Failed to update post %d: %v", id, err)
		return nil, err
	}

	s.invalidateCache(fmt.Sprintf("post:%d", id))
	s.invalidateCache("posts:*")
	s.invalidateCache("tags:*") // Thêm invalidation cho tag cache

	return post, nil
}

func (s *postService) UpdatePostStatus(id uint, status string) (*models.Post, error) {
	if !IsValidStatus(status) {
		return nil, errors.New("invalid status")
	}

	if err := s.postRepo.UpdatePostStatus(id, status); err != nil {
		log.Printf("Failed to update post status %d: %v", id, err)
		return nil, err
	}

	post, err := s.postRepo.GetPostByIDSimple(id)
	if err != nil {
		log.Printf("Failed to get updated post %d: %v", id, err)
		return nil, err
	}

	s.invalidateCache(fmt.Sprintf("post:%d", id))
	s.invalidateCache("posts:*")
	s.invalidateCache("tags:*") // Thêm invalidation cho tag cache

	return post, nil
}

func (s *postService) ListPosts(filters map[string]interface{}) ([]models.Post, int, error) {
	cacheKey := utils.GenerateCacheKey("posts:all", 0, filters)
	ctx := context.Background()

	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedData struct {
				Posts []models.Post
				Total int
			}
			if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
				log.Printf("Cache hit for %s", cacheKey)
				return cachedData.Posts, cachedData.Total, nil
			}
			log.Printf("Failed to unmarshal cache for %s: %v", cacheKey, err)
		}
		if err != redis.Nil {
			log.Printf("Redis error for %s (attempt %d): %v", cacheKey, attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	posts, total, err := s.postRepo.List(filters)
	if err != nil {
		log.Printf("Failed to list posts: %v", err)
		return nil, 0, err
	}

	cacheData := struct {
		Posts []models.Post
		Total int
	}{Posts: posts, Total: total}
	data, err := json.Marshal(cacheData)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("Cache set for %s with %d posts", cacheKey, len(posts))
		}
	} else {
		log.Printf("Failed to marshal posts for cache: %v", err)
	}

	return posts, total, nil
}

func (s *postService) GetAllPosts(filters map[string]interface{}) ([]models.Post, int, error) {
	cacheKey := utils.GenerateCacheKey("posts:all", 0, filters)
	ctx := context.Background()

	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedData struct {
				Posts []models.Post
				Total int
			}
			if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
				log.Printf("Cache hit for %s", cacheKey)
				return cachedData.Posts, cachedData.Total, nil
			}
			log.Printf("Failed to unmarshal cache for %s: %v", cacheKey, err)
		}
		if err != redis.Nil {
			log.Printf("Redis error for %s (attempt %d): %v", cacheKey, attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	posts, total, err := s.postRepo.GetAllPosts(filters)
	if err != nil {
		log.Printf("Failed to get all posts: %v", err)
		return nil, 0, err
	}

	cacheData := struct {
		Posts []models.Post
		Total int
	}{Posts: posts, Total: total}
	data, err := json.Marshal(cacheData)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("Cache set for %s with %d posts", cacheKey, len(posts))
		}
	} else {
		log.Printf("Failed to marshal posts for cache: %v", err)
	}

	return posts, total, nil
}

func (s *postService) invalidateCache(pattern string) {
	ctx := context.Background()
	var keys []string
	var cursor uint64
	for {
		batch, nextCursor, err := s.redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("Failed to scan cache keys for pattern %s: %v", pattern, err)
			return
		}
		keys = append(keys, batch...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	if len(keys) == 0 {
		log.Printf("No cache keys found for pattern %s", pattern)
		return
	}
	if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
		log.Printf("Failed to delete cache keys %v: %v", keys, err)
	} else {
		log.Printf("Deleted cache keys %v", keys)
	}
}
