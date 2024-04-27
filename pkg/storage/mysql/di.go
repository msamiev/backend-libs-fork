package mysql

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/trace"

	"github.com/jmoiron/sqlx"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/config"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
)

func Setup(c *di.Container) {
	for _, db := range []string{
		MainMaster,
		MainReplicaOLTP,
		MainReplicaOLAP,
		FooDB,
		CommentsMaster,
		CommentsReplica,
	} {
		db := db
		di.SetNamed(c, db, di.OptInit(func() (Config, error) {
			conf, err := ConfigFromEnv(db)
			conf.OTELTraceProvider = di.Get[trace.TracerProvider](c)
			return conf, err
		}))
		di.SetNamed(c, db, di.OptInit(func() (*sqlx.DB, error) {
			return NewDB(
				di.Get[context.Context](c),
				di.GetNamed[string](c, config.AppName)+normalizeIdentifier(db),
				di.GetNamed[Config](c, db),
			)
		}), di.OptDeinit(func(db *sqlx.DB) error { return db.Close() }))
	}
}

func normalizeIdentifier(name string) string {
	return "_" + strings.ToLower(strings.Trim(name, "_"))
}
