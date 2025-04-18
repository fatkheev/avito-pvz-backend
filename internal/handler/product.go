package handler

import (
	"net/http"

	"avito-pvz-service/internal/metrics"
	"avito-pvz-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type AddProductRequest struct {
	Type  string `json:"type" binding:"required,oneof=электроника одежда обувь"`
	PVZId string `json:"pvzId" binding:"required,uuid"`
}

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

	// Привязываем тело запроса
	var req AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}

	// Добавляем товар в БД
	product, err := repository.AddProduct(req.PVZId, req.Type)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Бизнес‑метрика: увеличиваем счётчик добавленных товаров
	metrics.ProductsCreatedTotal.Inc()

	c.JSON(http.StatusCreated, product)
}

func DeleteLastProductHandler(c *gin.Context) {
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

	pvzId := c.Param("pvzId")
	if pvzId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing PVZ id in URL"})
		return
	}

	if err := repository.DeleteLastProduct(pvzId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
