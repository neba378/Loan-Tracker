package Usecases

import (
	"Loan_Tracker/Domain"
	repository "Loan_Tracker/Repository"
	"Loan_Tracker/infrastructure"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtKey []byte

// Initialize the jwtKey from the .env file
func init() {
	// Load the .env file
	err := godotenv.Load(".env") // Adjust the path if necessary
	if err != nil {
		panic("Error loading .env file")
	}

	// Get the JWT secret key from the environment variable
	jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtKey) == 0 {
		panic("JWT_SECRET_KEY not set in .env file")
	}
}

type UserUsecase interface {
	Register(input Domain.RegisterInput) (*Domain.User, error)
	DeleteUser(username string) error
	Login(c *gin.Context, LoginUser *Domain.LoginInput) (string, string, error)
	Logout(tokenString string) error
	ForgotPassword(c *gin.Context, username string) (string, error)
	Reset(c *gin.Context, token string) (string, error)
	UpdatePassword(username string, newPassword string) error
	Verify(token string) error
	FindUser(id string) (Domain.User, error)
	InsertToken(username string, accessToken string, refreshToken string) error
	GetAllUsers() ([]Domain.User, error)
}

type userUsecase struct {
	userRepo        repository.UserRepository
	logRepo         repository.LogRepository
	emailService    *infrastructure.EmailService
	passwordService *infrastructure.PasswordService
}

func NewUserUsecase(userRepo repository.UserRepository, logRepo repository.LogRepository, emailService *infrastructure.EmailService) UserUsecase {
	return &userUsecase{
		userRepo:        userRepo,
		logRepo:         logRepo,
		emailService:    emailService,
		passwordService: infrastructure.NewPasswordService(),
	}
}

const (
	passwordMinLength = 8
	passwordMaxLength = 20
)

func (u *userUsecase) InsertToken(username string, accessToken string, refreshToken string) error {
	err := u.userRepo.InsertToken(username, accessToken, refreshToken)
	if err != nil {
		return err

	}
	return nil
}

func (u *userUsecase) Register(input Domain.RegisterInput) (*Domain.User, error) {
	// Validate username
	if strings.Contains(input.Username, "@") {
		return nil, errors.New("username must not contain '@'")
	}

	// Check if username already exists
	if _, err := u.userRepo.FindByUsername(input.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	// Validate email format
	if !isValidEmail(input.Email) {
		return nil, errors.New("invalid email format")
	}

	// Check if email already registered
	if _, err := u.userRepo.FindByEmail(input.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	// Validate password strength
	if err := validatePasswordStrength(input.Password); err != nil {
		return nil, err
	}

	// Hash the password
	hashedPassword, err := u.passwordService.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create new user
	user := &Domain.User{
		ID:             primitive.NewObjectID(),
		Name:           input.Name,
		Username:       input.Username,
		Email:          input.Email,
		Password:       string(hashedPassword),
		ProfilePicture: input.ProfilePicture,
		IsActive:       false, // Initially inactive
	}

	// Set user role based on database state
	if ok, err := u.userRepo.IsDbEmpty(); ok && err == nil {
		user.Role = "admin"
	} else {
		user.Role = "user"
	}

	// Save user to repository
	err = u.userRepo.Save(user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %v", err)
	}

	// Generate a verification token
	newToken, err := infrastructure.GenerateResetToken(user.Username, user.Role, jwtKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %v", err)
	}

	// Construct the email body
	subject := "Welcome to Our Service!"
	body := fmt.Sprintf("Hi %s,\n\nWelcome to our platform! Please verify your account by clicking the link below:\n\nhttp://localhost:8080/users/verify-email/%s\n\nThank you!", input.Name, newToken)

	// Send verification email
	err = u.emailService.SendEmail(input.Email, subject, body)
	if err != nil {
		return nil, fmt.Errorf("failed to send welcome email: %v", err)
	}

	return user, nil
}

func (u *userUsecase) UpdatePassword(username string, newPassword string) error {

	hashedPassword, err := u.passwordService.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	err = u.userRepo.Update(username, bson.M{"password": hashedPassword})
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	return nil
}

func (u *userUsecase) DeleteUser(id string) error {
	_, err := u.userRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	err = u.userRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}

func (u *userUsecase) Login(c *gin.Context, LoginUser *Domain.LoginInput) (string, string, error) {
	user, err := u.userRepo.FindByUsername(LoginUser.Username)
	if err != nil {
		// If not found by username, try to find by email
		user, err = u.userRepo.FindByEmail(LoginUser.Username)
		if err != nil {
			// Log failed login attempt
			log := &Domain.LogEntry{
				ID:        primitive.NewObjectID(),
				LogType:   "login_attempt",
				Timestamp: time.Now(),
				UserID:    "",
				Message:   fmt.Sprintf("Failed login attempt for username/email: %s", LoginUser.Username),
			}
			err = u.logRepo.Save(log)
			if err != nil {
				return "", "", fmt.Errorf("failed to log failed login attempt: %v", err)
			}
			return "", "", errors.New("invalid username or email or password")
		}
	}

	if err != nil {
		return "", "", errors.New("invalid username or password")
	}

	err = u.passwordService.ComparePasswords(user.Password, LoginUser.Password)
	if err != nil {
		// Log failed login attempt
		log := &Domain.LogEntry{
			ID:        primitive.NewObjectID(),
			LogType:   "login_attempt",
			Timestamp: time.Now(),
			UserID:    user.ID.Hex(),
			Message:   fmt.Sprintf("Failed login attempt for user %s", user.Username),
		}
		err = u.logRepo.Save(log)
		if err != nil {
			return "", "", fmt.Errorf("failed to log failed login attempt: %v", err)
		}
		return "", "", errors.New("invalid username or password")
	}

	accessToken, err := infrastructure.GenerateJWT(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %v", err)
	}

	refreshToken, err := infrastructure.GenerateRefreshToken(user.Username)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %v", err)
	}

	c.SetCookie("refresh_token", refreshToken, 3600, "/", "", false, true)

	err = u.userRepo.InsertToken(user.Username, accessToken, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to store tokens: %v", err)
	}

	if !user.IsActive {
		return "", "", fmt.Errorf("user not verified")
	}

	// Log successful login
	log := &Domain.LogEntry{
		ID:        primitive.NewObjectID(),
		LogType:   "login_attempt",
		Timestamp: time.Now(),
		UserID:    user.ID.Hex(),
		Message:   fmt.Sprintf("Successful login for user %s", user.Username),
	}
	err = u.logRepo.Save(log)
	if err != nil {
		return "", "", fmt.Errorf("failed to log successful login: %v", err)
	}

	return accessToken, refreshToken, nil
}

func (u *userUsecase) Logout(tokenString string) error {
	err := u.userRepo.ExpireToken(tokenString)
	if err != nil {
		return err
	}
	return nil
}

func (u *userUsecase) ForgotPassword(c *gin.Context, email string) (string, error) {
	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("user not found")
	}

	accessToken, err := infrastructure.GenerateJWT(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %v", err)
	}

	refreshToken, err := infrastructure.GenerateRefreshToken(user.Username)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %v", err)
	}

	c.SetCookie("refresh_token", refreshToken, 3600, "/", "", false, true)

	err = u.userRepo.InsertToken(user.Username, accessToken, refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to store tokens: %v", err)
	}

	subject := "Password Reset Request"
	body := fmt.Sprintf(`
	Hi %s,

	It seems like you requested a password reset. No worries, it happens to the best of us! You can reset your password by clicking the link below:

	<a href="http://localhost:8080/users/password-reset/%s">Reset Your Password</a>

	If you did not request a password reset, please ignore this email.

Best regards,
	Your Support Team
	`, user.Name, accessToken)

	err = u.emailService.SendEmail(user.Email, subject, body)
	if err != nil {
		return "", fmt.Errorf("failed to send reset email: %v", err)
	}

	// Log password reset request
	log := &Domain.LogEntry{
		ID:        primitive.NewObjectID(),
		LogType:   "password_reset_request",
		Timestamp: time.Now(),
		UserID:    user.ID.Hex(),
		Message:   fmt.Sprintf("Password reset requested for user %s", user.Username),
	}
	err = u.logRepo.Save(log)
	if err != nil {
		return "", fmt.Errorf("failed to log password reset request: %v", err)
	}
	return accessToken, nil
}

