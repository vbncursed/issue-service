package service

import (
	"time"

	"crypto/ed25519"

	"github.com/vbncursed/vkr/issue-service/internal/crypto"
)

// RealClock — продовая реализация Clock
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// JWSSigner — адаптер Signer поверх internal/crypto
type JWSSigner struct{}

func (JWSSigner) SignJWS(kid string, privateKey []byte, payload []byte) (string, []byte, error) {
	return crypto.SignJWS(kid, ed25519.PrivateKey(privateKey), payload)
}
