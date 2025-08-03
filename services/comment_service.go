package services

import (
	"Forum_BE/models"
	"Forum_BE/notification"
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
	GetAllComments(filters map[string]interface{}) ([]models.Comment, int, error)
	UpdateCommentStatus(id uint, status string) (*models.Comment, error)
}

type commentService struct {
	commentRepo repositories.CommentRepository
	postRepo    repositories.PostRepository
	answerRepo  repositories.AnswerRepository
	userRepo    repositories.UserRepository // Thêm UserRepository
	redisClient *redis.Client
	db          *gorm.DB
	novuClient  *notification.NovuClient // Thêm NovuClient
}

func NewCommentService(cRepo repositories.CommentRepository, pRepo repositories.PostRepository, aRepo repositories.AnswerRepository, userRepo repositories.UserRepository, redisClient *redis.Client, db *gorm.DB, novuClient *notification.NovuClient) CommentService {
	return &commentService{
		commentRepo: cRepo,
		postRepo:    pRepo,
		answerRepo:  aRepo,
		userRepo:    userRepo, // Khởi tạo UserRepository
		redisClient: redisClient,
		db:          db,
		novuClient:  novuClient, // Khởi tạo NovuClient
	}
}

func (s *commentService) CreateComment(content string, userID uint, postID *uint, answerID *uint, parentID *uint) (*models.Comment, error) {
	if content == "" {
		return nil, fmt.Errorf("Nội dung là bắt buộc")
	}

	if postID != nil {
		post, err := s.postRepo.GetPostByID(*postID)
		if err != nil {
			return nil, fmt.Errorf("Không tìm thấy bài viết: %v", err)
		}
		if post.Status != "approved" {
			return nil, fmt.Errorf("Không thể bình luận trên bài viết chưa được duyệt")
		}
	}

	if answerID != nil {
		_, err := s.answerRepo.GetAnswerByID(*answerID)
		if err != nil {
			return nil, fmt.Errorf("Không tìm thấy câu trả lời: %v", err)
		}
	}

	if parentID != nil {
		parent, err := s.commentRepo.GetCommentByID(*parentID)
		if err != nil {
			return nil, fmt.Errorf("Không tìm thấy bình luận cha: %v", err)
		}
		if parent.Status != "approved" {
			return nil, fmt.Errorf("Không thể trả lời bình luận chưa được duyệt")
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
		return nil, fmt.Errorf("Tạo bình luận thất bại: %v", err)
	}

	// Invalidate cache
	if postID != nil && parentID == nil {
		s.invalidateCache(fmt.Sprintf("comments:post:%d:*", *postID))
	}
	if answerID != nil && parentID == nil {
		s.invalidateCache(fmt.Sprintf("comments:answer:%d:*", *answerID))
	}
	if parentID != nil {
		s.invalidateCache(fmt.Sprintf("comments:comment:%d:*", *parentID))
		s.updateReplyCacheAfterCreate(comment, *parentID)
	}

	// Send notification
	commenter, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("Không lấy được thông tin người bình luận: %v", err)
	} else {
		if postID != nil {
			post, err := s.postRepo.GetPostByID(*postID)
			if err == nil && post.UserID != userID {
				workflowID := "new-post-comment-notification"
				message := fmt.Sprintf("%s đã bình luận trên bài viết của bạn: %s", commenter.FullName, post.Title)
				if err := s.novuClient.SendNotification(post.UserID, workflowID, message); err != nil {
					log.Printf("Gửi thông báo bình luận bài viết thất bại: %v", err)
				}
			}
		} else if parentID != nil {
			parent, err := s.commentRepo.GetCommentByID(*parentID)
			if err == nil && parent.UserID != userID {
				workflowID := "new-comment-reply-notification"

				message := fmt.Sprintf("%s đã trả lời bình luận của bạn: %s", commenter.FullName, utils.StripHTML(parent.Content))
				log.Printf("Gửi thông báo trả lời bình luận thất bại: %v", message)
				if err := s.novuClient.SendNotification(parent.UserID, workflowID, message); err != nil {
					log.Printf("Gửi thông báo trả lời bình luận thất bại: %v", err)
				}
			}
		} else if answerID != nil {
			answer, err := s.answerRepo.GetAnswerByID(*answerID)
			if err == nil && answer.UserID != userID {
				workflowID := "new-answer-comment-notification"
				message := fmt.Sprintf("%s đã bình luận trên câu trả lời của bạn: %s", commenter.FullName, answer.Title)
				if err := s.novuClient.SendNotification(answer.UserID, workflowID, message); err != nil {
					log.Printf("Gửi thông báo trả lời bình luận thất bại: %v", err)
				}
			}
		}
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
			log.Printf("Cache hit cho comment:%d", id)
			return &comment, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Lỗi Redis cho comment:%d: %v", id, err)
	}

	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		return nil, fmt.Errorf("Lấy bình luận thất bại: %v", err)
	}
	if comment.DeletedAt.Valid {
		return nil, fmt.Errorf("Không tìm thấy bình luận")
	}

	data, err := json.Marshal(comment)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Lưu cache cho comment:%d thất bại: %v", id, err)
		}
	}

	return comment, nil
}

