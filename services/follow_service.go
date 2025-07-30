package services

import (
	"Forum_BE/models"
	"Forum_BE/notification"
	"Forum_BE/notification"
	"Forum_BE/repositories"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm"
	"log"
	"time"
)

type FollowService interface {
	FollowTopic(userID, topicID uint) error
	UnfollowTopic(userID, topicID uint) error
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
	GetFollowedUsers(userID uint) ([]models.User, error)
	GetFollowingUsers(userID uint) ([]models.User, error)
	GetQuestionFollowStatus(userID uint, questionID uint) (bool, error)
	GetTopicFollowStatus(userID uint, topicID uint) (bool, error)
	GetUserFollowStatus(userID uint, followedUserID uint) (bool, error)
}

type followService struct {
	topicFollowRepo    repositories.TopicFollowRepository
	questionFollowRepo repositories.QuestionFollowRepository
	userFollowRepo     repositories.UserFollowRepository
	userRepo           repositories.UserRepository // Thêm UserRepository
	redisClient        *redis.Client
	db                 *gorm.DB
	novuClient         *notification.NovuClient
	topicFollowRepo    repositories.TopicFollowRepository
	questionFollowRepo repositories.QuestionFollowRepository
	userFollowRepo     repositories.UserFollowRepository
	userRepo           repositories.UserRepository // Thêm UserRepository
	redisClient        *redis.Client
	db                 *gorm.DB
	novuClient         *notification.NovuClient
}

