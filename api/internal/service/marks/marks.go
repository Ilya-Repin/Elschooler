package marks

import (
	"Elschool-API/internal/domain/models"
	"Elschool-API/internal/infra/metrics"
	"Elschool-API/internal/infra/storage"
	"Elschool-API/internal/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type StudentStorage interface {
	ReadStudent(ctx context.Context, studentToken string) (student models.Student, err error)
	CheckRelation(ctx context.Context, userToken, studentToken string) (err error)
}

type StudentAuth interface {
	AuthStudent(ctx context.Context, login, password string) (jwt string, err error)
	CheckToken(ctx context.Context, jwt string) (status bool, err error)
}

type TokenCache interface {
	FindToken(ctx context.Context, studentToken string) (jwt string, err error)
	AddToken(ctx context.Context, studentToken, jwt string) (err error)
	DeleteToken(ctx context.Context, studentToken string) (err error)
}

type Fetcher interface {
	FetchDayMarks(ctx context.Context, jwt, date string) (marks models.DayMarks, err error)
	FetchAverageMarks(ctx context.Context, jwt string, period int32) (marks models.AverageMarks, err error)
	FetchFinalMarks(ctx context.Context, jwt string) (marks models.FinalMarks, err error)
}

type MarksCache interface {
	SaveDayMarks(ctx context.Context, studentToken string, marks models.DayMarks) (err error)
	SaveAverageMarks(ctx context.Context, studentToken string, marks models.AverageMarks) (err error)
	SaveFinalMarks(ctx context.Context, studentToken string, marks models.FinalMarks) (err error)
	GetDayMarks(ctx context.Context, studentToken, date string) (marks models.DayMarks, err error)
	GetAverageMarks(ctx context.Context, studentToken string, period int32) (marks models.AverageMarks, err error)
	GetFinalMarks(ctx context.Context, studentToken string) (marks models.FinalMarks, err error)
}

type MarksService struct {
	log         *slog.Logger
	metrics     *metrics.Metrics
	studStorage StudentStorage
	tokenCache  TokenCache
	marksCache  MarksCache
	studAuth    StudentAuth
	fetcher     Fetcher
}

func New(log *slog.Logger, studStorage StudentStorage, tokenCache TokenCache, marksCache MarksCache, studAuth StudentAuth, fetcher Fetcher, metricsInfra *metrics.Metrics) *MarksService {
	return &MarksService{log: log, studStorage: studStorage, tokenCache: tokenCache, marksCache: marksCache, studAuth: studAuth, fetcher: fetcher, metrics: metricsInfra}
}

type Marks interface {
	GetDayMarks(ctx context.Context, userToken, studentToken, date string) (marks models.DayMarks, err error)
	GetAverageMarks(ctx context.Context, userToken, studentToken string, period int32) (marks models.AverageMarks, err error)
	GetFinalMarks(ctx context.Context, userToken, studentToken string) (marks models.FinalMarks, err error)
}

func (m *MarksService) GetDayMarks(ctx context.Context, userID, studID, date string) (marks models.DayMarks, err error) {
	const op = "services.marks.GetDayMarks"

	log := m.log.With(slog.String("op", op), slog.String("user", userID), slog.String("student", studID))
	log.Info("getting day marks")

	err = m.studStorage.CheckRelation(ctx, userID, studID)
	if err != nil {
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeDay, metrics.StatusErr).Inc()
		if !(errors.Is(err, storage.ErrUserNotFound) && errors.Is(err, storage.ErrStudentNotFound)) {
			log.Error("failed to check relation", "error", err)
			m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusErr).Inc()
			return models.DayMarks{}, fmt.Errorf("%s: %w", op, err)
		}

		m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusOk).Inc()
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("no such user", "error", err)
			return models.DayMarks{}, fmt.Errorf("%s: %w", op, service.ErrUserNotFound)
		}
		if errors.Is(err, storage.ErrStudentNotFound) {
			log.Error("no such student", "error", err)
			return models.DayMarks{}, fmt.Errorf("%s: %w", op, service.ErrStudentNotFound)
		}
	}
	m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusOk).Inc()

	marks, err = m.marksCache.GetDayMarks(ctx, studID, date)
	if err == nil {
		log.Info("day marks found in cache")
		m.metrics.MarksCacheRateTotal.WithLabelValues(metrics.TypeDay, metrics.StatusHit).Inc()
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeDay, metrics.StatusOk).Inc()
		return marks, nil
	}
	m.metrics.MarksCacheRateTotal.WithLabelValues(metrics.TypeDay, metrics.StatusMiss).Inc()

	log.Info("failed to get day marks from cache", "error", err)

	jwt, err := m.getToken(ctx, studID)

	if err != nil {
		log.Error("failed to get jwt", "error", err)
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeDay, metrics.StatusErr).Inc()
		return models.DayMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	start := time.Now()
	defer func() {
		m.metrics.ElschoolFetchDuration.WithLabelValues(metrics.TypeDay).
			Observe(time.Since(start).Seconds())
	}()

	marks, err = m.fetcher.FetchDayMarks(ctx, jwt, date)
	if err != nil {
		log.Error("failed to fetch day marks", "error", err)
		m.metrics.ElschoolFetchTotal.WithLabelValues(metrics.TypeDay, metrics.StatusErr).Inc()
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeDay, metrics.StatusErr).Inc()
		return models.DayMarks{}, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("day marks fetched")
	m.metrics.ElschoolFetchTotal.WithLabelValues(metrics.TypeDay, metrics.StatusOk).Inc()
	m.metrics.MarksRequests.WithLabelValues(metrics.TypeDay, metrics.StatusOk).Inc()

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		errCache := m.marksCache.SaveDayMarks(cacheCtx, studID, marks)
		if errCache != nil {
			log.Warn("failed to cache day marks", "error", errCache)
			m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionWrite, metrics.StatusErr).Inc()
		} else {
			m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionWrite, metrics.StatusOk).Inc()
		}
	}()

	return marks, nil
}

