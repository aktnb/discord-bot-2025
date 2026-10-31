package persistence

import (
	"context"
	"errors"
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
	pool interface {
		Begin(context.Context) (pgx.Tx, error)
	}
}

func NewTxManager(pool *pgxpool.Pool) db.TxManager {
	return &postgresTxManager{
		pool: pool,
	}
}

func (m *postgresTxManager) WithTx(ctx context.Context, fn func(ctx context.Context, tx db.Tx) error) (resultErr error) {
	pgxTx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Commit 後の Rollback は ErrTxClosed になるため無視する。panic 時も解放する。
	defer func() {
		if err := pgxTx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			resultErr = errors.Join(resultErr, fmt.Errorf("failed to rollback transaction: %w", err))
		}
	}()

	tx := &postgresTx{tx: pgxTx}

	if err := fn(ctx, tx); err != nil {
		return err
	}

	if err := pgxTx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
