package bootstrap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/garsue/watermillzap"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/config"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
)

func LoggerSetup(c *di.Container) {
	di.Set(c, di.OptInit(func() (*zap.Logger, error) {
		conf := di.Get[config.Introspection](c)

		logger, err := newLogger(conf.LogLevel)
		if err != nil {
			return nil, err
		}

		name := di.GetNamed[string](c, config.AppName)
		if envName := conf.Name; envName != "" {
			name = envName
		}

		return logger.With(zap.String("version", di.GetNamed[string](c, config.AppVersion))).
			With(zap.String("hostname", di.GetNamed[string](c, config.Hostname))).
			With(zap.String("name", name)), nil
	}), di.OptDeinit(func(logger *zap.Logger) error {
		_ = logger.Sync()
		return nil
	}))
	di.Set(c, di.OptInit(func() (watermill.LoggerAdapter, error) {
		log := di.Get[*zap.Logger](c)
		return watermillzap.NewLogger(log), nil //nolint: gocritic // it's correct order
	}))
}

func newLogger(lvl int8) (*zap.Logger, error) {
	level := zapcore.Level(lvl)

	if level == zapcore.DebugLevel {
		logger, err := zap.NewDevelopment()
		return logger, err
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)

	return config.Build()
}
