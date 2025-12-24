package middleware

import (
	"fmt"
	"sync"
	"time"
)

type RateLimiter struct {
	mu          sync.RWMutex
	attempts    map[string]*attemptRecord
	maxAttempts int
	window      time.Duration
}

type attemptRecord struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		attempts:    make(map[string]*attemptRecord),
		maxAttempts: maxAttempts,
		window:      window,
	}

	// 定期清理过期记录
	go rl.cleanupExpired()

	return rl
}

func (rl *RateLimiter) CheckAndIncrement(email string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := fmt.Sprintf("login_%s_%s", time.Now().Format("2006-01-02_15"), email)

	record, exists := rl.attempts[key]
	if !exists {
		rl.attempts[key] = &attemptRecord{
			count:     1,
			resetTime: time.Now().Add(rl.window),
		}
		return true
	}

	if time.Now().After(record.resetTime) {
		record.count = 1
		record.resetTime = time.Now().Add(rl.window)
		return true
	}

	if record.count >= rl.maxAttempts {
		return false
	}

	record.count++
	return true
}

func (rl *RateLimiter) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, record := range rl.attempts {
			if now.After(record.resetTime) {
				delete(rl.attempts, key)
			}
		}
		rl.mu.Unlock()
	}
}
