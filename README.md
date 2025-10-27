# issue-service

Сервис выпуска self-contained пропусков (JWS/COSE) с публикацией JWKS и (опц.) pickup-flow.

## Назначение
- Выпуск пропусков: формирование канонизированного payload v=1, подпись Ed25519 (JWS compact), запись в БД.
- Управление статусами: `Active`, `Revoked`, `Expired`.
- JWKS (`/.well-known/keys`) — публичные ключи эмитента для верификатора.
- Pickup-flow (опционально): безопасная выдача `payload` по одноразовому токену.

## Быстрый старт
### Локально (Go)
```bash
# переменные окружения (значения по умолчанию)
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/issue?sslmode=disable
export BIND=:8081
export MAX_TTL_H=24
export ENABLE_SWAGGER=true

# запуск
go run ./cmd/issue-service
```

### Docker Compose (общий, на корне репо)
```bash
# из корня монорепо
make up              # поднимет issue, verify и их БД
make seed-keys       # сгенерирует/активирует ключ эмитента для issue
# опционально
make seed-pass       # выпустит демо-пропуск напрямую в БД
```
Откройте Swagger UI: `http://localhost:8081/swagger/index.html`.

### Только этот сервис (локальный compose)
```bash
cd issue-service
make up
make seed-keys
```

## Переменные окружения
- `DATABASE_URL` — DSN PostgreSQL (по умолчанию: `postgres://postgres:postgres@localhost:5432/issue?sslmode=disable`).
- `BIND` — адрес HTTP (`:8081`).
- `MAX_TTL_H` — максимальный TTL пропуска в часах (лимит `exp-now`).
- `ENABLE_SWAGGER` — `true`/`1` для включения Swagger UI.

## Команды Makefile
- `make up|down` — поднять/остановить docker compose из каталога сервиса.
- `make run` — локальный запуск `go run`.
- `make seed-keys` — сгенерировать Ed25519‑ключ и сделать его активным (retire старые).
- `make seed-pass` — создать демо‑пропуск в БД (печатает ID и JWS).
- `make swagger` — сгенерировать Swagger (требуется установленный `swag`).

## Эндпоинты
Базовый префикс: `/api/v1`.

- GET `/healthz` — liveness.
- GET `/readyz` — readiness (пинг БД).
- GET `/.well-known/keys` — JWKS активных/retired ключей эмитента (OKP/Ed25519, `alg=EdDSA`).
- POST `/passes` — выпуск пропуска.
- POST `/passes/{id}/revoke` — отзыв пропуска (только из `Active`).
- POST `/passes/{id}/approve` — сгенерировать одноразовый pickup‑токен (TTL=1h).
- POST `/pickup` — получить `payload` по действующему pickup‑токену (и пометить его `used`).

### Примеры
Выпуск (окно валидно «сейчас» для macOS):
```bash
NOW_MINUS_5M=$(date -u -v-5M +%Y-%m-%dT%H:%M:%SZ); NOW_PLUS_1H=$(date -u -v+1H +%Y-%m-%dT%H:%M:%SZ)
curl -s -H 'Content-Type: application/json' -X POST http://localhost:8081/api/v1/passes -d "{
  \"org_id\":\"00000000-0000-0000-0000-000000000001\",
  \"policy_id\":\"standard\",
  \"subject_name\":\"Иван Иванов\",
  \"zone_id\":\"A1\",
  \"nbf\":\"$NOW_MINUS_5M\",
  \"exp\":\"$NOW_PLUS_1H\",
  \"one_time\":true,
  \"attrs\":{\"shift\":\"day\"}
}" | jq .
```
Отзыв:
```bash
curl -s -X POST http://localhost:8081/api/v1/passes/<PASS_ID>/revoke | jq .
```
Pickup:
```bash
curl -s -X POST http://localhost:8081/api/v1/passes/<PASS_ID>/approve | jq .
# возьмите pickup_token из ответа
curl -s -H 'Content-Type: application/json' -X POST http://localhost:8081/api/v1/pickup -d '{"token":"<TOKEN>"}' | jq .
```
JWKS:
```bash
curl -s http://localhost:8081/.well-known/keys | jq .
```

## Схема БД (миграции)
- `internal/migrations/0001_init.sql`:
  - `issuer_keys(key_id, alg, public_key, private_key, status)`
  - `passes(id, org_id, policy_id, subject_name, zone_id, nbf, exp, one_time, issuer_key_id, signature, payload, status)`
  - `pickup_tokens(token, pass_id, ttl_expires_at, used_at)`
  - Индексы: `passes(status)`, `passes(exp)`, `passes(org_id)`, `pickup_tokens(ttl_expires_at)`

Миграции применяются автоматически при старте.

## Канонизированный payload v=1
То, что подписывается и встраивается в QR (JWS compact). Без PII: `subject_name` хранится только в БД, для QR формируется `holder_hint`.
```json
{
  "v": 1,
  "pass": {
    "id": "UUID",
    "type": "<policy_id>",
    "level": "",
    "scopes": ["<zone_id>"],
    "one_time": true,
    "nbf": "ISO8601 UTC",
    "exp": "ISO8601 UTC",
    "attrs": {"shift":"day"},
    "holder_hint": "И.И."
  },
  "meta": {
    "org_id": "UUID",
    "policy_id": "string",
    "zone_context": "",
    "issued_at": "ISO8601 UTC",
    "nonce": "bytes(12)",
    "schema_version": 1
  },
  "issuer_key_id": "key-YYYY-MM"
}
```
Примечание: для совместимости `zone_id` кладётся как единственный элемент массива `pass.scopes`.

## Интеграция с verify-service
- verify берёт `payload` из клиента и проверяет подпись оффлайн, подгружая ключи по `KEYS_URL` с этого сервиса.
- В общем compose уже настроено `KEYS_URL=http://issue:8081/.well-known/keys` и `VERIFY_SKIP_SIGNATURE=false`.

## Безопасность
- PII (`subject_name`) не попадает в payload/QR; используется только `holder_hint`.
- Приватные ключи эмитента хранятся в таблице `issuer_keys` этого сервиса.
- Одноразовость обеспечивается на стороне verify-service через `pass_consumptions(pass_id, nonce)`.

## Swagger
- UI: `/swagger/index.html` (включается `ENABLE_SWAGGER=true`).
- Генерация: `make swagger` (нужен установленный `swag`: `go install github.com/swaggo/swag/cmd/swag@latest`).

## Лицензия
MIT (или укажите вашу).


