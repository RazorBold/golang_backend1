# IoT Platform Backend — Production Plan

## Overview

Platform IoT berbasis REST API menggunakan Go, dengan Redis sebagai cache & rate limiter,
RabbitMQ sebagai message broker async, PostgreSQL sebagai primary database,
Docker untuk containerisasi, dan Kubernetes untuk orkestrasi.

Semua device IoT berkomunikasi via HTTP ke backend ini. Data telemetry diproses
secara async melalui RabbitMQ sehingga ingest endpoint tetap cepat meskipun
downstream processing lambat.

---

## Tech Stack

| Layer               | Teknologi                                  |
|---------------------|--------------------------------------------|
| Language            | Go 1.22+                                   |
| HTTP Framework      | Fiber v2                                   |
| Auth                | JWT (access + refresh token)               |
| Primary DB          | PostgreSQL 16                              |
| Cache & Rate Limit  | Redis 7                                    |
| Message Broker      | RabbitMQ 3.13 (AMQP 0-9-1)               |
| Query Builder       | sqlc (type-safe SQL codegen)               |
| DB Migration        | golang-migrate                             |
| Container           | Docker (multi-stage) + Docker Compose      |
| Orchestration       | Kubernetes (K8s)                           |
| API Docs            | Swagger UI via swaggo/swag                 |
| Config              | Viper + .env                               |
| Logging             | Zerolog (structured JSON)                  |
| Metrics             | Prometheus + Grafana                       |
| Tracing             | OpenTelemetry + Jaeger                     |
| Edge / Security     | Cloudflare (WAF, DDoS, Tunnel)             |

---

## Domain / Entities

```
User          → login, register, manage API keys
Application   → logical grouping of devices (e.g., "Smart Farm", "Factory Floor")
Device        → physical IoT device, belongs to an Application
DeviceData    → time-series telemetry/readings dari Device
```

---

## Project Structure

```
golang_backend1/
├── cmd/
│   ├── server/
│   │   └── main.go                    # HTTP server entrypoint
│   └── worker/
│       └── main.go                    # RabbitMQ consumer entrypoint
├── internal/
│   ├── config/
│   │   └── config.go                  # load env via Viper
│   ├── server/
│   │   └── server.go                  # init Fiber, routes, middleware
│   ├── middleware/
│   │   ├── auth.go                    # JWT validation
│   │   ├── apikey.go                  # Device API key validation
│   │   ├── rate_limiter.go            # Redis sliding window rate limit
│   │   └── logger.go                  # request logging (zerolog)
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── application_handler.go
│   │   ├── device_handler.go
│   │   └── telemetry_handler.go
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── application_service.go
│   │   ├── device_service.go
│   │   └── telemetry_service.go
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── application_repo.go
│   │   ├── device_repo.go
│   │   └── telemetry_repo.go
│   ├── model/
│   │   ├── user.go
│   │   ├── application.go
│   │   ├── device.go
│   │   └── telemetry.go
│   ├── cache/
│   │   └── redis.go                   # Redis client wrapper
│   ├── broker/
│   │   ├── publisher.go               # publish pesan ke RabbitMQ
│   │   ├── consumer.go                # consume message dari queue
│   │   └── handler/
│   │       ├── telemetry_handler.go   # proses telemetry message
│   │       └── alert_handler.go       # proses alert message
│   └── pkg/
│       ├── password/                  # bcrypt helper
│       ├── token/                     # JWT helper
│       ├── response/                  # standard API response wrapper
│       └── validator/                 # input validation helper
├── db/
│   ├── migrations/                    # SQL files (up & down)
│   └── queries/                       # sqlc .sql query files
├── docs/                              # auto-generated swagger docs (swaggo)
├── deploy/
│   ├── docker/
│   │   └── Dockerfile
│   ├── docker-compose.yml             # local dev stack
│   └── k8s/
│       ├── namespace.yaml
│       ├── configmap.yaml
│       ├── secret.yaml
│       ├── deployment-api.yaml
│       ├── deployment-worker.yaml
│       ├── service.yaml
│       ├── ingress.yaml
│       ├── hpa.yaml
│       ├── redis/
│       │   ├── deployment.yaml
│       │   └── service.yaml
│       ├── rabbitmq/
│       │   ├── statefulset.yaml
│       │   └── service.yaml
│       └── postgres/
│           ├── statefulset.yaml
│           └── service.yaml
├── test/
│   ├── unit/                          # unit tests per package
│   ├── integration/                   # integration tests (testcontainers)
│   ├── load/                          # k6 load test scripts
│   └── api/                           # Bruno / Hurl API test collections
├── .env.example
├── .golangci.yml                      # linter config
├── sqlc.yaml
├── go.mod
├── go.sum
├── Makefile
└── plan.md
```

