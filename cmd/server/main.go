package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Service is running"})
	})

	log.Println("Server is running on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
