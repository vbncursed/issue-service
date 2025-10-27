package http

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/vbncursed/vkr/issue-service/internal/config"
	"github.com/vbncursed/vkr/issue-service/internal/crypto"
	"github.com/vbncursed/vkr/issue-service/internal/http/dto"
	im "github.com/vbncursed/vkr/issue-service/internal/models"
	"github.com/vbncursed/vkr/issue-service/internal/util"
)

// CreatePass — выпуск пропуска
// @Summary     Выпуск пропуска
// @Tags        passes
// @Accept      json
// @Produce     json
// @Param       request body dto.CreatePassRequest true "Create pass"
// @Success     201 {object} dto.CreatePassResponse
// @Failure     400 {object} APIError
// @Failure     503 {object} APIError
// @Router      /passes [post]
func CreatePass(pool *pgxpool.Pool, cfg config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreatePassRequest
		if err := c.Bind(&req); err != nil {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "malformed"})
		}
		// validations
		if strings.TrimSpace(req.ZoneID) == "" {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "zone_id required"})
		}
		if !req.NBF.Before(req.EXP) {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "nbf must be before exp"})
		}
		if req.EXP.Sub(time.Now().UTC()) > cfg.MaxTTL {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "exp exceeds max ttl"})
		}

		// get active key
		var (
			kid  string
			alg  string
			pub  []byte
			priv []byte
		)
		err := pool.QueryRow(c.Request().Context(), `SELECT key_id, alg, public_key, private_key FROM issuer_keys WHERE status='active' ORDER BY created_at DESC LIMIT 1`).Scan(&kid, &alg, &pub, &priv)
		if err != nil {
			return writeJSON(c, http.StatusServiceUnavailable, APIError{Code: "no_active_key", Message: "no active issuer key"})
		}
		if alg != "EdDSA" || len(priv) != ed25519.PrivateKeySize {
			return writeJSON(c, http.StatusServiceUnavailable, APIError{Code: "unsupported_alg", Message: "only EdDSA supported"})
		}

		passID := uuid.New().String()
		nonce := make([]byte, 12)
		if _, err := rand.Read(nonce); err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "nonce"})
		}
		holderHint := util.HolderHintFromName(req.SubjectName)

		body := im.SignedPayload{
			V: 1,
			Pass: im.PayloadPass{
				ID:         passID,
				Type:       req.PolicyID,
				Level:      "",
				Scopes:     []string{req.ZoneID},
				OneTime:    req.OneTime,
				NBF:        req.NBF.UTC(),
				EXP:        req.EXP.UTC(),
				Attrs:      req.Attrs,
				HolderHint: holderHint,
			},
			Meta: im.PayloadMeta{
				OrgID:         req.OrgID,
				PolicyID:      req.PolicyID,
				ZoneContext:   "",
				IssuedAt:      time.Now().UTC(),
				Nonce:         nonce,
				SchemaVersion: 1,
			},
			IssuerKeyID: kid,
		}
		payloadB, err := json.Marshal(body)
		if err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "marshal"})
		}

		compact, sig, err := crypto.SignJWS(kid, ed25519.PrivateKey(priv), payloadB)
		if err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "sign"})
		}

		// store
		cmd := `INSERT INTO passes (id, org_id, policy_id, subject_name, zone_id, nbf, exp, one_time, issuer_key_id, signature, payload, status)
                VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'Active')`
		_, err = pool.Exec(c.Request().Context(), cmd,
			passID, req.OrgID, req.PolicyID, req.SubjectName, req.ZoneID,
			req.NBF.UTC(), req.EXP.UTC(), req.OneTime, kid, sig, []byte(compact),
		)
		if err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}

		return writeJSON(c, http.StatusCreated, dto.CreatePassResponse{
			ID:          passID,
			Status:      "Active",
			IssuerKeyID: kid,
			Payload:     compact,
		})
	}
}

