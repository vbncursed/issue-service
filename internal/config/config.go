package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Bind          string
	DatabaseURL   string
	MaxTTL        time.Duration
	EnableSwagger bool
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	bind := getenv("BIND", ":8081")
	db := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/issue?sslmode=disable")
	ttlHoursStr := getenv("MAX_TTL_H", "24")
	ttlHours, err := strconv.Atoi(ttlHoursStr)
	if err != nil || ttlHours <= 0 {
		ttlHours = 24
	}
	swagEnv := getenv("ENABLE_SWAGGER", "false")
	swag := swagEnv == "true" || strings.EqualFold(swagEnv, "true")
	cfg := Config{
		Bind:          bind,
		DatabaseURL:   db,
		MaxTTL:        time.Duration(ttlHours) * time.Hour,
		EnableSwagger: swag,
	}
	log.Printf("config: bind=%s ttl=%s swagger=%v", cfg.Bind, cfg.MaxTTL, cfg.EnableSwagger)
	return cfg
}
