package models

import "time"

type ActivityType string

const (
	ActivityUserCreated    ActivityType = "user_created"
	ActivityPostCreated    ActivityType = "post_created"
	ActivityCommentCreated ActivityType = "comment_created"
	ActivityTopicCreated   ActivityType = "topic_created"
)

type ActivityItem struct {
	Type      ActivityType
	Data      interface{}
	CreatedAt time.Time
}
