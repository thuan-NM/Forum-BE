package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
	"time"
)

type TrafficSource struct {
	Source     string  `json:"source"`
	Count      int     `json:"count"`      // For aggregation
	Percentage float64 `json:"percentage"` // Calculated
	Change     float64 `json:"change"`     // Calculated or stored
}

type PostOrQuestion struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	UserID     uint   `json:"user_id"`
	Views      int    `json:"views"`      // Assume added column to Post/Question, or 0 if not
	Engagement int    `json:"engagement"` // Number of reactions + comments/answers
}

type DashboardRepository interface {
	GetTotalUsers() (int64, error)
	GetActiveUsers(lastDays int) (int64, error) // Users with status = active
	GetNewUsersThisWeek() (int64, error)
	GetUserGrowthRate(prevPeriod int, currPeriod int) (float64, error) // Compare periods
	GetTotalPosts() (int64, error)
	GetTotalQuestions() (int64, error)
	GetTotalAnswers() (int64, error)
	GetTotalTopics() (int64, error)
	GetDailyActiveUsers() (int64, error)     // Users logged in today
	GetAvgSessionDuration() (float64, error) // Return placeholder since no Session model
	GetRetentionRate() (float64, error)      // Simple: Active / Total * 100
	GetActivityTrends(startDate, endDate time.Time) ([]struct {
		Date          time.Time
		Registrations int
		Logins        int // If tracked via last_login or separate log
		Engagements   int // Posts + Comments + Answers + Reactions
	}, error)
	GetTrafficSources() ([]TrafficSource, error)
	// Use local struct
	GetPopularContent(limit int) ([]PostOrQuestion, error) // Use local struct
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetTotalUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

func (r *dashboardRepository) GetActiveUsers(lastDays int) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("status = ?", models.StatusActive).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) GetNewUsersThisWeek() (int64, error) {
	var count int64
	weekStart := time.Now().AddDate(0, 0, -7) // Adjust for exact week
	err := r.db.Model(&models.User{}).Where("created_at >= ?", weekStart).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) GetUserGrowthRate(prevPeriod, currPeriod int) (float64, error) {
	// Logic: (currNew - prevNew) / prevNew * 100
	prevStart := time.Now().AddDate(0, 0, -prevPeriod*2)
	prevEnd := time.Now().AddDate(0, 0, -prevPeriod)
	currStart := prevEnd
	currEnd := time.Now()

	var prevCount, currCount int64
	r.db.Model(&models.User{}).Where("created_at BETWEEN ? AND ?", prevStart, prevEnd).Count(&prevCount)
	r.db.Model(&models.User{}).Where("created_at BETWEEN ? AND ?", currStart, currEnd).Count(&currCount)

	if prevCount == 0 {
		return 0, nil
	}
	return float64(currCount-prevCount) / float64(prevCount) * 100, nil
}

func (r *dashboardRepository) GetTotalPosts() (int64, error) {
	var count int64
	err := r.db.Model(&models.Post{}).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) GetTotalQuestions() (int64, error) {
	var count int64
	err := r.db.Model(&models.Question{}).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) GetTotalAnswers() (int64, error) {
	var count int64
	err := r.db.Model(&models.Answer{}).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) GetTotalTopics() (int64, error) {
	var count int64
	err := r.db.Model(&models.Topic{}).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) GetDailyActiveUsers() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.Model(&models.User{}).Where("last_login >= ?", today).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) GetAvgSessionDuration() (float64, error) {
	return 8.5, nil // Or calculate from last_login if possible, but fake for now
}

func (r *dashboardRepository) GetRetentionRate() (float64, error) {
	total, _ := r.GetTotalUsers()
	active, _ := r.GetActiveUsers(30) // Use status active
	if total == 0 {
		return 0, nil
	}
	return float64(active) / float64(total) * 100, nil
}

func (r *dashboardRepository) GetActivityTrends(startDate, endDate time.Time) ([]struct {
	Date          time.Time
	Registrations int
	Logins        int
	Engagements   int
}, error) {
	var trends []struct {
		Date          time.Time
		Registrations int
		Logins        int
		Engagements   int
	}
	// Complex query with GROUP BY date, count registrations (user created_at), logins (last_login), engagements (union posts, comments, etc. created_at)
	// Use raw SQL for efficiency
	err := r.db.Raw(`
		SELECT date, 
		       COUNT(reg) AS registrations,
		       COUNT(log) AS logins,
		       COUNT(eng) AS engagements
		FROM (
			SELECT DATE(created_at) AS date, 1 AS reg, NULL AS log, NULL AS eng FROM users WHERE created_at BETWEEN ? AND ?
			UNION ALL
			SELECT DATE(last_login) AS date, NULL, 1 AS log, NULL FROM users WHERE last_login BETWEEN ? AND ?
			UNION ALL
			SELECT DATE(created_at) AS date, NULL, NULL, 1 AS eng FROM posts WHERE created_at BETWEEN ? AND ?
			UNION ALL
			SELECT DATE(created_at) AS date, NULL, NULL, 1 FROM comments WHERE created_at BETWEEN ? AND ?
			UNION ALL
			SELECT DATE(created_at) AS date, NULL, NULL, 1 FROM questions WHERE created_at BETWEEN ? AND ?
			UNION ALL
			SELECT DATE(created_at) AS date, NULL, NULL, 1 FROM answers WHERE created_at BETWEEN ? AND ?
			-- Add reactions if needed: SELECT DATE(created_at) AS date, NULL, NULL, 1 FROM reactions WHERE created_at BETWEEN ? AND ?
		) AS activity
		GROUP BY date
		ORDER BY date ASC
	`, startDate, endDate, startDate, endDate, startDate, endDate, startDate, endDate, startDate, endDate, startDate, endDate).Scan(&trends).Error
	return trends, err
}

func (r *dashboardRepository) GetTrafficSources() ([]TrafficSource, error) {
	// No model, hardcode or from logs. For now, return fake
	return []TrafficSource{
		{Source: "Direct", Percentage: 35, Change: 2.5},
		{Source: "Search", Percentage: 28, Change: 5.2},
		{Source: "Social", Percentage: 22, Change: -1.8},
		{Source: "Referral", Percentage: 12, Change: 3.4},
		{Source: "Email", Percentage: 3, Change: 0.5},
	}, nil
}

func (r *dashboardRepository) GetPopularContent(limit int) ([]PostOrQuestion, error) {
	// Union Post and Question, with different engagement
	// For Post: COUNT(comments) + COUNT(reactions)
	// For Question: COUNT(answers) + COUNT(follows)
	var content []PostOrQuestion
	err := r.db.Raw(`
		(SELECT id, title, 'Post' AS type, user_id, 0 AS views, 
		        (SELECT COUNT(*) FROM comments WHERE post_id = posts.id) + 
		        (SELECT COUNT(*) FROM reactions WHERE post_id = posts.id) AS engagement 
		 FROM posts
		 UNION
		 SELECT id, title, 'Question' AS type, user_id, 0 AS views, 
		        (SELECT COUNT(*) FROM answers WHERE question_id = questions.id) + 
		        (SELECT COUNT(*) FROM question_follows WHERE question_id = questions.id) AS engagement 
		 FROM questions)
		ORDER BY engagement DESC LIMIT ?
	`, limit).Scan(&content).Error // Assume no views column, set to 0
	return content, err
}
