package services

import (
	"Forum_BE/models"
	"Forum_BE/notification" // Giả sử có package notification với NovuClient
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"strings"
	"time"
)

type ReactionService interface {
	CreateReaction(userID uint, postID, commentID, answerID *uint) (*models.Reaction, error)
	GetReactionByID(id uint) (*models.Reaction, error)
	UpdateReaction(id, userID uint, postID, commentID, answerID *uint) (*models.Reaction, error)
	DeleteReaction(id, userID uint) error
	ListReactions(filters map[string]interface{}) ([]models.Reaction, int, error)
	GetReactionCount(postID, commentID, answerID *uint) (int64, error)
	CheckUserReaction(userID uint, postID, commentID, answerID *uint) (bool, *models.Reaction, error)
	ValidateReactionID(postID, commentID, answerID *uint) error
}

type reactionService struct {
	reactionRepo repositories.ReactionRepository
	answerRepo   repositories.AnswerRepository
	postRepo     repositories.PostRepository
	commentRepo  repositories.CommentRepository
	userRepo     repositories.UserRepository // Thêm UserRepository
	redisClient  *redis.Client
	novuClient   *notification.NovuClient // Thêm NovuClient
}

func NewReactionService(repo repositories.ReactionRepository, userRepo repositories.UserRepository, answerRepo repositories.AnswerRepository, postRepo repositories.PostRepository, commentRepo repositories.CommentRepository, redisClient *redis.Client, novuClient *notification.NovuClient) ReactionService {
	if repo == nil {
		log.Fatal("reaction repository is nil")
	}
	if redisClient == nil {
		log.Fatal("redis client is nil")
	}
	if userRepo == nil {
		log.Fatal("user repository is nil")
	}
	if answerRepo == nil {
		log.Fatal("user repository is nil")
	}
	if postRepo == nil {
		log.Fatal("user repository is nil")
	}
	if commentRepo == nil {
		log.Fatal("user repository is nil")
	}
	if novuClient == nil {
		log.Fatal("novu client is nil")
	}
	log.Printf("Initialized ReactionService with repo: %v, userRepo: %v, redis: %v, novu: %v", repo != nil, userRepo != nil, redisClient != nil, novuClient != nil)
	return &reactionService{reactionRepo: repo, userRepo: userRepo, answerRepo: answerRepo, postRepo: postRepo, commentRepo: commentRepo, redisClient: redisClient, novuClient: novuClient}
}

type ReactionListResponse struct {
	Reactions []models.Reaction `json:"reactions"`
	Total     int               `json:"total"`
}

