package mysql

import (
	"os"
	"testing"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/primitives"
)

func envSetter(envs map[string]string) (closer func()) {
	originalEnvs := map[string]string{}

	for name, value := range envs {
		if originalValue, ok := os.LookupEnv(name); ok {
			originalEnvs[name] = originalValue
		}
		_ = os.Setenv(name, value)
	}

	return func() {
		for name := range envs {
			origValue, has := originalEnvs[name]
			if has {
				_ = os.Setenv(name, origValue)
			} else {
				_ = os.Unsetenv(name)
			}
		}
	}
}

func fromEnv(env map[string]string) Config {
	var closer = envSetter(env)
	defer closer()
	return primitives.Must(func() (Config, error) {
		return ConfigFromEnv("TEST_")
	})
}

func Test_makeDSN(t *testing.T) {
	type args struct {
		conf Config
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"default",
			args{
				fromEnv(map[string]string{}),
			},
			"root:toor@tcp(db:3306)/?charset=utf8&parseTime=True",
		},
		{
			"clear options",
			args{
				fromEnv(map[string]string{
					"TEST_MYSQL_OPTIONS": "",
				}),
			},
			"root:toor@tcp(db:3306)/",
		},
		{
			"from HOST env",
			args{
				fromEnv(map[string]string{
					"TEST_MYSQL_USER":     "mysql",
					"TEST_MYSQL_PASSWORD": "secret",
					"TEST_MYSQL_HOST":     "myhost",
					"TEST_MYSQL_PORT":     "666",
					"TEST_MYSQL_DATABASE": "mydb",
					"TEST_MYSQL_OPTIONS":  "foo=bar&baz=boo",
				}),
			},
			"mysql:secret@tcp(myhost:666)/mydb?foo=bar&baz=boo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeDSN(
				tt.args.conf.User,
				tt.args.conf.Password,
				tt.args.conf.Host,
				tt.args.conf.Port,
				tt.args.conf.Database,
				tt.args.conf.Options,
			); got != tt.want {
				t.Errorf("makeDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}
