package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"time"
)

type FollowService interface {
	FollowTopic(userID, topicID uint) error
	UnfollowTopic(userID, topicID uint) error
	FollowQuestion(userID, questionID uint) error
	UnfollowQuestion(userID, questionID uint) error
	FollowUser(userID, followedUserID uint) error
	UnfollowUser(userID, followedUserID uint) error
	GetTopicFollows(topicID uint) ([]models.TopicFollow, error)
	GetQuestionFollows(questionID uint) ([]models.QuestionFollow, error)
	GetUserFollows(userID uint) ([]models.UserFollow, error)
	GetFollowedTopics(userID uint) ([]models.Topic, error)
	GetQuestionFollowStatus(userID uint, questionID uint) (bool, error)
}

type followService struct {
	topicFollowRepo    repositories.TopicFollowRepository
	questionFollowRepo repositories.QuestionFollowRepository
	userFollowRepo     repositories.UserFollowRepository
	redisClient        *redis.Client
	db                 *gorm.DB
}

func NewFollowService(tRepo repositories.TopicFollowRepository, qRepo repositories.QuestionFollowRepository, uRepo repositories.UserFollowRepository, redisClient *redis.Client, db *gorm.DB) FollowService {
	return &followService{
		topicFollowRepo:    tRepo,
		questionFollowRepo: qRepo,
		userFollowRepo:     uRepo,
		redisClient:        redisClient,
		db:                 db,
	}
}

