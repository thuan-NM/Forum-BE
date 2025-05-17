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

type AnswerService interface {
	CreateAnswer(content string, userID uint, questionID uint) (*models.Answer, error)
	GetAnswerByID(id uint) (*models.Answer, error)
	UpdateAnswer(id uint, content string) (*models.Answer, error)
	DeleteAnswer(id uint) error
	ListAnswers(filters map[string]interface{}) ([]models.Answer, error)
}

type answerService struct {
	answerRepo      repositories.AnswerRepository
	questionRepo    repositories.QuestionRepository
	questionService QuestionService // Thêm để gọi invalidate cache
	redisClient     *redis.Client
}

func NewAnswerService(aRepo repositories.AnswerRepository, qRepo repositories.QuestionRepository, qService QuestionService, redisClient *redis.Client) AnswerService {
	return &answerService{
		answerRepo:      aRepo,
		questionRepo:    qRepo,
		questionService: qService,
		redisClient:     redisClient,
	}
}

func (s *answerService) CreateAnswer(content string, userID uint, questionID uint) (*models.Answer, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}

	question, err := s.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		log.Printf("Failed to get question %d: %v", questionID, err)
		return nil, errors.New("question not found")
	}

	if question.Status != models.StatusApproved {
		log.Printf("Cannot answer question %d: status is %s", questionID, question.Status)
		return nil, errors.New("cannot answer a question that is not approved")
	}

	answer := &models.Answer{
		Content:    content,
		UserID:     userID,
		QuestionID: questionID,
	}

	if err := s.answerRepo.CreateAnswer(answer); err != nil {
		log.Printf("Failed to create answer for question %d: %v", questionID, err)
		return nil, err
	}

	// Xóa cache cho answers
	for attempt := 1; attempt <= 3; attempt++ {
		s.invalidateCache(fmt.Sprintf("answers:question:%d:*", questionID))
		log.Printf("Cache invalidated for answers:question:%d (attempt %d)", questionID, attempt)
		if attempt == 3 {
			log.Printf("Max retries reached for invalidating answers cache of question %d", questionID)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Xóa cache cho questions (gọi qua QuestionService)
	s.invalidateCache("questions:*")
	log.Printf("Cache invalidated for questions:* due to new answer for question %d", questionID)

	log.Printf("Answer %d created successfully for question %d", answer.ID, questionID)
	return answer, nil
}

func (s *answerService) GetAnswerByID(id uint) (*models.Answer, error) {
	cacheKey := fmt.Sprintf("answer:%d", id)
	ctx := context.Background()

	var answer *models.Answer
	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cached), &answer); err == nil {
				log.Printf("Cache hit for answer:%d", id)
				return answer, nil
			}
			log.Printf("Failed to unmarshal cache for answer %d: %v", id, err)
		}
		if err != redis.Nil {
			log.Printf("Redis error for answer:%d (attempt %d): %v", id, attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	answer, err := s.answerRepo.GetAnswerByID(id)
	if err != nil {
		log.Printf("Failed to get answer %d: %v", id, err)
		return nil, err
	}

	data, err := json.Marshal(answer)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for answer:%d: %v", id, err)
		}
	}

	return answer, nil
}

func (s *answerService) UpdateAnswer(id uint, content string) (*models.Answer, error) {
	answer, err := s.answerRepo.GetAnswerByID(id)
	if err != nil {
		log.Printf("Failed to get answer %d: %v", id, err)
		return nil, err
	}

	if content != "" {
		answer.Content = content
	}

	if err := s.answerRepo.UpdateAnswer(answer); err != nil {
		log.Printf("Failed to update answer %d: %v", id, err)
		return nil, err
	}

	// Xóa cache
	for attempt := 1; attempt <= 3; attempt++ {
		s.invalidateCache(fmt.Sprintf("answer:%d", id))
		s.invalidateCache(fmt.Sprintf("answers:question:%d:*", answer.QuestionID))
		log.Printf("Cache cleared for answer:%d (attempt %d)", id, attempt)
		if attempt == 3 {
			log.Printf("Max retries reached for invalidating cache of answer %d", id)
		}
		time.Sleep(100 * time.Millisecond)
	}

	return answer, nil
}

func (s *answerService) DeleteAnswer(id uint) error {
	answer, err := s.answerRepo.GetAnswerByID(id)
	if err != nil {
		log.Printf("Failed to get answer %d: %v", id, err)
		return err
	}

	err = s.answerRepo.DeleteAnswer(id)
	if err != nil {
		log.Printf("Failed to delete answer %d: %v", id, err)
		return err
	}

	// Xóa cache
	for attempt := 1; attempt <= 3; attempt++ {
		s.invalidateCache(fmt.Sprintf("answer:%d", id))
		s.invalidateCache(fmt.Sprintf("answers:question:%d:*", answer.QuestionID))
		log.Printf("Cache cleared for answer:%d (attempt %d)", id, attempt)
		if attempt == 3 {
			log.Printf("Max retries reached for invalidating cache of answer %d", id)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Xóa cache cho questions
	s.invalidateCache("questions:*")
	log.Printf("Cache invalidated for questions:* due to deleted answer for question %d", answer.QuestionID)

	return nil
}

func (s *answerService) ListAnswers(filters map[string]interface{}) ([]models.Answer, error) {
	questionID, ok := filters["question_id"].(uint)
	if !ok {
		return nil, errors.New("question_id is required")
	}

	cacheKey := utils.GenerateCacheKey("answers:question", questionID, filters)
	ctx := context.Background()

	var answers []models.Answer
	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cached), &answers); err == nil {
				log.Printf("Cache hit for answers:question:%d", questionID)
				return answers, nil
			}
			log.Printf("Failed to unmarshal cache for question %d: %v", questionID, err)
		}
		if err != redis.Nil {
			log.Printf("Redis error for answers:question:%d (attempt %d): %v", questionID, attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	answers, err := s.answerRepo.ListAnswers(filters)
	if err != nil {
		log.Printf("Failed to list answers for question %d: %v", questionID, err)
		return nil, err
	}

	data, err := json.Marshal(answers)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for answers:question:%d: %v", questionID, err)
		}
	}

	return answers, nil
}

func (s *answerService) invalidateCache(pattern string) {
	ctx := context.Background()
	var keys []string
	for attempt := 1; attempt <= 3; attempt++ {
		var err error
		keys, err = s.redisClient.Keys(ctx, pattern).Result()
		if err == nil {
			log.Printf("Found %d cache keys to invalidate for pattern %s", len(keys), pattern)
			break
		}
		log.Printf("Failed to get cache keys for pattern %s (attempt %d): %v", pattern, attempt, err)
		if attempt == 3 {
			log.Printf("Max retries reached for invalidating cache of pattern %s", pattern)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	if len(keys) > 0 {
		for attempt := 1; attempt <= 3; attempt++ {
			if err := s.redisClient.Del(ctx, keys...).Err(); err == nil {
				log.Printf("Deleted cache keys %v", keys)
				break
			}
			log.Printf("Failed to delete cache keys %v (attempt %d): %v", keys, attempt)
			if attempt == 3 {
				log.Printf("Max retries reached for deleting cache keys %v", keys)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
