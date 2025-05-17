package app

import (
	grpcapp "Elschool-API/internal/app/grpc"
	"Elschool-API/internal/config"
	"Elschool-API/internal/infra/auth"
	cache "Elschool-API/internal/infra/cache/redis"
	"Elschool-API/internal/infra/fetcher"
	"Elschool-API/internal/infra/metrics"
	"Elschool-API/internal/infra/storage/postgres"
	"Elschool-API/internal/infra/storage/transaction"
	"Elschool-API/internal/service/marks"
	"Elschool-API/internal/service/student"
	"Elschool-API/internal/service/user"
	"database/sql"
	"github.com/go-redis/redis/v8"
	"log/slog"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func New(log *slog.Logger, db *sql.DB, rclient *redis.Client, metricsInfra *metrics.Metrics, infraCfg config.InfraConfig, grpcPort int) *App {
	storageInfra := postgres.New(db)
	txManager := transaction.NewTransactionManager(db)
	cacheInfra := cache.New(rclient)
	authInfra := auth.New(infraCfg.Url)
	fetcherInfra := fetcher.New(infraCfg.Url)

	userService := user.New(log, storageInfra, metricsInfra)
	studentService := student.New(log, storageInfra, storageInfra, authInfra, txManager, metricsInfra)
	marksService := marks.New(log, storageInfra, cacheInfra, cacheInfra, authInfra, fetcherInfra, metricsInfra)

	grpcApp := grpcapp.New(log, userService, studentService, marksService, grpcPort)

	return &App{GRPCsrv: grpcApp}
}
