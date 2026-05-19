package middleware

import (
	"fmt"
	"strings"

	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/pkg/response"
	"github.com/RazorBold/golang_backend1/internal/pkg/token"
	"github.com/gofiber/fiber/v2"
)

// JWTAuth validates Bearer token dan cek blacklist Redis.
// Claims disimpan di fiber.Locals: user_id, role, jti, expires_at.
func JWTAuth(jwtSecret string, redis *cache.RedisClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return response.Unauthorized(c, "missing or invalid authorization header")
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := token.Validate(tokenStr, jwtSecret)
		if err != nil {
			return response.Unauthorized(c, "invalid or expired token")
		}

		// cek apakah token sudah di-blacklist (logout)
		blacklisted, _ := redis.Exists(c.Context(), fmt.Sprintf("blacklist:%s", claims.JTI))
		if blacklisted {
			return response.Unauthorized(c, "token has been revoked")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("jti", claims.JTI)
		c.Locals("expires_at", claims.ExpiresAt.Time)

		return c.Next()
	}
}
