package handler

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/djalben/xplr-core/backend/cron"
)

// Handler is the Vercel Cron entry point for syncing card balances.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Verify cron secret (Vercel sends this header for cron jobs)
	if r.Header.Get("Authorization") != "Bearer "+os.Getenv("CRON_SECRET") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbURL := os.Getenv("POSTGRES_DSN")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		http.Error(w, "POSTGRES_DSN/DATABASE_URL not set", http.StatusInternalServerError)
		return
	}

	rows, err := cron.SyncCardBalances(context.Background(), dbURL)
	if err != nil {
		http.Error(w, "Sync error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK: balances synced, updated rows: " + strconv.FormatInt(rows, 10)))
}
