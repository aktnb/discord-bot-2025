package db

import (
	"context"
)

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

type CommandTag interface {
	RowsAffected() int64
	String() string
}

type Tx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
}

type LockKey string

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
	WithKeyLock(ctx context.Context, key LockKey, fn func(ctx context.Context, tx Tx) error) error
}