func (s *reactionService) CreateReaction(userID uint, postID, commentID, answerID *uint) (*models.Reaction, error) {
	if err := s.reactionRepo.ValidateReactionID(postID, commentID, answerID); err != nil {
		return nil, err
	}
	filters := map[string]interface{}{
		"user_id": userID,
	}
	if postID != nil {
		filters["post_id"] = *postID
	} else if commentID != nil {
		filters["comment_id"] = *commentID
	} else if answerID != nil {
		filters["answer_id"] = *answerID
	}
	existingReactions, _, err := s.reactionRepo.ListReactions(filters)
	if err != nil {
		log.Printf("Failed to check existing reaction: %v", err)
		return nil, err
	}
	if len(existingReactions) > 0 {
		return nil, errors.New("reaction already exists for this user and entity")
	}
	reaction := &models.Reaction{
		UserID:    userID,
		PostID:    postID,
		CommentID: commentID,
		AnswerID:  answerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.reactionRepo.CreateReaction(reaction); err != nil {
		log.Printf("Failed to create reaction: %v", err)
		return nil, err
	}
	s.updateCacheAfterCreate(reaction, postID, commentID, answerID)
	cacheKey := fmt.Sprintf("user_reaction:%d:%s", userID, generateCacheKey(filters))
	s.invalidateCache(cacheKey)
	s.invalidateCache(fmt.Sprintf("reaction:%d", reaction.ID))
	if postID != nil {
		s.invalidateCache(fmt.Sprintf("reactions:post:%d:*", *postID))
		s.invalidateCache(fmt.Sprintf("reaction_count:post:%d", *postID))
	} else if commentID != nil {
		s.invalidateCache(fmt.Sprintf("reactions:comment:%d:*", *commentID))
		s.invalidateCache(fmt.Sprintf("reaction_count:comment:%d", *commentID))
	} else if answerID != nil {
		s.invalidateCache(fmt.Sprintf("reactions:answer:%d:*", *answerID))
		s.invalidateCache(fmt.Sprintf("reaction_count:answer:%d", *answerID))
	}
	s.invalidateCache("reactions:all:*")

	// Send notification
	reactor, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		log.Printf("Không lấy được thông tin người phản ứng: %v", err)
	} else {
		if postID != nil {
			// Lấy thông tin post để gửi thông báo cho chủ sở hữu
			post, err := s.postRepo.GetPostByID(*postID) // Giả sử có method này trong repository
			if err == nil && post.UserID != userID {
				workflowID := "new-post-reaction-notification"
				message := fmt.Sprintf("%s đã thích bài viết của bạn: %s", reactor.FullName, post.Title)
				if err := s.novuClient.SendNotification(post.UserID, workflowID, message); err != nil {
					log.Printf("Gửi thông báo phản ứng bài viết thất bại: %v", err)
				}
			}
		} else if commentID != nil {
			comment, err := s.commentRepo.GetCommentByID(*commentID)
			if err == nil && comment.UserID != userID {
				workflowID := "new-comment-reaction-notification"
				message := fmt.Sprintf("%s đã thích bình luận của bạn: %s", reactor.FullName, utils.StripHTML(comment.Content))
				if err := s.novuClient.SendNotification(comment.UserID, workflowID, message); err != nil {
					log.Printf("Gửi thông báo thích bình luận thất bại: %v", err)
				}
			}
		} else if answerID != nil {
			answer, err := s.answerRepo.GetAnswerByID(*answerID)
			if err == nil && answer.UserID != userID {
				workflowID := "new-answer-reaction-notification"
				message := fmt.Sprintf("%s đã thích câu trả lời của bạn: %s", reactor.FullName, answer.Title)
				if err := s.novuClient.SendNotification(answer.UserID, workflowID, message); err != nil {
					log.Printf("Gửi thông báo thích câu trả lời thất bại: %v", err)
				}
			}
		}
	}

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
	if reaction.DeletedAt.Valid {
		return nil, fmt.Errorf("reaction not found")
	}
	data, err := json.Marshal(reaction)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for reaction:%d: %v", id, err)
		}
	}
	return reaction, nil
}

func (s *reactionService) UpdateReaction(id, userID uint, postID, commentID, answerID *uint) (*models.Reaction, error) {
	reaction, err := s.reactionRepo.GetReactionByID(id)
	if err != nil {
		log.Printf("Failed to get reaction %d: %v", id, err)
		return nil, err
	}
	if reaction.DeletedAt.Valid {
		return nil, fmt.Errorf("reaction not found")
	}
	if reaction.UserID != userID {
		return nil, errors.New("unauthorized: cannot update another user's reaction")
	}
	if err := s.reactionRepo.ValidateReactionID(postID, commentID, answerID); err != nil {
		return nil, err
	}
	reaction.PostID = postID
	reaction.CommentID = commentID
	reaction.AnswerID = answerID
	reaction.UpdatedAt = time.Now()
	if err := s.reactionRepo.UpdateReaction(reaction); err != nil {
		log.Printf("Failed to update reaction %d: %v", id, err)
		return nil, err
	}
	s.invalidateCache(fmt.Sprintf("reaction:%d", id))
	if postID != nil {
		s.invalidateCache(fmt.Sprintf("reactions:post:%d:*", *postID))
		s.invalidateCache(fmt.Sprintf("reaction_count:post:%d", *postID))
	} else if commentID != nil {
		s.invalidateCache(fmt.Sprintf("reactions:comment:%d:*", *commentID))
		s.invalidateCache(fmt.Sprintf("reaction_count:comment:%d", *commentID))
	} else if answerID != nil {
		s.invalidateCache(fmt.Sprintf("reactions:answer:%d:*", *answerID))
		s.invalidateCache(fmt.Sprintf("reaction_count:answer:%d", *answerID))
	}
	s.invalidateCache("reactions:all:*")
	return reaction, nil
}

