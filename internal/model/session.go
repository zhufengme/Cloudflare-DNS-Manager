package model

import "time"

type Session struct {
	CloudflareEmail string
	UserAPIKey      string
	CreatedAt       time.Time
	ExpiresAt       time.Time
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
