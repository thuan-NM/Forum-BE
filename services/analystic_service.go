package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/responses"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"time"
)

type DashboardService interface {
	GetDashboardData(period string) (*responses.DashboardResponse, error)
}

type dashboardService struct {
	dashRepo    repositories.DashboardRepository
	redisClient *redis.Client
	db          *gorm.DB // Add db to service for fetching user
}

func NewDashboardService(repo repositories.DashboardRepository, redisClient *redis.Client, db *gorm.DB) DashboardService {
	return &dashboardService{dashRepo: repo, redisClient: redisClient, db: db}
}

func (s *dashboardService) GetDashboardData(period string) (*responses.DashboardResponse, error) {
	cacheKey := fmt.Sprintf("dashboard:%s", period)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var data responses.DashboardResponse
		if err := json.Unmarshal([]byte(cached), &data); err == nil {
			return &data, nil
		}
	}

	// Calculate period days
	days := 7
	switch period {
	case "30days":
		days = 30
	case "90days":
		days = 90
	}
	startDate := time.Now().AddDate(0, 0, -days)

	// Fetch data
	totalUsers, _ := s.dashRepo.GetTotalUsers()
	activeUsers, _ := s.dashRepo.GetActiveUsers(days)
	newThisWeek, _ := s.dashRepo.GetNewUsersThisWeek()
	growthRate, _ := s.dashRepo.GetUserGrowthRate(days, days) // Prev same period

	totalPosts, _ := s.dashRepo.GetTotalPosts()
	totalQuestions, _ := s.dashRepo.GetTotalQuestions()
	totalAnswers, _ := s.dashRepo.GetTotalAnswers()
	totalTopics, _ := s.dashRepo.GetTotalTopics()

	dailyActive, _ := s.dashRepo.GetDailyActiveUsers()
	avgSession, _ := s.dashRepo.GetAvgSessionDuration()
	retention, _ := s.dashRepo.GetRetentionRate()

	trends, _ := s.dashRepo.GetActivityTrends(startDate, time.Now())
	var dates []string
	var regs, logs, engs []int
	for _, t := range trends {
		dates = append(dates, t.Date.Format("2006-01-02"))
		regs = append(regs, t.Registrations)
		logs = append(logs, t.Logins)
		engs = append(engs, t.Engagements)
	}

	traffic, _ := s.dashRepo.GetTrafficSources()
	var trafficRes []responses.TrafficSource
	for _, src := range traffic {
		// Calculate percentage, change - fake or from data
		trafficRes = append(trafficRes, responses.TrafficSource{
			Source:     src.Source,
			Percentage: src.Percentage,
			Change:     src.Change,
		})
	}

	popular, _ := s.dashRepo.GetPopularContent(5)
	var popularRes []responses.PopularContent
	for _, p := range popular {
		author := "Unknown"
		// Fetch author username
		var user models.User
		s.db.First(&user, p.UserID)
		author = user.Username
		popularRes = append(popularRes, responses.PopularContent{
			Title:      p.Title,
			Type:       p.Type,
			Author:     author,
			Views:      p.Views,
			Engagement: p.Engagement,
			ID:         p.ID,
		})
	}

	resp := &responses.DashboardResponse{
		UserGrowth: responses.UserGrowthData{
			TotalUsers:  int(totalUsers),
			ActiveUsers: int(activeUsers),
			NewThisWeek: int(newThisWeek),
			GrowthRate:  growthRate,
		},
		ContentActivity: responses.ContentActivityData{
			TotalPosts:     int(totalPosts),
			TotalQuestions: int(totalQuestions),
			TotalAnswers:   int(totalAnswers),
			TotalTopics:    int(totalTopics),
		},
		Engagement: responses.EngagementData{
			DailyActiveUsers: int(dailyActive),
			AvgSessionMin:    avgSession,
			RetentionRate:    retention,
		},
		ActivityTrends: responses.ActivityTrendsData{
			Dates:         dates,
			Registrations: regs,
			Logins:        logs,
			Engagements:   engs,
		},
		TrafficSources: trafficRes,
		PopularContent: popularRes,
	}

	data, _ := json.Marshal(resp)
	s.redisClient.Set(ctx, cacheKey, data, 10*time.Minute)

	return resp, nil
}
