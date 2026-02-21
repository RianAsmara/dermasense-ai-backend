package entity

import "time"

// Scan represents a skin scan analysis result.
type Scan struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	ImageURL        string    `json:"image_url"`
	Diagnosis       string    `json:"diagnosis"`
	Confidence      float64   `json:"confidence"`
	Severity        string    `json:"severity"`
	Explanation     string    `json:"explanation"`
	Recommendations []string  `json:"recommendations"`
	CreatedAt       time.Time `json:"created_at"`
}

// ScanResult represents the structured AI analysis response.
type ScanResult struct {
	Diagnosis       string   `json:"diagnosis"`
	Confidence      float64  `json:"confidence"`
	Severity        string   `json:"severity"`
	Explanation     string   `json:"explanation"`
	Recommendations []string `json:"recommendations"`
}
