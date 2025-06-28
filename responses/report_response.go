package responses

import (
	"Forum_BE/models"
	"errors"
	"time"
)

type UserDataResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type ReportResponse struct {
	ID             string            `json:"id"`
	Reason         string            `json:"reason"`
	Details        string            `json:"details,omitempty"`
	Reporter       UserDataResponse  `json:"reporter"`
	ContentType    string            `json:"contentType"`
	ContentID      string            `json:"contentId"`
	ContentPreview string            `json:"contentPreview"`
	Status         string            `json:"status"`
	ResolvedBy     *UserDataResponse `json:"resolvedBy,omitempty"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt,omitempty"`
}

func ToReportResponse(report *models.Report) (ReportResponse, error) {
	if report.Reporter.ID == 0 {
		return ReportResponse{}, errors.New("reporter data not loaded")
	}

	reporter := UserDataResponse{
		ID:       report.Reporter.ID,
		Username: report.Reporter.Username,
		FullName: report.Reporter.FullName,
	}

	var resolvedBy *UserDataResponse
	if report.ResolvedBy != nil && report.ResolvedBy.ID != 0 {
		resolvedByUser := UserDataResponse{
			ID:       report.ResolvedBy.ID,
			Username: report.ResolvedBy.Username,
			FullName: report.ResolvedBy.FullName,
		}
		resolvedBy = &resolvedByUser
	}

	if report.CreatedAt.IsZero() {
		return ReportResponse{}, errors.New("invalid created_at timestamp")
	}

	return ReportResponse{
		ID:             report.ID,
		Reason:         report.Reason,
		Details:        report.Details,
		Reporter:       reporter,
		ContentType:    report.ContentType,
		ContentID:      report.ContentID,
		ContentPreview: report.ContentPreview,
		Status:         string(report.Status),
		ResolvedBy:     resolvedBy,
		CreatedAt:      report.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      report.UpdatedAt.Format(time.RFC3339),
	}, nil
}