func (m *MarksService) GetAverageMarks(ctx context.Context, userID, studID string, period int32) (marks models.AverageMarks, err error) {
	const op = "services.marks.GetAverageMarks"

	log := m.log.With(slog.String("op", op), slog.String("user", userID), slog.String("student", studID))
	log.Info("getting average marks")

	err = m.studStorage.CheckRelation(ctx, userID, studID)
	if err != nil {
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeAverage, metrics.StatusErr).Inc()
		if !(errors.Is(err, storage.ErrUserNotFound) && errors.Is(err, storage.ErrStudentNotFound)) {
			log.Error("failed to check relation", "error", err)
			m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusErr).Inc()
			return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
		}

		m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusOk).Inc()
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("no such user", "error", err)
			return models.AverageMarks{}, fmt.Errorf("%s: %w", op, service.ErrUserNotFound)
		}
		if errors.Is(err, storage.ErrStudentNotFound) {
			log.Error("no such student", "error", err)
			return models.AverageMarks{}, fmt.Errorf("%s: %w", op, service.ErrStudentNotFound)
		}
	}
	m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusOk).Inc()

	marks, errCache := m.marksCache.GetAverageMarks(ctx, studID, period)
	if errCache == nil {
		log.Info("average marks found in cache")
		m.metrics.MarksCacheRateTotal.WithLabelValues(metrics.TypeAverage, metrics.StatusHit).Inc()
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeAverage, metrics.StatusOk).Inc()
		return marks, nil
	}
	m.metrics.MarksCacheRateTotal.WithLabelValues(metrics.TypeAverage, metrics.StatusMiss).Inc()

	log.Info("failed to get average marks from cache", "error", errCache)

	jwt, err := m.getToken(ctx, studID)

	if err != nil {
		log.Error("failed to get jwt", "error", err)
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeAverage, metrics.StatusErr).Inc()
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	start := time.Now()
	defer func() {
		m.metrics.ElschoolFetchDuration.WithLabelValues(metrics.TypeAverage).
			Observe(time.Since(start).Seconds())
	}()

	marks, err = m.fetcher.FetchAverageMarks(ctx, jwt, period)
	if err != nil {
		log.Error("failed to fetch average marks", "error", err)
		m.metrics.ElschoolFetchTotal.WithLabelValues(metrics.TypeAverage, metrics.StatusErr).Inc()
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeAverage, metrics.StatusErr).Inc()
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("average marks fetched")
	m.metrics.ElschoolFetchTotal.WithLabelValues(metrics.TypeAverage, metrics.StatusOk).Inc()
	m.metrics.MarksRequests.WithLabelValues(metrics.TypeAverage, metrics.StatusOk).Inc()

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		errCache = m.marksCache.SaveAverageMarks(cacheCtx, studID, marks)
		if errCache != nil {
			log.Warn("failed to cache average marks", "error", errCache)
			m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionWrite, metrics.StatusErr).Inc()
		} else {
			m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionWrite, metrics.StatusOk).Inc()
		}
	}()

	return marks, nil
}

