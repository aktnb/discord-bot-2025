package persistence

import (
	"context"
	"fmt"

	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresTx struct {
	tx pgx.Tx
}

func (t *postgresTx) Exec(ctx context.Context, sql string, arguments ...any) (db.CommandTag, error) {
	return t.tx.Exec(ctx, sql, arguments...)
}

func (t *postgresTx) Query(ctx context.Context, sql string, args ...any) (db.Rows, error) {
	return t.tx.Query(ctx, sql, args...)
}

func (t *postgresTx) QueryRow(ctx context.Context, sql string, args ...any) db.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}

type postgresTxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) db.TxManager {
	return &postgresTxManager{
		pool: pool,
	}
}

func (m *postgresTxManager) WithTx(ctx context.Context, fn func(ctx context.Context, tx db.Tx) error) error {
	pgxTx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	tx := &postgresTx{tx: pgxTx}

	if err := fn(ctx, tx); err != nil {
		if rbErr := pgxTx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction (original error: %w): %w", err, rbErr)
		}
		return err
	}

	if err := pgxTx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m *postgresTxManager) WithKeyLock(ctx context.Context, key db.LockKey, fn func(ctx context.Context, tx db.Tx) error) error {
	return m.WithTx(ctx, func(ctx context.Context, tx db.Tx) error {
		lockSQL := "SELECT pg_advisory_xact_lock(hashtext($1))"
		if _, err := tx.Exec(ctx, lockSQL, string(key)); err != nil {
			return fmt.Errorf("failed to acquire advisory lock: %w", err)
		}

		return fn(ctx, tx)
	})
}
