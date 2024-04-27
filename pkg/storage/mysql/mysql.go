package mysql

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	// mysql driver.
	_ "github.com/go-sql-driver/mysql"

	"go.opentelemetry.io/otel/trace"

	env "github.com/caarlos0/env/v6"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
)

const (
	MainMaster      Kind = "MASTER_"
	MainReplicaOLTP Kind = "REPLICA_OLTP_" // OLTP means Online Transaction Processing
	MainReplicaOLAP Kind = "REPLICA_OLAP_" // OLAP means Online Analytical Processing
	FooDB           Kind = "FOO_"
	CommentsMaster  Kind = "COMMENTS_MASTER_"
	CommentsReplica Kind = "COMMENTS_REPLICA_"
)

type (
	Kind   = string
	Config struct {
		User              string               `env:"MYSQL_USER"          envDefault:"root"`
		Password          string               `env:"MYSQL_PASSWORD"      envDefault:"toor"`
		Host              string               `env:"MYSQL_HOST"          envDefault:"db"`
		Port              int                  `env:"MYSQL_PORT"          envDefault:"3306"`
		Database          string               `env:"MYSQL_DATABASE"`
		Options           []string             `env:"MYSQL_OPTIONS"       envDefault:"charset=utf8&parseTime=True" envSeparator:"&"`
		MaxIdleTime       time.Duration        `env:"MYSQL_MAX_IDLE_TIME" envDefault:"60s"`
		MaxLifetime       time.Duration        `env:"MYSQL_MAX_LIFETIME"  envDefault:"5m"`
		MaxTotal          int                  `env:"MYSQL_MAX_TOTAL"     envDefault:"32"`
		MaxIdle           int                  `env:"MYSQL_MAX_IDLE"      envDefault:"8"`
		OTELTraceProvider trace.TracerProvider `env:"-"`
	}
)

func ConfigFromEnv(kind Kind) (conf Config, _ error) {
	var (
		opts = env.Options{Prefix: kind}
		err  = env.Parse(&conf, opts)
	)

	return conf, err
}

func makeDSN(user, password, host string, port int, database string, options []string) string {
	var opts string
	if len(options) > 0 {
		opts = "?" + strings.Join(options, "&")
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		database,
		opts,
	)
}

//nolint:gocritic // huge param is ok since called only once during bootstrap
func NewDB(ctx context.Context, name string, conf Config) (db *sqlx.DB, err error) {
	if db, err = otelsqlx.ConnectContext(
		ctx,
		"mysql",
		makeDSN(conf.User, conf.Password, conf.Host, conf.Port, conf.Database, conf.Options),
		otelsql.WithTracerProvider(conf.OTELTraceProvider),
	); err != nil {
		dsn := makeDSN(conf.User, "***", conf.Host, conf.Port, conf.Database, conf.Options)
		return nil, fmt.Errorf("failed to ping %s: %w", dsn, err)
	}

	db.SetConnMaxIdleTime(conf.MaxIdleTime)
	db.SetConnMaxLifetime(conf.MaxLifetime)
	db.SetMaxOpenConns(conf.MaxTotal)
	db.SetMaxIdleConns(conf.MaxIdle)

	prometheus.MustRegister(collectors.NewDBStatsCollector(db.DB, name))

	return db, nil
}
