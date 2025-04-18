// cmd/server/main.go
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

func main() {
    if err := database.Init(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    log.Println("Database connection established.")

    // Запуск gRPC-сервера
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        grpcSrv.RunGRPCServer()
    }()

    // Запуск Prometheus‑метрик
    wg.Add(1)
    go func() {
        defer wg.Done()
        log.Println("Metrics server is running on :9000")
        metrics.RunMetricsServer()
    }()

    // Запуск HTTP-сервера Gin
    router := gin.New()
    router.Use(gin.Recovery(), metrics.GinMiddleware())

    router.POST("/dummyLogin", handler.DummyLoginHandler)
    router.POST("/register", handler.RegisterHandler)
    router.POST("/login", handler.LoginHandler)

    auth := router.Group("", middleware.JWTMiddleware())
    {
        auth.POST("/pvz", handler.CreatePVZHandler)
        auth.GET("/pvz", handler.PVZListHandler)
        auth.POST("/pvz/:pvzId/close_last_reception", handler.CloseReceptionHandler)
        auth.POST("/pvz/:pvzId/delete_last_product", handler.DeleteLastProductHandler)
        auth.POST("/receptions", handler.CreateReceptionHandler)
        auth.POST("/products", handler.AddProductHandler)
    }

    log.Println("HTTP server is running on port 8080")
    if err := router.Run(":8080"); err != nil {
        log.Fatalf("Error starting HTTP server: %v", err)
    }

    wg.Wait()
}