---

## Database Schema (PostgreSQL)

```sql
-- users
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    role        VARCHAR(20) DEFAULT 'user',        -- 'admin' | 'user'
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- applications
CREATE TABLE applications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    api_key     VARCHAR(64) UNIQUE NOT NULL,       -- device auth key
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- devices
CREATE TABLE devices (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name           VARCHAR(100) NOT NULL,
    type           VARCHAR(50),                    -- 'sensor' | 'actuator' | 'gateway'
    status         VARCHAR(20) DEFAULT 'inactive', -- 'active' | 'inactive' | 'error'
    last_seen_at   TIMESTAMPTZ,
    metadata       JSONB,                          -- custom attributes per device
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

-- device_data (telemetry, time-series)
CREATE TABLE device_data (
    id          BIGSERIAL PRIMARY KEY,
    device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    payload     JSONB NOT NULL,                    -- { "temperature": 28.5, "humidity": 60 }
    received_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_device_data_device_id_received ON device_data(device_id, received_at DESC);

-- refresh_tokens
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       TEXT UNIQUE NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
```

---

## API Endpoints

### Auth
```
POST   /api/v1/auth/register                → register user baru
POST   /api/v1/auth/login                   → login → return access + refresh token
POST   /api/v1/auth/refresh                 → refresh access token
POST   /api/v1/auth/logout                  → revoke refresh token
```

### Users  `[JWT required]`
```
GET    /api/v1/users/me                     → get current user profile
PUT    /api/v1/users/me                     → update profile
```

### Applications  `[JWT required]`
```
POST   /api/v1/applications                 → create application
GET    /api/v1/applications                 → list applications milik user
GET    /api/v1/applications/:id             → detail application
PUT    /api/v1/applications/:id             → update application
DELETE /api/v1/applications/:id             → delete application
POST   /api/v1/applications/:id/regenerate-key  → regenerate API key
```

### Devices  `[JWT required]`
```
POST   /api/v1/applications/:app_id/devices          → create device
GET    /api/v1/applications/:app_id/devices          → list devices
GET    /api/v1/applications/:app_id/devices/:id      → detail device
PUT    /api/v1/applications/:app_id/devices/:id      → update device
DELETE /api/v1/applications/:app_id/devices/:id      → delete device
```

### Telemetry — Device → Backend  `[X-API-Key required]`
```
POST   /api/v1/ingest/:device_id            → device kirim data sensor
GET    /api/v1/devices/:device_id/data      → query telemetry (?from=&to=&limit=)
GET    /api/v1/devices/:device_id/latest    → data terbaru dari device
```

### Health & Observability
```
GET    /health                              → liveness probe (K8s)
GET    /ready                               → readiness probe (cek DB, Redis, RabbitMQ)
GET    /metrics                             → Prometheus metrics
GET    /swagger/*                           → Swagger UI (docs)
```

---

## RabbitMQ Architecture

### Topology

```
Exchange: iot.telemetry  (type: topic, durable)
Exchange: iot.alerts     (type: direct, durable)
Exchange: iot.dlx        (Dead Letter Exchange, type: direct)

Queues:
  telemetry.persist      → routing key: telemetry.#
                           consumer: simpan ke PostgreSQL
  telemetry.process      → routing key: telemetry.#
                           consumer: threshold check, anomaly detection
  alerts.notify          → routing key: alert.*
                           consumer: kirim notifikasi (email/webhook)
  dlq.telemetry          → dead letter queue (pesan gagal setelah 3x retry)
```

### Flow Telemetry

```
Device
  │
  ▼  HTTP POST /api/v1/ingest/:device_id
API Server
  │  validate X-API-Key via Redis cache
  │  update device.last_seen_at (async, non-blocking)
  │
  ├──► Publish ke iot.telemetry exchange
  │     routing key: telemetry.<app_id>.<device_id>
  │     payload: { device_id, app_id, received_at, data }
  │
  ▼  Response 201 (cepat, tidak tunggu processing)

Worker (cmd/worker)
  ├── Consumer telemetry.persist → INSERT INTO device_data
  ├── Consumer telemetry.process → cek threshold, publish ke iot.alerts jika anomali
  └── Consumer alerts.notify    → kirim webhook/email notifikasi
```

