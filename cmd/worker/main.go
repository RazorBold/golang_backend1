package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RazorBold/golang_backend1/internal/broker"
	"github.com/RazorBold/golang_backend1/internal/config"
	pgRepo "github.com/RazorBold/golang_backend1/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	log.Info().Msg("starting IoT telemetry worker")

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

	// ── RabbitMQ Consumer ─────────────────────────────────────
	consumer, err := broker.NewConsumer(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	defer consumer.Close()
	log.Info().Msg("connected to rabbitmq")

	dataRepo := pgRepo.NewDeviceDataRepo(db)

	// ── Graceful shutdown ─────────────────────────────────────
	runCtx, runCancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Info().Msg("worker shutting down...")
		runCancel()
	}()

	log.Info().Str("queue", broker.QueuePersist).Msg("worker consuming")

	// telemetry persist handler
	err = consumer.Consume(runCtx, broker.QueuePersist, func(ctx context.Context, body []byte) error {
		var msg broker.TelemetryMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal telemetry message")
			return err
		}

		if err := dataRepo.Create(ctx, msg.DeviceID, msg.Data, msg.ReceivedAt); err != nil {
			log.Error().Err(err).Str("device_id", msg.DeviceID).Msg("failed to persist telemetry")
			return err
		}

		log.Info().Str("device_id", msg.DeviceID).Str("event_id", msg.EventID).Msg("telemetry persisted")
		return nil
	})

	if err != nil && err != context.Canceled {
		log.Error().Err(err).Msg("consumer exited with error")
	}

	log.Info().Msg("worker exited")
}
