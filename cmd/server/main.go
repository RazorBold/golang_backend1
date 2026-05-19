package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/config"
	"github.com/RazorBold/golang_backend1/internal/server"
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
		// production: JSON log tanpa console writer
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Str("env", cfg.App.Env).Str("port", cfg.App.Port).Msg("starting IoT Platform API")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("postgres ping failed")
	}
	log.Info().Msg("connected to postgres")

	redis, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redis.Close()
	log.Info().Str("addr", cfg.Redis.Addr).Msg("connected to redis")

	srv := server.New(cfg, db, redis)

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
