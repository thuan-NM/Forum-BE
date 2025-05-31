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
	"gorm.io/gorm"
	"log"
	"time"
)

type CommentService interface {
	CreateComment(content string, userID uint, postID *uint, answerID *uint, parentID *uint) (*models.Comment, error)
	GetCommentByID(id uint) (*models.Comment, error)
	UpdateComment(id uint, content string) (*models.Comment, error)
	DeleteComment(id uint) error
	ListComments(filters map[string]interface{}) ([]models.Comment, int, error)
	ListReplies(parentID uint, filters map[string]interface{}) ([]models.Comment, int, error)
	GetAllComments(filters map[string]interface{}) ([]models.Comment, int, error) // Thêm method mới
	UpdateCommentStatus(id uint, status string) (*models.Comment, error)
}

type commentService struct {
	commentRepo  repositories.CommentRepository
	questionRepo repositories.QuestionRepository
	answerRepo   repositories.AnswerRepository
	redisClient  *redis.Client
	db           *gorm.DB
}

func NewCommentService(cRepo repositories.CommentRepository, qRepo repositories.QuestionRepository, aRepo repositories.AnswerRepository, redisClient *redis.Client, db *gorm.DB) CommentService {
	return &commentService{
		commentRepo:  cRepo,
		questionRepo: qRepo,
		answerRepo:   aRepo,
		redisClient:  redisClient,
		db:           db,
	}
}

func (s *commentService) CreateComment(content string, userID uint, postID *uint, answerID *uint, parentID *uint) (*models.Comment, error) {
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	if postID == nil && answerID == nil {
		return nil, fmt.Errorf("either post_id or answer_id must be provided")
	}

	if postID != nil {
		question, err := s.questionRepo.GetQuestionByID(*postID)
		if err != nil {
			return nil, fmt.Errorf("post not found: %v", err)
		}
		if question.Status != models.StatusApproved {
			return nil, fmt.Errorf("cannot comment on an unapproved post")
		}
	}

	if answerID != nil {
		_, err := s.answerRepo.GetAnswerByID(*answerID)
		if err != nil {
			return nil, fmt.Errorf("answer not found: %v", err)
		}
	}

	if parentID != nil {
		parent, err := s.commentRepo.GetCommentByID(*parentID)
		if err != nil {
			return nil, fmt.Errorf("parent comment not found: %v", err)
		}
		if parent.Status != "approved" {
			return nil, fmt.Errorf("cannot reply to an unapproved comment")
		}
	}

	comment := &models.Comment{
		Content:  content,
		UserID:   userID,
		PostID:   postID,
		AnswerID: answerID,
		ParentID: parentID,
		Status:   "approved",
		Metadata: []byte(`{"has_replies": false}`),
	}

	if err := s.commentRepo.CreateComment(comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %v", err)
	}

	if postID != nil && parentID == nil {
		s.updateCacheAfterCreate(comment, *postID, true)
	}
	if answerID != nil && parentID == nil {
		s.updateCacheAfterCreate(comment, *answerID, false)
	}
	if parentID != nil {
		s.updateReplyCacheAfterCreate(comment, *parentID)
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
		return nil, fmt.Errorf("failed to get comment: %v", err)
	}
	if comment.DeletedAt.Valid {
		return nil, fmt.Errorf("comment not found")
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
		return nil, fmt.Errorf("failed to get comment: %v", err)
	}
	if comment.DeletedAt.Valid {
		return nil, fmt.Errorf("comment not found")
	}

	if content != "" {
		comment.Content = content
	}

	if err := s.commentRepo.UpdateComment(comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %v", err)
	}

	s.invalidateCache(fmt.Sprintf("comment:%d", id))
	if comment.PostID != nil {
		s.invalidateCache(fmt.Sprintf("comments:question:%d:*", *comment.PostID))
	}
	if comment.AnswerID != nil {
		s.invalidateCache(fmt.Sprintf("comments:answer:%d:*", *comment.AnswerID))
	}
	if comment.ParentID != nil {
		s.invalidateCache(fmt.Sprintf("replies:comment:%d:*", *comment.ParentID))
	}

	return comment, nil
}