func (s *followService) FollowTopic(userID, topicID uint) error {
	var topic models.Topic
	if err := s.db.Model(&models.Topic{}).Where("id = ?", topicID).First(&topic).Error; err != nil {
		return fmt.Errorf("topic with ID %d does not exist", topicID)
	}

	follows, err := s.topicFollowRepo.GetFollowsByTopic(topicID)
	if err != nil {
		return err
	}
	for _, f := range follows {
		if f.UserID == userID {
			return fmt.Errorf("user already follows this topic")
		}
	}

	follow := &models.TopicFollow{
		TopicID:   topicID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	err = s.topicFollowRepo.CreateFollow(follow)
	if err != nil {
		return err
	}

	s.db.Model(&topic).Update("followers_count", gorm.Expr("followers_count + 1"))
	s.invalidateCache(fmt.Sprintf("topic:%d", topicID))
	s.invalidateCache("topics:*")

	return nil
}

func (s *followService) UnfollowTopic(userID, topicID uint) error {
	var topic models.Topic
	if err := s.db.Model(&models.Topic{}).Where("id = ?", topicID).First(&topic).Error; err != nil {
		return fmt.Errorf("topic with ID %d does not exist", topicID)
	}

	err := s.topicFollowRepo.DeleteFollow(userID, topicID)
	if err != nil {
		return err
	}

	if topic.FollowersCount > 0 {
		s.db.Model(&topic).Update("followers_count", gorm.Expr("followers_count - 1"))
	}
	s.invalidateCache(fmt.Sprintf("topic:%d", topicID))
	s.invalidateCache("topics:*")

	return nil
}

func (s *followService) FollowQuestion(userID, questionID uint) error {
	var question models.Question
	if err := s.db.Model(&models.Question{}).Where("id = ?", questionID).First(&question).Error; err != nil {
		return fmt.Errorf("question with ID %d does not exist", questionID)
	}

	follows, err := s.questionFollowRepo.GetFollowsByQuestion(questionID)
	if err != nil {
		return err
	}
	for _, f := range follows {
		if f.UserID == userID {
			return fmt.Errorf("user already follows this question")
		}
	}

	follow := &models.QuestionFollow{
		QuestionID: questionID,
		UserID:     userID,
		CreatedAt:  time.Now(),
	}
	err = s.questionFollowRepo.CreateFollow(follow)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("follows:question:%d:user:%d", questionID, userID)
	s.invalidateCache(cacheKey) // Invalidate cache trạng thái cụ thể
	log.Printf("Invalidated cache for key: %s", cacheKey)
	s.invalidateCache(fmt.Sprintf("question:%d", questionID))
	s.invalidateCache("questions:*")

	return nil
}

func (s *followService) UnfollowQuestion(userID, questionID uint) error {
	var question models.Question
	if err := s.db.Model(&models.Question{}).Where("id = ?", questionID).First(&question).Error; err != nil {
		return fmt.Errorf("question with ID %d does not exist", questionID)
	}

	err := s.questionFollowRepo.DeleteFollow(userID, questionID)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("follows:question:%d:user:%d", questionID, userID)
	s.invalidateCache(cacheKey) // Invalidate cache trạng thái cụ thể
	log.Printf("Invalidated cache for key: %s", cacheKey)
	s.invalidateCache(fmt.Sprintf("question:%d", questionID))
	s.invalidateCache("questions:*")

	return nil
}

func (s *followService) GetQuestionFollowStatus(userID, questionID uint) (bool, error) {
	// Luôn kiểm tra database trước, chỉ dùng cache như một lớp tăng tốc
	cacheKey := fmt.Sprintf("follows:question:%d:user:%d", questionID, userID)
	ctx := context.Background()

	// Kiểm tra cache (tùy chọn, có thể bỏ nếu không cần)
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var isFollowing bool
		if err := json.Unmarshal([]byte(cached), &isFollowing); err == nil {
			log.Printf("Cache hit for follows:question:%d:user:%d", questionID, userID)
			// Kiểm tra lại database để đảm bảo đồng bộ
			dbCheck, dbErr := s.questionFollowRepo.ExistsByQuestionAndUser(questionID, userID)
			if dbErr != nil {
				log.Printf("Database check failed, falling back to cache: %v", dbErr)
				return isFollowing, nil
			}
			if dbCheck != isFollowing {
				log.Printf("Cache outdated, updating cache for key: %s", cacheKey)
				data, _ := json.Marshal(dbCheck)
				s.redisClient.Set(ctx, cacheKey, data, 30*time.Second)
				return dbCheck, nil
			}
			return isFollowing, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for follows:question:%d:user:%d: %v", questionID, userID, err)
	}

	// Lấy từ database nếu không có cache hoặc cache lỗi
	isFollowing, err := s.questionFollowRepo.ExistsByQuestionAndUser(questionID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %v", err)
	}

	// Lưu vào cache với TTL ngắn
	data, err := json.Marshal(isFollowing)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 30*time.Second).Err(); err != nil {
			log.Printf("Failed to set cache for follows:question:%d:user:%d: %v", questionID, userID, err)
		} else {
			log.Printf("Cache set for key: %s with value: %v", cacheKey, isFollowing)
		}
	}

	return isFollowing, nil
}
func (s *followService) FollowUser(userID, followedUserID uint) error {
	var user models.User
	if err := s.db.Model(&models.User{}).Where("id = ?", followedUserID).First(&user).Error; err != nil {
		return fmt.Errorf("user with ID %d does not exist", followedUserID)
	}

	follows, err := s.userFollowRepo.GetFollowsByUser(followedUserID)
	if err != nil {
		return err
	}
	for _, f := range follows {
		if f.UserID == userID {
			return fmt.Errorf("user already follows this user")
		}
	}

	follow := &models.UserFollow{
		FollowedUserID: followedUserID,
		UserID:         userID,
		CreatedAt:      time.Now(),
	}
	err = s.userFollowRepo.CreateFollow(follow)
	if err != nil {
		return err
	}

	s.db.Model(&models.User{}).Where("id = ?", userID).Update("following_count", gorm.Expr("following_count + 1"))
	s.db.Model(&models.User{}).Where("id = ?", followedUserID).Update("followers_count", gorm.Expr("followers_count + 1"))

	s.invalidateCache(fmt.Sprintf("user:%d", userID))
	s.invalidateCache(fmt.Sprintf("user:%d", followedUserID))
	s.invalidateCache("users:*")

	return nil
}

