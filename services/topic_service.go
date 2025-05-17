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

type TopicService interface {
	CreateTopic(name, description string, createdBy uint) (*models.Topic, error)
	ProposeTopic(name, description string, userID uint) (*models.Topic, error)
	GetTopicByID(id uint) (*models.Topic, error)
	UpdateTopic(id uint, name, description string) (*models.Topic, error)
	DeleteTopic(id uint) error
	ListTopics(filters map[string]interface{}) ([]models.Topic, error)
	ApproveTopic(id uint) (*models.Topic, error)
	RejectTopic(id uint) (*models.Topic, error)
	AddQuestionToTopic(questionID, topicID uint) error
	RemoveQuestionFromTopic(questionID, topicID uint) error
}

type topicService struct {
	topicRepo   repositories.TopicRepository
	redisClient *redis.Client
}

func NewTopicService(tRepo repositories.TopicRepository, redisClient *redis.Client) TopicService {
	return &topicService{topicRepo: tRepo, redisClient: redisClient}
}

func (s *topicService) CreateTopic(name, description string, createdBy uint) (*models.Topic, error) {
	if name == "" {
		return nil, fmt.Errorf("topic name is required")
	}

	// Kiểm tra trùng tên
	existingTopic, err := s.topicRepo.GetTopicByName(name)
	if err == nil && existingTopic != nil {
		return nil, fmt.Errorf("topic with name %s already exists", name)
	}

	topic := &models.Topic{
		Name:        name,
		Description: description,
		Status:      models.TopicStatusApproved, // Tạo tự động thì auto duyệt
		CreatedBy:   createdBy,
	}

	if err := s.topicRepo.CreateTopic(topic); err != nil {
		log.Printf("Failed to create topic: %v", err)
		return nil, err
	}

	// Xóa cache
	s.invalidateCache("topics:*")

	return topic, nil
}

func (s *topicService) ProposeTopic(name, description string, userID uint) (*models.Topic, error) {
	if name == "" {
		return nil, fmt.Errorf("topic name is required")
	}

	// Kiểm tra trùng tên
	existingTopic, err := s.topicRepo.GetTopicByName(name)
	if err == nil && existingTopic != nil {
		return nil, fmt.Errorf("topic with name %s already exists", name)
	}

	topic := &models.Topic{
		Name:        name,
		Description: description,
		Status:      models.TopicStatusPending,
		CreatedBy:   userID,
	}

	if err := s.topicRepo.CreateTopic(topic); err != nil {
		log.Printf("Failed to propose topic: %v", err)
		return nil, err
	}

	// Xóa cache
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
		// Kiểm tra trùng tên
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

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("topic:%d", id))
	s.invalidateCache("topics:*")

	return topic, nil
}

func (s *topicService) DeleteTopic(id uint) error {
	err := s.topicRepo.DeleteTopic(id)
	if err != nil {
		log.Printf("Failed to delete topic %d: %v", id, err)
		return err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("topic:%d", id))
	s.invalidateCache("topics:*")

	return nil
}

func (s *topicService) ListTopics(filters map[string]interface{}) ([]models.Topic, error) {
	cacheKey := utils.GenerateCacheKey("topics", 0, filters)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var topics []models.Topic
		if err := json.Unmarshal([]byte(cached), &topics); err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return topics, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	topics, err := s.topicRepo.ListTopics(filters)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(topics)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}

	return topics, nil
}

func (s *topicService) ApproveTopic(id uint) (*models.Topic, error) {
	topic, err := s.topicRepo.GetTopicByID(id)
	if err != nil {
		return nil, err
	}

	if topic.Status != models.TopicStatusPending {
		return nil, fmt.Errorf("topic is not pending")
	}

	topic.Status = models.TopicStatusApproved

	if err := s.topicRepo.UpdateTopic(topic); err != nil {
		log.Printf("Failed to approve topic %d: %v", id, err)
		return nil, err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("topic:%d", id))
	s.invalidateCache("topics:*")

	return topic, nil
}

func (s *topicService) RejectTopic(id uint) (*models.Topic, error) {
	topic, err := s.topicRepo.GetTopicByID(id)
	if err != nil {
		return nil, err
	}

	if topic.Status != models.TopicStatusPending {
		return nil, fmt.Errorf("topic is not pending")
	}

	topic.Status = models.TopicStatusRejected

	if err := s.topicRepo.UpdateTopic(topic); err != nil {
		log.Printf("Failed to reject topic %d: %v", id, err)
		return nil, err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("topic:%d", id))
	s.invalidateCache("topics:*")

	return topic, nil
}

func (s *topicService) AddQuestionToTopic(questionID, topicID uint) error {
	err := s.topicRepo.AddQuestionToTopic(questionID, topicID)
	if err != nil {
		log.Printf("Failed to add question %d to topic %d: %v", questionID, topicID, err)
		return err
	}

	// Xóa cache
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

	// Xóa cache
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
