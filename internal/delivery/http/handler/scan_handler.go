package handler

import (
	"errors"
	"strconv"

	"github.com/dermasense/backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

// ScanHandler handles scan endpoints.
type ScanHandler struct {
	scanService  *service.ScanService
	quotaService *service.QuotaService
}

// NewScanHandler creates a new ScanHandler.
func NewScanHandler(scanService *service.ScanService, quotaService *service.QuotaService) *ScanHandler {
	return &ScanHandler{
		scanService:  scanService,
		quotaService: quotaService,
	}
}

// CreateScan handles POST /api/v1/scans.
func (h *ScanHandler) CreateScan(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "image file is required",
		})
	}

	scan, err := h.scanService.CreateScan(c.Context(), userID, file)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if errors.Is(err, service.ErrQuotaExhausted) {
			statusCode = fiber.StatusPaymentRequired
		} else if errors.Is(err, service.ErrFileTooLarge) || errors.Is(err, service.ErrInvalidFileType) {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    scan,
	})
}

// GetScans handles GET /api/v1/scans.
func (h *ScanHandler) GetScans(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	scans, err := h.scanService.GetUserScans(c.Context(), userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch scans",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    scans,
	})
}

// GetScanByID handles GET /api/v1/scans/:id.
func (h *ScanHandler) GetScanByID(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	scanID := c.Params("id")
	if scanID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "scan id is required",
		})
	}

	scan, err := h.scanService.GetScanByID(c.Context(), scanID, userID)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "scan not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    scan,
	})
}

// GetQuota handles GET /api/v1/quota.
func (h *ScanHandler) GetQuota(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	quota, err := h.quotaService.GetQuota(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch quota",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    quota,
	})
}

// AddQuotaRequest is the request body for adding quota.
type AddQuotaRequest struct {
	Amount int `json:"amount"`
}

// AddQuota handles POST /api/v1/quota/add (after watching a rewarded ad).
func (h *ScanHandler) AddQuota(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req AddQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Amount <= 0 {
		req.Amount = 1 // Default to 1 scan per ad
	}

	quota, err := h.quotaService.AddScans(c.Context(), userID, req.Amount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    quota,
	})
}