func (s *followService) UnfollowUser(userID, followedUserID uint) error {
	var user models.User
	if err := s.db.Model(&models.User{}).Where("id = ?", followedUserID).First(&user).Error; err != nil {
		return fmt.Errorf("user with ID %d does not exist", followedUserID)
	}

	err := s.userFollowRepo.DeleteFollow(userID, followedUserID)
	if err != nil {
		return err
	}

	if user.FollowersCount > 0 {
		s.db.Model(&models.User{}).Where("id = ?", followedUserID).Update("followers_count", gorm.Expr("followers_count - 1"))
	}
	if s.db.Model(&models.User{}).Where("id = ?", userID).Select("following_count").Row().Scan(&user.FollowingCount); user.FollowingCount > 0 {
		s.db.Model(&models.User{}).Where("id = ?", userID).Update("following_count", gorm.Expr("following_count - 1"))
	}

	s.invalidateCache(fmt.Sprintf("user:%d", userID))
	s.invalidateCache(fmt.Sprintf("user:%d", followedUserID))
	s.invalidateCache("users:*")

	return nil
}

func (s *followService) GetTopicFollows(topicID uint) ([]models.TopicFollow, error) {
	cacheKey := fmt.Sprintf("follows:topic:%d", topicID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var follows []models.TopicFollow
		if err := json.Unmarshal([]byte(cached), &follows); err == nil {
			log.Printf("Cache hit for follows:topic:%d", topicID)
			return follows, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for follows:topic:%d: %v", topicID, err)
	}

	follows, err := s.topicFollowRepo.GetFollowsByTopic(topicID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(follows)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for follows:topic:%d: %v", topicID, err)
		}
	}

	return follows, nil
}

func (s *followService) GetQuestionFollows(questionID uint) ([]models.QuestionFollow, error) {
	cacheKey := fmt.Sprintf("follows:question:%d", questionID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var follows []models.QuestionFollow
		if err := json.Unmarshal([]byte(cached), &follows); err == nil {
			log.Printf("Cache hit for follows:question:%d", questionID)
			return follows, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for follows:question:%d: %v", questionID, err)
	}

	follows, err := s.questionFollowRepo.GetFollowsByQuestion(questionID)
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

func (s *followService) GetUserFollows(userID uint) ([]models.UserFollow, error) {
	cacheKey := fmt.Sprintf("follows:user:%d", userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var follows []models.UserFollow
		if err := json.Unmarshal([]byte(cached), &follows); err == nil {
			log.Printf("Cache hit for follows:user:%d", userID)
			return follows, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for follows:user:%d: %v", userID, err)
	}

	follows, err := s.userFollowRepo.GetFollowsByUser(userID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(follows)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for follows:user:%d: %v", userID, err)
		}
	}

	return follows, nil
}

func (s *followService) GetFollowedTopics(userID uint) ([]models.Topic, error) {
	cacheKey := fmt.Sprintf("followed_topics:user:%d", userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var topics []models.Topic
		if err := json.Unmarshal([]byte(cached), &topics); err == nil {
			log.Printf("Cache hit for followed_topics:user:%d", userID)
			return topics, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for followed_topics:user:%d: %v", userID, err)
	}

	var topicFollows []models.TopicFollow
	if err := s.db.Where("user_id = ?", userID).Find(&topicFollows).Error; err != nil {
		return nil, err
	}

	var topicIDs []uint
	for _, follow := range topicFollows {
		topicIDs = append(topicIDs, follow.TopicID)
	}

	var topics []models.Topic
	if len(topicIDs) > 0 {
		if err := s.db.Preload("Questions").Find(&topics, topicIDs).Error; err != nil {
			return nil, err
		}
	}

	data, err := json.Marshal(topics)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for followed_topics:user:%d: %v", userID, err)
		}
	}

	return topics, nil
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
