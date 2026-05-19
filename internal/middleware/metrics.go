package middleware

import (
	"strconv"
	"time"

	"github.com/RazorBold/golang_backend1/internal/metrics"
	"github.com/gofiber/fiber/v2"
)

// Metrics records HTTP request duration and count per route.
func Metrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()

		method := c.Method()
		// Use route path template (e.g. /api/v1/devices/:id) to avoid high cardinality.
		path := c.Route().Path
		status := strconv.Itoa(c.Response().StatusCode())

		metrics.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()

		return err
	}
}
