package main

import (
	"log"

	"avito-pvz-service/internal/database"
	"avito-pvz-service/internal/handler"
	"avito-pvz-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database connection established.")

	router := gin.Default()

	router.POST("/dummyLogin", handler.DummyLoginHandler)
	router.POST("/register", handler.RegisterHandler)
	router.POST("/login", handler.LoginHandler)

	// Защищённые эндпоинты
	// Создание ПВЗ (только для модераторов)
	router.POST("/pvz", middleware.JWTMiddleware(), handler.CreatePVZHandler)
	// Создание приёмки товаров (только для сотрудников, роль "staff")
	router.POST("/receptions", middleware.JWTMiddleware(), handler.CreateReceptionHandler)
	// Защищённый эндпоинт для закрытия последней приёмки (только для сотрудников ПВЗ)
	router.POST("/pvz/:pvzId/close_last_reception", middleware.JWTMiddleware(), handler.CloseReceptionHandler)
	// Добавление товара (только для сотрудников, роль "staff")
	router.POST("/products", middleware.JWTMiddleware(), handler.AddProductHandler)
	// Удаление последнего товара (только для сотрудников, роль "staff")
	router.POST("/pvz/:pvzId/delete_last_product", middleware.JWTMiddleware(), handler.DeleteLastProductHandler)

	log.Println("Server is running on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
