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

type CacheActivityItem struct {
	Type      models.ActivityType `json:"type"`
	Data      json.RawMessage     `json:"data"`
	CreatedAt time.Time           `json:"created_at"`
}

type ActivityService interface {
	GetRecentActivities(limit int) ([]models.ActivityItem, error)
}

type activityService struct {
	activityRepo repositories.ActivityRepository
	redisClient  *redis.Client
}

func NewActivityService(repo repositories.ActivityRepository, redisClient *redis.Client) ActivityService {
	return &activityService{activityRepo: repo, redisClient: redisClient}
}

func (s *activityService) GetRecentActivities(limit int) ([]models.ActivityItem, error) {
	if limit <= 0 {
		limit = 20 // Default
	}
	cacheKey := fmt.Sprintf("recent_activities:%d", limit)
	ctx := context.Background()

	// Check cache
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cacheItems []CacheActivityItem
		if err := json.Unmarshal([]byte(cached), &cacheItems); err == nil {
			var activities []models.ActivityItem
			for _, ci := range cacheItems {
				var data interface{}
				switch ci.Type {
				case models.ActivityUserCreated:
					var user models.User
					if err := json.Unmarshal(ci.Data, &user); err == nil {
						data = &user
					}
				case models.ActivityPostCreated:
					var post models.Post
					if err := json.Unmarshal(ci.Data, &post); err == nil {
						data = &post
					}
				case models.ActivityCommentCreated:
					var comment models.Comment
					if err := json.Unmarshal(ci.Data, &comment); err == nil {
						data = &comment
					}
				case models.ActivityTopicCreated:
					var topic models.Topic
					if err := json.Unmarshal(ci.Data, &topic); err == nil {
						data = &topic
					}
				}
				if data != nil {
					activities = append(activities, models.ActivityItem{
						Type:      ci.Type,
						Data:      data,
						CreatedAt: ci.CreatedAt,
					})
				}
			}
			if len(activities) > 0 {
				log.Printf("Cache hit for %s", cacheKey)
				return activities, nil
			}
		}
	}

	// Query DB
	activities, err := s.activityRepo.GetRecentActivities(limit)
	if err != nil {
		return nil, err
	}

	// Prepare for cache
	var cacheItems []CacheActivityItem
	for _, act := range activities {
		dataJSON, err := json.Marshal(act.Data)
		if err != nil {
			continue
		}
		cacheItems = append(cacheItems, CacheActivityItem{
			Type:      act.Type,
			Data:      dataJSON,
			CreatedAt: act.CreatedAt,
		})
	}
	if len(cacheItems) > 0 {
		cacheData, err := json.Marshal(cacheItems)
		if err == nil {
			s.redisClient.Set(ctx, cacheKey, cacheData, 5*time.Minute)
		}
	}

	return activities, nil
}
