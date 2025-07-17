package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
	"log"
	"time"
)

type ReportRepository interface {
	CreateReport(report *models.Report) error
	GetReportByID(id string) (*models.Report, error)
	UpdateReport(report *models.Report) error
	UpdateReportStatus(id string, status string, resolvedByID *uint) error
	DeleteReport(id string) error
	BatchDeleteReports(ids []string) error
	List(filters map[string]interface{}) ([]models.Report, int, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) CreateReport(report *models.Report) error {
	report.CreatedAt = time.Now()
	if err := r.db.Create(report).Error; err != nil {
		return err
	}
	if err := r.db.Preload("Reporter").First(report, "id = ?", report.ID).Error; err != nil {
		return nil
	}
	return nil
}

func (r *reportRepository) GetReportByID(id string) (*models.Report, error) {
	var report models.Report
	if err := r.db.
		Preload("Reporter").
		Preload("ResolvedBy").
		First(&report, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) UpdateReport(report *models.Report) error {
	report.UpdatedAt = time.Now()
	return r.db.Save(report).Error
}

func (r *reportRepository) UpdateReportStatus(id string, status string, resolvedByID *uint) error {
	var report models.Report
	if err := r.db.First(&report, "id = ?", id).Error; err != nil {
		return err
	}
	return r.db.Model(&report).Updates(map[string]interface{}{
		"status":         status,
		"resolved_by_id": resolvedByID,
		"updated_at":     time.Now(),
	}).Error
}

func (r *reportRepository) DeleteReport(id string) error {
	return r.db.Delete(&models.Report{}, "id = ?", id).Error
}

func (r *reportRepository) BatchDeleteReports(ids []string) error {
	return r.db.Where("id IN ?", ids).Delete(&models.Report{}).Error
}

func (r *reportRepository) List(filters map[string]interface{}) ([]models.Report, int, error) {
	var reports []models.Report

	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	countQuery := r.db.Model(&models.Report{})
	query := r.db.Preload("Reporter").Preload("ResolvedBy")

	for key, value := range filters {
		if key != "limit" && key != "page" && key != "sort_by" && key != "order" {
			switch key {
			case "search":
				countQuery = countQuery.Where("reason LIKE ? OR content_preview LIKE ? OR details LIKE ?", "%"+value.(string)+"%", "%"+value.(string)+"%", "%"+value.(string)+"%")
				query = query.Where("reason LIKE ? OR content_preview LIKE ? OR details LIKE ?", "%"+value.(string)+"%", "%"+value.(string)+"%", "%"+value.(string)+"%")
			case "status":
				countQuery = countQuery.Where("status = ?", value)
				query = query.Where("status = ?", value)
			case "content_type":
				countQuery = countQuery.Where("content_type = ?", value)
				query = query.Where("content_type = ?", value)
			case "reporter_id":
				countQuery = countQuery.Where("reporter_id = ?", value)
				query = query.Where("reporter_id = ?", value)
			}
		}
	}

	if sortBy, ok := filters["sort_by"].(string); ok && sortBy != "" {
		order := "DESC"
		if ord, ok := filters["order"].(string); ok && ord == "ASC" {
			order = "ASC"
		}
		query = query.Order(sortBy + " " + order)
	} else {
		query = query.Order("created_at DESC")
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		log.Printf("Error counting reports: %v", err)
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&reports).Error; err != nil {
		log.Printf("Error fetching reports: %v", err)
		return nil, 0, err
	}

	log.Printf("Found %d reports with total %d", len(reports), total)
	return reports, int(total), nil
}
