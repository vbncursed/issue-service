package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	imodels "github.com/vbncursed/vkr/issue-service/internal/models"
	"github.com/vbncursed/vkr/issue-service/internal/util"
)

// Service реализует use case'ы выпуска
type Service struct {
	keys   KeyRepository
	passes PassRepository
	clock  Clock
	signer Signer
}

func New(keys KeyRepository, passes PassRepository, clock Clock, signer Signer) *Service {
	return &Service{keys: keys, passes: passes, clock: clock, signer: signer}
}

// ошибки вынесены в errors.go

// IssuePass — основной сценарий выпуска
func (s *Service) IssuePass(ctx context.Context, cmd IssuePassCommand) (IssuePassResult, error) {
	kid, alg, _, priv, err := s.keys.GetActiveIssuerKey(ctx)
	if err != nil {
		return IssuePassResult{}, err
	}
	if alg != "EdDSA" {
		return IssuePassResult{}, ErrUnsupportedAlg
	}

	passID := uuid.New().String()
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return IssuePassResult{}, err
	}
	holderHint := util.HolderHintFromName(cmd.SubjectName)

	body := imodels.SignedPayload{
		V: 1,
		Pass: imodels.PayloadPass{
			ID:         passID,
			Type:       cmd.PolicyID,
			Level:      "",
			Scopes:     []string{cmd.ZoneID},
			OneTime:    cmd.OneTime,
			NBF:        cmd.NBF.UTC(),
			EXP:        cmd.EXP.UTC(),
			Attrs:      cmd.Attrs,
			HolderHint: holderHint,
		},
		Meta: imodels.PayloadMeta{
			OrgID:         cmd.OrgID,
			PolicyID:      cmd.PolicyID,
			ZoneContext:   "",
			IssuedAt:      s.clock.Now().UTC(),
			Nonce:         nonce,
			SchemaVersion: 1,
		},
		IssuerKeyID: kid,
	}
	payloadB, err := json.Marshal(body)
	if err != nil {
		return IssuePassResult{}, err
	}

	compact, sig, err := s.signer.SignJWS(kid, priv, payloadB)
	if err != nil {
		return IssuePassResult{}, err
	}

	rec := PassRecord{
		ID:          passID,
		OrgID:       cmd.OrgID,
		PolicyID:    cmd.PolicyID,
		SubjectName: cmd.SubjectName,
		ZoneID:      cmd.ZoneID,
		NBF:         cmd.NBF.UTC(),
		EXP:         cmd.EXP.UTC(),
		OneTime:     cmd.OneTime,
		IssuerKeyID: kid,
		Signature:   sig,
		Payload:     []byte(compact),
	}
	if err := s.passes.InsertPass(ctx, rec); err != nil {
		return IssuePassResult{}, err
	}
	return IssuePassResult{ID: passID, IssuerKeyID: kid, Payload: compact}, nil
}

// RevokePass — смена статуса на Revoked
func (s *Service) RevokePass(ctx context.Context, id string) error {
	return s.passes.RevokeActivePass(ctx, id)
}

type ApproveResult struct {
	Token     string
	ExpiresAt string
}

// ApprovePass — генерирует pickup-token
func (s *Service) ApprovePass(ctx context.Context, id string, ttl time.Duration) (ApproveResult, error) {
	st, err := s.passes.GetPassStatus(ctx, id)
	if err != nil {
		return ApproveResult{}, err
	}
	if st != string(imodels.StatusActive) {
		return ApproveResult{}, ErrConflict
	}
	// генерируем токен
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return ApproveResult{}, err
	}
	token := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
	exp := s.clock.Now().UTC().Add(ttl)
	if err := s.passes.InsertPickupToken(ctx, token, id, exp); err != nil {
		return ApproveResult{}, err
	}
	return ApproveResult{Token: token, ExpiresAt: exp.Format(time.RFC3339)}, nil
}

type PickupResult struct {
	Payload     string
	IssuerKeyID string
}

// Pickup — атомарно помечает токен использованным и возвращает payload
func (s *Service) Pickup(ctx context.Context, token string) (PickupResult, error) {
	payload, kid, err := s.passes.MarkTokenUsedAndGetPass(ctx, token)
	if err != nil {
		return PickupResult{}, err
	}
	return PickupResult{Payload: string(payload), IssuerKeyID: kid}, nil
}

// ListIssuerKeys — список ключей эмитента для JWKS
func (s *Service) ListIssuerKeys(ctx context.Context) ([]IssuerKey, error) {
	return s.keys.ListIssuerKeys(ctx)
}
