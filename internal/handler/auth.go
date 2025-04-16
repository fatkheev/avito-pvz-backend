package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DummyLoginRequest struct {
	Role string `json:"role" binding:"required"`
}

type DummyLoginResponse struct {
	Token string `json:"token"`
}

func DummyLoginHandler(c *gin.Context) {
	var req DummyLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Неверный JSON или отсутствует поле role"})
		return
	}

	if req.Role != "employee" && req.Role != "moderator" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Недопустимая роль"})
		return
	}

	token := "dummy-token-for-" + req.Role
	c.JSON(http.StatusOK, DummyLoginResponse{Token: token})
}
