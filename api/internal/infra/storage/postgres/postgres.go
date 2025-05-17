package postgres

import (
	"Elschool-API/internal/config"
	"Elschool-API/internal/domain/models"
	"Elschool-API/internal/infra/storage"
	"Elschool-API/internal/infra/storage/transaction"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func New(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) CreateUser(ctx context.Context, userID, service string) (err error) {
	const op = "infra.storage.postgres.CreateUser"

	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO users (id, service) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, userID, service)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *PostgresStorage) CreateStudent(ctx context.Context, studID, login, password string) (err error) {
	const op = "infra.storage.postgres.CreateStudent"

	txRef, err := s.getTransaction(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	tx := txRef.Tx

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO students (id, login, password) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, studID, login, password)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *PostgresStorage) DeleteStudent(ctx context.Context, studID string) error {
	const op = "infra.storage.postgres.DeleteStudent"

	txRef, err := s.getTransaction(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	tx := txRef.Tx

	stmt, err := tx.PrepareContext(ctx, "DELETE FROM students WHERE id = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, studID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrStudentNotFound)
	}

	return nil
}

func (s *PostgresStorage) HasStudentWithCredentials(ctx context.Context, login, password string) (userID string, err error) {
	const op = "infra.storage.postgres.HasStudentWithCredentials"

	txRef, err := s.getTransaction(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	tx := txRef.Tx

	stmt, err := tx.PrepareContext(ctx, "SELECT id FROM students WHERE login = $1 AND password = $2")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, login, password)

	err = row.Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrStudentNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *PostgresStorage) ReadStudent(ctx context.Context, studID string) (student models.Student, err error) {
	const op = "infra.storage.postgres.ReadStudent"

	stmt, err := s.db.PrepareContext(ctx, "SELECT id, login, password FROM students WHERE id = $1")
	if err != nil {
		return models.Student{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, studID)

	err = row.Scan(&student.Token, &student.Login, &student.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Student{}, fmt.Errorf("%s: %w", op, storage.ErrStudentNotFound)
		}

		return models.Student{}, fmt.Errorf("%s: %w", op, err)
	}

	return student, nil
}

func (s *PostgresStorage) AddRelation(ctx context.Context, userID, studID string) (err error) {
	const op = "infra.storage.postgres.AddRelation"

	txRef, err := s.getTransaction(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	tx := txRef.Tx

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO user_students (user_id, student_id) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, userID, studID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *PostgresStorage) DeleteRelation(ctx context.Context, userID, studID string) error {
	const op = "infra.storage.postgres.DeleteRelation"

	txRef, err := s.getTransaction(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	tx := txRef.Tx

	stmt, err := tx.PrepareContext(ctx, "DELETE FROM user_students WHERE user_id = $1 AND student_id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, userID, studID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrRelationNotFound)
	}

	return nil
}

func (s *PostgresStorage) GetStudentRelations(ctx context.Context, studID string) (users []string, err error) {
	const op = "infra.storage.postgres.GetStudentRelations"

	txRef, err := s.getTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	tx := txRef.Tx

	stmt, err := tx.PrepareContext(ctx, "SELECT 1 FROM students WHERE id = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var exists int
	row := stmt.QueryRowContext(ctx, studID)

	err = row.Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrStudentNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = tx.PrepareContext(ctx, "SELECT user_id FROM user_students WHERE student_id = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, studID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var userId string
		if err := rows.Scan(&userId); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users = append(users, userId)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func (s *PostgresStorage) CheckRelation(ctx context.Context, userID, studID string) (err error) {
	const op = "infra.storage.postgres.CheckRelation"

	stmt, err := s.db.PrepareContext(ctx, "SELECT 1 FROM user_students WHERE user_id = $1 AND student_id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var exists int
	row := stmt.QueryRowContext(ctx, userID, studID)

	err = row.Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			stmt, err = s.db.PrepareContext(ctx, "SELECT 1 FROM users WHERE id = $1")
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			defer stmt.Close()

			var exists int
			row := stmt.QueryRowContext(ctx, userID)

			err = row.Scan(&exists)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
				}
				return fmt.Errorf("%s: %w", op, err)
			}

			return fmt.Errorf("%s: %w", op, storage.ErrStudentNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *PostgresStorage) getTransaction(ctx context.Context) (tx *transaction.DbTransaction, err error) {
	const op = "infra.storage.postgres.getTransaction"

	var ok bool
	if tx, ok = ctx.Value("tx").(*transaction.DbTransaction); !ok {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrFailedToGetTX)
	}
	return tx, nil
}

func InitDB(cfg *config.StorageConfig) (db *sql.DB, err error) {
	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Dbname,
		cfg.Sslmode,
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
