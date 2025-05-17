package user

import (
	"Elschool-API/internal/infra/metrics"
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
)

type UserService struct {
	log        *slog.Logger
	metrics    *metrics.Metrics
	usrStorage UserStorage
}

type UserStorage interface {
	CreateUser(ctx context.Context, token, service string) (err error)
}

func New(log *slog.Logger, usrStorage UserStorage, metricsInfra *metrics.Metrics) *UserService {
	return &UserService{log: log, usrStorage: usrStorage, metrics: metricsInfra}
}

func (u *UserService) CreateUser(ctx context.Context, service string) (token string, err error) {
	const op = "service.user.CreateUser"

	log := u.log.With(slog.String("op", op))
	log.Info("creating user")

	uuid4, err := uuid.NewRandom()

	if err != nil {
		log.Error("failed to gen uuid", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token = uuid4.String()
	err = u.usrStorage.CreateUser(ctx, token, service)

	if err != nil {
		log.Error("failed to save user", "error", err)
		u.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceUser, metrics.ActionWrite, metrics.StatusErr).Inc()
		u.metrics.UserRegistrations.WithLabelValues(service, metrics.StatusErr).Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}
	u.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceUser, metrics.ActionWrite, metrics.StatusOk).Inc()
	u.metrics.UserRegistrations.WithLabelValues(service, metrics.StatusOk).Inc()
	log.Info("user created")

	return token, nil
}
