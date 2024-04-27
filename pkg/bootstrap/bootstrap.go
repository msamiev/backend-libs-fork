package bootstrap

import (
	"context"

	// 🤷.
	_ "go.uber.org/automaxprocs"
	// introspection.
	_ "net/http/pprof"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/config"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/di"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/http/client"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/http/server"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/storage/mysql"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/storage/postgres"
	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/tracing"
)

func Setup(
	ctx context.Context,
	c *di.Container,
	namespace, subsystem string,
	spec *openapi3.T,
) {
	config.Setup(c, namespace, subsystem)
	ContextSetup(c) //nolint:contextcheck // it provides independent context
	LoggerSetup(c)
	tracing.Setup(ctx, c)
	GracefulSetup(c) //nolint:contextcheck // "lazy" loaded context
	client.Setup(c)  //nolint:contextcheck // "lazy" loaded context
	server.Setup(ctx, c, spec)
	mysql.Setup(c)    //nolint:contextcheck // "lazy" loaded context
	postgres.Setup(c) //nolint:contextcheck // "lazy" loaded context
}
