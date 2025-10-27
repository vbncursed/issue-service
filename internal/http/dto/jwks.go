package dto

import (
	"encoding/base64"

	issvc "github.com/vbncursed/vkr/issue-service/internal/service"
)

type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	X   string `json:"x"`
}
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// FromIssuerKeys маппит доменные ключи в JWKSet
func FromIssuerKeys(keys []issvc.IssuerKey) JWKSet {
	out := JWKSet{Keys: make([]JWK, 0, len(keys))}
	for _, k := range keys {
		switch k.Alg {
		case "EdDSA":
			out.Keys = append(out.Keys, JWK{
				Kty: "OKP",
				Crv: "Ed25519",
				Kid: k.KID,
				Alg: "EdDSA",
				X:   base64.RawURLEncoding.EncodeToString(k.PublicKey),
			})
		default:
			// skip unsupported alg
		}
	}
	return out
}