func (s *reactionService) DeleteReaction(id, userID uint) error {
	reaction, err := s.reactionRepo.GetReactionByID(id)
	if err != nil {
		log.Printf("Failed to get reaction %d: %v", id, err)
		return err
	}
	if reaction.DeletedAt.Valid {
		return fmt.Errorf("reaction not found")
	}
	if reaction.UserID != userID {
		return errors.New("unauthorized: cannot delete another user's reaction")
	}
	if err := s.reactionRepo.DeleteReaction(id); err != nil {
		log.Printf("Failed to delete reaction %d: %v", id, err)
		return err
	}
	filters := map[string]interface{}{
		"user_id": userID,
	}
	if reaction.PostID != nil {
		filters["post_id"] = *reaction.PostID
		cacheKey := fmt.Sprintf("user_reaction:%d:%s", userID, generateCacheKey(filters))
		s.invalidateCache(cacheKey)
		s.invalidateCache(fmt.Sprintf("reaction_count:post:%d", *reaction.PostID))
	} else if reaction.CommentID != nil {
		filters["comment_id"] = *reaction.CommentID
		cacheKey := fmt.Sprintf("user_reaction:%d:%s", userID, generateCacheKey(filters))
		s.invalidateCache(cacheKey)
		s.invalidateCache(fmt.Sprintf("reaction_count:comment:%d", *reaction.CommentID))
	} else if reaction.AnswerID != nil {
		filters["answer_id"] = *reaction.AnswerID
		cacheKey := fmt.Sprintf("user_reaction:%d:%s", userID, generateCacheKey(filters))
		s.invalidateCache(cacheKey)
		s.invalidateCache(fmt.Sprintf("reaction_count:answer:%d", *reaction.AnswerID))
	}
	s.invalidateCache(fmt.Sprintf("reaction:%d", id))
	if reaction.PostID != nil {
		s.invalidateCache(fmt.Sprintf("reactions:post:%d:*", *reaction.PostID))
	} else if reaction.CommentID != nil {
		s.invalidateCache(fmt.Sprintf("reactions:comment:%d:*", *reaction.CommentID))
	} else if reaction.AnswerID != nil {
		s.invalidateCache(fmt.Sprintf("reactions:answer:%d:*", *reaction.AnswerID))
	}
	s.invalidateCache("reactions:all:*")
	return nil
}

func (s *reactionService) ListReactions(filters map[string]interface{}) ([]models.Reaction, int, error) {
	cacheKey := fmt.Sprintf("reactions:all:%s", generateCacheKey(filters))
	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var response ReactionListResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			if s.validateReactionsUserData(response.Reactions) {
				log.Printf("Cache hit for %s", cacheKey)
				return response.Reactions, response.Total, nil
			}
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
	response := ReactionListResponse{
		Reactions: reactions,
		Total:     total,
	}
	data, err := json.Marshal(response)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}
	return reactions, total, nil
}

