package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type ReactionService interface {
	CreateReaction(userID, reactableID uint, reactableType string) (*models.Reaction, error)
	GetReactionByID(id uint) (*models.Reaction, error)
	UpdateReaction(id, userID, reactableID uint, reactableType string) (*models.Reaction, error)
	DeleteReaction(id uint) error
	ListReactions(filters map[string]interface{}) ([]models.Reaction, int, error)
}

type reactionService struct {
	reactionRepo repositories.ReactionRepository
	redisClient  *redis.Client
}

func NewReactionService(repo repositories.ReactionRepository, redisClient *redis.Client) ReactionService {
	return &reactionService{reactionRepo: repo, redisClient: redisClient}
}

func (s *reactionService) CreateReaction(userID, reactableID uint, reactableType string) (*models.Reaction, error) {
	if reactableType != "Post" && reactableType != "Comment" && reactableType != "Answer" {
		return nil, errors.New("invalid reactable_type")
	}

	reaction := &models.Reaction{
		UserID:        userID,
		ReactableID:   reactableID,
		ReactableType: reactableType,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := s.reactionRepo.CreateReaction(reaction); err != nil {
		log.Printf("Failed to create reaction: %v", err)
		return nil, err
	}

	// Invalidate cache (nếu có)
	s.invalidateCache(fmt.Sprintf("reactions:%d", reactableID))
	s.invalidateCache("reactions:*")

	return reaction, nil
}

func (s *reactionService) GetReactionByID(id uint) (*models.Reaction, error) {
	cacheKey := fmt.Sprintf("reaction:%d", id)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var reaction models.Reaction
		if err := json.Unmarshal([]byte(cached), &reaction); err == nil {
			log.Printf("Cache hit for reaction:%d", id)
			return &reaction, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for reaction:%d: %v", id, err)
	}

	reaction, err := s.reactionRepo.GetReactionByID(id)
	if err != nil {
		log.Printf("Failed to get reaction %d: %v", id, err)
		return nil, err
	}

	data, err := json.Marshal(reaction)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for reaction:%d: %v", id, err)
		}
	}

	return reaction, nil
}

func (s *reactionService) UpdateReaction(id, userID, reactableID uint, reactableType string) (*models.Reaction, error) {
	reaction, err := s.reactionRepo.GetReactionByID(id)
	if err != nil {
		log.Printf("Failed to get reaction %d: %v", id, err)
		return nil, err
	}

	reaction.UserID = userID
	reaction.ReactableID = reactableID
	reaction.ReactableType = reactableType
	reaction.UpdatedAt = time.Now()

	if err := s.reactionRepo.UpdateReaction(reaction); err != nil {
		log.Printf("Failed to update reaction %d: %v", id, err)
		return nil, err
	}

	// Invalidate cache
	s.invalidateCache(fmt.Sprintf("reaction:%d", id))
	s.invalidateCache(fmt.Sprintf("reactions:%d", reactableID))
	s.invalidateCache("reactions:*")

	return reaction, nil
}

func (s *reactionService) DeleteReaction(id uint) error {
	reaction, err := s.reactionRepo.GetReactionByID(id)
	if err != nil {
		log.Printf("Failed to get reaction %d: %v", id, err)
		return err
	}

	if err := s.reactionRepo.DeleteReaction(id); err != nil {
		log.Printf("Failed to delete reaction %d: %v", id, err)
		return err
	}

	// Invalidate cache
	s.invalidateCache(fmt.Sprintf("reaction:%d", id))
	s.invalidateCache(fmt.Sprintf("reactions:%d", reaction.ReactableID))
	s.invalidateCache("reactions:*")

	return nil
}

func (s *reactionService) ListReactions(filters map[string]interface{}) ([]models.Reaction, int, error) {
	cacheKey := "reactions:all"
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedData struct {
			Reactions []models.Reaction
			Total     int
		}
		if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return cachedData.Reactions, cachedData.Total, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	reactions, total, err := s.reactionRepo.ListReactions(filters)
	if err != nil {
		log.Printf("Failed to list reactions: %v", err)
		return nil, 0, err
	}

	cacheData := struct {
		Reactions []models.Reaction
		Total     int
	}{Reactions: reactions, Total: total}
	data, err := json.Marshal(cacheData)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}

	return reactions, total, nil
}

func (s *reactionService) invalidateCache(pattern string) {
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Failed to invalidate cache for pattern %s: %v", pattern, err)
		return
	}
	if len(keys) > 0 {
		if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Failed to delete cache keys %v: %v", keys, err)
		} else {
			log.Printf("Deleted cache keys: %v", keys)
		}
	}
}
