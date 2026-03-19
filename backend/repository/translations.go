package repository

import (
	"fmt"
	"log"
	"time"
)

// EnsureTranslationsTable creates the translations table if it doesn't exist.
func EnsureTranslationsTable() error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(`
		CREATE TABLE IF NOT EXISTS translations (
			id        SERIAL PRIMARY KEY,
			msg_key   TEXT NOT NULL,
			lang      TEXT NOT NULL DEFAULT 'ru',
			value     TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(msg_key, lang)
		);
		CREATE INDEX IF NOT EXISTS idx_translations_key_lang ON translations(msg_key, lang);
	`)
	if err != nil {
		log.Printf("[I18N] Error creating translations table: %v", err)
		return err
	}
	log.Println("[I18N] ✅ translations table ensured")
	return nil
}

// T returns the translation for a given key and language.
// Falls back to the key itself if no translation is found.
func T(key, lang string) string {
	if GlobalDB == nil {
		return key
	}
	var value string
	err := GlobalDB.QueryRow(
		`SELECT value FROM translations WHERE msg_key = $1 AND lang = $2`,
		key, lang,
	).Scan(&value)
	if err != nil || value == "" {
		return key
	}
	return value
}

// TranslationRow represents a single translation entry for admin panel.
type TranslationRow struct {
	ID        int    `json:"id"`
	MsgKey    string `json:"msg_key"`
	Lang      string `json:"lang"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
}

// GetAllTranslations returns all translations, optionally filtered by lang.
func GetAllTranslations(langFilter string) ([]TranslationRow, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	q := `SELECT id, msg_key, lang, value, updated_at FROM translations`
	var args []interface{}
	if langFilter != "" {
		q += ` WHERE lang = $1`
		args = append(args, langFilter)
	}
	q += ` ORDER BY msg_key ASC, lang ASC LIMIT 500`

	rows, err := GlobalDB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TranslationRow
	for rows.Next() {
		var t TranslationRow
		var updatedAt time.Time
		if err := rows.Scan(&t.ID, &t.MsgKey, &t.Lang, &t.Value, &updatedAt); err != nil {
			continue
		}
		t.UpdatedAt = updatedAt.Format(time.RFC3339)
		result = append(result, t)
	}
	if result == nil {
		result = []TranslationRow{}
	}
	return result, nil
}

// UpsertTranslation creates or updates a translation.
func UpsertTranslation(msgKey, lang, value string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`INSERT INTO translations (msg_key, lang, value, updated_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (msg_key, lang) DO UPDATE SET value = $3, updated_at = NOW()`,
		msgKey, lang, value,
	)
	return err
}

// DeleteTranslation removes a translation by ID.
func DeleteTranslation(id int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(`DELETE FROM translations WHERE id = $1`, id)
	return err
}
