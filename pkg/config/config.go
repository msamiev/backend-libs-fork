package config

import (
	"os"
	"strings"
	"time"

	env "github.com/caarlos0/env/v6"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
)

var version = "dev"

const (
	AppName      = "$app_name"
	AppNamespace = "$app_namespace"
	AppSubsystem = "$app_subsystem"
	AppVersion   = "$app_version"
	Hostname     = "$hostname"
)

func Setup(c *di.Container, namespace, subsystem string) {
	di.SetNamed(c, AppVersion, di.OptInit(func() (string, error) {
		return version, nil
	}))
	di.SetNamed(c, AppNamespace, di.OptInit(func() (string, error) {
		return strings.ReplaceAll(namespace, "-", "_"), nil
	}))
	di.SetNamed(c, AppSubsystem, di.OptInit(func() (string, error) {
		return strings.ReplaceAll(subsystem, "-", "_"), nil
	}))
	di.SetNamed(c, AppName, di.OptInit(func() (string, error) {
		var (
			namespace = di.GetNamed[string](c, AppNamespace)
			subsystem = di.GetNamed[string](c, AppSubsystem)
		)
		return namespace + "_" + subsystem, nil
	}))
	di.SetNamed(c, Hostname, di.OptInit(os.Hostname))
	di.Set(c, di.OptInit(func() (conf Introspection, _ error) {
		return conf, env.Parse(&conf) //nolint:gocritic // it's correct evaluation order
	}))
}

type Introspection struct {
	Name        string        `env:"INTROSPECTION_NAME"       envDefault:""`
	Sock        string        `env:"INTROSPECTION_SOCK"       envDefault:"0.0.0.0:1984"`
	Timeout     time.Duration `env:"INTROSPECTION_TIMEOUT"    envDefault:"5s"`
	ShutdownNum int           `env:"INTROSPECTION_STOP_COUNT" envDefault:"2"`
	LogLevel    int8          `env:"LOG_LEVEL"                envDefault:"0"` // info
}
