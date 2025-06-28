package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/api/idtoken"
	"gopkg.in/gomail.v2"
)

type AuthService interface {
	Register(username, email, password, fullname string, isVerify bool) (*models.User, error)
	Login(email, password string) (string, *models.User, error)
	ResetToken(userID uint) (string, error)
	Logout(userID uint) error
	VerifyEmailToken(token string) (*models.User, error)
	ResendVerificationEmail(email string) error
	GetUserFromToken(token string) (*models.User, error)
	HandleGoogleIDToken(idToken string) (string, *models.User, error)
}

type authService struct {
	userService    UserService
	jwtSecret      string
	redisClient    *redis.Client
	smtpHost       string
	smtpPort       int
	smtpUsername   string
	smtpPassword   string
	googleClientID string
}

func NewAuthService(u UserService, secret string, redisClient *redis.Client, smtpHost, smtpUsername, smtpPassword string, smtpPort int) AuthService {

	googleClientID := os.Getenv("YOUR_GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		slog.Error("Google Client ID is not set in environment variables")
	}

	return &authService{
		userService:    u,
		jwtSecret:      secret,
		redisClient:    redisClient,
		smtpHost:       smtpHost,
		smtpPort:       smtpPort,
		smtpUsername:   smtpUsername,
		smtpPassword:   smtpPassword,
		googleClientID: googleClientID,
	}
}

func (s *authService) HandleGoogleIDToken(idToken string) (string, *models.User, error) {
	if s.googleClientID == "" {
		slog.Error("Google Client ID is not configured")
		return "", nil, errors.New("Server configuration error")
	}

	payload, err := idtoken.Validate(context.Background(), idToken, s.googleClientID)
	if err != nil {
		slog.Error("Failed to validate Google ID token", "error", err)
		return "", nil, errors.New("Invalid Google ID token")
	}

	email, emailOk := payload.Claims["email"].(string)
	name, nameOk := payload.Claims["name"].(string)
	googleID, idOk := payload.Claims["sub"].(string)
	if !emailOk || !nameOk || !idOk || email == "" {
		slog.Error("Missing or invalid claims in Google ID token", "claims", payload.Claims)
		return "", nil, errors.New("Invalid user data from Google")
	}

	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			username := "google_" + googleID
			user, err = s.userService.CreateUser(username, email, utils.GenerateRandomString(16), name, true)
			if err != nil {
				slog.Error("Failed to create user from Google", "email", email, "error", err)
				return "", nil, err
			}
			user.EmailVerified = true
			user, err = s.userService.UpdateUser(user.ID, user.Username, user.Email, "", string(user.Role), string(user.Status), user.EmailVerified)
			if err != nil {
				slog.Error("Failed to update user from Google", "email", email, "error", err)
				return "", nil, err
			}
		} else {
			slog.Error("Failed to check user by email", "email", email, "error", err)
			return "", nil, err
		}
	}

	jwtToken, err := utils.GenerateJWT(user.ID, s.jwtSecret)
	if err != nil {
		slog.Error("Failed to generate JWT for Google user", "email", email, "error", err)
		return "", nil, err
	}

	if s.redisClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("user:status:%d", user.ID)
		if err := s.redisClient.Set(ctx, cacheKey, "active", 1*time.Hour).Err(); err != nil {
			slog.Warn("Failed to set user status in Redis", "userID", user.ID, "error", err)
		}
		cacheUserKey := fmt.Sprintf("user:%d", user.ID)
		userJSON, err := json.Marshal(user)
		if err == nil {
			if err := s.redisClient.Set(ctx, cacheUserKey, userJSON, 24*time.Hour).Err(); err != nil {
				slog.Warn("Failed to cache user data in Redis", "userID", user.ID, "error", err)
			}
		} else {
			slog.Warn("Failed to marshal user for Redis", "userID", user.ID, "error", err)
		}
	}

	return jwtToken, user, nil
}