func (s *commentService) DeleteComment(id uint) error {
	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		return fmt.Errorf("failed to get comment: %v", err)
	}
	if comment.DeletedAt.Valid {
		return fmt.Errorf("comment not found")
	}

	childIDs, err := s.commentRepo.GetAllChildCommentIDs(id)
	if err != nil {
		return fmt.Errorf("failed to get child comment IDs: %v", err)
	}

	if err := s.commentRepo.DeleteComment(id); err != nil {
		return fmt.Errorf("failed to delete comment: %v", err)
	}

	s.invalidateCache(fmt.Sprintf("comment:%d", id))

	for _, childID := range childIDs {
		s.invalidateCache(fmt.Sprintf("comment:%d", childID))
		s.invalidateCache(fmt.Sprintf("replies:comment:%d:*", childID)) // Nếu comment con có replies
	}

	if comment.PostID != nil {
		s.invalidateCache(fmt.Sprintf("comments:question:%d:*", *comment.PostID))
	}
	if comment.AnswerID != nil {
		s.invalidateCache(fmt.Sprintf("comments:answer:%d:*", *comment.AnswerID))
	}
	if comment.ParentID != nil {
		s.invalidateCache(fmt.Sprintf("replies:comment:%d:*", *comment.ParentID))
	}
	s.invalidateCache("comments:all:*")

	log.Printf("Deleted comment %d and its children, invalidated cache", id)
	return nil
}

