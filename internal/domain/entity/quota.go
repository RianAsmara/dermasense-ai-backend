package entity

import "time"

// Quota represents a user's scan quota.
type Quota struct {
	UserID         string    `json:"user_id"`
	RemainingScans int       `json:"remaining_scans"`
	UpdatedAt      time.Time `json:"updated_at"`
}
