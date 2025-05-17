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

type CommentService interface {
	CreateComment(content string, userID uint, questionID *uint, answerID *uint) (*models.Comment, error)
	GetCommentByID(id uint) (*models.Comment, error)
	UpdateComment(id uint, content string) (*models.Comment, error)
	DeleteComment(id uint) error
	ListComments(filters map[string]interface{}) ([]models.Comment, error)
}

type commentService struct {
	commentRepo  repositories.CommentRepository
	questionRepo repositories.QuestionRepository
	answerRepo   repositories.AnswerRepository
	redisClient  *redis.Client
}

func NewCommentService(cRepo repositories.CommentRepository, qRepo repositories.QuestionRepository, aRepo repositories.AnswerRepository, redisClient *redis.Client) CommentService {
	return &commentService{
		commentRepo:  cRepo,
		questionRepo: qRepo,
		answerRepo:   aRepo,
		redisClient:  redisClient,
	}
}

func (s *commentService) CreateComment(content string, userID uint, questionID *uint, answerID *uint) (*models.Comment, error) {
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	if questionID == nil && answerID == nil {
		return nil, fmt.Errorf("either question_id or answer_id must be provided")
	}

	if questionID != nil {
		question, err := s.questionRepo.GetQuestionByID(*questionID)
		if err != nil {
			return nil, fmt.Errorf("question not found")
		}
		if question.Status != models.StatusApproved {
			return nil, fmt.Errorf("cannot comment on a question that is not approved")
		}
	}

	if answerID != nil {
		_, err := s.answerRepo.GetAnswerByID(*answerID)
		if err != nil {
			return nil, fmt.Errorf("answer not found")
		}
	}

	comment := &models.Comment{
		Content:    content,
		UserID:     userID,
		QuestionID: questionID,
		AnswerID:   answerID,
	}

	if err := s.commentRepo.CreateComment(comment); err != nil {
		return nil, err
	}

	// Cập nhật cache
	if questionID != nil {
		s.updateCacheAfterCreate(comment, *questionID, true)
	}
	if answerID != nil {
		s.updateCacheAfterCreate(comment, *answerID, false)
	}

	return comment, nil
}

func (s *commentService) GetCommentByID(id uint) (*models.Comment, error) {
	cacheKey := fmt.Sprintf("comment:%d", id)

	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var comment models.Comment
		if err := json.Unmarshal([]byte(cached), &comment); err == nil {
			log.Printf("Cache hit for comment:%d", id)
			return &comment, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for comment:%d: %v", id, err)
	}

	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(comment)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for comment:%d: %v", id, err)
		}
	}

	return comment, nil
}

func (s *commentService) UpdateComment(id uint, content string) (*models.Comment, error) {
	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		return nil, err
	}

	if content != "" {
		comment.Content = content
	}

	if err := s.commentRepo.UpdateComment(comment); err != nil {
		return nil, err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("comment:%d", id))
	if comment.QuestionID != nil {
		s.invalidateCache(fmt.Sprintf("comments:question:%d:*", *comment.QuestionID))
	}
	if comment.AnswerID != nil {
		s.invalidateCache(fmt.Sprintf("comments:answer:%d:*", *comment.AnswerID))
	}

	return comment, nil
}

func (s *commentService) DeleteComment(id uint) error {
	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		return err
	}

	err = s.commentRepo.DeleteComment(id)
	if err != nil {
		return err
	}

	// Xóa cache
	s.invalidateCache(fmt.Sprintf("comment:%d", id))
	if comment.QuestionID != nil {
		s.invalidateCache(fmt.Sprintf("comments:question:%d:*", *comment.QuestionID))
	}
	if comment.AnswerID != nil {
		s.invalidateCache(fmt.Sprintf("comments:answer:%d:*", *comment.AnswerID))
	}

	return nil
}

func (s *commentService) ListComments(filters map[string]interface{}) ([]models.Comment, error) {
	var cacheKey string
	//var id uint
	//isQuestion := false

	if questionID, ok := filters["question_id"].(uint); ok {
		cacheKey = utils.GenerateCacheKey("comments:question", questionID, filters)
		//id = questionID
		//isQuestion = true
	} else if answerID, ok := filters["answer_id"].(uint); ok {
		cacheKey = utils.GenerateCacheKey("comments:answer", answerID, filters)
		//id = answerID
	} else {
		return nil, fmt.Errorf("question_id or answer_id is required")
	}

	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var comments []models.Comment
		if err := json.Unmarshal([]byte(cached), &comments); err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return comments, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	comments, err := s.commentRepo.ListComments(filters)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(comments)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}

	return comments, nil
}

func (s *commentService) updateCacheAfterCreate(comment *models.Comment, id uint, isQuestion bool) {
	prefix := "comments:answer"
	if isQuestion {
		prefix = "comments:question"
	}
	cachePattern := fmt.Sprintf("%s:%d:*", prefix, id)
	ctx := context.Background()

	pipe := s.redisClient.Pipeline()
	keys, err := s.redisClient.Keys(ctx, cachePattern).Result()
	if err != nil {
		log.Printf("Failed to get cache keys for pattern %s: %v", cachePattern, err)
		return
	}

	for _, key := range keys {
		cached, err := s.redisClient.Get(ctx, key).Result()
		if err == nil {
			var comments []models.Comment
			if err := json.Unmarshal([]byte(cached), &comments); err == nil {
				comments = append([]models.Comment{*comment}, comments...)
				data, err := json.Marshal(comments)
				if err == nil {
					pipe.Set(ctx, key, data, 2*time.Minute)
				}
			}
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("Failed to update cache for pattern %s: %v", cachePattern, err)
	}
}

func (s *commentService) invalidateCache(pattern string) {
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
