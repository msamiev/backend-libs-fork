package primitives

import (
	"context"
	"time"
)

type (
	Timer struct {
		fn EachFn
		d  time.Duration
	}
	EachFn func(context.Context) (oneMore bool, _ error)
)

func NewTimer(d time.Duration, fn EachFn) *Timer {
	return &Timer{fn: fn, d: d}
}

func (t *Timer) RunEach(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := t.run(ctx); err != nil {
		return err
	}

	for {
		if err := t.runEach(ctx); err != nil {
			return err
		}
	}
}

func (t *Timer) runEach(ctx context.Context) error {
	timer := time.NewTimer(t.d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		if err := t.run(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (t *Timer) run(ctx context.Context) (err error) {
	oneMore := true
	for oneMore {
		if oneMore, err = t.fn(ctx); err != nil {
			return err
		}
	}

	return nil
}
