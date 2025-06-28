package controllers

import (
	"Forum_BE/models"
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ReportController struct {
	reportService services.ReportService
}

func NewReportController(r services.ReportService) *ReportController {
	return &ReportController{reportService: r}
}

func (rc *ReportController) CreateReport(c *gin.Context) {
	var req struct {
		Reason         string                 `json:"reason" binding:"required"`
		Details        string                 `json:"details"`
		ContentType    string                 `json:"content_type" binding:"required"`
		ContentID      string                 `json:"content_id" binding:"required"`
		ContentPreview string                 `json:"content_preview" binding:"required"`
		Metadata       map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")

	report, err := rc.reportService.CreateReport(req.Reason, userID, req.ContentType, req.ContentID, req.ContentPreview, req.Details, req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := responses.ToReportResponse(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process report response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Report created successfully",
		"report":  resp,
	})
}

func (rc *ReportController) GetReportById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	report, err := rc.reportService.GetReportByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	resp, err := responses.ToReportResponse(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process report response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": resp,
	})
}

func (rc *ReportController) DeleteReport(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	if err := rc.reportService.DeleteReport(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Report deleted successfully",
	})
}

func (rc *ReportController) UpdateReport(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	var req struct {
		Reason     string `json:"reason"`
		Details    string `json:"details"`
		ResolvedBy *uint  `json:"resolved_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := rc.reportService.UpdateReport(id, req.Reason, req.Details, req.ResolvedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := responses.ToReportResponse(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process report response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Report updated successfully",
		"report":  resp,
	})
}

func (rc *ReportController) UpdateReportStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	var req struct {
		Status     string `json:"status" binding:"required"`
		ResolvedBy *uint  `json:"resolved_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	report, err := rc.reportService.UpdateReportStatus(id, req.Status, req.ResolvedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := responses.ToReportResponse(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process report response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Report status updated successfully",
		"report":  resp,
	})
}

func (rc *ReportController) ListReports(c *gin.Context) {
	filters := make(map[string]interface{})

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if status := c.Query("status"); status != "" {
		if !isValidStatus(status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		filters["status"] = status
	}
	if contentType := c.Query("content_type"); contentType != "" {
		if !isValidContentType(contentType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content type"})
			return
		}
		filters["content_type"] = contentType
	}
	if reporterID := c.Query("reporter_id"); reporterID != "" {
		if id, err := strconv.Atoi(reporterID); err == nil && id > 0 {
			filters["reporter_id"] = uint(id)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reporter ID"})
			return
		}
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filters["page"] = p
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters["limit"] = l
		}
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		if isValidSortBy(sortBy) {
			filters["sort_by"] = sortBy
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort_by field"})
			return
		}
	}
	if order := c.Query("order"); order != "" {
		if order == "ASC" || order == "DESC" {
			filters["order"] = order
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order value"})
			return
		}
	}

	reports, total, err := rc.reportService.ListReports(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reports"})
		return
	}

	var response []responses.ReportResponse
	for _, report := range reports {
		resp, err := responses.ToReportResponse(&report)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process report response"})
			return
		}
		response = append(response, resp)
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": response,
		"total":   total,
	})
}

func (rc *ReportController) BatchDeleteReports(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no report IDs provided"})
		return
	}

	if err := rc.reportService.BatchDeleteReports(req.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reports deleted successfully",
	})
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

func isValidContentType(ct string) bool {
	validTypes := []string{"post", "comment", "user", "question", "answer"}
	for _, t := range validTypes {
		if ct == t {
			return true
		}
	}
	return false
}

func isValidSortBy(sortBy string) bool {
	validFields := []string{"created_at", "updated_at", "reason", "status"}
	for _, f := range validFields {
		if sortBy == f {
			return true
		}
	}
	return false
}
