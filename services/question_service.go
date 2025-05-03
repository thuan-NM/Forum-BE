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

type QuestionService interface {
	CreateQuestion(title string, userID uint) (*models.Question, error)
	GetQuestionByID(id uint) (*models.Question, error)
	UpdateQuestion(id uint, title string) (*models.Question, error)
	DeleteQuestion(id uint) error
	ListQuestions(filters map[string]interface{}) ([]models.Question, error)
	ApproveQuestion(id uint) (*models.Question, error)
	RejectQuestion(id uint) (*models.Question, error)
}

type questionService struct {
	questionRepo repositories.QuestionRepository
	redisClient  *redis.Client
}

func NewQuestionService(qRepo repositories.QuestionRepository, redisClient *redis.Client) QuestionService {
	return &questionService{questionRepo: qRepo, redisClient: redisClient}
}

func (s *questionService) CreateQuestion(title string, userID uint) (*models.Question, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	question := &models.Question{
		Title:  title,
		UserID: userID,
		Status: models.StatusPending,
	}

	if err := s.questionRepo.CreateQuestion(question); err != nil {
		log.Printf("Failed to create question: %v", err)
		return nil, err
	}

	// Xóa cache
	s.invalidateCache("questions:*")

	return question, nil
}

func (s *questionService) GetQuestionByID(id uint) (*models.Question, error) {
	cacheKey := fmt.Sprintf("question:%d", id)

	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var question models.Question
		if err := json.Unmarshal([]byte(cached), &question); err == nil {
			log.Printf("Cache hit for question:%d", id)
			return &question, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for question:%d: %v", id, err)
	}

	question, err := s.questionRepo.GetQuestionByID(id)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(question)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for question:%d: %v", id, err)
		}
	}

	return question, nil
}

func (s *questionService) UpdateQuestion(id uint, title string) (*models.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(id)
	if err != nil {
		return nil, err
	}

	if title != "" {
		question.Title = title
	}

	if err := s.questionRepo.UpdateQuestion(question); err != nil {
		log.Printf("Failed to update question %d: %v", id, err)
		return nil, err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("question:%d", id))
	s.invalidateCache("questions:*")

	return question, nil
}

func (s *questionService) DeleteQuestion(id uint) error {
	err := s.questionRepo.DeleteQuestion(id)
	if err != nil {
		log.Printf("Failed to delete question %d: %v", id, err)
		return err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("question:%d", id))
	s.invalidateCache("questions:*")

	return nil
}

func (s *questionService) ListQuestions(filters map[string]interface{}) ([]models.Question, error) {
	cacheKey := utils.GenerateCacheKey("questions", 0, filters)

	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var questions []models.Question
		if err := json.Unmarshal([]byte(cached), &questions); err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return questions, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	questions, err := s.questionRepo.ListQuestions(filters)
	if err != nil {
		return nil, err
	}

	if tagID, ok := filters["tag_id"]; ok {
		var filtered []models.Question
		for _, q := range questions {
			for _, t := range q.Tags {
				if t.ID == tagID.(uint) {
					filtered = append(filtered, q)
					break
				}
			}
		}
		questions = filtered
	}

	if userIDRaw, ok := filters["user_id"]; ok {
		userID := userIDRaw.(uint)
		passedIDs, err := s.questionRepo.GetPassedQuestionIDs(userID)
		if err != nil {
			return nil, err
		}

		passedMap := make(map[uint]bool)
		for _, id := range passedIDs {
			passedMap[id] = true
		}

		var visibleQuestions []models.Question
		for _, q := range questions {
			if !passedMap[q.ID] {
				visibleQuestions = append(visibleQuestions, q)
			}
		}
		questions = visibleQuestions
	}

	data, err := json.Marshal(questions)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}

	return questions, nil
}

func (s *questionService) ApproveQuestion(id uint) (*models.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(id)
	if err != nil {
		return nil, err
	}

	if question.Status != models.StatusPending {
		return nil, fmt.Errorf("question is not pending")
	}

	question.Status = models.StatusApproved

	if err := s.questionRepo.UpdateQuestion(question); err != nil {
		log.Printf("Failed to approve question %d: %v", id, err)
		return nil, err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("question:%d", id))
	s.invalidateCache("questions:*")

	return question, nil
}

func (s *questionService) RejectQuestion(id uint) (*models.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(id)
	if err != nil {
		return nil, err
	}

	if question.Status != models.StatusPending {
		return nil, fmt.Errorf("question is not pending")
	}

	question.Status = models.StatusRejected

	if err := s.questionRepo.UpdateQuestion(question); err != nil {
		log.Printf("Failed to reject question %d: %v", id, err)
		return nil, err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("question:%d", id))
	s.invalidateCache("questions:*")

	return question, nil
}

func (s *questionService) invalidateCache(pattern string) {
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
