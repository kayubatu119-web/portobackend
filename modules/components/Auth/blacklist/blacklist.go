// File: modules/components/Auth/blacklist/blacklist.go
package blacklist

import (
	"sync"
	"time"
)

// TokenBlacklist untuk menyimpan token yang sudah logout
type TokenBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

// Global blacklist instance - PENTING: harus singleton
var instance *TokenBlacklist
var once sync.Once

// GetInstance mengembalikan singleton instance dari blacklist
func GetInstance() *TokenBlacklist {
	once.Do(func() {
		instance = &TokenBlacklist{
			tokens: make(map[string]time.Time),
		}

		// Jalankan cleanup job setiap 1 jam
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				instance.CleanupExpired()
			}
		}()
	})
	return instance
}

// IsBlacklisted mengecek apakah token ada di blacklist
func (tb *TokenBlacklist) IsBlacklisted(token string) bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if expiry, exists := tb.tokens[token]; exists {
		// Cek apakah token masih belum expired
		if time.Now().Before(expiry) {
			return true
		}
		// Token sudah expired, hapus dari blacklist
		tb.mu.RUnlock()
		tb.mu.Lock()
		delete(tb.tokens, token)
		tb.mu.Unlock()
		tb.mu.RLock()
	}
	return false
}

// Add menambahkan token ke blacklist
func (tb *TokenBlacklist) Add(token string, expiryTime time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.tokens[token] = expiryTime
}

// CleanupExpired membersihkan token yang sudah expired dari blacklist
func (tb *TokenBlacklist) CleanupExpired() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	for token, expiry := range tb.tokens {
		if now.After(expiry) {
			delete(tb.tokens, token)
		}
	}
}

// GetCount mengembalikan jumlah token di blacklist (untuk debugging)
func (tb *TokenBlacklist) GetCount() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return len(tb.tokens)
}

// IsTokenInList untuk debugging - cek apakah token ada di map
func (tb *TokenBlacklist) IsTokenInList(token string) (bool, time.Time) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	expiry, exists := tb.tokens[token]
	return exists, expiry
}
