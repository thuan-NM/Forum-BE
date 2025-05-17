package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type FollowService interface {
	FollowQuestion(userID, questionID uint) error
	UnfollowQuestion(userID, questionID uint) error
	GetFollowsByQuestionID(questionID uint) ([]models.Follow, error)
}

type followService struct {
	followRepo  repositories.FollowRepository
	redisClient *redis.Client
}

func NewFollowService(fRepo repositories.FollowRepository, redisClient *redis.Client) FollowService {
	return &followService{followRepo: fRepo, redisClient: redisClient}
}

func (s *followService) FollowQuestion(userID, questionID uint) error {
	follows, err := s.followRepo.GetFollowsByQuestionID(questionID)
	if err != nil {
		return err
	}
	for _, f := range follows {
		if f.UserID == userID {
			return fmt.Errorf("user already follows this question")
		}
	}

	follow := &models.Follow{
		UserID:     userID,
		QuestionID: questionID,
	}
	err = s.followRepo.CreateFollow(follow)
	if err != nil {
		return err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("follows:question:%d", questionID))
	s.invalidateCache("questions:*")

	return nil
}

func (s *followService) UnfollowQuestion(userID, questionID uint) error {
	err := s.followRepo.DeleteFollow(userID, questionID)
	if err != nil {
		return err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("follows:question:%d", questionID))
	s.invalidateCache("questions:*")

	return nil
}

func (s *followService) GetFollowsByQuestionID(questionID uint) ([]models.Follow, error) {
	cacheKey := fmt.Sprintf("follows:question:%d", questionID)

	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var follows []models.Follow
		if err := json.Unmarshal([]byte(cached), &follows); err == nil {
			log.Printf("Cache hit for follows:question:%d", questionID)
			return follows, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for follows:question:%d: %v", questionID, err)
	}

	follows, err := s.followRepo.GetFollowsByQuestionID(questionID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(follows)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for follows:question:%d: %v", questionID, err)
		}
	}

	return follows, nil
}

func (s *followService) invalidateCache(pattern string) {
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
