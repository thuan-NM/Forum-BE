package responses

import (
	"Forum_BE/models"
	"time"
)

type UserResponse struct {
	ID             uint    `json:"id"`
	Username       string  `json:"username"`
	Email          string  `json:"email"`
	Role           string  `json:"role"`
	Status         string  `json:"status"`
	FullName       string  `json:"fullName"`
	Avatar         *string `json:"avatar,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	Location       *string `json:"location,omitempty"`
	Reputation     uint    `json:"reputation"`
	FollowersCount uint    `json:"followersCount"`
	FollowingCount uint    `json:"followingCount"`
	LastLogin      *string `json:"lastLogin,omitempty"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
	EmailVerified  bool    `json:"emailVerified"`
	PostCount      int64   `json:"postCount"`
	AnswerCount    int64   `json:"answerCount"`
	QuestionCount  int64   `json:"questionCount"`
}

func ToUserResponse(user *models.User) UserResponse {
	var lastLogin *string
	if user.LastLogin != nil {
		formatted := user.LastLogin.Format(time.RFC3339)
		lastLogin = &formatted
	}

	return UserResponse{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Role:           string(user.Role),
		Status:         string(user.Status),
		Avatar:         user.Avatar,
		FullName:       user.FullName,
		Bio:            user.Bio,
		Location:       user.Location,
		Reputation:     user.Reputation,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
		LastLogin:      lastLogin,
		CreatedAt:      user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      user.UpdatedAt.Format(time.RFC3339),
		EmailVerified:  user.EmailVerified,
		PostCount:      user.PostCount,
		AnswerCount:    user.AnswerCount,
		QuestionCount:  user.QuestionCount,
	}
}
