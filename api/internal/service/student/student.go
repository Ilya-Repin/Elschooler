package student

import (
	"Elschool-API/internal/infra/metrics"
	"Elschool-API/internal/infra/storage"
	"Elschool-API/internal/service"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type Transaction interface {
	Commit() error
	Rollback() error
}

type TransactionManager interface {
	StartTransaction(context.Context) (Transaction, error)
}

type StudentStorage interface {
	CreateStudent(ctx context.Context, studentToken, login, password string) (err error)
	DeleteStudent(ctx context.Context, studentToken string) (err error)
	HasStudentWithCredentials(ctx context.Context, login, password string) (studentToken string, err error)
}

type UserStudentsStorage interface {
	AddRelation(ctx context.Context, userToken, studentToken string) (err error)
	DeleteRelation(ctx context.Context, userToken, studentToken string) (err error)
	GetStudentRelations(ctx context.Context, studentToken string) (users []string, err error)
}

type StudentAuthChecker interface {
	AuthStudent(ctx context.Context, login, password string) (jwt string, err error)
}

type StudentService struct {
	log             *slog.Logger
	metrics         *metrics.Metrics
	studStorage     StudentStorage
	usrStudStorage  UserStudentsStorage
	studAuthChecker StudentAuthChecker
	txManager       TransactionManager
}

func New(log *slog.Logger, studStorage StudentStorage, usrStudStorage UserStudentsStorage, studAuthChecker StudentAuthChecker, txManager TransactionManager, metricsInfra *metrics.Metrics) *StudentService {
	return &StudentService{log: log, studStorage: studStorage, usrStudStorage: usrStudStorage, studAuthChecker: studAuthChecker, txManager: txManager, metrics: metricsInfra}
}

func (s *StudentService) AddStudent(ctx context.Context, userID, login, password string) (studID string, err error) {
	student, err := s.addStudent(ctx, userID, login, password, nil)
	if err != nil {
		s.metrics.StudentActions.WithLabelValues(metrics.ActionAdd, metrics.StatusErr).Inc()
	} else {
		s.metrics.StudentActions.WithLabelValues(metrics.ActionAdd, metrics.StatusOk).Inc()
	}
	return student, err
}

func (s *StudentService) DeleteStudent(ctx context.Context, userID, studID string) (err error) {
	err = s.deleteStudent(ctx, userID, studID, nil)
	if err != nil {
		s.metrics.StudentActions.WithLabelValues(metrics.ActionDelete, metrics.StatusErr).Inc()
	} else {
		s.metrics.StudentActions.WithLabelValues(metrics.ActionDelete, metrics.StatusOk).Inc()
	}
	return err
}

func (s *StudentService) UpdateStudent(ctx context.Context, userID, studID, login, password string) (newStudID string, err error) {
	const op = "services.student.UpdateStudent"

	log := s.log.With(slog.String("op", op), slog.String("user", userID), slog.String("student", studID))
	log.Info("updating student")

	tx, err := s.txManager.StartTransaction(ctx)
	if err != nil {
		log.Error("failed to start transaction", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	ctx = context.WithValue(ctx, "tx", tx)

	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	relations, err := s.usrStudStorage.GetStudentRelations(ctx, studID)
	if err != nil {
		if errors.Is(err, storage.ErrStudentNotFound) {
			log.Error("no such student", "error", err)
			s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusOk)
			s.metrics.StudentActions.WithLabelValues(metrics.ActionUpdate, metrics.StatusErr).Inc()
			return "", fmt.Errorf("%s: %w", op, service.ErrStudentNotFound)
		}

		log.Error("failed to find student relations in storage", "error", err)
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusErr)
		s.metrics.StudentActions.WithLabelValues(metrics.ActionUpdate, metrics.StatusErr).Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}
	s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusOk)

	if !s.contains(relations, userID) {
		log.Error("no relations with that student", "error", err)
		s.metrics.StudentActions.WithLabelValues(metrics.ActionUpdate, metrics.StatusErr).Inc()
		return "", fmt.Errorf("%s: %w", op, service.ErrUserNotFound)
	}

	newStudID, err = s.addStudent(ctx, userID, login, password, tx)
	if err != nil {
		log.Error("failed to add student in storage", "error", err)
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusErr)
		s.metrics.StudentActions.WithLabelValues(metrics.ActionUpdate, metrics.StatusErr).Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("student added in storage during update")
	s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusOk)

	if newStudID != studID {
		err = s.deleteStudent(ctx, userID, studID, tx)
		if err != nil {
			log.Error("failed to delete old student in storage", "error", err)
			s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionDelete, metrics.StatusOk)
			s.metrics.StudentActions.WithLabelValues(metrics.ActionUpdate, metrics.StatusErr).Inc()
			return "", fmt.Errorf("%s: %w", op, err)
		}
		log.Info("relation to old student deleted")
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionDelete, metrics.StatusOk)
	}
	log.Info("student updated in storage")
	s.metrics.StudentActions.WithLabelValues(metrics.ActionUpdate, metrics.StatusOk).Inc()

	return newStudID, nil
}

