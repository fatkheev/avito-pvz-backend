package handler

import (
	"net/http"
	"os"
	"time"

	"avito-pvz-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default654пв6а45п46ва5п6ва5п46в5а1п6а5816а16в"
	}
	return []byte(secret)
}

func generateJWT(email, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  email,
		"role": role,
		"exp":  time.Now().Add(72 * time.Hour).Unix(), // токен действителен 72 часа
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// DummyLoginHandler для тестовой авторизации.
type DummyLoginRequest struct {
	Role string `json:"role" binding:"required"`
}

type DummyLoginResponse struct {
	Token string `json:"token"`
}

func DummyLoginHandler(c *gin.Context) {
	var req DummyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing role"})
		return
	}

	if req.Role != "client" && req.Role != "moderator" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role provided"})
		return
	}

	token, err := generateJWT(req.Role, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Token generation error"})
		return
	}

	c.JSON(http.StatusOK, DummyLoginResponse{Token: token})
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=client moderator"`
}

func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}

	user, err := repository.CreateUser(req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}

	user, err := repository.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	}

	// bcrypt для сравнения пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	}

	token, err := generateJWT(user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Token generation error"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
