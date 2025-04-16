package handler

import (
	"net/http"

	"avito-pvz-service/internal/repository"

	"github.com/gin-gonic/gin"
)

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

	token := "dummy-token-for-" + req.Role
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

	if user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	}

	// Для теста возвращаем токен в формате "dummy-token-for-{email}"
	token := "dummy-token-for-" + user.Email
	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
