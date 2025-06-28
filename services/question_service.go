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
	"strings"
	"time"
)

type QuestionService interface {
	CreateQuestion(title string, description string, userID, topicID uint) (*models.Question, error)
	GetQuestionByID(id uint) (*models.Question, error)
	UpdateQuestion(id uint, title string, description string, topicID uint) (*models.Question, error)
	DeleteQuestion(id uint) error
	ListQuestions(filters map[string]interface{}) ([]models.Question, int, error)
	UpdateQuestionStatus(id uint, status string) (*models.Question, error)
	UpdateInteractionStatus(id uint, status models.InteractionStatus, userID uint) (*models.Question, error)
}

type questionService struct {
	questionRepo repositories.QuestionRepository
	topicService TopicService
	redisClient  *redis.Client
}

func NewQuestionService(qRepo repositories.QuestionRepository, tService TopicService, redisClient *redis.Client) QuestionService {
	return &questionService{questionRepo: qRepo, topicService: tService, redisClient: redisClient}
}

func (s *questionService) CreateQuestion(title string, description string, userID, topicID uint) (*models.Question, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	question := &models.Question{
		Title:             title,
		Description:       description,
		UserID:            userID,
		TopicID:           topicID,
		Status:            models.StatusPending,
		InteractionStatus: models.InteractionOpened,
	}

	if err := s.questionRepo.CreateQuestion(question); err != nil {
		log.Printf("Failed to create question: %v", err)
		return nil, err
	}

	if topicID == 0 {
		s.suggestTopicForQuestion(question)
	}

	s.invalidateCache("questions:*")

	return question, nil
}

func (s *questionService) suggestTopicForQuestion(question *models.Question) {
	keywords := strings.Split(strings.ToLower(question.Title), " ")
	for _, keyword := range keywords {
		if len(keyword) < 3 {
			continue
		}
		topic, err := s.topicService.GetTopicByName(keyword)
		if err != nil && err.Error() == "topic not found" {
			topic, err = s.topicService.CreateTopic(keyword, "Auto-generated topic from question")
			if err != nil {
				log.Printf("Failed to suggest topic %s for question %d: %v", keyword, question.ID, err)
				continue
			}
			question.TopicID = topic.ID
			if err := s.questionRepo.UpdateQuestion(question); err != nil {
				log.Printf("Failed to update question %d with topic %d: %v", question.ID, topic.ID, err)
			}
			break
		}
	}
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

func (s *questionService) UpdateQuestion(id uint, title string, description string, topicID uint) (*models.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(id)
	if err != nil {
		return nil, err
	}

	if title != "" {
		question.Title = title
	}
	question.Description = description
	if topicID != 0 {
		question.TopicID = topicID
	}

	if err := s.questionRepo.UpdateQuestion(question); err != nil {
		log.Printf("Failed to update question %d: %v", id, err)
		return nil, err
	}

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

	s.invalidateCache(fmt.Sprintf("question:%d", id))
	s.invalidateCache("questions:*")

	return nil
}

func (s *questionService) ListQuestions(filters map[string]interface{}) ([]models.Question, int, error) {
	cacheKey := utils.GenerateCacheKey("questions:all", 0, filters)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedData struct {
			Questions []models.Question
			Total     int
		}
		if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return cachedData.Questions, cachedData.Total, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	var questions []models.Question
	var total int
	if userID, ok := filters["user_id"].(uint); ok && userID != 0 {
		questions, total, err = s.questionRepo.ListQuestionsExcludingPassed(filters)
	} else {
		questions, total, err = s.questionRepo.ListQuestions(filters)
	}
	if err != nil {
		return nil, 0, err
	}

	cacheData := struct {
		Questions []models.Question
		Total     int
	}{Questions: questions, Total: total}
	data, err := json.Marshal(cacheData)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("Cache set for %s with %d questions", cacheKey, len(questions))
		}
	} else {
		log.Printf("Failed to marshal questions for cache: %v", err)
	}

	return questions, total, nil
}

func (s *questionService) UpdateQuestionStatus(id uint, status string) (*models.Question, error) {
	if status != string(models.StatusApproved) && status != string(models.StatusPending) && status != string(models.StatusRejected) {
		return nil, fmt.Errorf("invalid question status")
	}

	if err := s.questionRepo.UpdateQuestionStatus(id, status); err != nil {
		log.Printf("Failed to update question status %d: %v", id, err)
		return nil, err
	}

	updatedQuestion, err := s.questionRepo.GetQuestionByIDMinimal(id)
	if err != nil {
		log.Printf("Failed to get updated question %d: %v", id, err)
		return nil, err
	}

	s.invalidateCache(fmt.Sprintf("question:%d", id))
	s.invalidateCache("questions:*")

	return updatedQuestion, nil
}

func (s *questionService) UpdateInteractionStatus(id uint, status models.InteractionStatus, userID uint) (*models.Question, error) {
	if status != models.InteractionOpened && status != models.InteractionSolved && status != models.InteractionClosed {
		return nil, fmt.Errorf("trạng thái tương tác không hợp lệ")
	}

	if err := s.questionRepo.UpdateInteractionStatus(id, string(status)); err != nil {
		log.Printf("Failed to update interaction status for question %d: %v", id, err)
		return nil, err
	}

	updatedQuestion, err := s.questionRepo.GetQuestionByIDMinimal(id)
	if err != nil {
		log.Printf("Failed to get updated question %d: %v", id, err)
		return nil, err
	}

	s.invalidateCache(fmt.Sprintf("question:%d", id))
	s.invalidateCache("questions:*")

	return updatedQuestion, nil
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
