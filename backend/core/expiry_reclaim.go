package core

import (
	"log"
	"time"

	"github.com/djalben/xplr-core/backend/repository"
)

// StartExpiryReclaimWorker — фоновый процесс: каждые 30 минут проверяет
// истёкшие карты и возвращает неиспользованный остаток лимита в Кошелёк.
func StartExpiryReclaimWorker() {
	log.Println("[EXPIRY-RECLAIM] Starting expiry reclaim worker...")

	// Первый запуск через 2 минуты после старта сервера
	time.Sleep(2 * time.Minute)

	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		// Первый запуск сразу после задержки
		repository.ReclaimExpiredCardLimits()

		// Затем по расписанию
		for range ticker.C {
			repository.ReclaimExpiredCardLimits()
		}
	}()

	log.Println("[EXPIRY-RECLAIM] Expiry reclaim worker started (checking every 30 minutes)")
}
