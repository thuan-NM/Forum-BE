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
	"log"
	"strconv"
	"time"
)

type ReportService interface {
	CreateReport(reason string, reporterID uint, contentType string, contentID string, contentPreview string, details string) (*models.Report, error)
	GetReportByID(id string) (*models.Report, error)
	UpdateReport(id string, reason string, details string, resolvedByID *uint) (*models.Report, error)
	UpdateReportStatus(id string, status string, resolvedByID *uint) (*models.Report, error)
	DeleteReport(id string) error
	BatchDeleteReports(ids []string) error
	ListReports(filters map[string]interface{}) ([]models.Report, int, error)
}

type reportService struct {
	repo         repositories.ReportRepository
	userRepo     repositories.UserRepository
	postRepo     repositories.PostRepository
	commentRepo  repositories.CommentRepository
	questionRepo repositories.QuestionRepository
	answerRepo   repositories.AnswerRepository
	redisClient  *redis.Client
}

func NewReportService(
	repo repositories.ReportRepository,
	userRepo repositories.UserRepository,
	postRepo repositories.PostRepository,
	commentRepo repositories.CommentRepository,
	questionRepo repositories.QuestionRepository,
	answerRepo repositories.AnswerRepository,
	redisClient *redis.Client,
) ReportService {
	return &reportService{
		repo:         repo,
		userRepo:     userRepo,
		postRepo:     postRepo,
		commentRepo:  commentRepo,
		questionRepo: questionRepo,
		answerRepo:   answerRepo,
		redisClient:  redisClient,
	}
}

func (s *reportService) CreateReport(reason string, reporterID uint, contentType string, contentID string, contentPreview string, details string) (*models.Report, error) {
	if reason == "" || contentType == "" || contentID == "" || contentPreview == "" {
		return nil, errors.New("reason, content type, content ID, and content preview are required")
	}
	if !isValidContentType(contentType) {
		return nil, errors.New("invalid content type")
	}

	// Chuyển ContentID sang uint
	id, err := strconv.ParseUint(contentID, 10, 32)
	if err != nil {
		return nil, errors.New("invalid content ID format")
	}

	// Kiểm tra ContentID tồn tại
	switch contentType {
	case "post":
		_, err := s.postRepo.GetPostByID(uint(id))
		if err != nil {
			return nil, errors.New("invalid post ID")
		}
	case "comment":
		_, err := s.commentRepo.GetCommentByID(uint(id))
		if err != nil {
			return nil, errors.New("invalid comment ID")
		}
	case "user":
		_, err := s.userRepo.GetUserByID(uint(id))
		if err != nil {
			return nil, errors.New("invalid user ID")
		}
	case "question":
		_, err := s.questionRepo.GetQuestionByID(uint(id))
		if err != nil {
			return nil, errors.New("invalid question ID")
		}
	case "answer":
		_, err := s.answerRepo.GetAnswerByID(uint(id))
		if err != nil {
			return nil, errors.New("invalid answer ID")
		}
	}

	report := &models.Report{
		ID:             fmt.Sprintf("report-%d-%d", time.Now().Unix(), reporterID),
		Reason:         reason,
		Details:        details,
		ReporterID:     reporterID,
		ContentType:    contentType,
		ContentID:      contentID,
		ContentPreview: contentPreview,
	}

	if err := s.repo.CreateReport(report); err != nil {
		log.Printf("Failed to create report: %v", err)
		return nil, err
	}

	s.invalidateCache("reports:*")
	log.Printf("Cache invalidated for reports:* due to new report %s", report.ID)

	return report, nil
}

func (s *reportService) GetReportByID(id string) (*models.Report, error) {
	cacheKey := fmt.Sprintf("report:%s", id)
	ctx := context.Background()

	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedData struct {
				Report     models.Report
				Reporter   models.User
				ResolvedBy *models.User
			}
			if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
				log.Printf("Cache hit for report:%s", id)
				cachedData.Report.Reporter = cachedData.Reporter
				cachedData.Report.ResolvedBy = cachedData.ResolvedBy
				return &cachedData.Report, nil
			}
			log.Printf("Failed to unmarshal cache for report %s: %v", id, err)
		}
		if err != redis.Nil {
			log.Printf("Redis error for report:%s (attempt %d): %v", id, attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	report, err := s.repo.GetReportByID(id)
	if err != nil {
		log.Printf("Failed to get report %s: %v", id, err)
		return nil, err
	}

	data, err := json.Marshal(struct {
		Report     models.Report
		Reporter   models.User
		ResolvedBy *models.User
	}{*report, report.Reporter, report.ResolvedBy})
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for report:%s: %v", id, err)
		}
	}

	return report, nil
}

