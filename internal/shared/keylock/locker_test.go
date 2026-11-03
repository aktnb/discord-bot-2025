package keylock

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLockCancellationAndIndependentKeys(t *testing.T) {
	l := New()
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- l.WithLock(context.Background(), "a", func() error { close(entered); <-release; return errors.New("callback") })
	}()
	<-entered
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.WithLock(ctx, "a", func() error { t.Fatal("cancelled callback ran"); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	waitCtx, stop := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer stop()
	if err := l.WithLock(waitCtx, "a", func() error { t.Fatal("same key overlapped"); return nil }); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	otherCtx, stopOther := context.WithTimeout(context.Background(), time.Second)
	defer stopOther()
	if err := l.WithLock(otherCtx, "b", func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-done; err == nil {
		t.Fatal("lost callback error")
	}
	if err := l.WithLock(context.Background(), "a", func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if len(l.entries) != 0 {
		t.Fatal("unused entries leaked")
	}
}
func TestPanicReleasesLock(t *testing.T) {
	l := New()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic swallowed")
			}
		}()
		_ = l.WithLock(context.Background(), "a", func() error { panic("test") })
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := l.WithLock(ctx, "a", func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if len(l.entries) != 0 {
		t.Fatal("entry leaked after panic")
	}
}
