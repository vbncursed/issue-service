package service

import (
	"context"
	"time"
)

// Clock — абстракция времени для тестируемости
type Clock interface {
	Now() time.Time
}

// Signer — абстракция подписи JWS
type Signer interface {
	SignJWS(kid string, privateKey []byte, payload []byte) (compact string, signature []byte, err error)
}

// KeyRepository — доступ к ключам эмитента
type KeyRepository interface {
	GetActiveIssuerKey(ctx context.Context) (kid string, alg string, publicKey []byte, privateKey []byte, err error)
	ListIssuerKeys(ctx context.Context) ([]IssuerKey, error)
}

// PassRepository — порт для всех операций над пропусками и токенами
type PassRepository interface {
	InsertPass(ctx context.Context, p PassRecord) error
	RevokeActivePass(ctx context.Context, id string) error
	GetPassStatus(ctx context.Context, id string) (string, error)
	InsertPickupToken(ctx context.Context, token, passID string, exp time.Time) error
	MarkTokenUsedAndGetPass(ctx context.Context, token string) (payload []byte, kid string, err error)
}

// PassRecord — данные для сохранения пропуска (write-модель)
type PassRecord struct {
	ID          string
	OrgID       string
	PolicyID    string
	SubjectName string
	ZoneID      string
	NBF         time.Time
	EXP         time.Time
	OneTime     bool
	IssuerKeyID string
	Signature   []byte
	Payload     []byte
}

// Команда и результат для кейса IssuePass
type IssuePassCommand struct {
	OrgID       string
	PolicyID    string
	SubjectName string
	ZoneID      string
	NBF         time.Time
	EXP         time.Time
	OneTime     bool
	Attrs       map[string]any
}

type IssuePassResult struct {
	ID          string
	IssuerKeyID string
	Payload     string
}

// IssuerKey — доменная проекция ключа эмитента
type IssuerKey struct {
	KID       string
	Alg       string
	PublicKey []byte
}
