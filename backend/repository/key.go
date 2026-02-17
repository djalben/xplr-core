package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
	
	"crypto/rand"
	"encoding/hex"
	"strings" // <-- Обязательный импорт
)

// GlobalDB не объявляется здесь, так как она объявлена в другом файле пакета repository (globals.go)

// GenerateAPIKey генерирует новый API ключ и сохраняет/обновляет его в базе данных (UPSERT).
func GenerateAPIKey(userID int) (string, error) {
	if GlobalDB == nil { 
		return "", fmt.Errorf("database connection not initialized")
	}

	// 1. Генерируем новый ключ (16 байт -> 32 hex) и форматируем как UUID для колонки UUID в БД
	hexKey, err := generateRandomString(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate random API key: %w", err)
	}
	apiKeyUUID := formatAsUUID(hexKey)

	// 2. INSERT (таблица api_keys: api_key UUID UNIQUE; один ключ на user — перезаписываем через отдельный запрос)
	// Сначала удаляем старый ключ пользователя, затем вставляем новый
	_, _ = GlobalDB.Exec("DELETE FROM api_keys WHERE user_id = $1", userID)
	const query = `INSERT INTO api_keys (api_key, user_id, created_at) VALUES ($1::uuid, $2, $3) RETURNING api_key::text`
	var insertedKey string
	err = GlobalDB.QueryRow(query, apiKeyUUID, userID, time.Now()).Scan(&insertedKey)
	if err != nil {
		log.Printf("Error during API key UPSERT for user %d: %v", userID, err)
		return "", fmt.Errorf("failed to process API key: %w", err)
	}

	return insertedKey, nil
}

// GetUserIDByAPIKey - Находит UserID по API ключу.
func GetUserIDByAPIKey(apiKey string) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}
    
    log.Println("DIAGNOSTIC: GetUserIDByAPIKey called.") 

    // КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ: Удаляем пробелы (whitespace) перед поиском 
    trimmedAPIKey := strings.TrimSpace(apiKey)

	var userID int
	
    // ИСПРАВЛЕНИЕ: ЯВНО приводим поле api_key к текстовому типу (::text) 
	query := `
		SELECT user_id 
		FROM api_keys 
		WHERE api_key::text = $1 AND COALESCE(is_active, TRUE) = TRUE
	`
	// Используем очищенный ключ
	err := GlobalDB.QueryRow(query, trimmedAPIKey).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("API key not found")
		}
		log.Printf("Database error during API key lookup: %v", err)
		return 0, fmt.Errorf("database error during API key lookup")
	}

	return userID, nil
}


// GetAPIKeyByUserID извлекает текущий активный API ключ пользователя по UserID.
func GetAPIKeyByUserID(userID int) (string, error) {
    if GlobalDB == nil {
        return "", fmt.Errorf("database connection not initialized")
    }

    var apiKey string
    query := `SELECT api_key FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`

    err := GlobalDB.QueryRow(query, userID).Scan(&apiKey)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return "", fmt.Errorf("no API key found for user %d", userID)
        }
        log.Printf("Database error fetching API key for user %d: %v", userID, err)
        return "", fmt.Errorf("database error")
    }

    return apiKey, nil
}


// --- Вспомогательные функции (Реализация) ---

// generateRandomBytes генерирует n случайных байт.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b) // Используем crypto/rand
	if err != nil {
		return nil, err
	}
	return b, nil
}

// generateRandomString генерирует случайную строку в шестнадцатеричном формате.
func generateRandomString(n int) (string, error) {
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// formatAsUUID форматирует 32 hex-символа в вид UUID (8-4-4-4-12) для PostgreSQL.
func formatAsUUID(hexStr string) string {
	if len(hexStr) != 32 {
		return hexStr
	}
	return hexStr[0:8] + "-" + hexStr[8:12] + "-" + hexStr[12:16] + "-" + hexStr[16:20] + "-" + hexStr[20:32]
}