func (s *StudentService) addStudent(ctx context.Context, userID, login, password string, tx Transaction) (studID string, err error) {
	const op = "services.student.addStudent"

	log := s.log.With(slog.String("op", op), slog.String("user", userID))
	log.Info("adding student")

	start := time.Now()
	_, err = s.studAuthChecker.AuthStudent(ctx, login, password)
	s.metrics.ElschoolAuthDuration.WithLabelValues(metrics.MethodCheck).Observe(time.Since(start).Seconds())

	if err != nil {
		log.Error("failed to check student credential", "error", err)
		s.metrics.ElschoolAuthTotal.WithLabelValues(metrics.MethodAuth, metrics.StatusErr).Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}
	s.metrics.ElschoolAuthTotal.WithLabelValues(metrics.MethodAuth, metrics.StatusOk).Inc()

	if tx == nil {
		if tx, err = s.txManager.StartTransaction(ctx); err != nil {
			log.Error("failed to start transaction", "error", err)
			return "", fmt.Errorf("%s: %w", op, err)
		}

		ctx = context.WithValue(ctx, "tx", tx)

		defer func() {
			if err == nil {
				tx.Commit()
			} else {
				tx.Rollback()
			}
		}()
	}

	studID, err = s.studStorage.HasStudentWithCredentials(ctx, login, password)

	if errors.Is(err, storage.ErrStudentNotFound) {
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusOk)
		uuid4, genErr := uuid.NewRandom()

		if genErr != nil {
			err = genErr
			log.Error("failed to gen uuid", "error", err)
			return "", fmt.Errorf("%s: %w", op, err)
		}
		studID = uuid4.String()

		err = s.studStorage.CreateStudent(ctx, studID, login, password)
		if err != nil {
			if errors.Is(err, storage.ErrStudentExists) {
				s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusOk).Inc()
				log.Error("student already exists", "error", err)
			} else {
				s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusErr).Inc()
				log.Error("failed to save student", "error", err)
			}
			return "", fmt.Errorf("%s: %w", op, err)
		}

		log.Info("new student added")

		err = s.usrStudStorage.AddRelation(ctx, userID, studID)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Error("no such user", "error", err)
				s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusOk).Inc()
				return "", fmt.Errorf("%s: %w", op, service.ErrUserNotFound)
			}

			log.Error("failed to add user-student relation", "error", err)
			s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusErr).Inc()
			return "", fmt.Errorf("%s: %w", op, err)
		}

		log.Info("new user-student relation added")
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusOk).Inc()
		return studID, nil
	} else if err != nil {
		log.Error("failed to find student in storage", "error", err)
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusErr).Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}
	s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusOk).Inc()

	log.Info("student already exists")

	relations, err := s.usrStudStorage.GetStudentRelations(ctx, studID)
	if err != nil {
		log.Error("failed to find student relations in storage", "error", err)
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusErr).Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}
	s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusOk).Inc()

	if !s.contains(relations, userID) {
		log.Info("no relations with that student")

		err = s.usrStudStorage.AddRelation(ctx, userID, studID)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Error("no such user", "error", err)
				s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusOk).Inc()
				return "", fmt.Errorf("%s: %w", op, service.ErrUserNotFound)
			}

			log.Error("failed to add relation", "error", err)
			s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusErr).Inc()
			return "", fmt.Errorf("%s: %w", op, err)
		}

		log.Info("new user-student relation added")
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionWrite, metrics.StatusOk).Inc()
	}

	return studID, nil
}

func (s *StudentService) deleteStudent(ctx context.Context, userID, studID string, tx Transaction) (err error) {
	const op = "services.student.deleteStudent"

	log := s.log.With(slog.String("op", op), slog.String("user", userID), slog.String("student", studID))
	log.Info("deleting student")

	if tx == nil {
		if tx, err = s.txManager.StartTransaction(ctx); err != nil {
			log.Error("failed to start transaction", "error", err)
			return fmt.Errorf("%s: %w", op, err)
		}

		ctx = context.WithValue(ctx, "tx", tx)

		defer func() {
			if err == nil {
				tx.Commit()
			} else {
				tx.Rollback()
			}
		}()
	}

	relations, err := s.usrStudStorage.GetStudentRelations(ctx, studID)

	if err != nil {
		if errors.Is(err, storage.ErrStudentNotFound) {
			log.Error("no such student", "error", err)
			s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusOk).Inc()
			return fmt.Errorf("%s: %w", op, service.ErrStudentNotFound)
		}

		log.Error("failed to find student relations in storage", "error", err)
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionRead, metrics.StatusErr).Inc()
		return fmt.Errorf("%s: %w", op, err)
	}

	if !s.contains(relations, userID) {
		log.Error("no relations with that student", "error", err)
		return fmt.Errorf("%s: %w", op, service.ErrUserNotFound)
	}

	if len(relations) > 1 {
		err = s.usrStudStorage.DeleteRelation(ctx, userID, studID)

		if err != nil {
			if errors.Is(err, storage.ErrRelationNotFound) {
				log.Error("no such student", "error", err)
				s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionDelete, metrics.StatusOk).Inc()
				return fmt.Errorf("%s: %w", op, service.ErrStudentNotFound)
			}

			log.Error("failed to delete relation in storage", "error", err)
			s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionDelete, metrics.StatusErr).Inc()
			return fmt.Errorf("%s: %w", op, err)
		}
	} else {
		err = s.studStorage.DeleteStudent(ctx, studID)

		if err != nil {
			log.Error("failed to delete student in storage", "error", err)
			s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionDelete, metrics.StatusErr).Inc()
			return fmt.Errorf("%s: %w", op, err)
		}
		s.metrics.StorageRequestsTotal.WithLabelValues(metrics.ServiceStudent, metrics.ActionDelete, metrics.StatusOk).Inc()
	}

	log.Info("student deleted")

	return nil
}

func (s *StudentService) contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
