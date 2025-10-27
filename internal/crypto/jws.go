package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
)

type JWSHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ,omitempty"`
}

// SignJWS создает compact JWS с alg EdDSA
func SignJWS(kid string, priv ed25519.PrivateKey, payload []byte) (compact string, sigRaw []byte, err error) {
	hdr := JWSHeader{Alg: "EdDSA", Kid: kid}
	hdrB, err := json.Marshal(hdr)
	if err != nil {
		return "", nil, err
	}
	hEnc := base64.RawURLEncoding.EncodeToString(hdrB)
	pEnc := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := hEnc + "." + pEnc
	// Ed25519 signs raw input; add random reader to keep api stable
	_ = rand.Reader // not used but keep import
	sig := ed25519.Sign(priv, []byte(signingInput))
	sEnc := base64.RawURLEncoding.EncodeToString(sig)
	return signingInput + "." + sEnc, sig, nil
}
