package middleware

import (
	"fmt"
	"time"

	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/pkg/response"
	"github.com/RazorBold/golang_backend1/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// APIKeyAuth validates X-API-Key header.
// Cache Redis (5 menit) → DB fallback.
// Menyimpan app_id di fiber.Locals("app_id").
func APIKeyAuth(redis *cache.RedisClient, appRepo repository.ApplicationRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return response.Unauthorized(c, "missing X-API-Key header")
		}

		cacheKey := fmt.Sprintf("apikey:%s", apiKey)

		// Redis cache hit
		appID, err := redis.Get(c.Context(), cacheKey)
		if err == nil && appID != "" {
			c.Locals("app_id", appID)
			return c.Next()
		}

		// Cache miss → query DB
		app, err := appRepo.FindByAPIKey(c.Context(), apiKey)
		if err != nil {
			return response.InternalError(c)
		}
		if app == nil {
			return response.Unauthorized(c, "invalid API key")
		}

		// Simpan ke Redis 5 menit
		_ = redis.Set(c.Context(), cacheKey, app.ID, 5*time.Minute)

		c.Locals("app_id", app.ID)
		return c.Next()
	}
}
