package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"avito-pvz-service/internal/handler"
)

func main() {
	router := gin.Default()

	// Регистрируем обработчик для /dummyLogin
	router.POST("/dummyLogin", handler.DummyLoginHandler)

	log.Println("Сервер запущен на порту 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
