package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"time"

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

	// migrate
	if err := repo.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("keygen: %v", err)
	}
	kid := fmt.Sprintf("key-%s", time.Now().UTC().Format("2006-01"))

	// retire existing active
	if _, err := pool.Exec(ctx, `UPDATE issuer_keys SET status='retired' WHERE status='active'`); err != nil {
		log.Fatalf("retire: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO issuer_keys (key_id, alg, public_key, private_key, status) VALUES ($1,$2,$3,$4,'active')`, kid, "EdDSA", []byte(pub), []byte(priv))
	if err != nil {
		log.Fatalf("insert key: %v", err)
	}
	log.Printf("seeded active key %s", kid)
}