type CommentListResponse struct {
	Comments []models.Comment `json:"comments"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	Limit    int              `json:"limit"`
}

func (s *commentService) ListComments(filters map[string]interface{}) ([]models.Comment, int, error) {

	var cacheKey string
	if postID, ok := filters["post_id"].(uint); ok {
		cacheKey = utils.GenerateCacheKey("comments:question", postID, filters)
	} else if answerID, ok := filters["answer_id"].(uint); ok {
		cacheKey = utils.GenerateCacheKey("comments:answer", answerID, filters)
	} else {
		return nil, 0, fmt.Errorf("post_id or answer_id is required")
	}

	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var response CommentListResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			if s.validateCommentsUserData(response.Comments) {
				log.Printf("Cache hit for %s", cacheKey)
				return response.Comments, response.Total, nil
			}
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	comments, total, err := s.commentRepo.ListComments(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list comments: %v", err)
	}

	response := CommentListResponse{
		Comments: comments,
		Total:    int(total),
	}

	data, err := json.Marshal(response)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}

	return comments, int(total), nil
}

func (s *commentService) ListReplies(parentID uint, filters map[string]interface{}) ([]models.Comment, int, error) {

	cacheKey := utils.GenerateCacheKey("comments:comment", parentID, filters)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var response CommentListResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			if s.validateCommentsUserData(response.Comments) {
				log.Printf("Cache hit for %s", cacheKey)
				return response.Comments, response.Total, nil
			}
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	replies, total, err := s.commentRepo.ListReplies(parentID, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list replies: %v", err)
	}

	response := CommentListResponse{
		Comments: replies,
		Total:    int(total),
	}

	data, err := json.Marshal(response)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}

	return replies, int(total), nil
}
func (s *commentService) GetAllComments(filters map[string]interface{}) ([]models.Comment, int, error) {

	cacheKey := utils.GenerateCacheKey("comments:all", 0, filters)

	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var response CommentListResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			if s.validateCommentsUserData(response.Comments) {
				log.Printf("Cache hit for %s", cacheKey)
				return response.Comments, response.Total, nil
			}
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	comments, total, err := s.commentRepo.GetAllComments(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all comments: %v", err)
	}

	response := CommentListResponse{
		Comments: comments,
		Total:    int(total),
	}

	data, err := json.Marshal(response)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}

	return comments, int(total), nil
}

func (s *commentService) UpdateCommentStatus(id uint, status string) (*models.Comment, error) {
	if !IsValidCommentStatus(status) {
		return nil, errors.New("Invalid status")
	}
	if err := s.commentRepo.UpdateCommentStatus(id, status); err != nil {
		log.Printf("Failed to update comment status: %v", err)
		return nil, err
	}
	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		log.Printf("Failed to get Updated comment %d: %v", id, err)

	}
	s.invalidateCache(fmt.Sprintf("comment:%d", id))
	if comment.PostID != nil {
		s.invalidateCache(fmt.Sprintf("comments:question:%d:*", *comment.PostID))
	}
	if comment.AnswerID != nil {
		s.invalidateCache(fmt.Sprintf("comments:answer:%d:*", *comment.AnswerID))
	}
	if comment.ParentID != nil {
		s.invalidateCache(fmt.Sprintf("replies:comment:%d:*", *comment.ParentID))
	}
	s.invalidateCache("comments:all:*")
	return comment, nil
}
func IsValidCommentStatus(status string) bool {
	validStatuses := []string{"approved", "pending", "spam"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}
func (s *commentService) validateCommentsUserData(comments []models.Comment) bool {
	for _, comment := range comments {
		if comment.User.ID == 0 || comment.DeletedAt.Valid {
			log.Printf("Invalid comment data: ID %d, UserID %d, Deleted %v", comment.ID, comment.User.ID, comment.DeletedAt.Valid)
			return false
		}
	}
	return true
}

func (s *commentService) updateCacheAfterCreate(comment *models.Comment, id uint, isQuestion bool) {
	prefix := "answer"
	if isQuestion {
		prefix = "question"
	}
	cachePattern := fmt.Sprintf("comments:%s:%d:*", prefix, id)
	ctx := context.Background()

	keys, err := s.redisClient.Keys(ctx, cachePattern).Result()
	if err != nil {
		log.Printf("Failed to get cache keys for pattern %s: %v", cachePattern, err)
		return
	}

	pipe := s.redisClient.Pipeline()
	for _, key := range keys {
		cached, err := s.redisClient.Get(ctx, key).Result()
		if err == nil {
			var response CommentListResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				response.Comments = append([]models.Comment{*comment}, response.Comments...)
				response.Total++
				data, err := json.Marshal(response)
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

func (s *commentService) updateReplyCacheAfterCreate(comment *models.Comment, parentID uint) {
	cachePattern := fmt.Sprintf("replies:comment:%d:*", parentID)
	ctx := context.Background()

	pipe := s.redisClient.Pipeline()
	keys, err := s.redisClient.Keys(ctx, cachePattern).Result()
	if err != nil {
		log.Printf("Failed to get cache keys for reply pattern %s: %v", cachePattern, err)
		return
	}

	for _, key := range keys {
		cached, err := s.redisClient.Get(ctx, key).Result()
		if err == nil {
			var response CommentListResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				response.Comments = append([]models.Comment{*comment}, response.Comments...)
				response.Total++
				data, err := json.Marshal(response)
				if err == nil {
					pipe.Set(ctx, key, data, 2*time.Minute)
				}
			}
		}
	}

	// Update parent comment's has_replies
	parentCachePattern := fmt.Sprintf("comments:*:*")
	parentKeys, err := s.redisClient.Keys(ctx, parentCachePattern).Result()
	if err != nil {
		log.Printf("Failed to get cache keys for parent pattern %s: %v", parentCachePattern, err)
		return
	}

	for _, key := range parentKeys {
		cached, err := s.redisClient.Get(ctx, key).Result()
		if err == nil {
			var response CommentListResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				for i, c := range response.Comments {
					if c.ID == parentID {
						response.Comments[i].Metadata = []byte(`{"has_replies": true}`)
					}
				}
				data, err := json.Marshal(response)
				if err == nil {
					pipe.Set(ctx, key, data, 2*time.Minute)
				}
			}
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("Failed to update cache for patterns %s and %s: %v", cachePattern, parentCachePattern, err)
	}
}

func (s *commentService) invalidateCache(pattern string) {
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Failed to get cache keys for pattern %s: %v", pattern, err)
		return
	}

	if len(keys) > 0 {
		if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Failed to delete cache keys %v: %v", keys, err)
		}
	}
}
