// cmd/server/main.go
package main

import (
    "log"
    "sync"

    "avito-pvz-service/internal/database"
    grpcSrv "avito-pvz-service/internal/grpc"
    "avito-pvz-service/internal/handler"
    "avito-pvz-service/internal/middleware"

    "github.com/gin-gonic/gin"
)

func main() {
    // 1) Инициализируем БД
    if err := database.Init(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    log.Println("Database connection established.")

    // 2) Запускаем gRPC‑сервер в горутине
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        grpcSrv.RunGRPCServer()
    }()

    // 3) Настраиваем HTTP‑сервер (Gin)
    router := gin.Default()

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
