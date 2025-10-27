package service

import "errors"

var (
	ErrUnsupportedAlg = errors.New("unsupported_alg")
	ErrNotFound       = errors.New("not_found")
	ErrConflict       = errors.New("conflict")
	ErrInvalidToken   = errors.New("invalid_token")
	ErrExpiredOrUsed  = errors.New("expired_or_used")
)
