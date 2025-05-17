package main

import (
	"Elschool-API/internal/app"
	"Elschool-API/internal/config"
	"Elschool-API/internal/infra/cache/redis"
	"Elschool-API/internal/infra/metrics"
	"Elschool-API/internal/infra/storage/postgres"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting server", slog.String("env", cfg.Env))

	db, err := postgres.InitDB(&cfg.StorageConfig)
	if err != nil {
		panic(err)
	}

	rclient, err := redis.InitCache(&cfg.CacheConfig)
	if err != nil {
		panic(err)
	}

	metr, err := metrics.New(&cfg.MetricsConfig)
	if err != nil {
		panic(err)
	}

	application := app.New(log, db, rclient, metr, cfg.InfraConfig, cfg.GRPCConfig.Port)

	go func() {
		application.GRPCsrv.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("received shutdown signal", slog.String("signal", sign.String()))

	application.GRPCsrv.Stop()

	log.Info("closing db connection")
	if err := db.Close(); err != nil {
		log.Error("Error closing database connection:", err)
	}

	log.Info("closing cache connection")
	if err := rclient.Close(); err != nil {
		log.Error("Error closing cache connection:", err)
	}

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
