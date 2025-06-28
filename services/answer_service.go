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
	CreateAnswer(content string, userID uint, questionID uint, tagNames []string) (*models.Answer, error)
	GetAnswerByID(id uint) (*models.Answer, error)
	UpdateAnswer(id uint, content string) (*models.Answer, error)
	DeleteAnswer(id uint) error
	ListAnswers(filters map[string]interface{}) ([]models.Answer, int, error)
	GetAllAnswers(filters map[string]interface{}) ([]models.Answer, int, error)
	UpdateAnswerStatus(id uint, status string) (*models.Answer, error)
	AcceptAnswer(id uint, userID uint) (*models.Answer, error)
}

type answerService struct {
	answerRepo      repositories.AnswerRepository
	questionRepo    repositories.QuestionRepository
	questionService QuestionService
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

func (s *answerService) GetAllAnswers(filters map[string]interface{}) ([]models.Answer, int, error) {
	cacheKey := utils.GenerateCacheKey("answers:all", 0, filters)
	ctx := context.Background()

	var answers []models.Answer
	answers, total, err := s.answerRepo.GetAllAnswers(filters)
	if err != nil {
		log.Printf("Failed to get all answers: %v", err)
		return nil, 0, err
	}

	data, err := json.Marshal(answers)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("Cache set for %s with %d answers", cacheKey, len(answers))
		}
	} else {
		log.Printf("Failed to marshal answers for cache: %v", err)
	}

	return answers, total, nil
}
func (s *answerService) CreateAnswer(content string, userID uint, questionID uint, tagNames []string) (*models.Answer, error) {
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

	if err := s.answerRepo.CreateAnswer(answer, tagNames); err != nil {
		log.Printf("Failed to create answer for question %d: %v", questionID, err)
		return nil, err
	}

	for attempt := 1; attempt <= 3; attempt++ {
		s.invalidateCache(fmt.Sprintf("answers:question:%d:*", questionID))
		log.Printf("Cache invalidated for answers:question:%d (attempt %d)", questionID, attempt)
		if attempt == 3 {
			log.Printf("Max retries reached for invalidating answers cache of question %d", questionID)
		}
		time.Sleep(100 * time.Millisecond)
	}

	s.invalidateCache("questions:*")
	s.invalidateCache("tags:*") // Thêm invalidation cho tag cache
	log.Printf("Cache invalidated for questions:* and tags:* due to new answer for question %d", questionID)

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

	s.invalidateCache(fmt.Sprintf("answer:%d", id))
	s.invalidateCache(fmt.Sprintf("answers:question:%d:*", answer.QuestionID))

	return answer, nil
}

func (s *answerService) DeleteAnswer(id uint) error {
	answer, err := s.answerRepo.GetAnswerByIDSimple(id)
	if err != nil {
		log.Printf("Failed to get answer %d: %v", id, err)
		return err
	}

	err = s.answerRepo.DeleteAnswer(id)
	if err != nil {
		log.Printf("Failed to delete answer %d: %v", id, err)
		return err
	}

	s.invalidateCache(fmt.Sprintf("answer:%d", id))
	s.invalidateCache(fmt.Sprintf("answers:question:%d:*", answer.QuestionID))
	s.invalidateCache("answers:all:*")
	s.invalidateCache("questions:*")
	s.invalidateCache("tags:*") // Thêm invalidation cho tag cache
	log.Printf("Cache invalidated for questions:* and tags:* due to deleted answer for question %d", answer.QuestionID)

	return nil
}

func (s *answerService) ListAnswers(filters map[string]interface{}) ([]models.Answer, int, error) {
	questionID, ok := filters["question_id"].(uint)
	if !ok {
		return nil, 0, errors.New("question_id is required")
	}

	limit, _ := filters["limit"].(int)
	page, _ := filters["page"].(int)
	if limit == 0 {
		limit = 10 // Default limit
	}
	if page == 0 {
		page = 1 // Default page
	}

	cacheKey := utils.GenerateCacheKey("answers:question", questionID, map[string]interface{}{
		"question_id": questionID,
		"user_id":     filters["user_id"],
		"search":      filters["content LIKE ?"],
		"limit":       limit,
		"page":        page,
	})
	ctx := context.Background()

	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedData struct {
				Answers []models.Answer
				Total   int
			}
			if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
				log.Printf("Cache hit for answers:question:%d", questionID)
				return cachedData.Answers, cachedData.Total, nil
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

	answers, total, err := s.answerRepo.ListAnswers(filters)
	if err != nil {
		log.Printf("Failed to list answers for question %d: %v", questionID, err)
		return nil, 0, err
	}

	cacheData := struct {
		Answers []models.Answer
		Total   int
	}{Answers: answers, Total: total}
	data, err := json.Marshal(cacheData)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for answers:question:%d: %v", questionID, err)
		}
	}

	return answers, total, nil
}

func (s *answerService) invalidateCache(pattern string) {
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

func IsValidStatus(status string) bool {
	return status == "approved" || status == "pending" || status == "rejected"
}

func (s *answerService) UpdateAnswerStatus(id uint, status string) (*models.Answer, error) {
	if !IsValidStatus(status) {
		return nil, errors.New("invalid status")
	}

	if err := s.answerRepo.UpdateAnswerStatus(id, status); err != nil {
		log.Printf("Failed to update answer status %d: %v", id, err)
		return nil, err
	}

	answer, err := s.answerRepo.GetAnswerByIDSimple(id)
	if err != nil {
		log.Printf("Failed to get updated answer %d: %v", id, err)
		return nil, err
	}

	s.invalidateCache(fmt.Sprintf("answer:%d", id))
	s.invalidateCache(fmt.Sprintf("answers:question:%d:*", answer.QuestionID))
	s.invalidateCache("tags:*") // Thêm invalidation cho tag cache

	return answer, nil
}

func (s *answerService) AcceptAnswer(id uint, userID uint) (*models.Answer, error) {
	answer, err := s.answerRepo.GetAnswerByIDSimple(id)
	if err != nil {
		log.Printf("Failed to get answer %d: %v", id, err)
		return nil, err
	}

	question, err := s.questionRepo.GetQuestionByID(answer.QuestionID)
	if err != nil {
		log.Printf("Failed to get question %d: %v", answer.QuestionID, err)
		return nil, errors.New("question not found")
	}
	if question.UserID != userID {
		log.Printf("User %d is not authorized to accept answer %d", userID, id)
		return nil, errors.New("only question owner can accept answer")
	}

	if answer.Status != "approved" {
		log.Printf("Cannot accept answer %d: status is %s", id, answer.Status)
		return nil, errors.New("only approved answers can be accepted")
	}

	answer.Accepted = true
	if err := s.answerRepo.UpdateAnswer(answer); err != nil {
		log.Printf("Failed to accept answer %d: %v", id, err)
		return nil, err
	}

	s.invalidateCache(fmt.Sprintf("answer:%d", id))
	s.invalidateCache(fmt.Sprintf("answers:question:%d:*", answer.QuestionID))
	s.invalidateCache("questions:*")
	s.invalidateCache("tags:*") // Thêm invalidation cho tag cache
	log.Printf("Answer %d accepted successfully for question %d", id, answer.QuestionID)
	return answer, nil
}