func (s *authService) Register(username, email, password, fullname string, isVerify bool) (*models.User, error) {
	user, err := s.userService.CreateUser(username, email, password, fullname, isVerify)
	if err != nil {
		slog.Error("Failed to create user", "username", username, "error", err)
		return nil, err
	}

	verifyToken := utils.GenerateRandomString(32)
	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("verify:%s", verifyToken)
		verifyData := map[string]interface{}{
			"userID": user.ID,
			"email":  email,
		}
		verifyJSON, err := json.Marshal(verifyData)
		if err != nil {
			slog.Error("Failed to marshal verification data", "userID", user.ID, "error", err)
			return nil, err
		}
		if err := s.redisClient.Set(context.Background(), cacheKey, verifyJSON, 24*time.Hour).Err(); err != nil {
			slog.Error("Failed to set verification token in Redis", "userID", user.ID, "error", err)
			return nil, err
		}
	}

	if err := s.sendVerificationEmail(email, verifyToken, user.Username, user.FullName); err != nil {
		slog.Error("Failed to send verification email", "email", email, "error", err)
		return nil, err
	}

	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("user:%d", user.ID)
		userJSON, err := json.Marshal(user)
		if err != nil {
			slog.Warn("Failed to marshal user for Redis", "userID", user.ID, "error", err)
		} else if err := s.redisClient.Set(context.Background(), cacheKey, userJSON, 24*time.Hour).Err(); err != nil {
			slog.Warn("Failed to cache user in Redis", "userID", user.ID, "error", err)
		}
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, *models.User, error) {
	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		slog.Warn("Failed to get user by email", "email", email, "error", err)
		return "", nil, errors.New("Invalid email or password")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		slog.Warn("Invalid password attempt", "email", email)
		return "", nil, errors.New("Invalid email or password")
	}

	if !user.EmailVerified {
		slog.Warn("User email not verified", "email", email)
		return "", nil, errors.New("Email not verified. Please verify your email before logging in.")
	}

	if user.Status == "banned" {
		slog.Warn("User email is banned", "email", email)
		return "", nil, errors.New("Email is banned. Please contact admin to log in.")

	}

	_, err = s.userService.UpdateUser(user.ID, "", user.Email, "", string(user.Role), string(models.StatusActive), user.EmailVerified)
	if err != nil {
		slog.Error("Failed to update user status", "userID", user.ID, "error", err)
		return "", nil, err
	}

	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("user:status:%d", user.ID)
		if err := s.redisClient.Set(context.Background(), cacheKey, "active", 1*time.Hour).Err(); err != nil {
			slog.Warn("Failed to set user status in Redis", "userID", user.ID, "error", err)
		}
	}

	token, err := utils.GenerateJWT(user.ID, s.jwtSecret)
	if err != nil {
		slog.Error("Failed to generate JWT", "userID", user.ID, "error", err)
		return "", nil, err
	}

	if s.redisClient != nil {
		tokenKey := fmt.Sprintf("user:token:%d", user.ID)
		if err := s.redisClient.Set(context.Background(), tokenKey, token, 1*time.Hour).Err(); err != nil {
			slog.Warn("Failed to set token in Redis", "userID", user.ID, "error", err)
		}
	}

	user, err = s.userService.GetUserByID(user.ID)
	if err != nil {
		slog.Error("Failed to get user by ID after login", "userID", user.ID, "error", err)
		return "", nil, err
	}

	return token, user, nil
}

func (s *authService) ResetToken(userID uint) (string, error) {
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		slog.Error("Failed to get user by ID", "userID", userID, "error", err)
		return "", err
	}

	if !user.EmailVerified {
		slog.Warn("User email not verified", "userID", userID)
		return "", errors.New("Email not verified. Please verify your email before resetting token.")
	}

	_, err = s.userService.UpdateUser(user.ID, "", user.Email, "", string(user.Role), string(models.StatusActive), user.EmailVerified)
	if err != nil {
		slog.Error("Failed to update user status", "userID", userID, "error", err)
		return "", err
	}

	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("user:status:%d", userID)
		if err := s.redisClient.Set(context.Background(), cacheKey, "active", 1*time.Hour).Err(); err != nil {
			slog.Warn("Failed to set user status in Redis", "userID", userID, "error", err)
		}
	}

	token, err := utils.GenerateJWT(userID, s.jwtSecret)
	if err != nil {
		slog.Error("Failed to generate JWT", "userID", userID, "error", err)
		return "", err
	}

	if s.redisClient != nil {
		tokenKey := fmt.Sprintf("user:token:%d", userID)
		if err := s.redisClient.Set(context.Background(), tokenKey, token, 1*time.Hour).Err(); err != nil {
			slog.Warn("Failed to set token in Redis", "userID", userID, "error", err)
		}
	}

	return token, nil
}

func (s *authService) Logout(userID uint) error {
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		slog.Error("Failed to get user by ID", "userID", userID, "error", err)
		return err
	}

	if !user.EmailVerified {
		slog.Warn("User email not verified", "userID", userID)
		return errors.New("Email not verified. Cannot log out until verified.")
	}

	_, err = s.userService.UpdateUser(user.ID, "", user.Email, "", string(user.Role), string(models.StatusInactive), user.EmailVerified)
	if err != nil {
		slog.Error("Failed to update user status", "userID", userID, "error", err)
		return err
	}

	if s.redisClient != nil {
		ctx := context.Background()
		cacheStatusKey := fmt.Sprintf("user:status:%d", userID)
		cacheTokenKey := fmt.Sprintf("user:token:%d", userID)
		cacheUserKey := fmt.Sprintf("user:%d", user.ID)
		if err := s.redisClient.Del(ctx, cacheStatusKey, cacheTokenKey, cacheUserKey).Err(); err != nil {
			slog.Warn("Failed to delete Redis keys", "userID", userID, "error", err)
		}
	}

	return nil
}