func NewFollowService(tRepo repositories.TopicFollowRepository, qRepo repositories.QuestionFollowRepository, uRepo repositories.UserFollowRepository, userRepo repositories.UserRepository, redisClient *redis.Client, db *gorm.DB, novuClient *notification.NovuClient) FollowService {
	return &followService{
		topicFollowRepo:    tRepo,
		questionFollowRepo: qRepo,
		userFollowRepo:     uRepo,
		userRepo:           userRepo, // Khởi tạo UserRepository
		redisClient:        redisClient,
		db:                 db,
		novuClient:         novuClient,
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
	s.invalidateCache(fmt.Sprintf("followed_topics:user:%d", userID))
	s.invalidateCache(fmt.Sprintf("follows:topic:%d:user:%d", topicID, userID))

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
	s.invalidateCache(fmt.Sprintf("followed_topics:user:%d", userID))
	s.invalidateCache(fmt.Sprintf("follows:topic:%d:user:%d", topicID, userID))

	return nil
func NewFollowService(tRepo repositories.TopicFollowRepository, qRepo repositories.QuestionFollowRepository, uRepo repositories.UserFollowRepository, userRepo repositories.UserRepository, redisClient *redis.Client, db *gorm.DB, novuClient *notification.NovuClient) FollowService {
	return &followService{
		topicFollowRepo:    tRepo,
		questionFollowRepo: qRepo,
		userFollowRepo:     uRepo,
		userRepo:           userRepo, // Khởi tạo UserRepository
		redisClient:        redisClient,
		db:                 db,
		novuClient:         novuClient,
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
	s.invalidateCache(fmt.Sprintf("followed_topics:user:%d", userID))
	s.invalidateCache(fmt.Sprintf("follows:topic:%d:user:%d", topicID, userID))

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
	s.invalidateCache(fmt.Sprintf("followed_topics:user:%d", userID))
	s.invalidateCache(fmt.Sprintf("follows:topic:%d:user:%d", topicID, userID))

	return nil
}

func (s *followService) FollowQuestion(userID, questionID uint) error {
	var question models.Question
	if err := s.db.Model(&models.Question{}).Where("id = ?", questionID).First(&question).Error; err != nil {
		return fmt.Errorf("question with ID %d does not exist", questionID)
	}

	follows, err := s.questionFollowRepo.GetFollowsByQuestion(questionID)
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
	follow := &models.QuestionFollow{
		QuestionID: questionID,
		UserID:     userID,
		CreatedAt:  time.Now(),
		UserID:     userID,
		CreatedAt:  time.Now(),
	}
	err = s.questionFollowRepo.CreateFollow(follow)
	err = s.questionFollowRepo.CreateFollow(follow)
	if err != nil {
		return err
	}

	// Send notification to the question owner with follower's full name
	if question.UserID != userID {
		follower, err := s.userRepo.GetUserByID(userID)
		if err != nil {
			log.Printf("Failed to get follower info: %v", err)
		} else {
			workflowID := "new-question-follow-notification"
			message := fmt.Sprintf("%s has followed your question: %s", follower.FullName, question.Title)
			if err := s.novuClient.SendNotification(question.UserID, workflowID, message); err != nil {
				log.Printf("Failed to send notification for question follow: %v", err)
			}
		}
	}

	cacheKey := fmt.Sprintf("follows:question:%d:user:%d", questionID, userID)
	s.invalidateCache(cacheKey)
	log.Printf("Invalidated cache for key: %s", cacheKey)
	s.invalidateCache(fmt.Sprintf("question:%d", questionID))
	// Send notification to the question owner with follower's full name
	if question.UserID != userID {
		follower, err := s.userRepo.GetUserByID(userID)
		if err != nil {
			log.Printf("Failed to get follower info: %v", err)
		} else {
			workflowID := "new-question-follow-notification"
			message := fmt.Sprintf("%s has followed your question: %s", follower.FullName, question.Title)
			if err := s.novuClient.SendNotification(question.UserID, workflowID, message); err != nil {
				log.Printf("Failed to send notification for question follow: %v", err)
			}
		}
	}

	cacheKey := fmt.Sprintf("follows:question:%d:user:%d", questionID, userID)
	s.invalidateCache(cacheKey)
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
	var question models.Question
	if err := s.db.Model(&models.Question{}).Where("id = ?", questionID).First(&question).Error; err != nil {
		return fmt.Errorf("question with ID %d does not exist", questionID)
	}

	err := s.questionFollowRepo.DeleteFollow(userID, questionID)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("follows:question:%d:user:%d", questionID, userID)
	s.invalidateCache(cacheKey)
	log.Printf("Invalidated cache for key: %s", cacheKey)
	s.invalidateCache(fmt.Sprintf("question:%d", questionID))
	cacheKey := fmt.Sprintf("follows:question:%d:user:%d", questionID, userID)
	s.invalidateCache(cacheKey)
	log.Printf("Invalidated cache for key: %s", cacheKey)
	s.invalidateCache(fmt.Sprintf("question:%d", questionID))
	s.invalidateCache("questions:*")

	return nil
}

func (s *followService) GetQuestionFollowStatus(userID, questionID uint) (bool, error) {
	cacheKey := fmt.Sprintf("follows:question:%d:user:%d", questionID, userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var isFollowing bool
		if err := json.Unmarshal([]byte(cached), &isFollowing); err == nil {
			log.Printf("Cache hit for follows:question:%d:user:%d", questionID, userID)
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

	isFollowing, err := s.questionFollowRepo.ExistsByQuestionAndUser(questionID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %v", err)
	}

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

func (s *followService) GetTopicFollowStatus(userID, topicID uint) (bool, error) {
	cacheKey := fmt.Sprintf("follows:topic:%d:user:%d", topicID, userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var isFollowing bool
		if err := json.Unmarshal([]byte(cached), &isFollowing); err == nil {
			log.Printf("Cache hit for follows:topic:%d:user:%d", topicID, userID)
			dbCheck, dbErr := s.topicFollowRepo.ExistsByTopicAndUser(topicID, userID)
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
		log.Printf("Redis error for follows:topic:%d:user:%d: %v", topicID, userID, err)
	}

	isFollowing, err := s.topicFollowRepo.ExistsByTopicAndUser(topicID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %v", err)
	}

	data, err := json.Marshal(isFollowing)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 30*time.Second).Err(); err != nil {
			log.Printf("Failed to set cache for follows:topic:%d:user:%d: %v", topicID, userID, err)
		} else {
			log.Printf("Cache set for key: %s with value: %v", cacheKey, isFollowing)
		}
	}

	return isFollowing, nil
}

func (s *followService) GetUserFollowStatus(userID, followedUserID uint) (bool, error) {
	cacheKey := fmt.Sprintf("follows:user:%d:user:%d", followedUserID, userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var isFollowing bool
		if err := json.Unmarshal([]byte(cached), &isFollowing); err == nil {
			log.Printf("Cache hit for follows:user:%d:user:%d", followedUserID, userID)
			dbCheck, dbErr := s.userFollowRepo.ExistsByUserAndFollower(followedUserID, userID)
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
		log.Printf("Redis error for follows:user:%d:user:%d: %v", followedUserID, userID, err)
	}

	isFollowing, err := s.userFollowRepo.ExistsByUserAndFollower(followedUserID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %v", err)
	}

	data, err := json.Marshal(isFollowing)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 30*time.Second).Err(); err != nil {
			log.Printf("Failed to set cache for follows:user:%d:user:%d: %v", followedUserID, userID, err)
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

	// Send notification to the followed user with follower's full name
	follower, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to get follower info: %v", err)
	} else {
		workflowID := "new-user-follow-notification"
		message := fmt.Sprintf("%s has followed you", follower.FullName)
		if err := s.novuClient.SendNotification(followedUserID, workflowID, message); err != nil {
			log.Printf("Failed to send notification for user follow: %v", err)
		}
	}

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
func (s *followService) GetQuestionFollowStatus(userID, questionID uint) (bool, error) {
	cacheKey := fmt.Sprintf("follows:question:%d:user:%d", questionID, userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var isFollowing bool
		if err := json.Unmarshal([]byte(cached), &isFollowing); err == nil {
			log.Printf("Cache hit for follows:question:%d:user:%d", questionID, userID)
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

	isFollowing, err := s.questionFollowRepo.ExistsByQuestionAndUser(questionID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %v", err)
	}

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

func (s *followService) GetTopicFollowStatus(userID, topicID uint) (bool, error) {
	cacheKey := fmt.Sprintf("follows:topic:%d:user:%d", topicID, userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var isFollowing bool
		if err := json.Unmarshal([]byte(cached), &isFollowing); err == nil {
			log.Printf("Cache hit for follows:topic:%d:user:%d", topicID, userID)
			dbCheck, dbErr := s.topicFollowRepo.ExistsByTopicAndUser(topicID, userID)
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
		log.Printf("Redis error for follows:topic:%d:user:%d: %v", topicID, userID, err)
	}

	isFollowing, err := s.topicFollowRepo.ExistsByTopicAndUser(topicID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %v", err)
	}

	data, err := json.Marshal(isFollowing)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 30*time.Second).Err(); err != nil {
			log.Printf("Failed to set cache for follows:topic:%d:user:%d: %v", topicID, userID, err)
		} else {
			log.Printf("Cache set for key: %s with value: %v", cacheKey, isFollowing)
		}
	}

	return isFollowing, nil
}

func (s *followService) GetUserFollowStatus(userID, followedUserID uint) (bool, error) {
	cacheKey := fmt.Sprintf("follows:user:%d:user:%d", followedUserID, userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var isFollowing bool
		if err := json.Unmarshal([]byte(cached), &isFollowing); err == nil {
			log.Printf("Cache hit for follows:user:%d:user:%d", followedUserID, userID)
			dbCheck, dbErr := s.userFollowRepo.ExistsByUserAndFollower(followedUserID, userID)
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
		log.Printf("Redis error for follows:user:%d:user:%d: %v", followedUserID, userID, err)
	}

	isFollowing, err := s.userFollowRepo.ExistsByUserAndFollower(followedUserID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %v", err)
	}

	data, err := json.Marshal(isFollowing)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 30*time.Second).Err(); err != nil {
			log.Printf("Failed to set cache for follows:user:%d:user:%d: %v", followedUserID, userID, err)
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

	// Send notification to the followed user with follower's full name
	follower, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to get follower info: %v", err)
	} else {
		workflowID := "new-user-follow-notification"
		message := fmt.Sprintf("%s has followed you", follower.FullName)
		if err := s.novuClient.SendNotification(followedUserID, workflowID, message); err != nil {
			log.Printf("Failed to send notification for user follow: %v", err)
		}
	}

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
func (s *followService) GetFollowedUsers(userID uint) ([]models.User, error) {
	cacheKey := fmt.Sprintf("followed_users:user:%d", userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var users []models.User
		if err := json.Unmarshal([]byte(cached), &users); err == nil {
			log.Printf("Cache hit for followed_users:user:%d", userID)
			return users, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for followed_users:user:%d: %v", userID, err)
	}

	var userFollows []models.UserFollow
	if err := s.db.Where("followed_user_id = ?", userID).Find(&userFollows).Error; err != nil {
		return nil, err
	}

	var userIDs []uint
	for _, follow := range userFollows {
		userIDs = append(userIDs, follow.UserID)
	}

	var users []models.User
	if len(userIDs) > 0 {
		if err := s.db.Find(&users, userIDs).Error; err != nil {
			return nil, err
		}
	}

	data, err := json.Marshal(users)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for followed_users:user:%d: %v", userID, err)
		}
	}

	return users, nil
}
func (s *followService) GetFollowingUsers(userID uint) ([]models.User, error) {
	cacheKey := fmt.Sprintf("following_users:user:%d", userID)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var users []models.User
		if err := json.Unmarshal([]byte(cached), &users); err == nil {
			log.Printf("Cache hit for followed_users:user:%d", userID)
			return users, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for followed_users:user:%d: %v", userID, err)
	}

	var userFollows []models.UserFollow
	if err := s.db.Where("user_id = ?", userID).Find(&userFollows).Error; err != nil {
		return nil, err
	}

	var userIDs []uint
	for _, follow := range userFollows {
		userIDs = append(userIDs, follow.FollowedUserID)
	}

	var users []models.User
	if len(userIDs) > 0 {
		if err := s.db.Find(&users, userIDs).Error; err != nil {
			return nil, err
		}
	}

	data, err := json.Marshal(users)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for followed_users:user:%d: %v", userID, err)
		}
	}

	return users, nil
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
