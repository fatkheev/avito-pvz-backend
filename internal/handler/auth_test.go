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

func TestDummyLoginHandler_ValidRoles(t *testing.T) {
    gin.SetMode(gin.TestMode)
    cases := []struct {
        Role       string
        StatusCode int
    }{
        {"client", http.StatusOK},
        {"moderator", http.StatusOK},
    }
    for _, c := range cases {
        body, _ := json.Marshal(gin.H{"role": c.Role})
        req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()

        ctx, _ := gin.CreateTestContext(w)
        ctx.Request = req

        DummyLoginHandler(ctx)

        assert.Equal(t, c.StatusCode, w.Code)
        var resp map[string]string
        assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
        assert.NotEmpty(t, resp["token"])
    }
}

func TestDummyLoginHandler_InvalidRole(t *testing.T) {
    gin.SetMode(gin.TestMode)
    body, _ := json.Marshal(gin.H{"role": "unknown"})
    req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    ctx, _ := gin.CreateTestContext(w)
    ctx.Request = req

    DummyLoginHandler(ctx)

    assert.Equal(t, http.StatusBadRequest, w.Code)
    var resp map[string]string
    assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
    assert.Equal(t, "Invalid role provided", resp["message"])
}

func TestDummyLoginHandler_MissingRole(t *testing.T) {
    gin.SetMode(gin.TestMode)

    reqBody := []byte(`{}`) // нет поля role
    req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(reqBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    ctx, _ := gin.CreateTestContext(w)
    ctx.Request = req

    DummyLoginHandler(ctx)

    assert.Equal(t, http.StatusBadRequest, w.Code)
    var resp map[string]string
    assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
    assert.Equal(t, "Invalid JSON or missing role", resp["message"])
}

func TestDummyLoginHandler_InvalidJSON(t *testing.T) {
    gin.SetMode(gin.TestMode)

    reqBody := []byte(`{role: staff}`) // некорректный JSON (нет кавычек вокруг ключа)
    req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(reqBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    ctx, _ := gin.CreateTestContext(w)
    ctx.Request = req

    DummyLoginHandler(ctx)

    assert.Equal(t, http.StatusBadRequest, w.Code)
    var resp map[string]string
    assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
    assert.Equal(t, "Invalid JSON or missing role", resp["message"])
}

func TestDummyLoginHandler_StaffRole(t *testing.T) {
    gin.SetMode(gin.TestMode)

    reqBody, _ := json.Marshal(gin.H{"role": "staff"})
    req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(reqBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    ctx, _ := gin.CreateTestContext(w)
    ctx.Request = req

    DummyLoginHandler(ctx)

    assert.Equal(t, http.StatusOK, w.Code)
    var resp map[string]string
    assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
    assert.NotEmpty(t, resp["token"])
}

func TestRegisterHandler_InvalidInput(t *testing.T) {
    gin.SetMode(gin.TestMode)
    body := []byte(`{"email": "not-an-email", "password": "", "role": "moderator"}`)
    req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    ctx, _ := gin.CreateTestContext(w)
    ctx.Request = req

    RegisterHandler(ctx)

    assert.Equal(t, http.StatusBadRequest, w.Code)
}
