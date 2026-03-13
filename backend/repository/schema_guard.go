package repository

import (
	"fmt"
	"log"
)

// requiredColumn defines a column that must exist on a table.
type requiredColumn struct {
	Table        string
	Column       string
	Definition   string // e.g. "NUMERIC(20,4) DEFAULT 0.0000"
}

// allRequiredColumns is the single source of truth for columns the backend needs.
var allRequiredColumns = []requiredColumn{
	// --- users ---
	{"users", "balance", "NUMERIC(20,4) DEFAULT 0.0000"},
	{"users", "balance_rub", "NUMERIC(20,4) DEFAULT 0.0000 NOT NULL"},
	{"users", "balance_arbitrage", "NUMERIC(20,4) DEFAULT 0.0000"},
	{"users", "balance_personal", "NUMERIC(20,4) DEFAULT 0.0000"},
	{"users", "kyc_status", "VARCHAR(50) DEFAULT 'pending'"},
	{"users", "active_mode", "VARCHAR(50) DEFAULT 'personal'"},
	{"users", "status", "VARCHAR(50) DEFAULT 'ACTIVE'"},
	{"users", "telegram_chat_id", "BIGINT DEFAULT NULL"},
	{"users", "telegram_id", "BIGINT"},
	{"users", "is_admin", "BOOLEAN DEFAULT FALSE"},
	{"users", "is_verified", "BOOLEAN DEFAULT FALSE"},
	{"users", "role", "VARCHAR(20) DEFAULT 'user'"},
	{"users", "referral_code", "VARCHAR(20)"},
	{"users", "referred_by", "INTEGER"},
	{"users", "verification_token", "VARCHAR(255)"},
	{"users", "auth_method_preference", "VARCHAR(20) DEFAULT 'email'"},
	// --- cards ---
	{"cards", "service_slug", "VARCHAR(50) DEFAULT 'arbitrage'"},
	{"cards", "spending_limit", "NUMERIC(20,4) DEFAULT 0.0000"},
	{"cards", "spent_from_wallet", "NUMERIC(20,4) DEFAULT 0.0000"},
	{"cards", "expiry_date", "TIMESTAMP WITH TIME ZONE"},
	{"cards", "default_max_limit", "NUMERIC(20,4) DEFAULT 0.0000"},
	// --- transactions ---
	{"transactions", "source_type", "VARCHAR(50) DEFAULT 'card_charge'"},
	{"transactions", "source_id", "INTEGER"},
	{"transactions", "currency", "VARCHAR(10) DEFAULT 'USD'"},
	{"transactions", "provider_tx_id", "VARCHAR(255)"},
	// --- internal_balances ---
	{"internal_balances", "auto_topup_enabled", "BOOLEAN DEFAULT FALSE"},
}

// RunSchemaGuard checks all required columns exist and creates missing ones.
// Should be called once at backend startup after DB connection is established.
func RunSchemaGuard() {
	if GlobalDB == nil {
		log.Println("[SCHEMA-GUARD] ⚠️  Skipped: DB not initialized")
		return
	}

	log.Println("[SCHEMA-GUARD] 🔍 Checking database schema integrity...")

	created := 0
	skipped := 0
	errors := 0

	for _, col := range allRequiredColumns {
		exists, err := columnExists(col.Table, col.Column)
		if err != nil {
			log.Printf("[SCHEMA-GUARD] ❌ Error checking %s.%s: %v", col.Table, col.Column, err)
			errors++
			continue
		}
		if exists {
			skipped++
			continue
		}

		// Column missing — create it
		ddl := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", col.Table, col.Column, col.Definition)
		if _, err := GlobalDB.Exec(ddl); err != nil {
			log.Printf("[SCHEMA-GUARD] ❌ Failed to create %s.%s: %v", col.Table, col.Column, err)
			errors++
		} else {
			log.Printf("[SCHEMA-GUARD] ✅ Created missing column: %s.%s", col.Table, col.Column)
			created++
		}
	}

	// Ensure RLS is disabled on key tables
	rlsTables := []string{
		"users", "cards", "transactions", "user_grades", "internal_balances",
		"api_keys", "teams", "team_members", "referrals", "referral_commissions",
		"verification_tokens", "password_reset_tokens", "support_tickets",
		"admin_logs", "commission_config",
	}
	for _, t := range rlsTables {
		if _, err := GlobalDB.Exec(fmt.Sprintf("ALTER TABLE IF EXISTS %s DISABLE ROW LEVEL SECURITY", t)); err != nil {
			log.Printf("[SCHEMA-GUARD] ⚠️  Could not disable RLS on %s: %v", t, err)
		}
	}

	log.Printf("[SCHEMA-GUARD] ✅ Done: %d created, %d already OK, %d errors", created, skipped, errors)
}

// columnExists checks if a column exists on a table via information_schema.
func columnExists(table, column string) (bool, error) {
	var exists bool
	err := GlobalDB.QueryRow(
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		)`, table, column,
	).Scan(&exists)
	return exists, err
}
