# Backend infra libraries

It's a collection of independent reusable components which are wrapped by
dependency injection container. You can use it directly, without DI, but
it provides a good enough defaults and pre-setuped metrics, healthchecks,
swagger, graceful-shutdown and etc.

Endpoints working out of the box:
- `<host>:1984/readiness`, manage graceful shutdown;
- `<host>:1984/metrics`, prometheus metrics;
- `<host>:1984/debug/pprof`, profiler.

Default application port - **8080**.

Examples are available at [folder](./examples).

## How to

At first look into [go.dev/doc](https://go.dev/doc/faq#git_https) to make it
works locally by ssh.

Next, setup `.env` by puttig into it:

- `GH_USER=<user-name>`, username to fetch private lib, usually it's
  `invqauser`, but, you can setup you're own;
- `GH_TOKEN=<token-val>`, user's
  [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).

`.env` will redistribute it over docker-build/compose/integration
tests/tools/etc.

### Versions

- https://go.dev/blog/v2-go-modules#publishing-v2-and-beyond
- https://go.dev/ref/mod#workspaces

To get a specific version of `backend-infra-libraries` in your application, you can use the command like:

```shell
GOPRIVATE="github.com/fusionmedialimited/*" go get -v github.com/fusionmedialimited/backend-infra-libraries/v3@v3.0.0
```
