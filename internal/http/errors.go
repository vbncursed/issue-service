package http

import (
	"errors"
	"net/http"

	"github.com/vbncursed/vkr/issue-service/internal/http/dto"
	issvc "github.com/vbncursed/vkr/issue-service/internal/service"
)

// MapError переводит доменные/DTO ошибки в HTTP статус и тело APIError
func MapError(err error) (int, APIError) {
	switch {
	// DTO validation
	case errors.Is(err, dto.ErrZoneRequired):
		return http.StatusBadRequest, APIError{Code: "invalid_request", Message: "zone_id required"}
	case errors.Is(err, dto.ErrNbfAfterExp):
		return http.StatusBadRequest, APIError{Code: "invalid_request", Message: "nbf must be before exp"}
	case errors.Is(err, dto.ErrExpExceedsMaxTTL):
		return http.StatusBadRequest, APIError{Code: "invalid_request", Message: "exp exceeds max ttl"}
	case errors.Is(err, dto.ErrTokenRequired):
		return http.StatusBadRequest, APIError{Code: "invalid_request", Message: "token required"}

	// Service errors
	case errors.Is(err, issvc.ErrUnsupportedAlg):
		return http.StatusServiceUnavailable, APIError{Code: "unsupported_alg", Message: "only EdDSA supported"}
	case errors.Is(err, issvc.ErrNotFound):
		return http.StatusNotFound, APIError{Code: "not_found", Message: "pass not found"}
	case errors.Is(err, issvc.ErrConflict):
		return http.StatusConflict, APIError{Code: "conflict", Message: "not Active"}
	case errors.Is(err, issvc.ErrExpiredOrUsed):
		return http.StatusBadRequest, APIError{Code: "invalid_token", Message: "expired_or_used"}
	case errors.Is(err, issvc.ErrInvalidToken):
		return http.StatusBadRequest, APIError{Code: "invalid_token", Message: "invalid"}
	}
	return http.StatusInternalServerError, APIError{Code: "internal", Message: "internal error"}
}
