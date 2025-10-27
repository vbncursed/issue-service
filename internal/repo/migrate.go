package repo

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vbncursed/vkr/issue-service/internal/migrations"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations(
  id TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`)
	if err != nil {
		return err
	}

	ents, err := migrations.Files.ReadDir(".")
	if err != nil {
		return err
	}
	var files []string
	for _, e := range ents {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		var exists bool
		if err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE id=$1)", f).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		b, err := migrations.Files.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := conn.Exec(ctx, string(b)); err != nil {
			return fmt.Errorf("migration %s: %w", f, err)
		}
		if _, err := conn.Exec(ctx, "INSERT INTO schema_migrations(id) VALUES($1)", f); err != nil {
			return err
		}
	}
	return nil
}
