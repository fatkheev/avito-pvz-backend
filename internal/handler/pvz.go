package handler

import (
	"net/http"

	"avito-pvz-service/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type CreatePVZRequest struct {
	City string `json:"city" binding:"required"`
}

func CreatePVZHandler(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// Приводим к типу jwt.MapClaims
	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
		return
	}

	role, ok := jwtClaims["role"].(string)
	if !ok || role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен: требуется роль модератора"})
		return
	}

	var req CreatePVZRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing city"})
		return
	}

	pvz, err := repository.CreatePVZ(req.City)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pvz)
}
