package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aalabin/xplr/core"
	"github.com/aalabin/xplr/models"
	"github.com/shopspring/decimal"
)

// TestAuthorizeCardHandler - Тестовый хендлер для симуляции авторизации (Задача 2.1)
func TestAuthorizeCardHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Преобразование модели запроса в модель Core
	coreReq := core.AuthorizeCardRequest{
        CardID: req.CardID,
        Amount: req.Amount,
        MerchantName: req.MerchantName, 
    }

	// 2. Вызов функции Core с правильной структурой-аргументом
	response := core.AuthorizeCard(coreReq) 

	// 3. Возвращение результата
	statusCode := http.StatusOK
    if !response.Success {
        statusCode = http.StatusPaymentRequired 
    }
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
    }
}

// TestAuthorizeCard - Базовый тест (заглушка)
func TestAuthorizeCard(t *testing.T) {
	// Создаем тестовое тело запроса (используем models.AuthRequest)
	authData := models.AuthRequest{CardID: 12345, Amount: decimal.NewFromFloat(100.50), MerchantName: "Test Merchant"}
	body, _ := json.Marshal(authData)

	// Создаем тестовый HTTP-запрос
	req := httptest.NewRequest("POST", "/api/v1/authorize", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Создаем тестовый ResponseRecorder
	rr := httptest.NewRecorder()
	
	// Вызываем хендлер 
	TestAuthorizeCardHandler(rr, req) 

	// Проверяем статус код (ожидаем 402, так как баланс пустой)
	if status := rr.Code; status != http.StatusPaymentRequired {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusPaymentRequired)
	}

	// Проверяем тело ответа (базовая проверка)
	var resp models.AuthResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}
}