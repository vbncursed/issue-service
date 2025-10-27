package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/vbncursed/vkr/issue-service/internal/crypto"
	im "github.com/vbncursed/vkr/issue-service/internal/models"
	"github.com/vbncursed/vkr/issue-service/internal/repo"
)

func main() {
	var dbURL string
	flag.StringVar(&dbURL, "db", "postgres://postgres:postgres@localhost:5432/issue?sslmode=disable", "database url")
	flag.Parse()

	ctx := context.Background()
	pool, err := repo.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if err := repo.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	var kid, alg string
	var priv []byte
	if err := pool.QueryRow(ctx, `SELECT key_id, alg, private_key FROM issuer_keys WHERE status='active' ORDER BY created_at DESC LIMIT 1`).Scan(&kid, &alg, &priv); err != nil {
		log.Fatalf("no active key: %v", err)
	}

	passID := uuid.New().String()
	now := time.Now().UTC()
	nbf := now.Add(1 * time.Hour)
	exp := now.Add(10 * time.Hour)
	nonce := make([]byte, 12)
	_, _ = rand.Read(nonce)
	body := im.SignedPayload{
		V: 1,
		Pass: im.PayloadPass{
			ID:         passID,
			Type:       "demo",
			Level:      "",
			Scopes:     []string{"A1"},
			OneTime:    true,
			NBF:        nbf,
			EXP:        exp,
			Attrs:      map[string]any{"shift": "day"},
			HolderHint: "D.D.",
		},
		Meta: im.PayloadMeta{
			OrgID:         uuid.New().String(),
			PolicyID:      "demo",
			ZoneContext:   "",
			IssuedAt:      now,
			Nonce:         nonce,
			SchemaVersion: 1,
		},
		IssuerKeyID: kid,
	}
	payloadB, _ := json.Marshal(body)
	compact, sig, err := crypto.SignJWS(kid, ed25519.PrivateKey(priv), payloadB)
	if err != nil {
		log.Fatalf("sign: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO passes (id, org_id, policy_id, subject_name, zone_id, nbf, exp, one_time, issuer_key_id, signature, payload, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'Active')`,
		passID, uuid.New().String(), "demo", "Demo User", "A1", nbf, exp, true, kid, sig, []byte(compact))
	if err != nil {
		log.Fatalf("insert pass: %v", err)
	}
	fmt.Println("Demo pass ID:", passID)
	fmt.Println("Payload (JWS):", compact)
}
