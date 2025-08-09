package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
	"sort"
)

type ActivityRepository interface {
	GetRecentActivities(limit int) ([]models.ActivityItem, error)
}

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) GetRecentActivities(limit int) ([]models.ActivityItem, error) {
	var activities []models.ActivityItem

	// Query recent users
	var users []models.User
	r.db.Order("created_at desc").Limit(limit).Find(&users)
	for _, u := range users {
		activities = append(activities, models.ActivityItem{Type: models.ActivityUserCreated, Data: &u, CreatedAt: u.CreatedAt})
	}

	// Query recent posts (preload User)
	var posts []models.Post
	r.db.Preload("User").Order("created_at desc").Limit(limit).Find(&posts)
	for _, p := range posts {
		activities = append(activities, models.ActivityItem{Type: models.ActivityPostCreated, Data: &p, CreatedAt: p.CreatedAt})
	}

	// Query recent comments (preload User và Post)
	var comments []models.Comment // Giả sử model Comment có UserID, PostID, CreatedAt
	r.db.Preload("User").Preload("Post").Order("created_at desc").Limit(limit).Find(&comments)
	for _, c := range comments {
		activities = append(activities, models.ActivityItem{Type: models.ActivityCommentCreated, Data: &c, CreatedAt: c.CreatedAt})
	}

	// Query recent topics
	var topics []models.Topic
	r.db.Order("created_at desc").Limit(limit).Find(&topics)
	for _, t := range topics {
		activities = append(activities, models.ActivityItem{Type: models.ActivityTopicCreated, Data: &t, CreatedAt: t.CreatedAt})
	}

	// Sort all by created_at desc (sử dụng sort slice)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.After(activities[j].CreatedAt)
	})

	// Lấy top limit
	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}
