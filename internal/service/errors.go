package service

import "errors"

var (
	// ErrQuotaExhausted is returned when a user has no scans remaining.
	ErrQuotaExhausted = errors.New("scan quota exhausted, watch a rewarded ad to unlock more scans")

	// ErrFileTooLarge is returned when the uploaded file exceeds the max size.
	ErrFileTooLarge = errors.New("file too large, maximum size is 10MB")

	// ErrInvalidFileType is returned when the uploaded file is not a valid image.
	ErrInvalidFileType = errors.New("invalid file type, only JPEG, PNG, WebP, and HEIC are allowed")

	// ErrUnauthorized is returned when a user tries to access another user's resource.
	ErrUnauthorized = errors.New("unauthorized access")
)