### Message Payload

```json
{
  "event_id":    "uuid-v4",
  "device_id":   "uuid-v4",
  "app_id":      "uuid-v4",
  "received_at": "2026-05-19T10:00:00Z",
  "data": {
    "temperature": 28.5,
    "humidity": 60.2,
    "battery": 95
  }
}
```

### RabbitMQ Best Practices yang diterapkan

| Practice                  | Detail                                                     |
|---------------------------|------------------------------------------------------------|
| Durable queues            | Pesan survive RabbitMQ restart                             |
| Persistent messages       | `delivery_mode: 2`                                         |
| Publisher confirms        | API tidak return 201 sebelum RabbitMQ ack publish          |
| Consumer ack manual       | Ack setelah DB insert berhasil, bukan setelah receive      |
| Dead Letter Queue (DLQ)   | Pesan gagal 3x masuk dlq.telemetry untuk investigasi       |
| Prefetch count            | `prefetch_count: 10` per consumer agar tidak overwhelm     |
| Connection pooling        | Satu connection, banyak channel                            |

---

## Redis Usage

| Use Case               | Key Pattern                        | TTL       |
|------------------------|------------------------------------|-----------|
| JWT Blacklist          | `blacklist:<jti>`                  | sisa TTL token |
| Rate Limiting          | `ratelimit:<ip>:<minute>`          | 60s       |
| Device Status Cache    | `device:<id>:status`               | 30s       |
| Latest Telemetry Cache | `device:<id>:latest`               | 10s       |
| API Key → App ID Cache | `apikey:<key>`                     | 5 menit   |
| Pub/Sub Realtime       | channel `telemetry.<device_id>`    | -         |

---

## Authentication Flow

```
1. Register → bcrypt hash password (cost 12) → simpan ke DB

2. Login → validasi password → generate:
   - Access Token  (JWT HS256, TTL 15 menit, claims: user_id, role, jti)
   - Refresh Token (random UUID, TTL 7 hari, simpan di tabel refresh_tokens)

3. Protected endpoint → JWT middleware:
   - Verify signature & expiry
   - Cek Redis blacklist (key: blacklist:<jti>)

4. Token expired → POST /auth/refresh dengan refresh_token
   - Validasi di DB (ada & belum expired)
   - Rotate: hapus token lama, generate pasangan baru

5. Logout → hapus refresh_token dari DB + masukkan JTI ke Redis blacklist

Device Auth Flow:
   - Device kirim X-API-Key di header
   - Middleware cek Redis cache (apikey:<key>)
   - Cache miss → query DB applications → cache 5 menit
   - Validasi device_id milik application tersebut
```

---

## Testing Stack

### 1. Unit Tests — `testing` + `testify` + `mockery`

```
Packages yang ditest:
  - service/*   → business logic (mock repository)
  - pkg/token   → JWT generate & validate
  - pkg/password → bcrypt hash & compare
  - middleware/* → unit test middleware logic

Tools:
  go test ./internal/...
  github.com/stretchr/testify/assert
  github.com/stretchr/testify/mock
  github.com/vektra/mockery/v2        (generate mock dari interface)
```

### 2. Integration Tests — `testcontainers-go`

```
Spin up real PostgreSQL, Redis, RabbitMQ via Docker di dalam test.
Tidak ada mock untuk infrastruktur.

Test scenarios:
  - Auth flow end-to-end (register → login → refresh → logout)
  - Telemetry ingest → RabbitMQ publish → consumer → DB insert
  - Rate limiter dengan real Redis
  - API key cache hit/miss

Tool: github.com/testcontainers/testcontainers-go
File: test/integration/
```

### 3. HTTP / API Tests — `Bruno` (open-source)

```
Bruno adalah Postman alternative yang file-based (commit ke git).

Struktur collections:
  test/api/
  ├── auth/
  │   ├── register.bru
  │   ├── login.bru
  │   └── refresh.bru
  ├── applications/
  │   ├── create.bru
  │   └── list.bru
  ├── devices/
  │   ├── create.bru
  │   └── ingest.bru
  └── environments/
      ├── local.bru
      └── staging.bru

Run via CLI: bru run test/api/ --env local
```

