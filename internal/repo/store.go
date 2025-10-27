package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	im "github.com/vbncursed/vkr/issue-service/internal/models"
	"github.com/vbncursed/vkr/issue-service/internal/service"
)

// Store — адаптер Postgres, реализующий порты service.*
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// KeyRepository
func (s *Store) GetActiveIssuerKey(ctx context.Context) (kid string, alg string, publicKey []byte, privateKey []byte, err error) {
	err = s.pool.QueryRow(ctx, `SELECT `+colKeyID+`, `+colAlg+`, `+colPublicKey+`, `+colPrivateKey+` FROM `+tableIssuerKeys+` WHERE `+colStatus+`='active' ORDER BY `+colCreatedAt+` DESC LIMIT 1`).
		Scan(&kid, &alg, &publicKey, &privateKey)
	return
}

// ListIssuerKeys — активные и retired ключи эмитента
func (s *Store) ListIssuerKeys(ctx context.Context) ([]service.IssuerKey, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+colKeyID+`, `+colAlg+`, `+colPublicKey+` FROM `+tableIssuerKeys+` WHERE `+colStatus+` IN ('active','retired')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []service.IssuerKey
	for rows.Next() {
		var k service.IssuerKey
		if err := rows.Scan(&k.KID, &k.Alg, &k.PublicKey); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, nil
}

// PassWriter
func (s *Store) InsertPass(ctx context.Context, p service.PassRecord) error {
	cmd := `INSERT INTO ` + tablePasses + ` (` +
		colID + `, ` + colOrgID + `, ` + colPolicyID + `, ` + colSubjectName + `, ` + colZoneID + `, ` +
		colNbf + `, ` + colExp + `, ` + colOneTime + `, ` + colIssuerKeyID + `, ` + colSignature + `, ` + colPayload + `, ` + colStatus + `)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := s.pool.Exec(ctx, cmd,
		p.ID, p.OrgID, p.PolicyID, p.SubjectName, p.ZoneID,
		p.NBF, p.EXP, p.OneTime, p.IssuerKeyID, p.Signature, p.Payload,
		string(im.StatusActive),
	)
	return err
}

// RevokeActivePass — устанавливает статус Revoked, возвращает ErrNotFound/ErrConflict
func (s *Store) RevokeActivePass(ctx context.Context, id string) error {
	cmd := `UPDATE ` + tablePasses + ` SET ` + colStatus + `=$1 WHERE ` + colID + `=$2 AND ` + colStatus + `=$3`
	tag, err := s.pool.Exec(ctx, cmd, string(im.StatusRevoked), id, string(im.StatusActive))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM "+tablePasses+" WHERE "+colID+"=$1)", id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return service.ErrNotFound
	}
	return service.ErrConflict
}

// GetPassStatus — возвращает статус или ErrNotFound
func (s *Store) GetPassStatus(ctx context.Context, id string) (string, error) {
	var status string
	if err := s.pool.QueryRow(ctx, "SELECT "+colStatus+" FROM "+tablePasses+" WHERE "+colID+"=$1", id).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", service.ErrNotFound
		}
		return "", err
	}
	return status, nil
}

// InsertPickupToken — сохраняет pickup-token
func (s *Store) InsertPickupToken(ctx context.Context, token, passID string, exp time.Time) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO `+tablePickupTokens+` (`+colToken+`, `+colPassID+`, `+colTTLExpiresAt+`) VALUES ($1,$2,$3)`, token, passID, exp)
	return err
}

// MarkTokenUsedAndGetPass — атомарно помечает токен и возвращает payload
func (s *Store) MarkTokenUsedAndGetPass(ctx context.Context, token string) ([]byte, string, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var passID string
	if err := tx.QueryRow(ctx, `UPDATE `+tablePickupTokens+` SET `+colUsedAt+`=now() WHERE `+colToken+`=$1 AND `+colUsedAt+` IS NULL AND `+colTTLExpiresAt+` > now() RETURNING `+colPassID+``, token).Scan(&passID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", service.ErrExpiredOrUsed
		}
		return nil, "", err
	}
	var payload []byte
	var kid string
	if err := tx.QueryRow(ctx, `SELECT `+colPayload+`, `+colIssuerKeyID+` FROM `+tablePasses+` WHERE `+colID+`=$1`, passID).Scan(&payload, &kid); err != nil {
		return nil, "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, "", err
	}
	return payload, kid, nil
}