func (s *commentService) UpdateComment(id uint, content string) (*models.Comment, error) {
	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		return nil, fmt.Errorf("Lấy bình luận thất bại: %v", err)
	}
	if comment.DeletedAt.Valid {
		return nil, fmt.Errorf("Không tìm thấy bình luận")
	}

	if content != "" {
		comment.Content = content
	}

	if err := s.commentRepo.UpdateComment(comment); err != nil {
		return nil, fmt.Errorf("Cập nhật bình luận thất bại: %v", err)
	}

	s.invalidateCache(fmt.Sprintf("comment:%d", id))
	if comment.PostID != nil {
		s.invalidateCache(fmt.Sprintf("comments:post:%d:*", *comment.PostID))
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
		return fmt.Errorf("Lấy bình luận thất bại: %v", err)
	}
	if comment.DeletedAt.Valid {
		return fmt.Errorf("Không tìm thấy bình luận")
	}

	childIDs, err := s.commentRepo.GetAllChildCommentIDs(id)
	if err != nil {
		return fmt.Errorf("Lấy ID bình luận con thất bại: %v", err)
	}

	if err := s.commentRepo.DeleteComment(id); err != nil {
		return fmt.Errorf("Xóa bình luận thất bại: %v", err)
	}

	s.invalidateCache(fmt.Sprintf("comment:%d", id))

	for _, childID := range childIDs {
		s.invalidateCache(fmt.Sprintf("comment:%d", childID))
		s.invalidateCache(fmt.Sprintf("replies:comment:%d:*", childID))
	}

	if comment.PostID != nil {
		s.invalidateCache(fmt.Sprintf("comments:post:%d:*", *comment.PostID))
	}
	if comment.AnswerID != nil {
		s.invalidateCache(fmt.Sprintf("comments:answer:%d:*", *comment.AnswerID))
	}
	if comment.ParentID != nil {
		s.invalidateCache(fmt.Sprintf("replies:comment:%d:*", *comment.ParentID))
	}
	s.invalidateCache("comments:all:*")

	log.Printf("Đã xóa bình luận %d và các con của nó, xóa cache", id)
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
		cacheKey = utils.GenerateCacheKey("comments:post", postID, filters)
	} else if answerID, ok := filters["answer_id"].(uint); ok {
		cacheKey = utils.GenerateCacheKey("comments:answer", answerID, filters)
	} else {
		return nil, 0, fmt.Errorf("post_id hoặc answer_id là bắt buộc")
	}

	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var response CommentListResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			if s.validateCommentsUserData(response.Comments) {
				log.Printf("Cache hit cho %s", cacheKey)
				return response.Comments, response.Total, nil
			}
		}
	}
	if err != redis.Nil {
		log.Printf("Lỗi Redis cho %s: %v", cacheKey, err)
	}

	comments, total, err := s.commentRepo.ListComments(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("Lấy danh sách bình luận thất bại: %v", err)
	}

	response := CommentListResponse{
		Comments: comments,
		Total:    int(total),
	}

	data, err := json.Marshal(response)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Lưu cache cho %s thất bại: %v", cacheKey, err)
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
				log.Printf("Cache hit cho %s", cacheKey)
				return response.Comments, response.Total, nil
			}
		}
	}
	if err != redis.Nil {
		log.Printf("Lỗi Redis cho %s: %v", cacheKey, err)
	}

	replies, total, err := s.commentRepo.ListReplies(parentID, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("Lấy danh sách trả lời thất bại: %v", err)
	}

	response := CommentListResponse{
		Comments: replies,
		Total:    int(total),
	}

	data, err := json.Marshal(response)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Lưu cache cho %s thất bại: %v", cacheKey, err)
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
				log.Printf("Cache hit cho %s", cacheKey)
				return response.Comments, response.Total, nil
			}
		}
	}
	if err != redis.Nil {
		log.Printf("Lỗi Redis cho %s: %v", cacheKey, err)
	}

	comments, total, err := s.commentRepo.GetAllComments(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("Lấy tất cả bình luận thất bại: %v", err)
	}

	response := CommentListResponse{
		Comments: comments,
		Total:    int(total),
	}

	data, err := json.Marshal(response)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Lưu cache cho %s thất bại: %v", cacheKey, err)
		}
	}

	return comments, int(total), nil
}

