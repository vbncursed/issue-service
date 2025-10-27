// @title         issue-service API
// @version       1.0
// @description   Сервис выпуска одноразовых пропусков.
// @BasePath      /api/v1
// @schemes       http
// @host          localhost:8081
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/vbncursed/vkr/issue-service/docs"
	icfg "github.com/vbncursed/vkr/issue-service/internal/config"
	ih "github.com/vbncursed/vkr/issue-service/internal/http"
	"github.com/vbncursed/vkr/issue-service/internal/repo"
)

func main() {
	cfg := icfg.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := repo.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if err := repo.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	e := ih.Router(pool, cfg)

	srv := &http.Server{
		Addr:              cfg.Bind,
		Handler:           e,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("issue-service listening on %s", cfg.Bind)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	shutdownCtx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	_ = srv.Shutdown(shutdownCtx)
}
