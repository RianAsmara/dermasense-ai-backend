package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/dermasense/backend/internal/config"
	"github.com/dermasense/backend/internal/domain/entity"
	"github.com/dermasense/backend/internal/domain/repository"
	"github.com/google/uuid"
)

// ScanService orchestrates the skin scan analysis flow.
type ScanService struct {
	scanRepo  repository.ScanRepository
	quotaRepo repository.QuotaRepository
	openAISvc *OpenAIService
	cfg       *config.Config
}

// NewScanService creates a new ScanService.
func NewScanService(
	scanRepo repository.ScanRepository,
	quotaRepo repository.QuotaRepository,
	openAISvc *OpenAIService,
	cfg *config.Config,
) *ScanService {
	return &ScanService{
		scanRepo:  scanRepo,
		quotaRepo: quotaRepo,
		openAISvc: openAISvc,
		cfg:       cfg,
	}
}

// CreateScan processes a new skin scan.
func (s *ScanService) CreateScan(ctx context.Context, userID string, file *multipart.FileHeader) (*entity.Scan, error) {
	quota, err := s.quotaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check quota: %w", err)
	}
	if quota.RemainingScans <= 0 {
		return nil, ErrQuotaExhausted
	}

	if file.Size > s.cfg.App.MaxFileSize {
		return nil, ErrFileTooLarge
	}

	contentType := file.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		return nil, ErrInvalidFileType
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	imageData, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	scanID := uuid.New().String()
	imageURL, err := s.saveImage(scanID, file.Filename, imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	result, err := s.openAISvc.AnalyzeSkinImage(ctx, imageData)
	if err != nil {
		return nil, fmt.Errorf("ai analysis failed: %w", err)
	}

	scan := &entity.Scan{
		ID:              scanID,
		UserID:          userID,
		ImageURL:        imageURL,
		Diagnosis:       result.Diagnosis,
		Confidence:      result.Confidence,
		Severity:        result.Severity,
		Explanation:     result.Explanation,
		Recommendations: result.Recommendations,
	}

	if err := s.scanRepo.Create(ctx, scan); err != nil {
		return nil, fmt.Errorf("failed to store scan: %w", err)
	}

	if err := s.quotaRepo.DecrementScans(ctx, userID); err != nil {
		fmt.Printf("warning: failed to decrement quota for user %s: %v\n", userID, err)
	}

	return scan, nil
}

// GetUserScans retrieves scan history for a user.
func (s *ScanService) GetUserScans(ctx context.Context, userID string, limit, offset int) ([]*entity.Scan, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.scanRepo.GetByUserID(ctx, userID, limit, offset)
}

// GetScanByID retrieves a specific scan.
func (s *ScanService) GetScanByID(ctx context.Context, scanID, userID string) (*entity.Scan, error) {
	scan, err := s.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		return nil, err
	}
	if scan.UserID != userID {
		return nil, ErrUnauthorized
	}
	return scan, nil
}

func (s *ScanService) saveImage(scanID, originalName string, data []byte) (string, error) {
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".jpg"
	}

	filename := scanID + ext
	dirPath := filepath.Join(s.cfg.App.UploadDir, "scans")

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return "", err
	}

	filePath := filepath.Join(dirPath, filename)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", err
	}

	return "/uploads/scans/" + filename, nil
}

func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/heic": true,
	}
	return validTypes[contentType]
}
