package service

import (
	"context"
	"fmt"

	"github.com/dermasense/backend/internal/domain/entity"
	"github.com/dermasense/backend/internal/domain/repository"
)

// QuotaService handles scan quota management.
type QuotaService struct {
	quotaRepo repository.QuotaRepository
}

// NewQuotaService creates a new QuotaService.
func NewQuotaService(quotaRepo repository.QuotaRepository) *QuotaService {
	return &QuotaService{quotaRepo: quotaRepo}
}

// GetQuota retrieves the current quota for a user.
func (s *QuotaService) GetQuota(ctx context.Context, userID string) (*entity.Quota, error) {
	quota, err := s.quotaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}
	return quota, nil
}

// AddScans adds scans to a user's quota (e.g., after watching a rewarded ad).
func (s *QuotaService) AddScans(ctx context.Context, userID string, amount int) (*entity.Quota, error) {
	if amount <= 0 || amount > 5 {
		return nil, fmt.Errorf("invalid scan amount: must be between 1 and 5")
	}

	if err := s.quotaRepo.IncrementScans(ctx, userID, amount); err != nil {
		return nil, fmt.Errorf("failed to add scans: %w", err)
	}

	return s.quotaRepo.GetByUserID(ctx, userID)
}
