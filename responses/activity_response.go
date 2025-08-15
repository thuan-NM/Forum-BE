package responses

import (
	"Forum_BE/models"
	"fmt"
	"time"
)

type ActivityResponse struct {
	Type        models.ActivityType `json:"type"`
	IconColor   string              `json:"icon_color"` // e.g., "blue", "purple", "green", "orange"
	Description string              `json:"description"`
	TimeAgo     string              `json:"time_ago"`
	CreatedAt   string              `json:"created_at"` // RFC3339 format
}

func ToActivityResponses(activities []models.ActivityItem) []ActivityResponse {
	var res []ActivityResponse
	now := time.Now()
	for _, act := range activities {
		timeAgo := humanizeTime(now.Sub(act.CreatedAt))
		var desc string
		var iconColor string
		switch act.Type {
		case models.ActivityUserCreated:
			user := act.Data.(*models.User)
			desc = fmt.Sprintf("%s Tạo tài khoản", user.FullName)
			iconColor = "blue"
		case models.ActivityPostCreated:
			post := act.Data.(*models.Post)
			desc = fmt.Sprintf("%s đăng \"%s\"", post.User.FullName, post.Title)
			iconColor = "purple"
		case models.ActivityCommentCreated:
			comment := act.Data.(*models.Comment)    // Giả sử bạn có model Comment
			postTitle := "Bài viết không có tiêu đề" // Default nếu nil
			if comment.Post != nil {
				postTitle = comment.Post.Title // Preload Post
			}
			desc = fmt.Sprintf("%s bình luận trên \"%s\"", comment.User.FullName, postTitle)
			iconColor = "green"
		case models.ActivityTopicCreated:
			topic := act.Data.(*models.Topic)
			desc = fmt.Sprintf("Admin đã thêm \"%s\" chủ đề", topic.Name) // Giả sử admin tạo topic
			iconColor = "orange"
		}
		res = append(res, ActivityResponse{
			Type:        act.Type,
			IconColor:   iconColor,
			Description: desc,
			TimeAgo:     timeAgo,
			CreatedAt:   act.CreatedAt.Format(time.RFC3339),
		})
	}
	return res
}

// Helper để format time ago (giống human-readable)
func humanizeTime(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	} else if d < time.Hour {
		return fmt.Sprintf("%d phút trước", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d giờ trước", int(d.Hours()))
	} else {
		return fmt.Sprintf("%d ngày trước", int(d.Hours()/24))
	}
}
