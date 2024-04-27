package primitives

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimerCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	timer := NewTimer(time.Second, func(context.Context) (bool, error) {
		t.FailNow()
		return true, nil
	})
	err := timer.RunEach(ctx)

	if !errors.Is(err, context.Canceled) {
		t.FailNow()
	}
}

func TestTimerOneMore(t *testing.T) {
	var (
		ctx, cancel  = context.WithCancel(context.Background())
		count, limit = 0, 10
	)

	timer := NewTimer(time.Nanosecond, func(context.Context) (bool, error) {
		cancel()

		if count == limit {
			return false, nil
		}
		count++
		return true, nil
	})
	err := timer.RunEach(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Fail()
	}
	if count != limit {
		t.Fail()
	}
}

func TestTimerNooneMore(t *testing.T) {
	var (
		ctx, cancel  = context.WithCancel(context.Background())
		count, limit = 0, 10
	)

	timer := NewTimer(time.Nanosecond, func(context.Context) (bool, error) {
		if count == limit {
			cancel()
			return false, nil
		}
		count++
		return false, nil
	})
	err := timer.RunEach(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Fail()
	}
	if count != limit {
		t.Fail()
	}
}

func TestTimerError(t *testing.T) {
	var target = errors.New("test")

	timer := NewTimer(time.Nanosecond, func(context.Context) (bool, error) {
		return true, target
	})
	err := timer.RunEach(context.Background())

	if !errors.Is(err, target) {
		t.Fail()
	}
}
