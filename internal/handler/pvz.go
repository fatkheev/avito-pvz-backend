package handler

import (
	"log"
	"net/http"

	"avito-pvz-service/internal/metrics"
	"avito-pvz-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type CreatePVZRequest struct {
	City string `json:"city" binding:"required"`
}

func CreatePVZHandler(c *gin.Context) {
	log.Println("Создание ПВЗ: начало")

	claims, exists := c.Get("user")
	if !exists {
		log.Println("Создание ПВЗ: неавторизован")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		log.Println("Создание ПВЗ: некорректные токен-клеймы")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
		return
	}
	role, ok := jwtClaims["role"].(string)
	if !ok || role != "moderator" {
		log.Println("Создание ПВЗ: нет прав (нужна роль модератор)")
		c.JSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен: требуется роль модератора"})
		return
	}

	// привязка JSON
	var req CreatePVZRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Создание ПВЗ: неверный запрос:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing city"})
		return
	}
	log.Printf("Создание ПВЗ: город=%s\n", req.City)

	// создание ПВЗ в БД
	pvz, err := repository.CreatePVZ(req.City)
	if err != nil {
		log.Println("Создание ПВЗ: ошибка создания:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// метрика
	metrics.PVZCreatedTotal.Inc()
	log.Printf("Создание ПВЗ: успешно, id=%s, город=%s\n", pvz.ID, pvz.City)

	c.JSON(http.StatusCreated, pvz)
}