func (u *userUsecase) Reset(c *gin.Context, token string) (string, error) {

	claims, err := infrastructure.ParseResetToken(token, jwtKey)
	if err != nil {
		fmt.Println("Error parsing token:", err)
		return "", err
	}

	user, err := u.userRepo.FindByUsername(claims.Username)

	if err != nil {
		return "", errors.New("user not found")
	}

	accessToken, err := infrastructure.GenerateJWT(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %v", err)
	}

	refreshToken, err := infrastructure.GenerateRefreshToken(user.Username)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %v", err)
	}

	c.SetCookie("refresh_token", refreshToken, 3600, "/", "", false, true)

	err = u.userRepo.InsertToken(user.Username, accessToken, refreshToken)
	if err != nil {
		return "", errors.New("user not found")
	}
	access_token, err := infrastructure.GenerateJWT(user.ID.Hex(), user.Username, user.Role)

	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %v", err)
	}

	// Log password reset completion
	log := &Domain.LogEntry{
		ID:        primitive.NewObjectID(),
		LogType:   "password_reset_completion",
		Timestamp: time.Now(),
		UserID:    user.ID.Hex(),
		Message:   fmt.Sprintf("Password reset completed for user %s", user.Username),
	}
	err = u.logRepo.Save(log)
	if err != nil {
		return "", fmt.Errorf("failed to log password reset completion: %v", err)
	}

	return access_token, nil
}

func (u *userUsecase) Verify(token string) error {
	claims, err := infrastructure.ParseResetToken(token, jwtKey)
	if err != nil {
		fmt.Println("Error parsing token:", err)
	}

	user, err := u.userRepo.FindByUsername(claims.Username)
	if err != nil {
		return errors.New("user not found")
	}
	err = u.userRepo.Update(user.Username, bson.M{"is_active": true})
	if err != nil {
		return fmt.Errorf("failed to verify user: %v", err)
	}
	return nil
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func (u *userUsecase) FindUser(id string) (Domain.User, error) {

	user, err := u.userRepo.ShowUser(id)
	if err != nil {
		return Domain.User{}, err
	}
	return user, nil
}

func validatePasswordStrength(password string) error {
	if len(password) < passwordMinLength || len(password) > passwordMaxLength {
		return fmt.Errorf("password must be between %d and %d characters", passwordMinLength, passwordMaxLength)
	}

	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case c == '@' || c == '#' || c == '$' || c == '%' || c == '^' || c == '&' || c == '*':
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

func (uc *userUsecase) GetAllUsers() ([]Domain.User, error) {
	users, err := uc.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}
