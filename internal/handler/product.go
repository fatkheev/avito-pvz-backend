package handler

import (
	"log"
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
	log.Println("Добавление товара: начало")
	claims, exists := c.Get("user")
	if !exists {
		log.Println("Добавление товара: неавторизован")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		log.Println("Добавление товара: неверные данные токена")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
		return
	}
	role, ok := jwtClaims["role"].(string)
	if !ok || role != "staff" {
		log.Println("Добавление товара: доступ запрещён, роль не сотрудник ПВЗ")
		c.JSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен: требуется роль сотрудника ПВЗ"})
		return
	}

	// привязка JSON
	var req AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Добавление товара: неверный запрос:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON or missing fields"})
		return
	}
	log.Printf("Добавление товара: PVZ=%s, тип=%s\n", req.PVZId, req.Type)

	// создание записи
	product, err := repository.AddProduct(req.PVZId, req.Type)
	if err != nil {
		log.Println("Добавление товара: ошибка добавления:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// метрика
	metrics.ProductsCreatedTotal.Inc()
	log.Printf("Добавление товара: успешно, ID=%s\n", product.ID)

	c.JSON(http.StatusCreated, product)
}

func DeleteLastProductHandler(c *gin.Context) {
	log.Println("Удаление товара: начало")
	claims, exists := c.Get("user")
	if !exists {
		log.Println("Удаление товара: неавторизован")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		log.Println("Удаление товара: неверные данные токена")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
		return
	}
	role, ok := jwtClaims["role"].(string)
	if !ok || role != "staff" {
		log.Println("Удаление товара: доступ запрещён, роль не сотрудник ПВЗ")
		c.JSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен: требуется роль сотрудника ПВЗ"})
		return
	}

	pvzId := c.Param("pvzId")
	if pvzId == "" {
		log.Println("Удаление товара: отсутствует PVZ id в URL")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing PVZ id in URL"})
		return
	}
	log.Println("Удаление товара: PVZ =", pvzId)

	if err := repository.DeleteLastProduct(pvzId); err != nil {
		log.Println("Удаление товара: ошибка удаления:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	log.Println("Удаление товара: успешно удалён последний товар")
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