### 4. Load / Stress Tests — `k6`

```javascript
// test/load/ingest_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '1m', target: 100 },   // ramp up
    { duration: '3m', target: 500 },   // sustained load
    { duration: '1m', target: 0 },     // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],   // 95% request < 200ms
    http_req_failed: ['rate<0.01'],     // error rate < 1%
  },
};

// Target: 500 concurrent devices kirim telemetry simultan
export default function () {
  http.post(`${__ENV.BASE_URL}/api/v1/ingest/${DEVICE_ID}`,
    JSON.stringify({ temperature: 28.5 }),
    { headers: { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' } }
  );
  sleep(1);
}
```

### 5. Linting & Static Analysis

```
golangci-lint run ./...

Linters aktif (.golangci.yml):
  - errcheck       → pastikan semua error di-handle
  - govet          → suspicious constructs
  - staticcheck    → advanced static analysis
  - gosec          → security issues
  - gofumpt        → strict formatter
  - exhaustruct    → struct fully initialized
  - revive         → style consistency
```

### 6. Coverage Target

```
Unit test coverage    : ≥ 80% untuk internal/service dan internal/pkg
Integration coverage  : semua happy path + critical error path
Load test threshold   : p95 < 200ms pada 500 concurrent devices
```

---

## API Documentation (Swagger / OpenAPI)

### Setup dengan `swaggo/swag`

```go
// cmd/server/main.go
// @title           IoT Platform API
// @version         1.0
// @description     REST API untuk platform IoT — device management & telemetry
// @host            api.iot-platform.com
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
```

```go
// Contoh anotasi di handler
// @Summary      Ingest telemetry data
// @Description  Device mengirim data sensor ke platform
// @Tags         telemetry
// @Accept       json
// @Produce      json
// @Param        device_id  path      string            true  "Device ID"
// @Param        payload    body      model.IngestReq   true  "Sensor data"
// @Success      201        {object}  response.Success
// @Failure      401        {object}  response.Error
// @Failure      429        {object}  response.Error
// @Security     ApiKeyAuth
// @Router       /ingest/{device_id} [post]
func (h *TelemetryHandler) Ingest(c *fiber.Ctx) error { ... }
```

```
Generate: make swag     → swag init -g cmd/server/main.go -o docs/
Access  : GET /swagger/ → Swagger UI interaktif
Output  : docs/swagger.json, docs/swagger.yaml
```

---

## Cloudflare — Perlu atau Tidak?

**Rekomendasi: YA, gunakan Cloudflare** untuk IoT platform di production.

### Alasan

| Masalah IoT                              | Solusi Cloudflare                              |
|------------------------------------------|------------------------------------------------|
| Device di-spoof, traffic palsu masuk     | WAF rules blokir request tanpa X-API-Key valid |
| DDoS dari banyak device compromised      | DDoS protection di edge, sebelum masuk K8s     |
| Expose LoadBalancer IP publik di K8s     | Cloudflare Tunnel: tidak perlu IP publik sama sekali |
| Rate limiting di level aplikasi kurang   | Edge rate limiting + Redis rate limiting = 2 layer |
| SSL/TLS termination di K8s mahal         | Cloudflare handle SSL gratis                   |

### Arsitektur dengan Cloudflare

```
Device / Browser
      │
      ▼
 Cloudflare Edge
  ├── DDoS protection (L3/L4/L7)
  ├── WAF (block SQLi, bad bots, invalid headers)
  ├── Edge Rate Limiting (tambahan layer sebelum Redis)
  └── SSL termination
      │
      ▼
 Cloudflare Tunnel (cloudflared daemon di dalam K8s)
      │  Tidak perlu expose port / LoadBalancer IP publik
      ▼
 K8s Ingress (nginx)
      │
      ▼
 iot-backend Pods
```

### Apa yang TIDAK digantikan Cloudflare

- Redis rate limiting tetap diperlukan (untuk per-device, per-API-key granularity)
- JWT auth tetap di aplikasi
- WAF aplikasi (input validation) tetap di Go code

### Kapan skip Cloudflare

- Kalau deploy di internal network / private cloud (tidak ada internet-facing)
- Kalau device pakai mTLS (mutual TLS) langsung ke K8s Ingress

---

## Docker Setup

### Dockerfile (multi-stage, non-root)