// RevokePass — отзыв пропуска
// @Summary     Отзыв пропуска
// @Tags        passes
// @Produce     json
// @Param       id  path string true "Pass ID"
// @Success     200 {object} dto.RevokeResponse
// @Failure     404 {object} APIError
// @Failure     409 {object} APIError
// @Router      /passes/{id}/revoke [post]
func RevokePass(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := strings.TrimSpace(c.Param("id"))
		if id == "" {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "id"})
		}
		cmd := `UPDATE passes SET status='Revoked' WHERE id=$1 AND status='Active'`
		tag, err := pool.Exec(c.Request().Context(), cmd, id)
		if err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}
		if tag.RowsAffected() == 0 {
			// check if exists
			var exists bool
			if err := pool.QueryRow(c.Request().Context(), "SELECT EXISTS(SELECT 1 FROM passes WHERE id=$1)", id).Scan(&exists); err != nil {
				return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
			}
			if !exists {
				return writeJSON(c, http.StatusNotFound, APIError{Code: "not_found", Message: "pass not found"})
			}
			return writeJSON(c, http.StatusConflict, APIError{Code: "conflict", Message: "not Active"})
		}
		return writeJSON(c, http.StatusOK, dto.RevokeResponse{ID: id, Status: "Revoked"})
	}
}

// ApprovePass — выдаёт pickup-token для забора payload
// @Summary     Сгенерировать pickup-token
// @Tags        pickup
// @Produce     json
// @Param       id  path string true "Pass ID"
// @Success     200 {object} dto.ApproveResponse
// @Failure     404 {object} APIError
// @Failure     409 {object} APIError
// @Router      /passes/{id}/approve [post]
func ApprovePass(pool *pgxpool.Pool, cfg config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := strings.TrimSpace(c.Param("id"))
		if id == "" {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "id"})
		}
		// ensure exists and Active
		var status string
		if err := pool.QueryRow(c.Request().Context(), "SELECT status FROM passes WHERE id=$1", id).Scan(&status); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return writeJSON(c, http.StatusNotFound, APIError{Code: "not_found", Message: "pass not found"})
			}
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}
		if status != "Active" {
			return writeJSON(c, http.StatusConflict, APIError{Code: "conflict", Message: "not Active"})
		}
		// generate opaque token
		raw := make([]byte, 16)
		if _, err := rand.Read(raw); err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "token"})
		}
		token := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
		exp := time.Now().UTC().Add(1 * time.Hour)
		_, err := pool.Exec(c.Request().Context(), `INSERT INTO pickup_tokens (token, pass_id, ttl_expires_at) VALUES ($1,$2,$3)`, token, id, exp)
		if err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}
		return writeJSON(c, http.StatusOK, dto.ApproveResponse{ID: id, PickupToken: token, ExpiresAt: exp.Format(time.RFC3339)})
	}
}

// Pickup — вернуть payload по действующему pickup-токену
// @Summary     Получить payload по pickup-token
// @Tags        pickup
// @Accept      json
// @Produce     json
// @Param       request body dto.PickupRequest true "Pickup"
// @Success     200 {object} dto.PickupResponse
// @Failure     400 {object} APIError
// @Router      /pickup [post]
func Pickup(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.PickupRequest
		if err := c.Bind(&req); err != nil {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "malformed"})
		}
		tok := strings.TrimSpace(req.Token)
		if tok == "" {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_request", Message: "token required"})
		}
		// lookup
		var (
			passID string
			expAt  time.Time
			usedAt *time.Time
		)
		err := pool.QueryRow(c.Request().Context(), `SELECT pass_id, ttl_expires_at, used_at FROM pickup_tokens WHERE token=$1`, tok).Scan(&passID, &expAt, &usedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_token", Message: "invalid"})
			}
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}
		now := time.Now().UTC()
		if usedAt != nil || now.After(expAt) {
			return writeJSON(c, http.StatusBadRequest, APIError{Code: "invalid_token", Message: "expired_or_used"})
		}
		// mark used and return payload
		tx, err := pool.BeginTx(c.Request().Context(), pgx.TxOptions{})
		if err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "tx"})
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		if _, err := tx.Exec(c.Request().Context(), `UPDATE pickup_tokens SET used_at=now() WHERE token=$1`, tok); err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}
		var payload []byte
		var kid string
		if err := tx.QueryRow(c.Request().Context(), `SELECT payload, issuer_key_id FROM passes WHERE id=$1`, passID).Scan(&payload, &kid); err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}
		if err := tx.Commit(c.Request().Context()); err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}
		return writeJSON(c, http.StatusOK, dto.PickupResponse{Payload: string(payload), IssuerKeyID: kid})
	}
}
