package responses

type DashboardResponse struct {
	UserGrowth      UserGrowthData      `json:"user_growth"`
	ContentActivity ContentActivityData `json:"content_activity"`
	Engagement      EngagementData      `json:"engagement"`
	ActivityTrends  ActivityTrendsData  `json:"activity_trends"`
	TrafficSources  []TrafficSource     `json:"traffic_sources"`
	PopularContent  []PopularContent    `json:"popular_content"`
}

type UserGrowthData struct {
	TotalUsers  int     `json:"total_users"`
	ActiveUsers int     `json:"active_users"`
	NewThisWeek int     `json:"new_this_week"`
	GrowthRate  float64 `json:"growth_rate"` // e.g., 5.8
}

type ContentActivityData struct {
	TotalPosts     int `json:"total_posts"`
	TotalQuestions int `json:"total_questions"`
	TotalAnswers   int `json:"total_answers"`
	TotalTopics    int `json:"total_topics"`
}

type EngagementData struct {
	DailyActiveUsers int     `json:"daily_active_users"`
	AvgSessionMin    float64 `json:"avg_session_min"`
	RetentionRate    float64 `json:"retention_rate"` // e.g., 68.7
}

type ActivityTrendsData struct {
	Dates         []string `json:"dates"`         // e.g., ["2025-08-01", "2025-08-02", ...]
	Registrations []int    `json:"registrations"` // Counts per date
	Logins        []int    `json:"logins"`        // If tracked
	Engagements   []int    `json:"engagements"`   // e.g., posts + comments + reactions
}

type TrafficSource struct {
	Source     string  `json:"source"`     // "Direct", "Search", etc.
	Percentage float64 `json:"percentage"` // 35
	Change     float64 `json:"change"`     // 2.5
}

type PopularContent struct {
	Title      string `json:"title"`
	Type       string `json:"type"`   // "Post" or "Question"
	Author     string `json:"author"` // Username
	Views      int    `json:"views"`
	Engagement int    `json:"engagement"` // Comments + Reactions
	ID         uint   `json:"id"`         // For view action
}
