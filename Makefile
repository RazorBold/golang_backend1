.PHONY: run build test lint swag mock tidy \
        migrate-up migrate-down migrate-status \
        docker-up docker-down docker-logs \
        k6

# ── Development ─────────────────────────────────────────────
run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

tidy:
	go mod tidy

# ── Testing ──────────────────────────────────────────────────
test:
	go test ./... -v -count=1

test-unit:
	go test ./internal/... -v -count=1

test-integration:
	go test ./test/integration/... -v -timeout 120s

coverage:
	go test ./internal/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# ── Linting ──────────────────────────────────────────────────
lint:
	golangci-lint run ./...

# ── Code Generation ──────────────────────────────────────────
swag:
	swag init -g cmd/server/main.go -o docs/

mock:
	mockery --all --dir internal/repository --output internal/mocks

sqlc:
	sqlc generate

# ── Database Migration ───────────────────────────────────────
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

migrate-status:
	migrate -path db/migrations -database "$(DATABASE_URL)" version

# ── Docker ───────────────────────────────────────────────────
docker-up:
	docker compose up -d postgres redis rabbitmq

docker-up-all:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api

# ── Load Test ────────────────────────────────────────────────
k6:
	k6 run test/load/ingest_test.js --env BASE_URL=http://localhost:8080
