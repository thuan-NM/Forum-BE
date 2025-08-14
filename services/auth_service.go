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
	ChangePassword(userId uint, oldPassword, newPassword string) (*models.User, error)
}

type authService struct {
	userService    UserService
	userRepo       repositories.UserRepository
	jwtSecret      string
	redisClient    *redis.Client
	smtpHost       string
	smtpPort       int
	smtpUsername   string
	smtpPassword   string
	googleClientID string
}

func NewAuthService(u UserService, uRepo repositories.UserRepository, secret string, redisClient *redis.Client, smtpHost, smtpUsername, smtpPassword string, smtpPort int) AuthService {
	googleClientID := os.Getenv("YOUR_GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		slog.Error("Google Client ID is not set in environment variables")
	}

	return &authService{
		userService:    u,
		userRepo:       uRepo,
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
		return "", nil, errors.New("Cấu hình server lỗi")
	}

	payload, err := idtoken.Validate(context.Background(), idToken, s.googleClientID)
	if err != nil {
		slog.Error("Không thể xác thực Google ID token", "error", err)
		return "", nil, errors.New("Google ID token không hợp lệ")
	}

	email, emailOk := payload.Claims["email"].(string)
	name, nameOk := payload.Claims["name"].(string)
	googleID, idOk := payload.Claims["sub"].(string)
	if !emailOk || !nameOk || !idOk || email == "" {
		slog.Error("Thiếu hoặc dữ liệu không hợp lệ trong Google ID token", "claims", payload.Claims)
		return "", nil, errors.New("Dữ liệu người dùng từ Google không hợp lệ")
	}

	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			username := "google_" + googleID
			user, err = s.userService.CreateUser(username, email, utils.GenerateRandomString(16), name, true)
			if err != nil {
				slog.Error("Không thể tạo người dùng từ Google", "email", email, "error", err)
				return "", nil, err
			}
			updateDTO := UpdateUserDTO{
				EmailVerified: &user.EmailVerified,
			}
			user, err = s.userService.UpdateUser(user.ID, updateDTO)
			if err != nil {
				slog.Error("Không thể cập nhật người dùng từ Google", "email", email, "error", err)
				return "", nil, err
			}
		} else {
			slog.Error("Không thể kiểm tra người dùng qua email", "email", email, "error", err)
			return "", nil, err
		}
	}

	jwtToken, err := utils.GenerateJWT(user.ID, s.jwtSecret)
	if err != nil {
		slog.Error("Không thể tạo JWT cho người dùng Google", "email", email, "error", err)
		return "", nil, err
	}

	if s.redisClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("user:status:%d", user.ID)
		if err := s.redisClient.Set(ctx, cacheKey, "active", 1*time.Hour).Err(); err != nil {
			slog.Warn("Không thể lưu trạng thái người dùng vào Redis", "userID", user.ID, "error", err)
		}
		cacheUserKey := fmt.Sprintf("user:%d", user.ID)
		userJSON, err := json.Marshal(user)
		if err == nil {
			if err := s.redisClient.Set(ctx, cacheUserKey, userJSON, 24*time.Hour).Err(); err != nil {
				slog.Warn("Không thể lưu dữ liệu người dùng vào Redis", "userID", user.ID, "error", err)
			}
		} else {
			slog.Warn("Không thể mã hóa người dùng cho Redis", "userID", user.ID, "error", err)
		}
	}

	return jwtToken, user, nil
}

