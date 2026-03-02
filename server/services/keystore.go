package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"vibecodepc/server/db"
)

// Set encrypts value and stores it under name in the settings table.
func Set(name, value string) error {
	encrypted, err := db.Encrypt(value)
	if err != nil {
		return fmt.Errorf("keystore: encrypt %q: %w", name, err)
	}

	now := time.Now().Unix()
	_, err = db.DB().Exec(
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
         ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		name, encrypted, now,
	)
	if err != nil {
		return fmt.Errorf("keystore: upsert %q: %w", name, err)
	}
	return nil
}

// Get retrieves and decrypts the value stored under name.
// Returns ("", false) if the key does not exist.
func Get(name string) (string, bool) {
	var encrypted string
	err := db.DB().QueryRow(`SELECT value FROM settings WHERE key = ?`, name).Scan(&encrypted)
	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		return "", false
	}

	plaintext, err := db.Decrypt(encrypted)
	if err != nil {
		return "", false
	}
	return plaintext, true
}

// Del removes the key from the settings table.
func Del(name string) error {
	_, err := db.DB().Exec(`DELETE FROM settings WHERE key = ?`, name)
	if err != nil {
		return fmt.Errorf("keystore: delete %q: %w", name, err)
	}
	return nil
}

// Exists returns true if the key is present in the settings table.
func Exists(name string) bool {
	var count int
	_ = db.DB().QueryRow(`SELECT COUNT(*) FROM settings WHERE key = ?`, name).Scan(&count)
	return count > 0
}

// Mask returns a partially redacted version of value showing only the last 4 characters.
// Example: "sk-ant-api03-abc123xyz" → "sk-ant-••••••••xyz"
func Mask(value string) string {
	if value == "" {
		return ""
	}
	runes := []rune(value)
	total := utf8.RuneCountInString(value)
	if total <= 4 {
		return strings.Repeat("•", total)
	}
	visible := string(runes[total-4:])
	prefix := ""
	// Preserve a recognisable prefix if the key has a known format.
	for _, sep := range []string{"-api", "_api", "sk-", "AIza"} {
		if strings.Contains(value[:minInt(20, len(value))], sep) {
			idx := strings.Index(value, sep)
			if idx >= 0 && idx+len(sep) < total-4 {
				prefix = value[:idx+len(sep)]
			}
			break
		}
	}
	masked := strings.Repeat("•", 8)
	if prefix != "" {
		return prefix + masked + visible
	}
	return masked + visible
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
