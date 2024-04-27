package main

import (
	"context"
	"net/http"
	"time"
	"io"

	"go.uber.org/zap"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/bootstrap"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/primitives"
)

func main() {
	var (
		c           = di.New()
		ctx, cancel = context.WithCancel(context.Background())
	)
	bootstrap.Setup(ctx, c, "example", "worker", nil)
	defer primitives.Must(func() (any, error) {
		cancel()
		err := c.Release()
		return nil, err
	})

	var (
		logger = di.Get[*zap.Logger](c)
		client = di.Get[*http.Client](c)
	)

	t := time.NewTicker(time.Second)
	for {
		select {
		case <-di.Get[context.Context](c).Done():
			return
		case <-t.C:
			req := primitives.Must[*http.Request](func() (*http.Request, error) {
				return http.NewRequestWithContext(ctx, http.MethodGet, "http://api:8080/test", nil)
			})
			resp, err := client.Do(req)
			if err != nil {
				logger.Error("error", zap.Error(err))
				continue
			}
			result := primitives.Must[[]byte](func() ([]byte, error) {
				return io.ReadAll(resp.Body)
			})
			_ = resp.Body.Close()
			logger.Info("tick", zap.ByteString("result", result))
		}
	}
}
