package handler

import (
	"net/http"

	"avito-pvz-service/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type CreateReceptionRequest struct {
	PVZId string `json:"pvzId" binding:"required,uuid"`
}

func CreateReceptionHandler(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
		return
	}

	role, ok := jwtClaims["role"].(string)
	if !ok || role != "staff" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен: требуется роль сотрудника ПВЗ"})
		return
	}

	var req CreateReceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing pvzId"})
		return
	}

	reception, err := repository.CreateReception(req.PVZId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reception)
}

func CloseReceptionHandler(c *gin.Context) {
	pvzId := c.Param("pvzId")
	if pvzId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing PVZ id in URL"})
		return
	}

	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
		return
	}

	role, ok := jwtClaims["role"].(string)
	if !ok || role != "staff" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен: требуется роль сотрудника ПВЗ"})
		return
	}

	reception, err := repository.CloseReception(pvzId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reception)
}