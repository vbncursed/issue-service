.PHONY: up down run lint test seed-keys seed-pass curl-create curl-revoke curl-approve curl-pickup jwks demo swagger

up:
	docker compose up --build

down:
	docker compose down -v

run:
	go run ./cmd/issue-service

lint:
	@golangci-lint run || echo "lint skipped (golangci-lint not installed)"

test:
	go test ./...

swagger:
	swag init -g main.go -o ./docs -d ./cmd/issue-service,./internal/http,./internal/models

# helper: run Go command either in docker network (if db is running) or locally
define RUN_GO
	@CID=$$(docker compose ps -q db 2>/dev/null); \
	if [ -n "$$CID" ]; then \
	  NET=$$(docker inspect -f '{{range $$k,$$v := .NetworkSettings.Networks}}{{printf "%s" $$k}}{{end}}' $$CID); \
	  docker run --rm --network $$NET -v "$$PWD":/app -w /app golang:1.25 $(1); \
	else \
	  $(1); \
	fi
endef

# seed-keys: сгенерировать и активировать Ed25519 ключ (retire старые active)
# Usage: make seed-keys [DSN=postgres://postgres:postgres@localhost:5433/issue?sslmode=disable]
seed-keys:
	$(call RUN_GO, go run ./cmd/seed-keys -db "$(or $(DSN),postgres://postgres:postgres@db:5432/issue?sslmode=disable)")

# seed-pass: создать демонстрационный пропуск напрямую в БД
# Usage: make seed-pass [DSN=postgres://postgres:postgres@localhost:5433/issue?sslmode=disable]
seed-pass:
	$(call RUN_GO, go run ./cmd/seed-pass -db "$(or $(DSN),postgres://postgres:postgres@db:5432/issue?sslmode=disable)")

# curl helpers
BASE?=http://localhost:8081

curl-create:
	@curl -s -H 'Content-Type: application/json' -X POST $(BASE)/api/v1/passes \
	  -d '{"org_id":"00000000-0000-0000-0000-000000000001","policy_id":"standard","subject_name":"Иван Иванов","zone_id":"A1","nbf":"2025-10-16T08:00:00Z","exp":"2025-10-16T18:00:00Z","one_time":true,"attrs":{"shift":"day"}}' | jq .

# Usage: make curl-revoke ID=<uuid>
curl-revoke:
	@[ -n "$(ID)" ] || (echo "Usage: make curl-revoke ID=<uuid>" && exit 2)
	@curl -s -X POST $(BASE)/api/v1/passes/$(ID)/revoke | jq .

# Usage: make curl-approve ID=<uuid>
curl-approve:
	@[ -n "$(ID)" ] || (echo "Usage: make curl-approve ID=<uuid>" && exit 2)
	@curl -s -H 'Content-Type: application/json' -X POST $(BASE)/api/v1/passes/$(ID)/approve | jq .

# Usage: make curl-pickup TOKEN=<token>
curl-pickup:
	@[ -n "$(TOKEN)" ] || (echo "Usage: make curl-pickup TOKEN=<token>" && exit 2)
	@curl -s -H 'Content-Type: application/json' -X POST $(BASE)/api/v1/pickup -d '{"token":"$(TOKEN)"}' | jq .

jwks:
	@curl -s $(BASE)/.well-known/keys | jq .

# demo: прогон — ключ, выпуск через API и JWKS
demo:
	$(MAKE) seed-keys
	$(MAKE) curl-create
	$(MAKE) jwks


