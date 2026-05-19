package server

import (
	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/config"
	"github.com/RazorBold/golang_backend1/internal/handler"
	"github.com/RazorBold/golang_backend1/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	app *fiber.App
	cfg *config.Config
}

func New(cfg *config.Config, db *pgxpool.Pool, redis *cache.RedisClient) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "IoT Platform API v1.0",
		ErrorHandler: customErrorHandler,
	})

	// global middleware
	app.Use(recover.New())
	app.Use(middleware.Logger())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key",
	}))

	// health & readiness probes
	healthHandler := handler.NewHealthHandler(db, redis)
	app.Get("/health", healthHandler.Live)
	app.Get("/ready", healthHandler.Ready)

	// /api/v1 group — handler lain ditambahkan di phase berikutnya
	// v1 := app.Group("/api/v1")

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
