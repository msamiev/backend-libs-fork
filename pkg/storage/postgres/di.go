package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/config"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
)

func Setup(c *di.Container) {
	di.Set(c, di.OptInit(func() (Config, error) {
		return ConfigFromEnv(Custom)
	}))
	di.Set(c, di.OptInit(func() (*sqlx.DB, error) {
		return NewDB(
			di.Get[context.Context](c),
			di.GetNamed[string](c, config.AppName),
			di.Get[Config](c),
		)
	}), di.OptDeinit(func(db *sqlx.DB) error { return db.Close() }))
}