```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o worker ./cmd/worker

# Stage 2: Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/worker .
USER app
EXPOSE 8080
CMD ["./server"]
```

### docker-compose.yml (local dev)

```yaml
services:
  api:
    build: .
    command: ./server
    ports: ["8080:8080"]
    env_file: .env
    depends_on: [postgres, redis, rabbitmq]

  worker:
    build: .
    command: ./worker
    env_file: .env
    depends_on: [postgres, redis, rabbitmq]

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: iot_platform
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: secret
    volumes: [pgdata:/var/lib/postgresql/data]
    ports: ["5432:5432"]

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes: [redisdata:/data]
    ports: ["6379:6379"]

  rabbitmq:
    image: rabbitmq:3.13-management-alpine
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    ports:
      - "5672:5672"   # AMQP
      - "15672:15672" # Management UI
    volumes: [rabbitmqdata:/var/lib/rabbitmq]

  prometheus:
    image: prom/prometheus:latest
    volumes: [./deploy/prometheus.yml:/etc/prometheus/prometheus.yml]
    ports: ["9090:9090"]

  grafana:
    image: grafana/grafana:latest
    ports: ["3000:3000"]
    depends_on: [prometheus]

volumes:
  pgdata:
  redisdata:
  rabbitmqdata:
```

---

## Kubernetes Architecture

```
Namespace: iot-platform
│
├── Deployment: iot-api     (replicas: 3, stateless)
│   ├── Resources: req 100m CPU/128Mi, limit 500m/512Mi
│   ├── Liveness:  GET /health
│   ├── Readiness: GET /ready (cek DB, Redis, RabbitMQ)
│   └── HPA: min 3 → max 10 pods @ CPU 70%
│
├── Deployment: iot-worker  (replicas: 2, stateless)
│   ├── Resources: req 200m CPU/256Mi, limit 1/512Mi
│   └── HPA: min 2 → max 8 pods @ custom metric (queue depth)
│
├── Service: iot-api-svc (ClusterIP)
│
├── Ingress: iot-ingress (nginx)
│   └── /api/v1/* → iot-api-svc:8080
│
├── StatefulSet: postgres    (replicas: 1, PVC 20Gi)
│   └── Service: postgres-svc (ClusterIP)
│
├── StatefulSet: rabbitmq    (replicas: 1, PVC 10Gi)
│   └── Service: rabbitmq-svc (ClusterIP, port 5672)
│
├── Deployment: redis        (replicas: 1, PVC 5Gi)
│   └── Service: redis-svc (ClusterIP)
│
├── ConfigMap: iot-config    (APP_ENV, LOG_LEVEL, dll)
└── Secret: iot-secret       (DB_PASSWORD, JWT_SECRET, RABBITMQ_PASS, dll)
```

---

## Implementation Phases

### Phase 1 — Foundation (Week 1)
- [ ] Setup project structure & go.mod dependencies
- [ ] Config loader (Viper + .env)
- [ ] PostgreSQL connection + migrations (golang-migrate)
- [ ] Redis connection wrapper
- [ ] Standard response format (`pkg/response`)
- [ ] Input validator (`pkg/validator` pakai go-playground/validator)
- [ ] Structured logging setup (zerolog)
- [ ] Health & readiness endpoints

### Phase 2 — Auth & User (Week 1-2)
- [ ] Register & login endpoint
- [ ] JWT access token + refresh token logic
- [ ] JWT middleware
- [ ] Blacklist logout di Redis
- [ ] Refresh token rotation
- [ ] Rate limiter middleware (Redis sliding window)
- [ ] Unit tests untuk auth service & token pkg

### Phase 3 — Core Domain (Week 2-3)
- [ ] CRUD Applications + API key generation (crypto/rand)
- [ ] CRUD Devices
- [ ] API Key middleware (Redis cache + DB fallback)
- [ ] Telemetry ingest endpoint → publish ke RabbitMQ
- [ ] RabbitMQ topology setup (exchange, queue, binding, DLQ)
- [ ] Worker: consumer telemetry.persist → INSERT device_data
- [ ] Worker: consumer telemetry.process → threshold check
- [ ] Cache device status & latest telemetry di Redis
- [ ] Query telemetry dengan filter waktu

