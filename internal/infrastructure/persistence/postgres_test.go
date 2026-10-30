package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/aktnb/discord-bot-go/internal/interfaces/db"
	"github.com/jackc/pgx/v5"
)

type stubBeginner struct {
	tx  pgx.Tx
	err error
}

func (b stubBeginner) Begin(context.Context) (pgx.Tx, error) { return b.tx, b.err }

type lifecycleTx struct {
	pgx.Tx
	commits, rollbacks     int
	commitErr, rollbackErr error
}

func (tx *lifecycleTx) Commit(context.Context) error {
	tx.commits++
	return tx.commitErr
}

func (tx *lifecycleTx) Rollback(context.Context) error {
	tx.rollbacks++
	return tx.rollbackErr
}

func TestWithTxLifecycle(t *testing.T) {
	cause := errors.New("callback failed")
	cleanupErr := errors.New("rollback failed")
	for _, tt := range []struct {
		name                                          string
		beginErr, callbackErr, commitErr, rollbackErr error
		commits, rollbacks                            int
	}{
		{name: "success", rollbackErr: pgx.ErrTxClosed, commits: 1, rollbacks: 1},
		{name: "begin failure", beginErr: cause},
		{name: "callback failure", callbackErr: cause, rollbacks: 1},
		{name: "commit failure", commitErr: cause, rollbacks: 1, commits: 1},
		{name: "rollback failure", callbackErr: cause, rollbackErr: cleanupErr, rollbacks: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tx := &lifecycleTx{commitErr: tt.commitErr, rollbackErr: tt.rollbackErr}
			manager := &postgresTxManager{pool: stubBeginner{tx: tx, err: tt.beginErr}}
			ctx := context.Background()
			called := false
			err := manager.WithTx(ctx, func(callbackCtx context.Context, wrapped db.Tx) error {
				called = true
				if callbackCtx != ctx || wrapped.(*postgresTx).tx != tx {
					t.Fatal("callback received wrong context or transaction")
				}
				return tt.callbackErr
			})
			if called != (tt.beginErr == nil) || tx.commits != tt.commits || tx.rollbacks != tt.rollbacks {
				t.Fatalf("called=%v, commits=%d, rollbacks=%d", called, tx.commits, tx.rollbacks)
			}
			for _, want := range []error{tt.beginErr, tt.callbackErr, tt.commitErr, tt.rollbackErr} {
				if want != nil && !errors.Is(want, pgx.ErrTxClosed) && !errors.Is(err, want) {
					t.Errorf("error = %v, missing cause %v", err, want)
				}
			}
			if tt.name == "success" && err != nil {
				t.Fatalf("success returned %v", err)
			}
		})
	}
}

func TestWithTxRollsBackOnPanic(t *testing.T) {
	tx := &lifecycleTx{}
	manager := &postgresTxManager{pool: stubBeginner{tx: tx}}
	marker := errors.New("panic in callback")
	defer func() {
		if recovered := recover(); recovered != marker {
			t.Errorf("panic = %v, want original panic", recovered)
		}
		if tx.rollbacks != 1 || tx.commits != 0 {
			t.Errorf("commits=%d, rollbacks=%d", tx.commits, tx.rollbacks)
		}
	}()
	_ = manager.WithTx(context.Background(), func(context.Context, db.Tx) error { panic(marker) })
}
