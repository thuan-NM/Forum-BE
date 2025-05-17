package services

import (
	"Forum_BE/repositories"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type PassService interface {
	PassQuestion(userID, questionID uint) error
	IsQuestionPassed(userID, questionID uint) (bool, error)
	GetPassedIDs(userID uint) ([]uint, error)
}

type passService struct {
	repo        repositories.PassRepository
	redisClient *redis.Client
}

func NewPassService(r repositories.PassRepository, redisClient *redis.Client) PassService {
	return &passService{repo: r, redisClient: redisClient}
}

func (s *passService) PassQuestion(userID, questionID uint) error {
	// Kiểm tra xem câu hỏi đã được bỏ qua chưa
	isPassed, err := s.IsQuestionPassed(userID, questionID)
	if err != nil {
		log.Printf("Failed to check if question %d is passed for user %d: %v", questionID, userID, err)
		return err
	}
	if isPassed {
		log.Printf("Question %d already passed by user %d", questionID, userID)
		return fmt.Errorf("question already passed")
	}

	err = s.repo.PassQuestion(userID, questionID)
	if err != nil {
		log.Printf("Failed to pass question %d for user %d: %v", questionID, userID, err)
		return err
	}

	// Cập nhật cache
	for attempt := 1; attempt <= 3; attempt++ {
		if err := s.updateCacheAfterPass(userID, questionID); err == nil {
			log.Printf("Cache updated for passed:user:%d", userID)
			break
		}
		log.Printf("Failed to update cache for user %d, question %d (attempt %d): %v", userID, questionID, attempt, err)
		time.Sleep(100 * time.Millisecond)
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("passed:user:%d", userID))
	s.invalidateCache("questions:*")

	return nil
}

func (s *passService) IsQuestionPassed(userID, questionID uint) (bool, error) {
	isPassed, err := s.repo.IsPassed(userID, questionID)
	if err != nil {
		log.Printf("Failed to check if question %d is passed for user %d: %v", questionID, userID, err)
	}
	return isPassed, err
}

func (s *passService) GetPassedIDs(userID uint) ([]uint, error) {
	cacheKey := fmt.Sprintf("passed:user:%d", userID)
	ctx := context.Background()

	var passedIDs []uint
	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(cached), &passedIDs); err == nil {
				log.Printf("Cache hit for passed:user:%d", userID)
				return passedIDs, nil
			}
			log.Printf("Failed to unmarshal cache for user %d: %v", userID, err)
		}
		if err != redis.Nil {
			log.Printf("Redis error for passed:user:%d (attempt %d): %v", userID, attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	passedIDs, err := s.repo.GetPassedQuestionIDs(userID)
	if err != nil {
		log.Printf("Failed to get passed questions for user %d: %v", userID, err)
		return nil, err
	}

	data, err := json.Marshal(passedIDs)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for passed:user:%d: %v", userID, err)
		}
	}

	return passedIDs, nil
}

func (s *passService) updateCacheAfterPass(userID, questionID uint) error {
	cacheKey := fmt.Sprintf("passed:user:%d", userID)
	ctx := context.Background()

	var passedIDs []uint
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cached), &passedIDs); err != nil {
			log.Printf("Failed to unmarshal cache for user %d: %v", userID, err)
			passedIDs = []uint{questionID}
		} else {
			passedIDs = append([]uint{questionID}, passedIDs...)
		}
	} else if err == redis.Nil {
		passedIDs = []uint{questionID}
	} else {
		log.Printf("Redis error for user %d: %v", userID, err)
		return err
	}

	data, err := json.Marshal(passedIDs)
	if err != nil {
		log.Printf("Failed to marshal cache for user %d: %v", userID, err)
		return err
	}

	if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
		log.Printf("Failed to set cache for user %d: %v", userID, err)
		return err
	}

	return nil
}

func (s *passService) invalidateCache(pattern string) {
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
