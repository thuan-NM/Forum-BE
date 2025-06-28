package responses

import (
	"Forum_BE/models"
	"time"
)

type PermissionResponse struct {
	ID        uint   `json:"id"`
	Role      string `json:"role"`
	Resource  string `json:"resource"`
	Action    string `json:"action"`
	Allowed   bool   `json:"allowed"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func ToPermissionResponse(permission *models.Permission) PermissionResponse {
	return PermissionResponse{
		ID:        permission.ID,
		Role:      string(permission.Role),
		Resource:  permission.Resource,
		Action:    permission.Action,
		Allowed:   permission.Allowed,
		CreatedAt: permission.CreatedAt.Format(time.RFC3339),
		UpdatedAt: permission.UpdatedAt.Format(time.RFC3339),
	}
}