func (s *authService) Register(username, email, password, fullname string, isVerify bool) (*models.User, error) {
	user, err := s.userService.CreateUser(username, email, password, fullname, isVerify)
	if err != nil {
		slog.Error("Không thể tạo người dùng", "username", username, "error", err)
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
			slog.Error("Không thể mã hóa dữ liệu xác thực", "userID", user.ID, "error", err)
			return nil, err
		}
		if err := s.redisClient.Set(context.Background(), cacheKey, verifyJSON, 24*time.Hour).Err(); err != nil {
			slog.Error("Không thể lưu token xác thực vào Redis", "userID", user.ID, "error", err)
			return nil, err
		}
	}

	if err := s.sendVerificationEmail(email, verifyToken, user.Username, user.FullName); err != nil {
		slog.Error("Không thể gửi email xác thực", "email", email, "error", err)
		return nil, err
	}

	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("user:%d", user.ID)
		userJSON, err := json.Marshal(user)
		if err != nil {
			slog.Warn("Không thể mã hóa người dùng cho Redis", "userID", user.ID, "error", err)
		} else if err := s.redisClient.Set(context.Background(), cacheKey, userJSON, 24*time.Hour).Err(); err != nil {
			slog.Warn("Không thể lưu người dùng vào Redis", "userID", user.ID, "error", err)
		}
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, *models.User, error) {
	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		slog.Warn("Không thể lấy người dùng qua email", "email", email, "error", err)
		return "", nil, errors.New("Email hoặc mật khẩu không hợp lệ")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		slog.Warn("Mật khẩu không đúng", "email", email)
		return "", nil, errors.New("Email hoặc mật khẩu không hợp lệ")
	}

	if !user.EmailVerified {
		slog.Warn("Email người dùng chưa được xác thực", "email", email)
		return "", nil, errors.New("Email chưa được xác thực. Vui lòng xác thực email trước khi đăng nhập.")
	}

	if user.Status == "banned" {
		slog.Warn("Email người dùng đã bị cấm", "email", email)
		return "", nil, errors.New("Email đã bị cấm. Vui lòng liên hệ admin để đăng nhập.")
	}

	// Cập nhật LastLogin
	now := time.Now()
	//lastLoginStr := now.Format(time.RFC3339)
	updateDTO := UpdateUserDTO{
		Status:    stringPtr(string(models.StatusActive)),
		LastLogin: &now,
	}
	_, err = s.userService.UpdateUser(user.ID, updateDTO)
	if err != nil {
		slog.Error("Không thể cập nhật trạng thái và last_login người dùng", "userID", user.ID, "error", err)
		return "", nil, err
	}

	if s.redisClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("user:status:%d", user.ID)
		if err := s.redisClient.Set(ctx, cacheKey, "active", 1*time.Hour).Err(); err != nil {
			slog.Warn("Không thể lưu trạng thái người dùng vào Redis", "userID", user.ID, "error", err)
		}
		cacheUserKey := fmt.Sprintf("user:%d", user.ID)
		userJSON, err := json.Marshal(user)
		if err == nil {
			if err := s.redisClient.Set(ctx, cacheUserKey, userJSON, 24*time.Hour).Err(); err != nil {
				slog.Warn("Không thể lưu dữ liệu người dùng vào Redis", "userID", user.ID, "error", err)
			}
		} else {
			slog.Warn("Không thể mã hóa người dùng cho Redis", "userID", user.ID, "error", err)
		}
	}

	token, err := utils.GenerateJWT(user.ID, s.jwtSecret)
	if err != nil {
		slog.Error("Không thể tạo JWT", "userID", user.ID, "error", err)
		return "", nil, err
	}

	if s.redisClient != nil {
		tokenKey := fmt.Sprintf("user:token:%d", user.ID)
		if err := s.redisClient.Set(context.Background(), tokenKey, token, 1*time.Hour).Err(); err != nil {
			slog.Warn("Không thể lưu token vào Redis", "userID", user.ID, "error", err)
		}
	}

	user, err = s.userService.GetUserByID(user.ID)
	if err != nil {
		slog.Error("Không thể lấy người dùng sau khi đăng nhập", "userID", user.ID, "error", err)
		return "", nil, err
	}

	return token, user, nil
}

func (s *authService) ChangePassword(userId uint, oldPassword, newPassword string) (*models.User, error) {
	// Lấy user từ DB
	user, err := s.userRepo.GetUserByIDWithPassword(userId)
	if err != nil {
		slog.Error("Không tìm thấy người dùng", "userID", userId, "error", err)
		return nil, errors.New("Không tìm thấy người dùng")
	}

	// Kiểm tra mật khẩu cũ
	if !utils.CheckPasswordHash(oldPassword, user.Password) {
		return nil, errors.New("Mật khẩu cũ không chính xác")
	}

	// Kiểm tra nếu mật khẩu mới trùng với mật khẩu cũ
	if oldPassword == newPassword {
		return nil, errors.New("Mật khẩu mới không được trùng với mật khẩu cũ")
	}

	// Hash mật khẩu mới
	hashedNewPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		slog.Error("Không thể hash mật khẩu mới", "userID", userId, "error", err)
		return nil, errors.New("Không thể xử lý mật khẩu mới")
	}

	// Cập nhật mật khẩu
	updateDTO := UpdateUserDTO{Password: &hashedNewPassword}
	if _, err := s.userService.UpdateUser(userId, updateDTO); err != nil {
		slog.Error("Cập nhật mật khẩu thất bại", "userID", userId, "error", err)
		return nil, errors.New("Đổi mật khẩu thất bại")
	}

	// Xóa cache Redis nếu có
	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("user:%d", userId)
		cacheStatusKey := fmt.Sprintf("user:status:%d", userId)
		cacheTokenKey := fmt.Sprintf("user:token:%d", userId)
		if err := s.redisClient.Del(context.Background(), cacheKey, cacheStatusKey, cacheTokenKey).Err(); err != nil {
			slog.Warn("Không thể xóa cache sau khi đổi mật khẩu", "userID", userId, "error", err)
		}
	}

	// Lấy lại thông tin user mới nhất
	updatedUser, err := s.userService.GetUserByID(userId)
	if err != nil {
		slog.Error("Không thể lấy lại người dùng sau khi đổi mật khẩu", "userID", userId, "error", err)
		return nil, err
	}

	return updatedUser, nil
}

