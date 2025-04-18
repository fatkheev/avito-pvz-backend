package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDummyLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ValidRoles", func(t *testing.T) {
		for _, role := range []string{"client", "staff", "moderator", "anything"} {
			// Формируем запрос
			body, _ := json.Marshal(gin.H{"role": role})
			req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			// Вызываем хендлер
			DummyLoginHandler(ctx)

			// Проверяем код и наличие токена
			assert.Equal(t, http.StatusOK, w.Code, "для роли %q должен быть 200", role)
			var resp DummyLoginResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "для роли %q тело должно парситься как JSON", role)
			assert.NotEmpty(t, resp.Token, "для роли %q token не должен быть пустым", role)
		}
	})

	t.Run("MissingRole", func(t *testing.T) {
		// Пустой JSON => нет поля role
		req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		DummyLoginHandler(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Неверный JSON или отсутствует роль", resp["message"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Некорректный JSON
		req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer([]byte(`{role: staff}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		DummyLoginHandler(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Неверный JSON или отсутствует роль", resp["message"])
	})
}
