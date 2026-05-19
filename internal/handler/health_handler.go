package handler

import (
	"context"
	"time"

	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	redis *cache.RedisClient
}

func NewHealthHandler(db *pgxpool.Pool, redis *cache.RedisClient) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// Live — liveness probe: server nyala atau tidak
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{"status": "alive"})
}

// Ready — readiness probe: cek koneksi DB dan Redis
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	checks := fiber.Map{}
	allOk := true

	if err := h.db.Ping(ctx); err != nil {
		checks["postgres"] = "unreachable"
		allOk = false
	} else {
		checks["postgres"] = "ok"
	}

	if err := h.redis.Ping(ctx); err != nil {
		checks["redis"] = "unreachable"
		allOk = false
	} else {
		checks["redis"] = "ok"
	}

	if !allOk {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not ready",
			"checks": checks,
		})
	}

	return response.OK(c, fiber.Map{"status": "ready", "checks": checks})
}