### Phase 4 — Testing (Week 3-4)
- [ ] Setup mockery, generate mocks untuk semua interface repo
- [ ] Unit tests service layer (≥80% coverage)
- [ ] Integration tests dengan testcontainers-go
- [ ] Bruno collection untuk semua endpoint
- [ ] k6 load test script untuk ingest endpoint
- [ ] Setup golangci-lint + CI pipeline (GitHub Actions)

### Phase 5 — API Docs & Observability (Week 4)
- [ ] Swagger annotations di semua handler
- [ ] Generate dan serve Swagger UI di /swagger/
- [ ] Prometheus metrics (latency histogram, ingest counter, queue depth)
- [ ] Grafana dashboard import
- [ ] OpenTelemetry tracing + Jaeger

### Phase 6 — Containerization (Week 5)
- [ ] Dockerfile multi-stage (non-root user)
- [ ] docker-compose dengan semua services
- [ ] Makefile commands lengkap
- [ ] CI: build & push Docker image ke registry

### Phase 7 — Kubernetes + Cloudflare (Week 5-6)
- [ ] K8s manifests (namespace, configmap, secret)
- [ ] Deployment API + Worker + HPA
- [ ] StatefulSet PostgreSQL + RabbitMQ dengan PVC
- [ ] Redis deployment
- [ ] Ingress nginx configuration
- [ ] Cloudflare Tunnel setup (cloudflared)
- [ ] Cloudflare WAF rules (block invalid headers, geo-block jika perlu)
- [ ] Smoke test end-to-end di staging

---

## Key Dependencies (go.mod)

```
# HTTP
github.com/gofiber/fiber/v2

# Auth
github.com/golang-jwt/jwt/v5
golang.org/x/crypto                      # bcrypt

# Database
github.com/jackc/pgx/v5                  # PostgreSQL driver
github.com/sqlc-dev/sqlc                 # codegen (tool)
github.com/golang-migrate/migrate/v4     # migrations

# Cache
github.com/redis/go-redis/v9

# Message Broker
github.com/rabbitmq/amqp091-go           # RabbitMQ AMQP client

# Config & Utils
github.com/spf13/viper
github.com/google/uuid
github.com/go-playground/validator/v10

# Logging & Observability
github.com/rs/zerolog
github.com/prometheus/client_golang
go.opentelemetry.io/otel
go.opentelemetry.io/otel/exporters/jaeger

# Docs
github.com/swaggo/swag                   # codegen (tool)
github.com/swaggo/fiber-swagger

# Testing
github.com/stretchr/testify
github.com/vektra/mockery/v2             # mock codegen (tool)
github.com/testcontainers/testcontainers-go
```

---

## Makefile

```makefile
.PHONY: run run-worker build test lint swag migrate-up migrate-down \
        mock docker-up docker-down k6

run:
	go run ./cmd/server

run-worker:
	go run ./cmd/worker

build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

test-unit:
	go test ./internal/... -v -count=1

test-integration:
	go test ./test/integration/... -v -timeout 120s

test-all: test-unit test-integration

coverage:
	go test ./internal/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

swag:
	swag init -g cmd/server/main.go -o docs/

mock:
	mockery --all --dir internal/repository --output internal/mocks

migrate-up:
	migrate -path db/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path db/migrations -database "$$DATABASE_URL" down 1

sqlc:
	sqlc generate

docker-up:
	docker compose up -d

docker-down:
	docker compose down

k6:
	k6 run test/load/ingest_test.js --env BASE_URL=http://localhost:8080
```

---

## Security Checklist

- [x] Password hashed bcrypt cost 12
- [x] JWT HS256 + secret dari env (min 32 byte)
- [x] Refresh token stored di DB + bisa direvoke
- [x] Token blacklist di Redis saat logout
- [x] Refresh token rotation (token lama langsung invalid)
- [x] Rate limiting: 2 layer (Cloudflare edge + Redis app-level)
- [x] Input validation di semua endpoint (go-playground/validator)
- [x] SQL injection prevention via parameterized query (sqlc/pgx)
- [x] API key tidak di-log (masked di zerolog)
- [x] Secret dari K8s Secret / env, tidak di-hardcode
- [x] Non-root user di Docker image
- [x] RabbitMQ pesan persistent (survive restart)
- [x] DLQ untuk pesan gagal (tidak hilang, bisa di-replay)
- [x] Cloudflare WAF: blokir request tanpa required headers
- [x] gosec linter untuk scan security issue di code
```
