package handler

import (
	"log"
	"net/http"

	"avito-pvz-service/internal/metrics"
	"avito-pvz-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type CreateReceptionRequest struct {
	PVZId string `json:"pvzId" binding:"required,uuid"`
}

func CreateReceptionHandler(c *gin.Context) {
	log.Println("Создание приёмки: начало")

	claims, exists := c.Get("user")
	if !exists {
		log.Println("Создание приёмки: неавторизован")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		log.Println("Создание приёмки: неверные данные токена")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
		return
	}
	role, ok := jwtClaims["role"].(string)
	if !ok || role != "staff" {
		log.Println("Создание приёмки: нет прав (нужна роль сотрудника ПВЗ)")
		c.JSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен: требуется роль сотрудника ПВЗ"})
		return
	}

	// привязка JSON
	var req CreateReceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Создание приёмки: неверный запрос:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing pvzId"})
		return
	}
	log.Printf("Создание приёмки: PVZ=%s\n", req.PVZId)

	// создание приёмки в репозитории
	reception, err := repository.CreateReception(req.PVZId)
	if err != nil {
		log.Println("Создание приёмки: ошибка создания:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// метрика
	metrics.ReceptionsCreatedTotal.Inc()
	log.Printf("Создание приёмки: успешно, id=%s\n", reception.ID)

	c.JSON(http.StatusCreated, reception)
}

func CloseReceptionHandler(c *gin.Context) {
	log.Println("Закрытие приёмки: начало")

	pvzId := c.Param("pvzId")
	if pvzId == "" {
		log.Println("Закрытие приёмки: отсутствует PVZ id в URL")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing PVZ id in URL"})
		return
	}
	log.Printf("Закрытие приёмки: PVZ=%s\n", pvzId)

	claims, exists := c.Get("user")
	if !exists {
		log.Println("Закрытие приёмки: неавторизован")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		log.Println("Закрытие приёмки: неверные данные токена")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
		return
	}
	role, ok := jwtClaims["role"].(string)
	if !ok || role != "staff" {
		log.Println("Закрытие приёмки: нет прав (нужна роль сотрудника ПВЗ)")
		c.JSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен: требуется роль сотрудника ПВЗ"})
		return
	}

	// закрытие через репозиторий
	reception, err := repository.CloseReception(pvzId)
	if err != nil {
		log.Println("Закрытие приёмки: ошибка закрытия:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	log.Printf("Закрытие приёмки: успешно, id=%s\n", reception.ID)
	c.JSON(http.StatusOK, reception)
}
