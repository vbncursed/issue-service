package http

import (
	"encoding/base64"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type jwk struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	X   string `json:"x"`
}
type jwkSet struct {
	Keys []jwk `json:"keys"`
}

// JWKS — отдать набор публичных ключей эмитента
// @Summary     JWKS набор ключей
// @Tags        keys
// @Produce     json
// @Success     200 {object} jwkSet
// @Router      /.well-known/keys [get]
func JWKS(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := pool.Query(c.Request().Context(), `SELECT key_id, alg, public_key FROM issuer_keys WHERE status IN ('active','retired')`)
		if err != nil {
			return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
		}
		defer rows.Close()
		out := jwkSet{}
		for rows.Next() {
			var kid, alg string
			var pub []byte
			if err := rows.Scan(&kid, &alg, &pub); err != nil {
				return writeJSON(c, http.StatusInternalServerError, APIError{Code: "internal", Message: "db"})
			}
			switch alg {
			case "EdDSA":
				out.Keys = append(out.Keys, jwk{
					Kty: "OKP",
					Crv: "Ed25519",
					Kid: kid,
					Alg: "EdDSA",
					X:   base64.RawURLEncoding.EncodeToString(pub),
				})
			default:
				// skip unsupported
			}
		}
		return writeJSON(c, http.StatusOK, out)
	}
}