func (s *authService) ResetToken(userID uint) (string, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		slog.Error("Không thể lấy người dùng qua ID", "userID", userID, "error", err)
		return "", err
	}

	if !user.EmailVerified {
		slog.Warn("Email người dùng chưa được xác thực", "userID", userID)
		return "", errors.New("Email chưa được xác thực. Vui lòng xác thực email trước khi đặt lại token.")
	}

	updateDTO := UpdateUserDTO{
		Status: stringPtr(string(models.StatusActive)),
	}
	_, err = s.userService.UpdateUser(user.ID, updateDTO)
	if err != nil {
		slog.Error("Không thể cập nhật trạng thái người dùng", "userID", userID, "error", err)
		return "", err
	}

	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("user:status:%d", userID)
		if err := s.redisClient.Set(context.Background(), cacheKey, "active", 1*time.Hour).Err(); err != nil {
			slog.Warn("Không thể lưu trạng thái người dùng vào Redis", "userID", userID, "error", err)
		}
	}

	token, err := utils.GenerateJWT(userID, s.jwtSecret)
	if err != nil {
		slog.Error("Không thể tạo JWT", "userID", userID, "error", err)
		return "", err
	}

	if s.redisClient != nil {
		tokenKey := fmt.Sprintf("user:token:%d", userID)
		if err := s.redisClient.Set(context.Background(), tokenKey, token, 1*time.Hour).Err(); err != nil {
			slog.Warn("Không thể lưu token vào Redis", "userID", userID, "error", err)
		}
	}

	return token, nil
}

func (s *authService) Logout(userID uint) error {
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		slog.Error("Không thể lấy người dùng qua ID", "userID", userID, "error", err)
		return err
	}

	if !user.EmailVerified {
		slog.Warn("Email người dùng chưa được xác thực", "userID", userID)
		return errors.New("Email chưa được xác thực. Không thể đăng xuất cho đến khi xác thực.")
	}

	updateDTO := UpdateUserDTO{
		Status: stringPtr(string(models.StatusInactive)),
	}
	_, err = s.userService.UpdateUser(user.ID, updateDTO)
	if err != nil {
		slog.Error("Không thể cập nhật trạng thái người dùng", "userID", userID, "error", err)
		return err
	}

	if s.redisClient != nil {
		ctx := context.Background()
		cacheStatusKey := fmt.Sprintf("user:status:%d", userID)
		cacheTokenKey := fmt.Sprintf("user:token:%d", userID)
		cacheUserKey := fmt.Sprintf("user:%d", user.ID)
		if err := s.redisClient.Del(ctx, cacheStatusKey, cacheTokenKey, cacheUserKey).Err(); err != nil {
			slog.Warn("Không thể xóa các khóa Redis", "userID", userID, "error", err)
		}
	}

	return nil
}

func (s *authService) VerifyEmailToken(token string) (*models.User, error) {
	if s.redisClient == nil {
		slog.Error("Redis client chưa được khởi tạo")
		return nil, errors.New("Redis client chưa được khởi tạo")
	}

	cacheKey := fmt.Sprintf("verify:%s", token)
	ctx := context.Background()
	verifyData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil {
		slog.Error("Không thể lấy token xác thực từ Redis", "token", token, "error", err)
		return nil, errors.New("Token xác thực không hợp lệ hoặc đã hết hạn")
	}
	slog.Info("Tìm thấy token xác thực trong Redis", "token", token)

	var data map[string]interface{}
	if err := json.Unmarshal(verifyData, &data); err != nil {
		slog.Error("Không thể giải mã dữ liệu xác thực", "token", token, "error", err)
		return nil, errors.New("Dữ liệu xác thực không hợp lệ")
	}

	userID, ok := data["userID"].(float64)
	if !ok {
		slog.Error("ID người dùng không hợp lệ trong dữ liệu xác thực", "token", token)
		return nil, errors.New("ID người dùng không hợp lệ trong dữ liệu xác thực")
	}

	user, err := s.userService.GetUserByID(uint(userID))
	if err != nil {
		slog.Error("Không thể lấy người dùng qua ID", "userID", userID, "error", err)
		return nil, err
	}

	if user.EmailVerified {
		slog.Info("Email đã được xác thực", "userID", userID)
		return nil, errors.New("Email đã được xác thực")
	}

	verified := true
	updateDTO := UpdateUserDTO{
		EmailVerified: &verified,
	}
	updatedUser, err := s.userService.UpdateUser(uint(userID), updateDTO)
	if err != nil {
		slog.Error("Không thể cập nhật người dùng", "userID", userID, "error", err)
		return nil, err
	}
	slog.Info("Xác thực email người dùng thành công", "userID", userID)

	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		slog.Warn("Không thể xóa token xác thực từ Redis", "token", token, "error", err)
	}

	return updatedUser, nil
}