func (s *reportService) UpdateReport(id string, reason string, details string, resolvedByID *uint) (*models.Report, error) {
	if resolvedByID != nil {
		_, err := s.userRepo.GetUserByID(*resolvedByID)
		if err != nil {
			return nil, errors.New("invalid resolved_by user")
		}
	}

	report, err := s.repo.GetReportByID(id)
	if err != nil {
		log.Printf("Failed to get report %s: %v", id, err)
		return nil, err
	}

	if reason != "" {
		report.Reason = reason
	}
	if details != "" {
		report.Details = details
	}
	if resolvedByID != nil {
		report.ResolvedByID = resolvedByID
	}

	if err := s.repo.UpdateReport(report); err != nil {
		log.Printf("Failed to update report %s: %v", id, err)
		return nil, err
	}

	s.invalidateCache(fmt.Sprintf("report:%s", id))
	s.invalidateCache("reports:*")
	log.Printf("Cache invalidated for reports:* due to updated report %s", id)

	return report, nil
}

func (s *reportService) UpdateReportStatus(id string, status string, resolvedByID *uint) (*models.Report, error) {
	if !isValidStatus(status) {
		return nil, errors.New("invalid status")
	}

	if resolvedByID != nil {
		_, err := s.userRepo.GetUserByID(*resolvedByID)
		if err != nil {
			return nil, errors.New("invalid resolved_by user")
		}
	}

	if err := s.repo.UpdateReportStatus(id, status, resolvedByID); err != nil {
		log.Printf("Failed to update report status %s: %v", id, err)
		return nil, err
	}

	report, err := s.repo.GetReportByID(id)
	if err != nil {
		log.Printf("Failed to get updated report %s: %v", id, err)
		return nil, err
	}

	s.invalidateCache(fmt.Sprintf("report:%s", id))
	s.invalidateCache("reports:*")
	log.Printf("Cache invalidated for reports:* due to status update of report %s", id)

	return report, nil
}

func (s *reportService) DeleteReport(id string) error {
	if err := s.repo.DeleteReport(id); err != nil {
		log.Printf("Failed to delete report %s: %v", id, err)
		return err
	}

	s.invalidateCache(fmt.Sprintf("report:%s", id))
	s.invalidateCache("reports:*")
	log.Printf("Cache invalidated for reports:* due to deleted report %s", id)

	return nil
}

func (s *reportService) BatchDeleteReports(ids []string) error {
	for _, id := range ids {
		if err := s.repo.DeleteReport(id); err != nil {
			log.Printf("Failed to delete report %s: %v", id, err)
			return fmt.Errorf("failed to delete report %s: %v", id, err)
		}
	}

	s.invalidateCache("reports:*")
	log.Printf("Cache invalidated for reports:* due to batch delete")

	return nil
}

func (s *reportService) ListReports(filters map[string]interface{}) ([]models.Report, int, error) {
	cacheKey := utils.GenerateCacheKey("reports:all", 0, filters)
	ctx := context.Background()

	for attempt := 1; attempt <= 3; attempt++ {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedData struct {
				Reports []models.Report
				Total   int
			}
			if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
				log.Printf("Cache hit for %s", cacheKey)
				return cachedData.Reports, cachedData.Total, nil
			}
			log.Printf("Failed to unmarshal cache for %s: %v", cacheKey, err)
		}
		if err != redis.Nil {
			log.Printf("Redis error for %s (attempt %d): %v", cacheKey, attempt, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	reports, total, err := s.repo.List(filters)
	if err != nil {
		log.Printf("Failed to list reports: %v", err)
		return nil, 0, err
	}

	cacheData := struct {
		Reports []models.Report
		Total   int
	}{Reports: reports, Total: total}
	data, err := json.Marshal(cacheData)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		} else {
			log.Printf("Cache set for %s with %d reports", cacheKey, len(reports))
		}
	} else {
		log.Printf("Failed to marshal reports for cache: %v", err)
	}

	return reports, total, nil
}

func isValidContentType(ct string) bool {
	validTypes := []string{"post", "comment", "user", "question", "answer"}
	for _, t := range validTypes {
		if ct == t {
			return true
		}
	}
	return false
}

func isValidStatus(status string) bool {
	validStatuses := []string{string(models.PendingStatus), string(models.ResolvedStatus), string(models.DismissedStatus)}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func (s *reportService) invalidateCache(pattern string) {
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Failed to scan cache keys for pattern %s: %v", pattern, err)
		return
	}
	if len(keys) == 0 {
		log.Printf("No cache keys found for pattern %s", pattern)
		return
	}
	if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
		log.Printf("Failed to delete cache keys %v: %v", keys, err)
	} else {
		log.Printf("Deleted cache keys %v", keys)
	}
}
