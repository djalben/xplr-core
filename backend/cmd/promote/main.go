package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// promote — standalone CLI tool to grant admin rights to a user.
// Usage:  go run backend/cmd/promote/main.go aalabin5@gmail.com
// The tool uses DATABASE_URL from .env (direct postgres connection, bypasses RLS).

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run backend/cmd/promote/main.go <email>")
		os.Exit(1)
	}
	email := strings.TrimSpace(os.Args[1])

	// Load env
	_ = godotenv.Load("backend/.env")
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	// 1. Check user exists
	var userID int
	var currentRole string
	var isAdmin bool
	err = db.QueryRowContext(ctx,
		`SELECT id, COALESCE(role, 'user'), COALESCE(is_admin, false) FROM users WHERE email = $1`, email,
	).Scan(&userID, &currentRole, &isAdmin)
	if err != nil {
		log.Fatalf("User '%s' not found: %v", email, err)
	}

	fmt.Printf("Found user: id=%d, email=%s, role=%s, is_admin=%v\n", userID, email, currentRole, isAdmin)

	if isAdmin && currentRole == "admin" {
		fmt.Println("✅ User is already admin. Nothing to do.")
		return
	}

	// 2. Promote
	_, err = db.ExecContext(ctx,
		`UPDATE users SET role = 'admin', is_admin = TRUE WHERE id = $1`, userID,
	)
	if err != nil {
		log.Fatalf("Failed to promote user: %v", err)
	}

	// 3. Verify
	err = db.QueryRowContext(ctx,
		`SELECT role, is_admin FROM users WHERE id = $1`, userID,
	).Scan(&currentRole, &isAdmin)
	if err != nil {
		log.Fatalf("Verify failed: %v", err)
	}

	fmt.Printf("✅ PROMOTED: id=%d, email=%s → role=%s, is_admin=%v\n", userID, email, currentRole, isAdmin)
}
