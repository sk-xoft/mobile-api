package authmysql

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

const successMessageTTL = 5 * time.Minute

type successMessageCache struct {
	mu            sync.RWMutex
	english, thai sql.NullString
	expiresAt     time.Time
}

// Message caches only static success descriptions. Account, password and session
// state always come from the database. Failed reads are not cached.
func (r *Repository) Message(ctx context.Context, code, language string) string {
	if code != "MA00000" {
		return r.Repository.Message(ctx, code, language)
	}
	cache := &r.successMessage
	cache.mu.RLock()
	if time.Now().Before(cache.expiresAt) {
		description := localizedMessage(cache.english, cache.thai, language)
		cache.mu.RUnlock()
		return description
	}
	cache.mu.RUnlock()
	// Coalesce concurrent cold reads without serving expired descriptions.
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if time.Now().Before(cache.expiresAt) {
		return localizedMessage(cache.english, cache.thai, language)
	}
	var english, thai sql.NullString
	if err := r.db.QueryRowContext(ctx,
		"SELECT MESSAGE_DESCRIPTION_EN, MESSAGE_DESCRIPTION_TH FROM m_message_code WHERE MESSAGE_CODE = ? LIMIT 1", code,
	).Scan(&english, &thai); err != nil {
		return "Message code : " + code + " is not matching in master."
	}
	cache.english, cache.thai = english, thai
	cache.expiresAt = time.Now().Add(successMessageTTL)
	return localizedMessage(english, thai, language)
}

func localizedMessage(english, thai sql.NullString, language string) string {
	if language == "EN" || language == "en" {
		return english.String
	}
	return thai.String
}
