package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB  *pgxpool.Pool
	testDSN string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// ── PostgreSQL container ───────────────────────────────────
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("iot_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer pgContainer.Terminate(ctx)

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("failed to get postgres DSN: %v\n", err)
		os.Exit(1)
	}
	testDSN = dsn

	// ── Redis container ───────────────────────────────────────
	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
	)
	if err != nil {
		fmt.Printf("failed to start redis container: %v\n", err)
		os.Exit(1)
	}
	defer redisContainer.Terminate(ctx)

	// ── DB connection + migrations ────────────────────────────
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Printf("failed to connect to postgres: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := runMigrations(ctx, pool); err != nil {
		fmt.Printf("failed to run migrations: %v\n", err)
		os.Exit(1)
	}
	testDB = pool

	os.Exit(m.Run())
}

// runMigrations jalankan DDL langsung tanpa memerlukan migrate CLI
func runMigrations(ctx context.Context, db *pgxpool.Pool) error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS users (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name       VARCHAR(100) NOT NULL,
		email      VARCHAR(255) UNIQUE NOT NULL,
		password   TEXT NOT NULL,
		role       VARCHAR(20) NOT NULL DEFAULT 'user',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS applications (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name        VARCHAR(100) NOT NULL,
		description TEXT,
		api_key     VARCHAR(64) UNIQUE NOT NULL,
		created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS devices (
		id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
		name           VARCHAR(100) NOT NULL,
		type           VARCHAR(50),
		status         VARCHAR(20) NOT NULL DEFAULT 'inactive',
		last_seen_at   TIMESTAMPTZ,
		metadata       JSONB,
		created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS device_data (
		id          BIGSERIAL PRIMARY KEY,
		device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
		payload     JSONB NOT NULL,
		received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token      TEXT UNIQUE NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	_, err := db.Exec(ctx, schema)
	return err
}