func (s *authService) VerifyEmailToken(token string) (*models.User, error) {
	if s.redisClient == nil {
		slog.Error("Redis client not initialized")
		return nil, errors.New("Redis client not initialized")
	}

	cacheKey := fmt.Sprintf("verify:%s", token)
	ctx := context.Background()
	verifyData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil {
		slog.Error("Failed to get verification token from Redis", "token", token, "error", err)
		return nil, errors.New("Invalid or expired verification token")
	}
	slog.Info("Verification token found in Redis", "token", token)

	var data map[string]interface{}
	if err := json.Unmarshal(verifyData, &data); err != nil {
		slog.Error("Failed to unmarshal verification data", "token", token, "error", err)
		return nil, errors.New("Invalid verification data")
	}

	userID, ok := data["userID"].(float64)
	if !ok {
		slog.Error("Invalid user ID in verification data", "token", token)
		return nil, errors.New("Invalid user ID in verification data")
	}

	user, err := s.userService.GetUserByID(uint(userID))
	if err != nil {
		slog.Error("Failed to get user by ID", "userID", userID, "error", err)
		return nil, err
	}

	if user.EmailVerified {
		slog.Info("Email already verified", "userID", userID)
		return nil, errors.New("Email already verified")
	}

	verified := true
	updatedUser, err := s.userService.UpdateUser(uint(userID), user.Username, user.Email, "", string(user.Role), string(user.Status), verified)
	if err != nil {
		slog.Error("Failed to update user", "userID", userID, "error", err)
		return nil, err
	}
	slog.Info("User email verified successfully", "userID", userID)

	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		slog.Warn("Failed to delete verification token from Redis", "token", token, "error", err)
	}

	return updatedUser, nil
}

func (s *authService) ResendVerificationEmail(email string) error {
	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		slog.Error("Failed to get user by email", "email", email, "error", err)
		return errors.New("User not found")
	}

	if user.EmailVerified {
		slog.Warn("Email already verified", "email", email)
		return errors.New("Email already verified")
	}

	verifyToken := utils.GenerateRandomString(32)
	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("verify:%s", verifyToken)
		verifyData := map[string]interface{}{
			"userID": user.ID,
			"email":  email,
		}
		verifyJSON, err := json.Marshal(verifyData)
		if err != nil {
			slog.Error("Failed to marshal verification data", "userID", user.ID, "error", err)
			return err
		}
		if err := s.redisClient.Set(context.Background(), cacheKey, verifyJSON, 24*time.Hour).Err(); err != nil {
			slog.Error("Failed to set verification token in Redis", "userID", user.ID, "error", err)
			return err
		}
	}

	if err := s.sendVerificationEmail(email, verifyToken, user.Username, user.FullName); err != nil {
		slog.Error("Failed to send verification email", "email", email, "error", err)
		return err
	}

	return nil
}

func (s *authService) sendVerificationEmail(email, verifyToken, username, fullName string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.smtpUsername)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Email Verification")
	verificationLink := fmt.Sprintf("http://localhost:5000/api/verify-email?token=%s", verifyToken)
	m.SetBody("text/html", fmt.Sprintf(`
		<div style="background-color: #f9f9f9; padding: 20px; font-family: Arial, sans-serif;">
			<table align="center" cellpadding="0" cellspacing="0" style="max-width: 600px; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);">
				<tr>
					<td style="text-align: center;">
						<h2 style="color: #2d3748; margin-bottom: 10px;">Hello %s,</h2>
						<p style="color: #666666; margin-top: 5px;">Congratulations on successfully registering!</p>
					</td>
				</tr>
				<tr>
					<td style="padding: 20px 0; font-size: 16px; line-height: 1.6; color: #333333;">
						<p>To verify your email and complete the registration process, please click the link below:</p>
					</td>
				</tr>
				<tr>
					<td style="text-align: center; padding: 20px;">
						<a href="%s"
							style="background-color: #4caf50; color: #ffffff; padding: 14px 28px; text-decoration: none; border-radius: 50px; font-size: 18px; box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1); display: inline-block;">
							Verify My Email
						</a>
					</td>
				</tr>
				<tr>
					<td style="padding-top: 20px; font-size: 14px; color: #888888; text-align: center; border-top: 1px solid #eeeeee;">
						<p>If you did not request this verification, please ignore this email.</p>
						<p>Thank you for trusting our services.</p>
					</td>
				</tr>
				<tr>
					<td style="text-align: center; font-size: 12px; color: #aaaaaa; padding-top: 20px;">
						<p>© %d KatzForum. All rights reserved.</p>
					</td>
				</tr>
			</table>
		</div>
	`, username, verificationLink, time.Now().Year()))

	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUsername, s.smtpPassword)
	if err := d.DialAndSend(m); err != nil {
		slog.Error("Failed to send verification email", "email", email, "error", err)
		return err
	}
	slog.Info("Verification email sent successfully", "email", email)
	return nil
}

func (s *authService) GetUserFromToken(token string) (*models.User, error) {
	claims, err := utils.ParseJWT(token, s.jwtSecret)
	if err != nil {
		slog.Error("Failed to parse JWT", "error", err)
		return nil, errors.New("Invalid token")
	}

	user, err := s.userService.GetUserByID(claims.UserID)
	if err != nil {
		slog.Error("Failed to get user by ID", "userID", claims.UserID, "error", err)
		return nil, err
	}

	return user, nil
}
