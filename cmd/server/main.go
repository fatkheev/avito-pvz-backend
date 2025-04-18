package main

import (
	"log"
	"sync"

	"avito-pvz-service/internal/database"
	grpcSrv "avito-pvz-service/internal/grpc"
	"avito-pvz-service/internal/handler"
	"avito-pvz-service/internal/metrics"
	"avito-pvz-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RunServer() {
	if err := database.Init(); err != nil {
		log.Fatalf("Не удалось инициализировать БД: %v", err)
	}
	log.Println("Соединение с БД установлено")

	var wg sync.WaitGroup

	// gRPC‑сервер
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("gRPC сервер запускается")
		grpcSrv.RunGRPCServer()
	}()

	// Prometheus‑метрики
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Metrics сервер слушает на :9000")
		metrics.RunMetricsServer()
	}()

	// HTTP‑сервер (Gin)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		gin.Recovery(),
		metrics.GinMiddleware(),
	)

	// Публичные ручки
	router.POST("/dummyLogin", handler.DummyLoginHandler)
	router.POST("/register", handler.RegisterHandler)
	router.POST("/login", handler.LoginHandler)

	// Защищённые через JWT
	auth := router.Group("/")
	auth.Use(middleware.JWTMiddleware())
	{
		auth.POST("/pvz", handler.CreatePVZHandler)
		auth.GET("/pvz", handler.PVZListHandler)
		auth.POST("/pvz/:pvzId/close_last_reception", handler.CloseReceptionHandler)
		auth.POST("/pvz/:pvzId/delete_last_product", handler.DeleteLastProductHandler)
		auth.POST("/receptions", handler.CreateReceptionHandler)
		auth.POST("/products", handler.AddProductHandler)
	}

	log.Println("HTTP сервер слушает на :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
	}

	wg.Wait()
}

func main() {
	RunServer()
}