func (s *reactionService) GetReactionCount(postID, commentID, answerID *uint) (int64, error) {
	count := 0
	if postID != nil {
		count++
	}
	if commentID != nil {
		count++
	}
	if answerID != nil {
		count++
	}
	if count != 1 {
		return 0, errors.New("exactly one of post_id, comment_id, or answer_id must be provided")
	}
	if err := s.reactionRepo.ValidateReactionID(postID, commentID, answerID); err != nil {
		return 0, err
	}
	var cacheKey string
	if postID != nil {
		cacheKey = fmt.Sprintf("reaction_count:post:%d", *postID)
	} else if commentID != nil {
		cacheKey = fmt.Sprintf("reaction_count:comment:%d", *commentID)
	} else if answerID != nil {
		cacheKey = fmt.Sprintf("reaction_count:answer:%d", *answerID)
	}
	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		countVal, err := strconv.ParseInt(cached, 10, 64)
		if err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return countVal, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}
	countVal, err := s.reactionRepo.GetReactionCount(postID, commentID, answerID)
	if err != nil {
		log.Printf("Failed to get reaction count: %v", err)
		return 0, err
	}
	if err := s.redisClient.Set(ctx, cacheKey, fmt.Sprintf("%d", countVal), 2*time.Minute).Err(); err != nil {
		log.Printf("Failed to set cache for %s: %v", cacheKey, err)
	}
	return countVal, nil
}

func (s *reactionService) CheckUserReaction(userID uint, postID, commentID, answerID *uint) (bool, *models.Reaction, error) {
	if err := s.reactionRepo.ValidateReactionID(postID, commentID, answerID); err != nil {
		return false, nil, err
	}
	filters := map[string]interface{}{
		"user_id": userID,
	}
	if postID != nil {
		filters["post_id"] = *postID
	} else if commentID != nil {
		filters["comment_id"] = *commentID
	} else if answerID != nil {
		filters["answer_id"] = *answerID
	}
	cacheKey := fmt.Sprintf("user_reaction:%d:%s", userID, generateCacheKey(filters))
	ctx := context.Background()
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var reaction models.Reaction
		if err := json.Unmarshal([]byte(cached), &reaction); err == nil {
			log.Printf("Cache hit for user reaction: %s", cacheKey)
			return true, &reaction, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}
	reactions, _, err := s.reactionRepo.ListReactions(filters)
	if err != nil {
		log.Printf("Failed to check user reaction: %v", err)
		return false, nil, err
	}
	if len(reactions) > 0 {
		data, err := json.Marshal(&reactions[0])
		if err == nil {
			if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
				log.Printf("Failed to set cache for %s: %v", cacheKey, err)
			}
		}
		return true, &reactions[0], nil
	}
	return false, nil, nil
}

func (s *reactionService) validateReactionsUserData(reactions []models.Reaction) bool {
	for _, reaction := range reactions {
		if reaction.User.ID == 0 || reaction.DeletedAt.Valid {
			log.Printf("Invalid reaction data: ID %d, UserID %d, Deleted %v", reaction.ID, reaction.User.ID, reaction.DeletedAt.Valid)
			return false
		}
	}
	return true
}

func (s *reactionService) updateCacheAfterCreate(reaction *models.Reaction, postID, commentID, answerID *uint) {
	var cachePattern string
	if postID != nil {
		cachePattern = fmt.Sprintf("reactions:post:%d:*", *postID)
	} else if commentID != nil {
		cachePattern = fmt.Sprintf("reactions:comment:%d:*", *commentID)
	} else if answerID != nil {
		cachePattern = fmt.Sprintf("reactions:answer:%d:*", *answerID)
	}
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
			var response ReactionListResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				response.Reactions = append([]models.Reaction{*reaction}, response.Reactions...)
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

func (s *reactionService) invalidateCache(pattern string) {
	if s.redisClient == nil {
		log.Printf("Redis client is not initialized, skipping cache invalidation")
		return
	}
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Failed to get cache keys for pattern %s: %v", pattern, err)
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

func generateCacheKey(filters map[string]interface{}) string {
	keys := []string{}
	for k, v := range filters {
		keys = append(keys, fmt.Sprintf("%s:%v", k, v))
	}
	return strings.Join(keys, ":")
}

func (s *reactionService) ValidateReactionID(postID, commentID, answerID *uint) error {
	return s.reactionRepo.ValidateReactionID(postID, commentID, answerID)
}