func (m *MarksService) GetFinalMarks(ctx context.Context, userID, studID string) (marks models.FinalMarks, err error) {
	const op = "services.marks.GetFinalMarks"

	log := m.log.With(slog.String("op", op), slog.String("user", userID), slog.String("student", studID))
	log.Info("getting final marks")

	err = m.studStorage.CheckRelation(ctx, userID, studID)
	if err != nil {
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeFinal, metrics.StatusErr).Inc()
		if !(errors.Is(err, storage.ErrUserNotFound) && errors.Is(err, storage.ErrStudentNotFound)) {
			log.Error("failed to check relation", "error", err)
			m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusErr).Inc()
			return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
		}

		m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusOk).Inc()
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("no such user", "error", err)
			return models.FinalMarks{}, fmt.Errorf("%s: %w", op, service.ErrUserNotFound)
		}
		if errors.Is(err, storage.ErrStudentNotFound) {
			log.Error("no such student", "error", err)
			return models.FinalMarks{}, fmt.Errorf("%s: %w", op, service.ErrStudentNotFound)
		}
	}
	m.metrics.StorageRequestsTotal.WithLabelValues(metrics.TypeFinal, metrics.ActionRead, metrics.StatusOk).Inc()

	marks, err = m.marksCache.GetFinalMarks(ctx, studID)
	if err == nil {
		log.Info("final marks found in cache")
		m.metrics.MarksCacheRateTotal.WithLabelValues(metrics.TypeFinal, metrics.StatusHit).Inc()
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeFinal, metrics.StatusOk).Inc()
		return marks, nil
	}
	m.metrics.MarksCacheRateTotal.WithLabelValues(metrics.TypeFinal, metrics.StatusMiss).Inc()

	log.Info("failed to get final marks from cache", "error", err)

	jwt, err := m.getToken(ctx, studID)

	if err != nil {
		log.Error("failed to get jwt", "error", err)
		m.metrics.MarksRequests.WithLabelValues(metrics.TypeFinal, metrics.StatusOk).Inc()
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	start := time.Now()
	defer func() {
		m.metrics.ElschoolFetchDuration.WithLabelValues(metrics.TypeFinal).
			Observe(time.Since(start).Seconds())
	}()

	marks, err = m.fetcher.FetchFinalMarks(ctx, jwt)
	if err != nil {
		log.Error("failed to fetch final marks", "error", err)
		m.metrics.ElschoolFetchTotal.WithLabelValues(metrics.TypeFinal, metrics.StatusErr).Inc()
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("final marks fetched")
	m.metrics.ElschoolFetchTotal.WithLabelValues(metrics.TypeFinal, metrics.StatusOk).Inc()
	m.metrics.MarksRequests.WithLabelValues(metrics.TypeFinal, metrics.StatusOk).Inc()

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		errCache := m.marksCache.SaveFinalMarks(cacheCtx, studID, marks)
		if errCache != nil {
			log.Warn("failed to cache final marks", "error", errCache)
			m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionWrite, metrics.StatusErr).Inc()
		} else {
			m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionWrite, metrics.StatusOk).Inc()
		}
	}()

	return marks, nil
}

func (m *MarksService) getToken(ctx context.Context, studID string) (token string, err error) {
	const op = "services.marks.getToken"

	log := m.log.With(slog.String("op", op), slog.String("student", studID))
	log.Info("getting token")

	token, err = m.tokenCache.FindToken(ctx, studID)
	if err == nil {
		log.Info("token found in cache")
		m.metrics.TokenCacheRateTotal.WithLabelValues(metrics.StatusHit)

		start := time.Now()
		status, err := m.studAuth.CheckToken(ctx, token)
		m.metrics.ElschoolAuthDuration.WithLabelValues(metrics.MethodCheck).Observe(time.Since(start).Seconds())

		if err != nil {
			log.Warn("failed check of cached token", "error", err)
			m.metrics.ElschoolAuthTotal.WithLabelValues(metrics.MethodCheck, metrics.StatusErr).Inc()
		} else if status {
			m.metrics.ElschoolAuthTotal.WithLabelValues(metrics.MethodCheck, metrics.StatusOk).Inc()
			return token, nil
		} else {
			log.Warn("wrong cached token", "error", err)

			go func() {
				errCache := m.tokenCache.DeleteToken(ctx, studID)
				if errCache != nil {
					log.Warn("failed invalidating of wrong cached token", "error", errCache)
					m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionDelete, metrics.StatusErr).Inc()
				} else {
					m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionDelete, metrics.StatusOk).Inc()
				}
			}()
		}
	}
	m.metrics.TokenCacheRateTotal.WithLabelValues(metrics.StatusMiss)

	log.Info("failed token search in cache", "error", err)

	student, err := m.studStorage.ReadStudent(ctx, studID)
	if err != nil {
		if errors.Is(err, storage.ErrStudentNotFound) {
			log.Error("no such student", "error", service.ErrStudentNotFound)
			m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusOk).Inc()
			return "", fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to find student in storage", "error", err)
		m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusErr).Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}
	m.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionRead, metrics.StatusOk).Inc()

	start := time.Now()
	token, err = m.studAuth.AuthStudent(ctx, student.Login, student.Password)
	m.metrics.ElschoolAuthDuration.WithLabelValues(metrics.MethodAuth).Observe(time.Since(start).Seconds())

	if err != nil {
		log.Error("failed to auth student", "error", err)
		m.metrics.ElschoolAuthTotal.WithLabelValues(metrics.MethodAuth, metrics.StatusErr).Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}
	m.metrics.ElschoolAuthTotal.WithLabelValues(metrics.MethodAuth, metrics.StatusOk).Inc()

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = m.tokenCache.AddToken(cacheCtx, studID, token)
		if err != nil {
			log.Warn("failed to cache token", "error", err)
			m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionWrite, metrics.StatusErr).Inc()
		} else {
			m.metrics.CacheModifyTotal.WithLabelValues(metrics.ServiceMarks, metrics.ActionWrite, metrics.StatusOk).Inc()
		}
	}()

	return token, nil
}
