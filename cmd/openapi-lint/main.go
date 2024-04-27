package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/bootstrap"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/primitives"
)

var ErrNoOpenAPI = errors.New("no openapi version")

func main() {
	var (
		c                  = di.New()
		ctx, cancel        = context.WithCancel(context.Background())
		acceptedExtensions = map[string]struct{}{
			".yaml": {},
			".yml":  {},
			".json": {},
		}
	)
	bootstrap.Setup(ctx, c, "openapi", "lint", nil)
	defer primitives.Must(func() (any, error) { cancel(); return nil, c.Release() }) //nolint:unparam // useless error

	var (
		logger = di.Get[*zap.Logger](c)
	)

	flag.CommandLine.Usage = func() {
		fmt.Println("./openapi-lint <spec dirs>")
		flag.CommandLine.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.CommandLine.Usage()
		logger.Panic("provide at least one positional argument")
	}

	for i := 0; i < flag.NArg(); i++ {
		specDir, err := filepath.Abs(filepath.Clean(flag.Arg(i)))
		if err != nil {
			logger.Panic("wrong path", zap.String("path", flag.Arg(i)), zap.Error(err))
		}

		err = filepath.Walk(specDir, func(path string, info fs.FileInfo, err error) error {
			skip := err != nil || info.IsDir()

			if _, ok := acceptedExtensions[filepath.Ext(path)]; skip || !ok {
				logger.Debug("Skip", zap.String("path", path))
				return err
			}

			logger.Debug("Checking", zap.String("path", path))

			if err = lint(ctx, path); !errors.Is(err, ErrNoOpenAPI) {
				return err
			}

			return nil
		})

		if err != nil {
			logger.Panic("Invalid spec", zap.Error(err))
		}
	}

	logger.Info("Success")
}

func lint(ctx context.Context, specPath string) error {
	spec, err := (&openapi3.Loader{ReadFromURIFunc: openapi3.ReadFromFile}).LoadFromFile(specPath)
	if err != nil {
		return fmt.Errorf("load %s: %w", specPath, err)
	}

	if spec.OpenAPI == "" {
		return fmt.Errorf("skip %s: %w", specPath, ErrNoOpenAPI)
	}

	if err := spec.Validate(ctx); err != nil {
		return fmt.Errorf("validate %s: %w", specPath, err)
	}

	return nil
}
