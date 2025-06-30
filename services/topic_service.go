package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"time"
)

type TopicService interface {
	CreateTopic(name, description string) (*models.Topic, error)
	ProposeTopic(name, description string, userID uint) (*models.Topic, error)
	GetTopicByID(id uint) (*models.Topic, error)
	GetTopicByName(name string) (*models.Topic, error)
	UpdateTopic(id uint, name, description string) (*models.Topic, error)
	DeleteTopic(id uint) error
	ListTopics(filters map[string]interface{}) ([]models.Topic, int, error)
	AddQuestionToTopic(questionID, topicID uint) error
	RemoveQuestionFromTopic(questionID, topicID uint) error
}

type topicService struct {
	topicRepo   repositories.TopicRepository
	redisClient *redis.Client
	db          *gorm.DB // Add db field to access GORM for follow cleanup
}

func NewTopicService(tRepo repositories.TopicRepository, redisClient *redis.Client, db *gorm.DB) TopicService {
	return &topicService{topicRepo: tRepo, redisClient: redisClient, db: db}
}

func (s *topicService) GetTopicByName(name string) (*models.Topic, error) {
	cacheKey := fmt.Sprintf("topic:name:%s", name)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var topic models.Topic
		if err := json.Unmarshal([]byte(cached), &topic); err == nil {
			log.Printf("Cache hit for topic:name:%s", name)
			return &topic, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for topic:name:%s: %v", name, err)
	}

	topic, err := s.topicRepo.GetTopicByName(name)
	if err != nil {
		return nil, fmt.Errorf("topic not found")
	}

	data, err := json.Marshal(topic)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for topic:name:%s: %v", name, err)
		}
	}

	return topic, nil
}

func (s *topicService) CreateTopic(name, description string) (*models.Topic, error) {
	if name == "" {
		return nil, fmt.Errorf("topic name is required")
	}

	existingTopic, err := s.topicRepo.GetTopicByName(name)
	if err == nil && existingTopic != nil {
		return nil, fmt.Errorf("topic with name %s already exists", name)
	}

	topic := &models.Topic{
		Name:        name,
		Description: description,
	}

	if err := s.topicRepo.CreateTopic(topic); err != nil {
		log.Printf("Failed to create topic: %v", err)
		return nil, err
	}

	// Invalidate cache
	s.invalidateCache("topics:*")

	return topic, nil
}

func (s *topicService) ProposeTopic(name, description string, userID uint) (*models.Topic, error) {
	if name == "" {
		return nil, fmt.Errorf("topic name is required")
	}

	existingTopic, err := s.topicRepo.GetTopicByName(name)
	if err == nil && existingTopic != nil {
		return nil, fmt.Errorf("topic with name %s already exists", name)
	}

	topic := &models.Topic{
		Name:        name,
		Description: description,
	}

	if err := s.topicRepo.CreateTopic(topic); err != nil {
		log.Printf("Failed to propose topic: %v", err)
		return nil, err
	}

	// Invalidate cache
	s.invalidateCache("topics:*")

	return topic, nil
}

func (s *topicService) GetTopicByID(id uint) (*models.Topic, error) {
	cacheKey := fmt.Sprintf("topic:%d", id)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var topic models.Topic
		if err := json.Unmarshal([]byte(cached), &topic); err == nil {
			log.Printf("Cache hit for topic:%d", id)
			return &topic, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for topic:%d: %v", id, err)
	}

	topic, err := s.topicRepo.GetTopicByID(id)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(topic)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for topic:%d: %v", id, err)
		}
	}

	return topic, nil
}

func (s *topicService) UpdateTopic(id uint, name, description string) (*models.Topic, error) {
	topic, err := s.topicRepo.GetTopicByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		existingTopic, err := s.topicRepo.GetTopicByName(name)
		if err == nil && existingTopic != nil && existingTopic.ID != id {
			return nil, fmt.Errorf("topic with name %s already exists", name)
		}
		topic.Name = name
	}
	if description != "" {
		topic.Description = description
	}

	if err := s.topicRepo.UpdateTopic(topic); err != nil {
		log.Printf("Failed to update topic %d: %v", id, err)
		return nil, err
	}

	// Invalidate cache
	s.invalidateCache(fmt.Sprintf("topic:%d", id))
	s.invalidateCache("topics:*")

	return topic, nil
}

func (s *topicService) DeleteTopic(id uint) error {
	// Delete related Follow records
	err := s.db.Where("followable_id = ? AND followable_type = ?", id, "Topic").Delete(&models.Follow{}).Error
	if err != nil {
		log.Printf("Failed to delete related follows for topic %d: %v", id, err)
		return err
	}

	err = s.topicRepo.DeleteTopic(id)
	if err != nil {
		log.Printf("Failed to delete topic %d: %v", id, err)
		return err
	}

	// Invalidate cache
	s.invalidateCache(fmt.Sprintf("topic:%d", id))
	s.invalidateCache("topics:*")
	s.invalidateCache("follows:Topic:*")

	return nil
}

func (s *topicService) ListTopics(filters map[string]interface{}) ([]models.Topic, int, error) {
	cacheKey := utils.GenerateCacheKey("topics:all", 0, filters)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedData struct {
			Topics []models.Topic
			Total  int
		}
		if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return cachedData.Topics, cachedData.Total, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	topics, total, err := s.topicRepo.ListTopics(filters)
	if err != nil {
		return nil, 0, err
	}

	cacheData := struct {
		Topics []models.Topic
		Total  int
	}{Topics: topics, Total: total}
	data, err := json.Marshal(cacheData)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("Cache set for %s with %d topics", cacheKey, len(topics))
		}
	} else {
		log.Printf("Failed to marshal topics for cache: %v", err)
	}

	return topics, total, nil
}

func (s *topicService) AddQuestionToTopic(questionID, topicID uint) error {
	err := s.topicRepo.AddQuestionToTopic(questionID, topicID)
	if err != nil {
		log.Printf("Failed to add question %d to topic %d: %v", questionID, topicID, err)
		return err
	}

	// Invalidate cache
	s.invalidateCache(fmt.Sprintf("topic:%d", topicID))
	s.invalidateCache("topics:*")
	s.invalidateCache(fmt.Sprintf("question:%d", questionID))
	s.invalidateCache("questions:*")

	return nil
}

func (s *topicService) RemoveQuestionFromTopic(questionID, topicID uint) error {
	err := s.topicRepo.RemoveQuestionFromTopic(questionID, topicID)
	if err != nil {
		log.Printf("Failed to remove question %d from topic %d: %v", questionID, topicID, err)
		return err
	}

	// Invalidate cache
	s.invalidateCache(fmt.Sprintf("topic:%d", topicID))
	s.invalidateCache("topics:*")
	s.invalidateCache(fmt.Sprintf("question:%d", questionID))
	s.invalidateCache("questions:*")

	return nil
}

func (s *topicService) invalidateCache(pattern string) {
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