func (s *commentService) UpdateCommentStatus(id uint, status string) (*models.Comment, error) {
	if !IsValidCommentStatus(status) {
		return nil, errors.New("Trạng thái không hợp lệ")
	}
	if err := s.commentRepo.UpdateCommentStatus(id, status); err != nil {
		log.Printf("Cập nhật trạng thái bình luận thất bại: %v", err)
		return nil, err
	}
	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		log.Printf("Lấy bình luận đã cập nhật %d thất bại: %v", id, err)
	}
	s.invalidateCache(fmt.Sprintf("comment:%d", id))
	if comment.PostID != nil {
		s.invalidateCache(fmt.Sprintf("comments:post:%d:*", *comment.PostID))
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
			log.Printf("Dữ liệu bình luận không hợp lệ: ID %d, UserID %d, Deleted %v", comment.ID, comment.User.ID, comment.DeletedAt.Valid)
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
		log.Printf("Lấy khóa cache cho pattern %s thất bại: %v", cachePattern, err)
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
		log.Printf("Cập nhật cache cho pattern %s thất bại: %v", cachePattern, err)
	}
}

func (s *commentService) updateReplyCacheAfterCreate(comment *models.Comment, parentID uint) {
	cachePattern := fmt.Sprintf("replies:comment:%d:*", parentID)
	ctx := context.Background()

	pipe := s.redisClient.Pipeline()
	keys, err := s.redisClient.Keys(ctx, cachePattern).Result()
	if err != nil {
		log.Printf("Lấy khóa cache cho reply pattern %s thất bại: %v", cachePattern, err)
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
		log.Printf("Lấy khóa cache cho parent pattern %s thất bại: %v", parentCachePattern, err)
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
		log.Printf("Cập nhật cache cho patterns %s và %s thất bại: %v", cachePattern, parentCachePattern, err)
	}
}

func (s *commentService) invalidateCache(pattern string) {
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Lấy khóa cache cho pattern %s thất bại: %v", pattern, err)
		return
	}

	if len(keys) > 0 {
		if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Xóa khóa cache %v thất bại: %v", keys, err)
		}
	}
}
