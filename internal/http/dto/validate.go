package dto

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrZoneRequired     = errors.New("zone_id required")
	ErrNbfAfterExp      = errors.New("nbf must be before exp")
	ErrExpExceedsMaxTTL = errors.New("exp exceeds max ttl")
	ErrTokenRequired    = errors.New("token required")
)

// Validate проверяет инварианты CreatePassRequest
func (r CreatePassRequest) Validate(now time.Time, maxTTL time.Duration) error {
	if strings.TrimSpace(r.ZoneID) == "" {
		return ErrZoneRequired
	}
	if !r.NBF.Before(r.EXP) {
		return ErrNbfAfterExp
	}
	if r.EXP.Sub(now.UTC()) > maxTTL {
		return ErrExpExceedsMaxTTL
	}
	return nil
}

// Validate проверяет инварианты PickupRequest
func (r PickupRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return ErrTokenRequired
	}
	return nil
}
