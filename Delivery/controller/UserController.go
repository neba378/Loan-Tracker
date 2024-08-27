package controller

import (
	"Loan_Tracker/Domain"
	Usecases "Loan_Tracker/Usecase"
	"Loan_Tracker/infrastructure"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

type UserController struct {
	UserUsecase Usecases.UserUsecase
}

// NewUserController creates a new instance of UserController
func NewUserController(userUsecase Usecases.UserUsecase) *UserController {
	controller := &UserController{
		UserUsecase: userUsecase,
	}
	return controller
}

// Register handles user registration
func (uc *UserController) Register(c *gin.Context) {
	var input Domain.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := uc.UserUsecase.Register(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// DeleteUser handles deleting a user
func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	err := uc.UserUsecase.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (uc *UserController) Login(c *gin.Context) {
	var input Domain.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	accessToken, refresh_token, err := uc.UserUsecase.Login(c, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.SetCookie("refresh_token", refresh_token, 60*60*24*7, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refresh_token})
}

func (uc *UserController) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token not found"})
		return
	}

	token, err := jwt.ParseWithClaims(refreshToken, &infrastructure.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}

	// Get the username from the token
	username, err := infrastructure.GetUsernameFromToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get username from token"})
		c.Abort()
		return
	}

	// Set token claims in context
	claims, ok := token.Claims.(*infrastructure.Claims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}
	claims.Username = username

	accessToken, err := infrastructure.GenerateJWT(claims.ID, claims.Username, claims.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	refreshToken, err = infrastructure.GenerateRefreshToken(claims.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uc.UserUsecase.InsertToken(claims.Username, accessToken, refreshToken)

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func (uc *UserController) Verify(c *gin.Context) {
	token := c.Param("token")
	err := uc.UserUsecase.Verify(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (uc *UserController) FindUser(c *gin.Context) {
	id := c.Param("id") // Get user ID from the URL parameters

	user, err := uc.UserUsecase.FindUser(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if (user == Domain.User{}) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (uc *UserController) Logout(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	var token string
	parts := strings.Split(tokenString, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		token = parts[1]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
		return
	}

	err := uc.UserUsecase.Logout(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}

func (uc *UserController) ChangePassword(c *gin.Context) {
	var input Domain.ChangePasswordInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	username := c.GetString("username")

	err := uc.UserUsecase.UpdatePassword(username, input.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (uc *UserController) ResetPassword(c *gin.Context) {
	reset_token := c.Param("token")

	new_token, err := uc.UserUsecase.Reset(c, reset_token)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": new_token})

}

func (uc *UserController) ForgotPassword(c *gin.Context) {

	var input Domain.ForgetPasswordInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := uc.UserUsecase.ForgotPassword(c, input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reset password link sent to your email"})

}

// userController.go
func (uc *UserController) GetAllUsers(c *gin.Context) {
	users, err := uc.UserUsecase.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}
