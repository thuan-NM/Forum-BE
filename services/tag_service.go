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

type TagService interface {
	CreateTag(name, description string) (*models.Tag, error)
	GetTagByID(id uint) (*models.Tag, error)
	GetTagByName(name string) (*models.Tag, error)
	UpdateTag(id uint, name, description string) (*models.Tag, error)
	DeleteTag(id uint) error
	ListTags(filters map[string]interface{}) ([]models.Tag, int, error)
	GetTagsByPostID(postID uint) ([]models.Tag, error)
	GetTagsByAnswerID(answerID uint) ([]models.Tag, error)
}

type tagService struct {
	tagRepo     repositories.TagRepository
	redisClient *redis.Client
}

func NewTagService(tRepo repositories.TagRepository, redisClient *redis.Client) TagService {
	return &tagService{tagRepo: tRepo, redisClient: redisClient}
}

func (s *tagService) CreateTag(name, description string) (*models.Tag, error) {
	if name == "" {
		return nil, fmt.Errorf("Tag name is required")
	}

	existingTag, err := s.tagRepo.GetTagByName(name)
	if err == nil && existingTag != nil {
		return nil, fmt.Errorf("Tag with name %s already exists", name)
	}

	tag := &models.Tag{
		Name:        name,
		Description: description,
	}

	if err := s.tagRepo.CreateTag(tag); err != nil {
		log.Printf("Failed to create tag: %v", err)
		return nil, err
	}

	s.invalidateCache("tags:*")

	return tag, nil
}

func (s *tagService) GetTagByID(id uint) (*models.Tag, error) {
	cacheKey := fmt.Sprintf("tag:%d", id)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var tag models.Tag
		if err := json.Unmarshal([]byte(cached), &tag); err == nil {
			log.Printf("Cache hit for tag:%d", id)
			return &tag, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for tag:%d: %v", id, err)
	}

	tag, err := s.tagRepo.GetTagByID(id)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(tag)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for tag:%d: %v", id, err)
		}
	}

	return tag, nil
}

func (s *tagService) GetTagByName(name string) (*models.Tag, error) {
	cacheKey := fmt.Sprintf("tag:name:%s", name)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var tag models.Tag
		if err := json.Unmarshal([]byte(cached), &tag); err == nil {
			log.Printf("Cache hit for tag:name:%s", name)
			return &tag, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for tag:name:%s: %v", name, err)
	}

	tag, err := s.tagRepo.GetTagByName(name)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(tag)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for tag:name:%s: %v", name, err)
		}
	}

	return tag, nil
}

func (s *tagService) UpdateTag(id uint, name, description string) (*models.Tag, error) {
	tag, err := s.tagRepo.GetTagByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		existingTag, err := s.tagRepo.GetTagByName(name)
		if err == nil && existingTag != nil && existingTag.ID != id {
			return nil, fmt.Errorf("Tag with name %s already exists", name)
		}
		tag.Name = name
	}
	if description != "" {
		tag.Description = description
	}

	if err := s.tagRepo.UpdateTag(tag); err != nil {
		log.Printf("Failed to update tag %d: %v", id, err)
		return nil, err
	}

	// Invalidate cache
	s.invalidateCache(fmt.Sprintf("tag:%d", id))
	s.invalidateCache(fmt.Sprintf("tag:name:%s", tag.Name))
	s.invalidateCache("tags:*")

	return tag, nil
}

func (s *tagService) DeleteTag(id uint) error {
	err := s.tagRepo.DeleteTag(id)
	if err != nil {
		log.Printf("Failed to delete tag %d: %v", id, err)
		return err
	}

	s.invalidateCache(fmt.Sprintf("tag:%d", id))
	s.invalidateCache("tags:*")

	return nil
}

func (s *tagService) ListTags(filters map[string]interface{}) ([]models.Tag, int, error) {
	cacheKey := utils.GenerateCacheKey("tags:all", 0, filters)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedData struct {
			Tags  []models.Tag
			Total int
		}
		if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return cachedData.Tags, cachedData.Total, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	tags, total, err := s.tagRepo.ListTags(filters)
	if err != nil {
		return nil, 0, err
	}

	cacheData := struct {
		Tags  []models.Tag
		Total int
	}{Tags: tags, Total: total}
	data, err := json.Marshal(cacheData)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("Cache set for %s with %d tags", cacheKey, len(tags))
		}
	} else {
		log.Printf("Failed to marshal tags for cache: %v", err)
	}

	return tags, total, nil
}

func (s *tagService) GetTagsByPostID(postID uint) ([]models.Tag, error) {
	cacheKey := fmt.Sprintf("tags:post:%d", postID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var tags []models.Tag
		if err := json.Unmarshal([]byte(cached), &tags); err == nil {
			log.Printf("Cache hit for tags:post:%d", postID)
			return tags, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for tags:post:%d: %v", postID, err)
	}

	tags, err := s.tagRepo.GetTagsByPostID(postID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(tags)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for tags:post:%d: %v", postID, err)
		}
	}

	return tags, nil
}

func (s *tagService) GetTagsByAnswerID(answerID uint) ([]models.Tag, error) {
	cacheKey := fmt.Sprintf("tags:answer:%d", answerID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var tags []models.Tag
		if err := json.Unmarshal([]byte(cached), &tags); err == nil {
			log.Printf("Cache hit for tags:answer:%d", answerID)
			return tags, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for tags:answer:%d: %v", answerID, err)
	}

	tags, err := s.tagRepo.GetTagsByAnswerID(answerID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(tags)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for tags:answer:%d: %v", answerID, err)
		}
	}

	return tags, nil
}

func (s *tagService) invalidateCache(pattern string) {
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
