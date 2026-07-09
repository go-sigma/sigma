# AGENTS.md

Guidance for AI agents working in the `sigma` repository.

## Project Overview

`sigma` is an OCI artifact storage and distribution system written in Go. It implements the [OCI Distribution Specification 1.1](https://github.com/opencontainers/distribution-spec/tree/v1.1.0) and acts as a private/public container registry. It supports docker registry v2 protocol, OCI images/artifacts/sboms, security scanning, registry proxy, namespace quotas, garbage collection, image signing (cosign), and image building via docker/podman/kubernetes.

- Module path: `github.com/go-sigma/sigma`
- Go version: 1.26.3 (godebug default=go1.26)
- License: Apache 2.0
- Entry point: `main.go` -> `cmd.Execute()`
- Default port: 3000
- Default credentials: `sigma` / `Admin@123`

## Build & Run

```bash
# Build the binary (requires CGO_ENABLED=1 for sqlite)
make build

# Run via docker
docker run --name sigma -p 3000:3000 --rm tosone/sigma:nightly-alpine

# Run locally (needs a config file, default /etc/sigma/config.yaml)
go run . server -c conf/config.yaml

# Subcommands: server, worker, builder, distribution, tools
```

Build tags required for tests/lint/build:

```text
netgo,timetzdata,exclude_graphdriver_btrfs,containers_image_openpgp
```

## Common Commands

```bash
# Lint
make lint                       # runs golangci-lint + hadolint
make lint-go                    # golangci-lint only (timeout 10m)

# Tests (set CI_DATABASE_TYPE=sqlite3|mysql|postgresql)
CI_DATABASE_TYPE=sqlite3 go test -parallel 1 -failfast \
  -tags "netgo,timetzdata,exclude_graphdriver_btrfs,containers_image_openpgp" \
  -timeout 30m ./...

# Code generation
make gormgen                    # regenerate gorm models/queries from DB
make swagen                     # regenerate swagger docs
make migration-create MIGRATION_NAME=<name>   # new SQL migration

# Formatting
make sql-format                 # format all SQL migrations
make addlicense                 # add Apache license headers

# Vendor / clean
make vendor
make clean
```

## Git Commit Messages

- Generated commit messages must use Conventional Commits and be written
  entirely in English, including the subject, body, footer, and verification
  notes.
- Include a body with what changed and why when the change needs more than a
  one-line subject, and put executed validation commands under `Verification:`.

## Test Conventions

- Tests use `github.com/stretchr/testify/require` (preferred over `assert`).
- For DI setup in tests, construct a fresh `dig.New()` container and provide `config.Configuration{}` plus `testkit.NewGin` before calling `factory{}.Initialize(digCon)`.
- Use `testkit.NewGin()` (in `pkg/testkit/gin.go`) to get a gin engine in `TestMode`; do not call `gin.New()` directly in tests.
- Repository tests must import the target repository domain package with a
  semantic alias, such as `registryrepo`, `namespacerepo`, or `userrepo`.
- Generate cryptographic keys (e.g., Ed25519) at runtime in tests; do not hardcode base64 keys.
- Tests are excluded for these packages (do not expect tests there): `pkg/testkit`, `pkg/dal/query`, `pkg/dal/cmd`, `pkg/handlers/apidocs`, and various `mocks` subpackages.
- Some storage tests (`pkg/storage/cos`, `pkg/storage/oss`) are skipped on PRs and require secret env vars (`COS_*`, `OSS_*`).
- Mocks are generated via `go.uber.org/mock` and live alongside the code as `*_mocks.go`.

## Code Conventions

### Import aliases

- Repository package aliases use the `repo` prefix followed by the domain name,
  such as `repouser`, `reponamespace`, `reporegistry`, and `repowebhook`.
- Service package aliases use the `svc` prefix followed by the domain name,
  such as `svcnamespace`, `svcrepository`, `svcmanifest`, and `svcaudit`.
- Do not use suffix-style aliases such as `userrepo`, `namespacesvc`, or
  `auditrepo`.

### Variable scope and unused parameters

- Do not shadow an outer `err` with `err := ...` in nested blocks, goroutines, or
  deferred cleanup functions. Use assignment (`err = ...`) when updating the
  outer error is intentional, or use a short local name such as `e` for isolated
  cleanup/logging errors.
- In deferred cleanup, log the cleanup error variable itself, not an unrelated
  outer `err`.
- If a required method signature includes an unused `context.Context`, name the
  parameter `_ context.Context` in the signature. Do not add `_ = ctx` inside the
  function body just to silence unused parameter checks.

### Error handling

- Use `errcode.AsType[T](err)` (in `pkg/server/errcode/astype.go`) instead of `errors.As(err, &e)` for typed error assertions. The canonical usage: `errcode.AsType[errcode.ErrCode](err)`.
- Handlers return `error` (echo-style) and are wrapped via `server.Wrap(h)` or `server.WrapHandle(r, method, path, h)`. See `pkg/server/adapter.go`.
- Return `errcode.NewHTTPError(c, ...)` from handlers for structured HTTP errors.

### Dependency Injection

- DI is powered by `go.uber.org/dig`. Handlers declare `dig.In` and request services by field.
- Do NOT use the old `utils.GetObjFromDigCon` / `utils.MustGetObjFromDigCon` helpers (removed). Call `digCon.Invoke(func(deps ...) error { ... })` directly.
- Handler factories register routes via `handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{})` in an `init()`.

### Handler structure

- Each handler package (e.g. `pkg/server/handlers/systems`) defines a `Handler` interface, a private `handler` struct embedding `dig.In`, and a `factory{}` with `Initialize(dig.Container) error`.
- The `systems` handler uses field `SystemsService` of type `service.SystemsServiceFactory`. Do not use a field named `config`.
- Routes are grouped under `consts.APIV1` (i.e. `/api/v1`).

### IDs & crypto

- All application IDs must be time-ordered UUIDv7 strings.
- Generate UUIDs through `pkg/utils/uuid.NewV7String()` instead of calling an
  external UUID package directly.
- While the project is on Go 1.26, `pkg/utils/uuid` is backed by
  `github.com/google/uuid`.
- After upgrading to Go 1.27 or later, switch the `pkg/utils/uuid`
  implementation to the standard library UUID implementation and remove the
  external UUID generator.

### Modernize Go

- Prefer modern Go idioms over legacy patterns. This project targets Go 1.26 and
  follows the modernizations that `go fix` (rewritten in Go 1.26 on top of the
  analysis framework) applies. Run `go fix ./...` after a toolchain bump to keep
  the codebase current (use `go fix -diff ./...` to preview first). See commit
  `feat: update with go fix modernize`.
- Concretely, prefer:
  - `any` over `interface{}`.
  - the `min`/`max` builtins over hand-written if/else comparisons.
  - `for i := range n` over C-style `for i := 0; i < n; i++` loops.
  - `slices` / `maps` stdlib helpers (e.g. `slices.Contains`, `slices.Sort`,
    `maps.Keys`) over manual loops and `sort.Slice`.
  - `strings.Cut` / `strings.CutPrefix` over `strings.Index` / `HasPrefix` +
    `TrimPrefix` slicing.
  - `fmt.Appendf` over `[]byte(fmt.Sprintf(...))`.
  - `reflect.TypeFor[T]()` over `reflect.TypeOf`.
  - `t.Context()` in tests over `context.WithCancel` boilerplate.
- These are also reinforced elsewhere in this file: `log/slog` for logging (see
  Logging) and `context.Context` threaded through the call chain (see Tracing).

### Configuration

- Config is loaded from YAML (default `/etc/sigma/config.yaml`, override with `-c`). See `conf/config.yaml` for the full schema.
- Supports `sqlite3`, `turso`, `mysql`, `postgresql`; `redis` (none/external); cache (inmemory/redis); workqueue (inmemory/redis/database); locker (inmemory/redis); storage (filesystem/s3/cos/oss); `trace` (enabled/endpoint/protocol/insecure/sampleRatio/serviceName).

### Logging

- Uses the standard library `log/slog`. Get the logger via `slog.Default()` (global)
  or `slog.FromContext(ctx)` (context-aware, falls back to default).
- Configure the level via `logger.SetLevel(levelStr)` from `pkg/logger` — installs
  a `slog.TextHandler` writing to stdout with `AddSource: true`.
- For fatal errors, use `logger.Fatal(msg, args...)` from `pkg/logger` (logs at Error
  then `os.Exit(1)`); do not call `slog.Error` followed by `os.Exit` directly.
- The gorm adapter `logger.ZLogger{}` is wired in `pkg/dal/dal.go`.
- Avoid re-introducing `log.Logger.WithContext(ctx)` calls — they are no longer needed;
  slog's default global logger is used everywhere.

### Tracing

- Distributed tracing is powered by OpenTelemetry. Initialization lives in
  `pkg/telemetry/trace.go` and is invoked from each long-lived subcommand
  (`server`, `worker`, `distribution`) via `telemetry.Init(*config.GetConfig())`.
  The `builder` subcommand is a one-shot CLI and intentionally skips trace init.
- The tracer provider is configured from the `trace:` section of the config:
  `enabled`, `endpoint` (OTLP collector), `protocol` (`grpc` or `http`),
  `insecure` (TLS off), `sampleRatio` (0..1), `serviceName`.
- Sampler is `ParentBased(TraceIDRatioBased(sampleRatio))` — root spans respect
  the ratio, child spans always inherit the parent's sampling decision.
- Propagator is W3C `TraceContext + Baggage`.
- Instrumentation is attached at the integration points (no manual spans in
  handlers/services):
  - Gin: `otelgin.Middleware(consts.AppName)` is the first middleware in
    `pkg/app/helper.go`.
  - GORM: `tracing.NewPlugin()` from `gorm.io/plugin/opentelemetry/tracing`
    is applied to every `gorm.Open` site in `pkg/dal/dal.go`.
  - Redis: `redisotel.InstrumentTracing(cli)` is called in
    `pkg/dal/redis/redis.go` and `pkg/infra/cache/cacher_redis.go`.
  - Outbound HTTP (resty): the `http.Client` transport is wrapped with
    `otelhttp.NewTransport(...)` in
    `pkg/distribution/clients/clients.go` and
    `pkg/background/daemon/webhook/webhook.go`.
- Always pass `context.Context` through the call chain so spans propagate.
  Do not call `context.Background()` mid-request unless you intentionally
  want to break the trace.

### Linting rules (`.golangci.yml`)

- Enabled linters: asciicheck, bidichk, bodyclose, copyloopvar, cyclop (max 50), gocritic, gosec, misspell, prealloc, unconvert, unparam, varnamelen, whitespace.
- Imports ordered by: standard, blank, default, `github.com/go-sigma/sigma` prefix, dot. (enforced by `gci`/`goimports`).
- `cyclop` is disabled for `*_test.go`.

## Project Structure

```text
cmd/                     CLI entrypoints (server, worker, builder, distribution, tools)
main.go                  Program entry; calls cmd.Execute()
pkg/
  app/                   Application assembly layer
    bootstrap/           Bootstrapping (dig container, signing keys, initial user)
    distribution/        Distribution-only process startup
    server/              Server process startup
    worker/              Worker process startup
  api/                   HTTP/API DTOs
    enums/               Shared API and configuration enums
  background/            Background execution domain
    builder/             Image build backends (docker, kubernetes, podman)
    cronjob/             Scheduled jobs
    daemon/              Background daemons (gc, scan, webhook, transfer, coderepo)
  config/                Config structs & defaults
  consts/                 Constants (API paths, regex, HTTP codes, settings)
  dal/                   Data access layer
    models/              GORM models
    query/               Generated gorm/gen queries (do not hand-edit)
    repository/          Repository layer grouped by domain
      analytics/         Analytics metrics repositories
      audit/             Audit log repositories
      builder/           Builder repositories
      coderepo/          Source code repository repositories
      daemon/            Background daemon repositories
      namespace/         Namespace and namespace member repositories
      registry/          OCI registry repositories (artifact/blob/repository/tag)
      setting/           System setting repositories
      user/              User repositories
      webhook/           Webhook repositories
    migrations/          SQL migrations per DB (mysql/postgresql/sqlite3)
    redis/               Redis client
  server/
    adapter.go           Handler wrap/error dispatch
    errcode/             Error types, AsType generic, HTTP error mapping
    handlers/            HTTP handlers grouped by domain
      <domain>/
        handler.go       Handler interface + factory + route registration
        handler_test.go  Factory/DI test
        <action>.go      Individual route handlers
    middlewares/         authn, authz, etag, extractor, healthz, metrics
  service/               Business logic layer (interfaces + impls + mocks)
    password/, token/    Auth helpers
  authz/                 Authorizer, permission, resolver, cache
  distribution/          OCI distribution protocol domain
    clients/             Distribution/proxy clients
    reference/           Image reference helpers
    signing/             Cosign signing integration
  infra/                 Infrastructure modules
    cache/               Cache implementations
    lock/                Locker implementations
    registry/            Generic factory/implementation registries
    timewheel/           Time wheel module
    workq/               Work queue implementations
  logger/                slog setup + glog adapter
  testkit/               Test helpers (NewGin, CI database setup)
  telemetry/             OpenTelemetry tracer provider init + shutdown
  utils/                 Misc utilities
  validators/            Request validators
  version/               Build-time version info (set via -ldflags)
conf/                    Example config files & TLS certs
deploy/sigma/            Helm chart
build/                   Dockerfiles (server, builder, web, base)
docs/                    Docusaurus documentation site
e2e/                     End-to-end test scripts
```

Directory rules:

- New repository code must be added to the matching
  `pkg/dal/repository/<domain>` package instead of the repository root.
- Handler package names must match their directory names.
- Handlers should not import `pkg/background/*` directly; expose background
  behavior through the service layer unless the code is an explicit startup or
  boundary adapter.

## CI Workflows (`.github/workflows/`)

- **lint.yml**: golangci-lint (Go 1.25) + hadolint on Dockerfiles. Triggers on `main`/`dev` PRs.
- **test.yml**: matrix runs for `sqlite3`, `mysql`, `postgresql` (Go 1.23). Uses MySQL 8.0, Postgres 15, Redis 7, Minio services. Uploads coverage to Codecov.
- **e2e.yml**, **image-build.yml**, **gh-pages.yml**, **codeql.yml**: end-to-end, image build, docs deploy, security scan.

## Important Files

- `pkg/server/adapter.go` - `HandlerFunc` + `Wrap`/`WrapHandle` + centralized error dispatch.
- `pkg/server/errcode/astype.go` - `AsType[T error]` generic helper.
- `pkg/server/errcode/xerrors.go` - `NewHTTPError`, `ErrCode`, `HTTPErrCodeInternalError`.
- `pkg/dal/dig.go` - DI registration for all repositories + authorizer.
- `pkg/server/handlers/systems/handler.go` - canonical handler example.
- `pkg/testkit/gin.go` - `NewGin()` test helper.
- `pkg/telemetry/trace.go` - OpenTelemetry tracer provider init + shutdown.
- `Makefile` - all build/lint/test/generation targets.
- `.golangci.yml` - lint configuration.

## Things to Avoid

- Do not reintroduce `GetObjFromDigCon` / `MustGetObjFromDigCon`; use `dig.Container.Invoke`.
- Do not use `errors.As` directly when an `errcode.AsType[T]` variant exists.
- Do not introduce non-UUID or unordered UUID generators. Use
  `pkg/utils/uuid.NewV7String()` so new IDs remain UUIDv7 and time-ordered.
- Do not add a `config` field to handler structs; use the appropriate service factory field (e.g., `SystemsService`).
- Do not import the repository root package in tests; import the target
  repository domain package with a semantic alias.
- Do not hand-edit generated files under `pkg/dal/query/*.gen.go`; regenerate with `make gormgen`.
- Do not create new test files that hardcode cryptographic material; generate at runtime.
- Do not import `github.com/rs/zerolog`; use `log/slog`.
- Do not use `interface{}`; use `any` (see Modernize Go).
- Do not add `log.Logger.WithContext(ctx)` calls; use the context directly or `slog.FromContext(ctx)`.
- Do not call `telemetry.Init` from the `builder` subcommand — it is a one-shot CLI and uses env vars, not the YAML config.
- Do not create manual spans in handlers/services; rely on the auto-instrumentation in `pkg/telemetry` and `pkg/app/helper.go`.
