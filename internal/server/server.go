package server

import (
	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/config"
	"github.com/RazorBold/golang_backend1/internal/handler"
	"github.com/RazorBold/golang_backend1/internal/middleware"
	"github.com/RazorBold/golang_backend1/internal/repository"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handlers struct {
	Health    *handler.HealthHandler
	Auth      *handler.AuthHandler
	App       *handler.ApplicationHandler
	Device    *handler.DeviceHandler
	Telemetry *handler.TelemetryHandler
}

type Server struct {
	app *fiber.App
	cfg *config.Config
}

func New(cfg *config.Config, handlers Handlers, redis *cache.RedisClient, appRepo repository.ApplicationRepository) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "IoT Platform API v1.0",
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(middleware.Logger())
	app.Use(middleware.Metrics())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key",
	}))

	jwtMW := middleware.JWTAuth(cfg.JWT.Secret, redis)
	apiKeyMW := middleware.APIKeyAuth(redis, appRepo)

	// ── Probes + observability (no auth) ──────────────────────
	app.Get("/health", handlers.Health.Live)
	app.Get("/ready", handlers.Health.Ready)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// ── Auth routes (no auth) ─────────────────────────────────
	auth := app.Group("/api/v1/auth")
	auth.Post("/register", handlers.Auth.Register)
	auth.Post("/login", handlers.Auth.Login)
	auth.Post("/refresh", handlers.Auth.Refresh)
	auth.Post("/logout", jwtMW, handlers.Auth.Logout)

	// ── v1 group tanpa group-level middleware ─────────────────
	// PENTING: jangan pass middleware ke Group() karena Fiber akan
	// memanggil app.Use(prefix, mw) yang berlaku untuk SEMUA path
	// berawalan /api/v1 — termasuk ingest yang pakai API key.
	v1 := app.Group("/api/v1")

	// Applications — JWT per route
	v1.Post("/applications", jwtMW, handlers.App.Create)
	v1.Get("/applications", jwtMW, handlers.App.List)
	v1.Get("/applications/:id", jwtMW, handlers.App.Get)
	v1.Put("/applications/:id", jwtMW, handlers.App.Update)
	v1.Delete("/applications/:id", jwtMW, handlers.App.Delete)
	v1.Post("/applications/:id/regenerate-key", jwtMW, handlers.App.RegenerateKey)

	// Devices — JWT per route
	v1.Post("/applications/:app_id/devices", jwtMW, handlers.Device.Create)
	v1.Get("/applications/:app_id/devices", jwtMW, handlers.Device.List)
	v1.Get("/applications/:app_id/devices/:id", jwtMW, handlers.Device.Get)
	v1.Put("/applications/:app_id/devices/:id", jwtMW, handlers.Device.Update)
	v1.Delete("/applications/:app_id/devices/:id", jwtMW, handlers.Device.Delete)

	// Telemetry query — JWT
	v1.Get("/devices/:device_id/data", jwtMW, handlers.Telemetry.GetData)
	v1.Get("/devices/:device_id/latest", jwtMW, handlers.Telemetry.GetLatest)

	// Telemetry ingest — API Key (BUKAN JWT, pisah dari group di atas)
	v1.Post("/ingest/:device_id", apiKeyMW, handlers.Telemetry.Ingest)

	return &Server{app: app, cfg: cfg}
}

func (s *Server) Start() error {
	return s.app.Listen(":" + s.cfg.App.Port)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"status":  "error",
		"message": err.Error(),
	})
}
