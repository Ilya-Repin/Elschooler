package transaction

import (
	"Elschool-API/internal/service/student"
	"context"
	"database/sql"
	"fmt"
)

type DbTransaction struct {
	Tx *sql.Tx
}

func (tx DbTransaction) Commit() (err error) {
	const op = "infra.storage.transaction.commit"
	if err = tx.Tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (tx DbTransaction) Rollback() (err error) {
	const op = "infra.storage.transaction.rollback"
	if err = tx.Tx.Rollback(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

type TransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

func (tm *TransactionManager) StartTransaction(ctx context.Context) (student.Transaction, error) {
	const op = "infra.storage.transaction.StartTransaction"

	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &DbTransaction{Tx: tx}, nil
}
