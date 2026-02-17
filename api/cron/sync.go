package handler

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/djalben/xplr-core/backend/handlers"
	"github.com/djalben/xplr-core/backend/repository"
)

// Handler is the Vercel Cron entry point for syncing card balances.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Verify cron secret (Vercel sends this header for cron jobs)
	if r.Header.Get("Authorization") != "Bearer "+os.Getenv("CRON_SECRET") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		http.Error(w, "DATABASE_URL not set", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		http.Error(w, "DB open error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	db.SetMaxOpenConns(3)
	if err = db.PingContext(context.Background()); err != nil {
		http.Error(w, "DB ping error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Wire DB
	handlers.GlobalDB = db
	repository.GlobalDB = db

	// Run sync
	wallesterRepo := repository.NewWallesterRepository()
	if wallesterRepo == nil {
		http.Error(w, "Wallester not configured", http.StatusInternalServerError)
		return
	}

	if err := wallesterRepo.SyncAllCardsBalances(); err != nil {
		log.Printf("Cron sync error: %v", err)
		http.Error(w, "Sync error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK: balances synced"))
}
