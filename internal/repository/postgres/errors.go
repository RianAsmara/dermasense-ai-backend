package postgres

import "errors"

// ErrNoQuotaRemaining is returned when a user has no remaining scans.
var ErrNoQuotaRemaining = errors.New("no quota remaining")