func (s *authService) ResendVerificationEmail(email string) error {
	user, err := s.userService.GetUserByEmail(email)
	if err != nil {
		slog.Error("Không thể lấy người dùng qua email", "email", email, "error", err)
		return errors.New("Không tìm thấy người dùng")
	}

	if user.EmailVerified {
		slog.Warn("Email đã được xác thực", "email", email)
		return errors.New("Email đã được xác thực")
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
			slog.Error("Không thể mã hóa dữ liệu xác thực", "userID", user.ID, "error", err)
			return err
		}
		if err := s.redisClient.Set(context.Background(), cacheKey, verifyJSON, 24*time.Hour).Err(); err != nil {
			slog.Error("Không thể lưu token xác thực vào Redis", "userID", user.ID, "error", err)
			return err
		}
	}

	if err := s.sendVerificationEmail(email, verifyToken, user.Username, user.FullName); err != nil {
		slog.Error("Không thể gửi email xác thực", "email", email, "error", err)
		return err
	}

	return nil
}

func (s *authService) sendVerificationEmail(email, verifyToken, username, fullName string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.smtpUsername)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Xác thực Email")
	verificationLink := fmt.Sprintf("http://localhost:5000/api/verify-email?token=%s", verifyToken)
	m.SetBody("text/html", fmt.Sprintf(`
       <div style="background-color: #f9f9f9; padding: 20px; font-family: Arial, sans-serif;">
          <table align="center" cellpadding="0" cellspacing="0" style="max-width: 600px; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);">
             <tr>
                <td style="text-align: center;">
                   <h2 style="color: #2d3748; margin-bottom: 10px;">Xin chào %s,</h2>
                   <p style="color: #666666; margin-top: 5px;">Chúc mừng bạn đã đăng ký thành công!</p>
                </td>
             </tr>
             <tr>
                <td style="padding: 20px 0; font-size: 16px; line-height: 1.6; color: #333333;">
                   <p>Để xác thực email và hoàn tất quá trình đăng ký, vui lòng nhấp vào liên kết bên dưới:</p>
                </td>
             </tr>
             <tr>
                <td style="text-align: center; padding: 20px;">
                   <a href="%s"
                      style="background-color: #4caf50; color: #ffffff; padding: 14px 28px; text-decoration: none; border-radius: 50px; font-size: 18px; box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1); display: inline-block;">
                      Xác thực Email
                   </a>
                </td>
             </tr>
             <tr>
                <td style="padding-top: 20px; font-size: 14px; color: #888888; text-align: center; border-top: 1px solid #eeeeee;">
                   <p>Nếu bạn không yêu cầu xác thực này, vui lòng bỏ qua email này.</p>
                   <p>Cảm ơn bạn đã tin tưởng dịch vụ của chúng tôi.</p>
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
		slog.Error("Không thể gửi email xác thực", "email", email, "error", err)
		return err
	}
	slog.Info("Gửi email xác thực thành công", "email", email)
	return nil
}

func (s *authService) GetUserFromToken(token string) (*models.User, error) {
	claims, err := utils.ParseJWT(token, s.jwtSecret)
	if err != nil {
		slog.Error("Không thể phân tích JWT", "error", err)
		return nil, errors.New("Token không hợp lệ")
	}

	user, err := s.userService.GetUserByID(claims.UserID)
	if err != nil {
		slog.Error("Không thể lấy người dùng qua ID", "userID", claims.UserID, "error", err)
		return nil, err
	}

	return user, nil
}

// Hàm tiện ích để tạo con trỏ string
func stringPtr(s string) *string {
	return &s
}
