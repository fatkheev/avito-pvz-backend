package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"avito-pvz-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func PVZListHandler(c *gin.Context) {
	log.Println("Получение списка ПВЗ: начало")

	startDateStr := c.Query("startDate")
	endDateStr   := c.Query("endDate")
	if startDateStr == "" || endDateStr == "" {
		log.Println("Получение списка ПВЗ: отсутствуют startDate или endDate")
		c.JSON(http.StatusBadRequest, gin.H{"message": "startDate and endDate parameters are required"})
		return
	}
	log.Printf("Получение списка ПВЗ: диапазон %s — %s\n", startDateStr, endDateStr)

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		log.Println("Получение списка ПВЗ: неверный формат startDate:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid startDate format"})
		return
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		log.Println("Получение списка ПВЗ: неверный формат endDate:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid endDate format"})
		return
	}

	// pagination
	pageStr  := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		log.Printf("Получение списка ПВЗ: некорректный page=%s, используем page=1\n", pageStr)
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		log.Printf("Получение списка ПВЗ: некорректный limit=%s, используем limit=10\n", limitStr)
		limit = 10
	}
	log.Printf("Получение списка ПВЗ: page=%d, limit=%d\n", page, limit)

	claims, exists := c.Get("user")
	if !exists {
		log.Println("Получение списка ПВЗ: неавторизован")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		log.Println("Получение списка ПВЗ: некорректный токен")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
		return
	}
	role, ok := jwtClaims["role"].(string)
	if !ok || (role != "staff" && role != "moderator") {
		log.Println("Получение списка ПВЗ: доступ запрещён для роли", role)
		c.JSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен"})
		return
	}
	log.Println("Получение списка ПВЗ: авторизован, роль =", role)

	// Вызов репозитория
	records, err := repository.GetPVZRecords(&startDate, &endDate, page, limit)
	if err != nil {
		log.Println("Получение списка ПВЗ: ошибка репозитория:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	log.Printf("Получение списка ПВЗ: найдено %d записей\n", len(records))
	c.JSON(http.StatusOK, records)
}
