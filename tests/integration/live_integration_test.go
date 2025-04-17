package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080"

func Test_FullFlow_LiveIntegration(t *testing.T) {
	// Получаем токен модератора
	modToken := getToken(t, "moderator")

	// Создаем ПВЗ
	pvzResp := postJSON(t, "/pvz", map[string]string{"city": "Москва"}, modToken)
	pvzID := pvzResp["id"].(string)

	// Получаем токен сотрудника
	staffToken := getToken(t, "staff")

	// Создаем приёмку
	recResp := postJSON(t, "/receptions", map[string]string{"pvzId": pvzID}, staffToken)
	assert.Equal(t, "in_progress", recResp["status"])

	// Добавляем 50 товаров
	for i := 0; i < 50; i++ {
		postJSON(t, "/products", map[string]string{
			"pvzId": pvzID,
			"type":  "электроника",
		}, staffToken)
	}

	// Закрываем приёмку
	url := fmt.Sprintf("/pvz/%s/close_last_reception", pvzID)
	closeResp := post(t, url, staffToken)
	assert.Equal(t, 200, closeResp.StatusCode)
}

func getToken(t *testing.T, role string) string {
	data := map[string]string{"role": role}
	body, _ := json.Marshal(data)

	resp, err := http.Post(baseURL+"/dummyLogin", "application/json", bytes.NewReader(body))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result["token"]
}

func postJSON(t *testing.T, path string, payload map[string]string, token string) map[string]interface{} {
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func post(t *testing.T, path string, token string) *http.Response {
	req, _ := http.NewRequest("POST", baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	return resp
}
