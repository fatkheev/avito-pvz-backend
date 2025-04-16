package handler

import (
	"net/http"

	"avito-pvz-service/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AddProductRequest описывает тело запроса для добавления товара.
type AddProductRequest struct {
	Type  string `json:"type" binding:"required,oneof=электроника одежда обувь"`
	PVZId string `json:"pvzId" binding:"required,uuid"`
}

// AddProductHandler обрабатывает запрос на добавление товара в рамках
// последней открытой приёмки для указанного PVZ.
// Доступен только для пользователей с ролью "staff".
func AddProductHandler(c *gin.Context) {
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

	var req AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}

	product, err := repository.AddProduct(req.PVZId, req.Type)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}
