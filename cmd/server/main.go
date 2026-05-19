// @title           IoT Platform API
// @version         1.0
// @description     Production-grade IoT backend — device telemetry ingest, real-time data query, JWT auth.
// @termsOfService  http://swagger.io/terms/

// @contact.name   RazorBold
// @contact.email  support@razorbold.dev

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer <token>"

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RazorBold/golang_backend1/internal/broker"
	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/config"
	"github.com/RazorBold/golang_backend1/internal/handler"
	pgRepo "github.com/RazorBold/golang_backend1/internal/repository/postgres"
	"github.com/RazorBold/golang_backend1/internal/server"
	"github.com/RazorBold/golang_backend1/internal/service"
	_ "github.com/RazorBold/golang_backend1/docs" // swagger docs
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	if cfg.App.Env == "production" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Str("env", cfg.App.Env).Str("port", cfg.App.Port).Msg("starting IoT Platform API")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ── Database ─────────────────────────────────────────────
	db, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("postgres ping failed")
	}
	log.Info().Msg("connected to postgres")

	// ── Redis ─────────────────────────────────────────────────
	redis, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redis.Close()
	log.Info().Str("addr", cfg.Redis.Addr).Msg("connected to redis")

	// ── RabbitMQ Publisher ────────────────────────────────────
	publisher, err := broker.NewPublisher(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	defer publisher.Close()
	log.Info().Msg("connected to rabbitmq")

	// ── Repositories ──────────────────────────────────────────
	userRepo := pgRepo.NewUserRepo(db)
	tokenRepo := pgRepo.NewRefreshTokenRepo(db)
	appRepo := pgRepo.NewApplicationRepo(db)
	deviceRepo := pgRepo.NewDeviceRepo(db)
	dataRepo := pgRepo.NewDeviceDataRepo(db)

	// ── Services ──────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, tokenRepo, redis, cfg.JWT)
	appSvc := service.NewApplicationService(appRepo, redis)
	deviceSvc := service.NewDeviceService(appRepo, deviceRepo)
	telemetrySvc := service.NewTelemetryService(deviceRepo, dataRepo, appRepo, publisher, redis)

	// ── Handlers ──────────────────────────────────────────────
	handlers := server.Handlers{
		Health:    handler.NewHealthHandler(db, redis),
		Auth:      handler.NewAuthHandler(authSvc),
		App:       handler.NewApplicationHandler(appSvc),
		Device:    handler.NewDeviceHandler(deviceSvc),
		Telemetry: handler.NewTelemetryHandler(telemetrySvc),
	}

	// ── Server ────────────────────────────────────────────────
	srv := server.New(cfg, handlers, redis, appRepo)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Error().Err(err).Msg("server stopped")
		}
	}()

	log.Info().Str("port", cfg.App.Port).Msg("server is running")
	<-quit

	log.Info().Msg("shutting down gracefully...")
	if err := srv.Shutdown(); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}
	log.Info().Msg("server exited")
}
