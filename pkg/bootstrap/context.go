package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
)

func ContextSetup(c *di.Container) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)
	go func() {
		select {
		case <-done:
			cancel()
		case <-ctx.Done():
		}
	}()

	di.Set(c, di.OptInit(func() (context.Context, error) {
		return ctx, nil
	}))
	di.Set(c, di.OptInit(func() (context.CancelFunc, error) {
		return cancel, nil
	}), di.OptDeinit(func(cancel context.CancelFunc) error {
		cancel()
		return nil
	}))
}
