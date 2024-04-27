package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	// postgresql driver.
	_ "github.com/lib/pq"

	env "github.com/caarlos0/env/v6"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const (
	Custom Kind = ""
)

type (
	Kind   = string
	Config struct {
		User        string        `env:"POSTGRES_USER"          envDefault:"postgres"`
		Password    string        `env:"POSTGRES_PASSWORD"      envDefault:"postgres"`
		Host        string        `env:"POSTGRES_HOST"          envDefault:"db"`
		Port        int           `env:"POSTGRES_PORT"          envDefault:"5432"`
		Database    string        `env:"POSTGRES_DATABASE"`
		Options     []string      `env:"POSTGRES_OPTIONS"       envDefault:"sslmode=disable" envSeparator:" "`
		MaxIdleTime time.Duration `env:"POSTGRES_MAX_IDLE_TIME" envDefault:"60s"`
		MaxLifetime time.Duration `env:"POSTGRES_MAX_LIFETIME"  envDefault:"5m"`
		MaxTotal    int           `env:"POSTGRES_MAX_TOTAL"     envDefault:"32"`
		MaxIdle     int           `env:"POSTGRES_MAX_IDLE"      envDefault:"8"`
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
	var dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		host, port, user, password, database)

	if len(options) > 0 {
		dsn = dsn + " " + strings.Join(options, " ")
	}
	return dsn
}

//nolint:gocritic // huge param is ok since called only once during bootstrap
func NewDB(ctx context.Context, name string, conf Config) (db *sqlx.DB, err error) {
	dsn := makeDSN(conf.User, conf.Password, conf.Host, conf.Port, conf.Database, conf.Options)
	db, err = sqlx.Open("postgres", dsn)
	if err != nil {
		dsn = makeDSN(conf.User, "***", conf.Host, conf.Port, conf.Database, conf.Options)
		return nil, fmt.Errorf("failed to connect %s: %w", dsn, err)
	}

	db.SetConnMaxIdleTime(conf.MaxIdleTime)
	db.SetConnMaxLifetime(conf.MaxLifetime)
	db.SetMaxOpenConns(conf.MaxTotal)
	db.SetMaxIdleConns(conf.MaxIdle)

	if err = db.PingContext(ctx); err != nil {
		dsn = makeDSN(conf.User, "***", conf.Host, conf.Port, conf.Database, conf.Options)
		return nil, fmt.Errorf("failed to ping %s: %w", dsn, err)
	}

	prometheus.MustRegister(collectors.NewDBStatsCollector(db.DB, name))

	return db, nil
}
