package handler

import (
	"log"
	"net/http"
	"os"
	"time"

	"avito-pvz-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default654пв6а45п46ва5п6ва5п46в5а1п6а5816а16в"
		log.Println("Используется секрет по умолчанию")
	}
	return []byte(secret)
}

func generateJWT(email, role string) (string, error) {
	log.Printf("Генерация токена для %s с ролью %s", email, role)
	claims := jwt.MapClaims{
		"sub":  email,
		"role": role,
		"exp":  time.Now().Add(72 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(getJWTSecret())
	if err != nil {
		log.Println("Ошибка при создании токена:", err)
	}
	return signed, err
}

type DummyLoginRequest struct {
	Role string `json:"role" binding:"required"`
}
type DummyLoginResponse struct{ Token string }

func DummyLoginHandler(c *gin.Context) {
	log.Println("Вызов dummyLogin")
	var req DummyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Некорректный запрос dummyLogin:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Неверный JSON или отсутствует роль"})
		return
	}

	token, err := generateJWT(req.Role, req.Role)
	if err != nil {
		log.Println("Не удалось сгенерировать токен:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Ошибка генерации токена"})
		return
	}

	log.Println("dummyLogin успешно, роль:", req.Role)
	c.JSON(http.StatusOK, DummyLoginResponse{Token: token})
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=client moderator"`
}

func RegisterHandler(c *gin.Context) {
	log.Println("Вызов регистрации")
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Некорректный запрос регистрации:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Неверный JSON или отсутствуют поля"})
		return
	}

	user, err := repository.CreateUser(req.Email, req.Password, req.Role)
	if err != nil {
		log.Println("Ошибка при создании пользователя:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	log.Println("Пользователь зарегистрирован, ID:", user.ID)
	c.JSON(http.StatusCreated, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
type LoginResponse struct{ Token string }

func LoginHandler(c *gin.Context) {
	log.Println("Вызов авторизации")
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Некорректный запрос авторизации:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Неверный JSON или отсутствуют поля"})
		return
	}

	user, err := repository.GetUserByEmail(req.Email)
	if err != nil {
		log.Println("Пользователь не найден:", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Неверные учетные данные"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Println("Неверный пароль для:", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Неверные учетные данные"})
		return
	}

	token, err := generateJWT(user.Email, user.Role)
	if err != nil {
		log.Println("Ошибка при генерации токена:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Ошибка генерации токена"})
		return
	}

	log.Println("Авторизация успешна:", req.Email)
	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
