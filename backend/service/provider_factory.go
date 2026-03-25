package service

import (
	"database/sql"
	"log"
	"os"
)

var globalProvider CardProvider

// InitCardProvider инициализирует глобальный провайдер карт
// Выбирает между MockProvider и ArmeniaProvider в зависимости от конфигурации
func InitCardProvider(db *sql.DB) {
	providerType := os.Getenv("CARD_PROVIDER") // "mock" или "armenia"
	
	if providerType == "armenia" {
		apiKey := os.Getenv("ARMENIA_API_KEY")
		apiBaseURL := os.Getenv("ARMENIA_API_URL")
		
		if apiKey == "" || apiBaseURL == "" {
			log.Println("⚠️  [PROVIDER] ARMENIA_API_KEY or ARMENIA_API_URL not set, falling back to MockProvider")
			globalProvider = NewMockProvider(db)
		} else {
			globalProvider = NewArmeniaProvider(db, apiKey, apiBaseURL)
			log.Printf("✅ [PROVIDER] Initialized ArmeniaProvider (API: %s)", apiBaseURL)
		}
	} else {
		// По умолчанию используем MockProvider
		globalProvider = NewMockProvider(db)
		log.Println("✅ [PROVIDER] Initialized MockProvider (using XPLR database)")
	}
}

// GetCardProvider возвращает глобальный экземпляр провайдера
func GetCardProvider() CardProvider {
	if globalProvider == nil {
		log.Fatal("🚨 [FATAL] Card provider not initialized! Call InitCardProvider first.")
	}
	return globalProvider
}

// SetCardProvider устанавливает провайдер (для тестирования)
func SetCardProvider(provider CardProvider) {
	globalProvider = provider
}
