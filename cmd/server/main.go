package main

import (
    "log"

    "avito-pvz-service/internal/database"
    "avito-pvz-service/internal/handler"

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

    log.Println("Server is running on port 8080")
    if err := router.Run(":8080"); err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
}
