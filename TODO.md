# High Availability and High Load Optimization TODO

> Architecture analysis based on the current codebase as of 2026-07-13, ordered by priority.

---

## 1. Missing Database Connection Pool Configuration (done)

**Current state**: [`pkg/dal/dal.go`](file:///Users/bytedance/code/github/sigma/pkg/dal/dal.go#L111-L194)

- MySQL (`connectMysql`) and PostgreSQL (`connectPostgres`) did not configure `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, or `SetConnMaxIdleTime`
- SQLite had basic settings (`MaxOpenConns=10`, `MaxIdleConns=3`), but MySQL and PostgreSQL used driver defaults. In practice, `MaxOpenConns=0` means unlimited connections, which can exhaust the database or overload it under high concurrency

**Recommendations**:

- [x] Add connection pool parameters for MySQL and PostgreSQL. Suggested defaults: `MaxOpenConns=50-100`, `MaxIdleConns=10-25`, `ConnMaxLifetime=1h`, `ConnMaxIdleTime=10m`
- [x] Expose these parameters in `config.yaml` so each environment can tune them

---

## 2. Full Aggregation Queries on the Push Hot Path (done)

**Current state**: [`pkg/dal/repository/registry/artifact.go`](file:///Users/bytedance/code/github/sigma/pkg/dal/repository/registry/artifact.go#L392-L410)

- `GetNamespaceSize` and `GetRepositorySize` executed `SELECT SUM(blobs_size) FROM artifacts WHERE namespace_id/repository_id = ?` after every completed push
- With tens of thousands of artifacts, a single `SUM` may still be fast, but maintaining strong consistency requires locking, which serializes concurrent pushes

**Recommendations**:

- [x] Introduce atomic incremental counters: update repository size inside the push transaction with `UPDATE repository SET size = size + ?`, which is O(1) instead of O(N)
- [x] Add a background reconcile task to periodically run full recalculation, through cronjob or daemon
- [x] Make quota checks use incremental counters instead of real-time `SUM`

---

## 3. Database Hot Spots in the Analytics Backend

**Current state**: `conf/config.yaml` defaults `counterBackend` to `database`

- In `database` mode, `RecordPush` synchronously executes `INSERT ... ON CONFLICT DO UPDATE`, causing row-lock contention when many pushes hit the same namespace
- Confirmed direction: the `redis` backend can use `HINCRBY` for in-memory aggregation plus a minute-level `flushLoop` to batch writes to the database and remove the hot spot

**Recommendations**:

- [ ] Remove the `database` counter backend, keeping only `inmemory` as the default and `redis`
- [ ] Make `inmemory` aggregate by hour with `sync.Map`, flush during graceful shutdown, and restore the current-hour counters from the database on startup
- [ ] Recommend `redis` mode for production

---

## 4. Insufficient Cache System (done)

**Result**: [`pkg/infra/cache/`](file:///Users/bytedance/code/github/sigma/pkg/infra/cache/), [`pkg/dal/repository/namespace/namespace_cache.go`](file:///Users/bytedance/code/github/sigma/pkg/dal/repository/namespace/namespace_cache.go), [`pkg/dal/repository/registry/repository_cache.go`](file:///Users/bytedance/code/github/sigma/pkg/dal/repository/registry/repository_cache.go), [`pkg/app/cache_prewarm.go`](file:///Users/bytedance/code/github/sigma/pkg/app/cache_prewarm.go)

- The generic cache was improved:
  - `Get` now merges concurrent fetches for the same missed key through `singleflight`, avoiding DB stampedes when hot keys expire
  - The in-memory cache supports TTL, and Redis cache continues to use Redis TTL. `Set(ctx, key, val, ttl)` works for both backends
  - Short-lived negative caching is supported, so deterministic misses such as `gorm.ErrRecordNotFound` do not repeatedly penetrate to the database
  - `cache_requests_total` and `cache_fetch_duration_seconds` expose low-cardinality metrics by backend, prefix, and result
- Namespace and repository metadata are cached through DAL cached decorators:
  - DI injects cached repositories by default. Transaction-scoped `New...Repository(tx)` still returns the raw DB implementation to avoid caching uncommitted data
  - `Get` and `GetByName` seed both `id` and `name` cache keys
  - List queries do not cache paginated results, but opportunistically seed returned objects
- Consistency is handled through active invalidation:
  - Namespace and repository create, update, delete, size increment/decrement, and size reconcile paths invalidate cache after successful writes
  - Manifest push/delete paths explicitly clean metadata cache after raw repository writes inside transactions
  - TTL remains a fallback. Multi-replica production deployments should use Redis cache so invalidation results are shared
- Startup prewarming was added:
  - After server/distribution dependency checks pass, the app prewarms recently updated namespaces and repositories up to `cache.prewarm.limit`
  - `cache.prewarm.enabled`, `limit`, and `timeout` are configurable. Prewarm failures are logged but do not block startup
- A local L1 cache plus Redis L2 cache was evaluated:
  - This release keeps the single-backend cache model selected by `cache.type` (`inmemory` or `redis`)
  - Multi-replica production deployments should use Redis cache. Single-node or development deployments can use inmemory cache

**Completed items**:

- [x] Add `singleflight` to avoid cache stampedes by merging concurrent requests into one DB fetch
- [x] Add startup prewarming for hot namespace and repository metadata
- [x] Add cache hit-rate metrics
- [x] Evaluate namespace/repository local cache plus Redis second-level cache; keep the single-backend cache model for now

---

## 5. Distributed Lock Reliability (done)

**Current state**: [`pkg/infra/lock/`](file:///Users/bytedance/code/github/sigma/pkg/infra/lock/)

- Redis locks are implemented with `github.com/go-redsync/redsync/v4` and reuse the existing `redis.UniversalClient`
- The project wraps `AcquireWithRenew` with a renewal watchdog. `redsync` itself provides manual `ExtendContext`, not automatic renewal
- The current implementation still uses a single Redis pool and does not introduce a multi-Redis quorum configuration

**Recommendations**:

- [x] Implement lock retry with exponential backoff through the redsync retry delay function
- [x] Evaluate whether Redlock or a Raft-based lock is needed. The current design reuses a single Redis connection and does not introduce a multi-node quorum. If future deployments require stronger multi-Redis consistency, evaluate multi-pool Redlock or a stronger Raft-based lock
- [x] Add metrics for lock wait time, lock hold time, and renewal results

---

## 6. Missing Streaming and Multipart Uploads in Storage

**Current state**: [`pkg/storage/`](file:///Users/bytedance/code/github/sigma/pkg/storage/)

- All storage drivers (`filesystem`, S3, COS, OSS) write data after reading the whole object
- Object stores such as S3, COS, and OSS support Multipart Upload natively, but the drivers do not use it
- There is no resumable upload support and no parallel multipart upload support

**Recommendations**:

- [ ] Implement Multipart Upload for S3, COS, and OSS drivers, including parallel part upload and `CompleteMultipartUpload`
- [ ] Automatically switch large blobs to multipart strategy, for example above 5 MB
- [ ] Support upload progress tracking and resumable uploads
- [ ] Add a `StreamingWriter` or `MultipartUploader` abstraction to the storage interface

---

## 7. Weak Observability

**Current state**: [`pkg/server/middlewares/metrics/metrics.go`](file:///Users/bytedance/code/github/sigma/pkg/server/middlewares/metrics/metrics.go)

- Only basic HTTP request counters and latency histograms exist: `http_requests_total`, `http_request_duration_seconds`
- Missing business metrics:
  - Blob upload/download throughput and size distributions
  - Manifest push/pull latency
  - GC task duration and cleanup volume
  - Storage usage
  - Workqueue backlog
- Logs do not have sampling, which can amplify log volume under high load
- There are no predefined alert rules

**Recommendations**:

- [ ] Add business-level Prometheus metrics for blob operations, GC, and queues
- [ ] Implement log sampling through `slog` with `NewSamplingHandler` or a custom handler
- [ ] Add a `/metrics` endpoint for Prometheus metrics, or confirm and document the existing endpoint
- [ ] Define key alert rules for error rate, P99 latency, storage usage, and queue backlog
- [ ] Add daemon and cronjob task metrics for success, failure, and duration

---

## 8. Missing Rate Limiting and Overload Protection

**Current state**: middleware packages include `authn`, `authz`, `etag`, `extractor`, `healthz`, and `metrics`

- There is no rate-limiting middleware, such as token bucket or sliding window
- There is no circuit breaking or degradation when downstream systems such as DB, Redis, or S3 fail
- Request body limits have been added through `http.bodyLimit`, split by regular API, manifest, and blob upload request classes
- Full HTTP timeouts are configured through `http.timeout`: `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`

**Recommendations**:

- [ ] Add rate-limiting middleware by IP, user, or namespace
- [ ] Add circuit breakers, such as `gobreaker` or `sony/gobreaker`, to protect DB and Redis calls
- [x] Configure full `http.Server` timeouts: `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`
- [x] Add `MaxBytesReader` request body limits
- [x] Support `storage.redirect` to redirect blob pull traffic to object storage, preventing large pull traffic from blocking application processes

---

## 9. Missing Fine-Grained Distributed Mutual Exclusion for Some Background Tasks (done)

**Result**:

- Workqueue uses channels in single-instance mode and Redis MQ in multi-instance mode. Multiple instances do not claim the same message at the same time, so consumer-level leader election is not required
- GC uses rule-level locks, builder uses builder-ID-level locks, and analytics flush uses hour-level locks. Audit, size, and vulnerability tasks use their own logical task keys
- If a lock cannot be acquired, the worker quickly skips the current fine-grained task. Builder also uses database conditional updates to atomically claim expired tasks
- Vulnerability tasks can recover stale `Doing` tasks after `staleTimeout`, with a default grace period of 6 hours
- Multi-replica Helm deployments use Redis workqueue and Redis locker together

**Completed items**:

- [x] Add Redis-based fine-grained distributed locks only where task deduplication is required
- [x] Name lock keys by task type and resource ID, avoiding daemon-level global locks
- [x] Skip the current task quickly when lock acquisition fails and continue processing the next task
- [x] Keep task operations idempotent and use database atomic status transitions as the final fallback where needed

---

## 10. Incomplete Graceful Shutdown and Startup Checks (done)

**Current state**: [`pkg/app/server/server.go`](file:///Users/tosone/code/github/sigma/pkg/app/server/server.go#L156-L164)

- Graceful shutdown had a 10-second timeout, while `graceful.Shutdown()` had another 30-second timeout, scattering the shutdown budget
- `http.Server.Shutdown()` could be canceled by `context.WithTimeout` before active connections completed
- `/healthz` returned 200 only and did not check DB or Redis connectivity. Readiness should be more meaningful

**Result**: [`pkg/graceful/graceful.go`](file:///Users/tosone/code/github/sigma/pkg/graceful/graceful.go), [`pkg/app/readiness.go`](file:///Users/tosone/code/github/sigma/pkg/app/readiness.go), [`pkg/app/drain.go`](file:///Users/tosone/code/github/sigma/pkg/app/drain.go)

- Fixed the `runAtShutdown` data race by adding a `sync.Mutex`; `Shutdown` copies a snapshot under the lock before execution
- Removed the unnecessary inner goroutine plus WaitGroup pattern and directly invokes shutdown functions with `defer recover()`
- Panic values are logged with structured slog fields (`name` and `err`) instead of being discarded
- `Shutdown` is idempotent through `sync.Once`; repeated calls are safe no-ops
- `Shutdown` now has the signature `func Shutdown(ctx context.Context) error`, returning `context.DeadlineExceeded` on timeout
- Server, worker, and distribution now share one 30-second `context.WithTimeout`, passed to both `httpServer.Shutdown(ctx)` and `graceful.Shutdown(ctx)`
- `/readyz` dependency checks now cover DB ping, Redis ping, and storage `Ping`. The new `StorageDriver.Ping` method is implemented by filesystem, S3, COS, and OSS by writing and deleting a probe object with a random suffix, making it safe for multiple replicas. Delete failures are ignored
- Pre-stop draining is supported. On SIGTERM, `SetDraining(true)` makes `/readyz` return 503 immediately, waits 5 seconds for drain, then shuts down. Deployment manifests also add `lifecycle.preStop` sleep as a fallback
- Startup validation runs `ValidateOnStartup` before `ListenAndServe`, retrying dependencies with 2-second backoff up to 60 seconds. Failure exits the process. Deployment manifests add `startupProbe` pointing to `/readyz`
- Deployment manifests set `terminationGracePeriodSeconds` to 45 seconds for server, worker, and distribution: 5 seconds drain, 30 seconds shutdown, plus buffer. `drainDelay` is configurable
- Tests were aligned with project conventions: `assert` was replaced with `require`, `time.Sleep` hacks were removed, idempotency and panic recovery cases were added, and timeout tests are driven by `context.WithTimeout`

**Recommendations**:

- [x] Unify graceful shutdown timeout handling, merging the HTTP 10-second and graceful 30-second budgets into one budget
- [x] Add dependency checks to `/readyz`: DB ping, Redis ping, and storage reachability
- [x] Add a pre-stop hook to remove the pod from load balancing before waiting for active requests to drain
- [x] Validate all required dependencies before accepting traffic

---

## 11. Outdated Kubernetes Deployment Configuration

**Result**: [`deploy/sigma/templates/server/hpa.yaml`](file:///Users/bytedance/code/github/sigma/deploy/sigma/templates/server/hpa.yaml)

- HPA has been upgraded to `autoscaling/v2`
- PodDisruptionBudget has been added
- Default resources are configured for server, worker, distribution, and web
- Go components use `/readyz` for readinessProbe and keep `/healthz` for livenessProbe
- `terminationGracePeriodSeconds` has been added

**Recommendations**:

- [x] Upgrade HPA to `autoscaling/v2`
- [x] Add PodDisruptionBudget
- [x] Set reasonable default resources, including requests and limits, for every component
- [x] Use `/readyz` for readinessProbe with dependency checks, and keep `/healthz` for livenessProbe as a process-liveness check
- [x] Add `terminationGracePeriodSeconds` to support graceful shutdown

---

## 12. Security Hardening

**Current state**: multiple security risks exist

- Default configuration files contain hardcoded JWT private keys and OAuth2 client secrets
- Database, Redis, and object storage credentials can now be injected through Kubernetes Secrets, but other sensitive configuration should continue to be consolidated
- There is no login failure rate limiting, so brute-force login attempts are possible
- `multiStatements=true` in the MySQL DSN allows multiple statements and increases SQL injection risk

**Recommendations**:

- [x] Support injecting database, Redis, and object storage credentials through Kubernetes Secrets and overriding YAML configuration through environment variables
- [ ] Add login failure rate limiting, for example 5 attempts per minute per IP
- [ ] Remove `multiStatements=true` from MySQL DSNs and use single-statement execution
- [ ] Add CSRF protection where applicable
- [ ] Rotate JWT signing keys periodically

---

## 13. Code Architecture Optimization

**Current state**:

- `pkg/dal/dal.go` no longer exposes a global `DB *gorm.DB`; database connections are injected through dig. The gorm/gen `query.Q` default query entrypoint is still retained
- The project had commented-out code blocks, such as old workqueue and web-related comments in `server.go`
- `config.yaml` defaults to `sqlite3`, which is intended for quick single-node startup, not high-load production. Production deployments should explicitly switch to MySQL/PostgreSQL plus Redis
- Audit is now implemented as Gin middleware that records request context consistently, instead of writing audit logs by hand in services and repositories
- `pkg/dal/repository/registry.RepositoryRepository.Create` still mixes auto-create namespace/repository, namespace member, webhook, and other business orchestration into DAL responsibilities

**Recommendations**:

- [x] Remove the global DB variable and inject database dependencies through DI
- [x] Remove commented-out dead code
- [x] Move audit to Gin middleware and record user information, operation path, HTTP method, and other request context there instead of hand-writing audit logic in services and repositories
- [ ] Split auto-create repository logic: DAL repositories should only handle repository table CRUD, while auto-create namespace/repository, namespace member, webhook, and other orchestration should move to the service layer

---

## 14. Unit Test Coverage Priorities

**Current state**: CI package coverage report for the PostgreSQL application suite and multi-database DAL suite

- Coverage is uneven across the most important runtime paths. Several protocol, service, and handler packages remain below 10% coverage
- The database repository matrix already runs across PostgreSQL, MySQL, SQLite, and Turso, so repository coverage should be improved selectively where behavior is complex
- Handler tests should focus on request binding, authorization decisions, route registration, service calls, and error mapping instead of trying to cover all service behavior through HTTP tests
- Service tests should be prioritized for state transitions, transactional behavior, queue messages, storage side effects, and OCI protocol edge cases

**P0 recommendations**:

- [ ] Add focused tests for `pkg/service/distribution/manifest`, especially manifest push/pull flows, digest validation, tag handling, delete behavior, and error mapping
- [ ] Add tests for `pkg/service/distribution/blob` and `pkg/service/distribution/upload`, covering blob reads, uploads, mounts, cleanup, and storage error handling
- [ ] Add tests for `pkg/server/handlers/distribution/base`, `blob`, `manifest`, and `upload`, covering HTTP methods, headers, status codes, range behavior, and distribution-spec errors
- [ ] Add service-layer tests for `pkg/service/repositories`, `pkg/service/namespaces`, and `pkg/service/users`, covering permissions, quotas, member changes, login flows, and state updates
- [ ] Add tests for `pkg/service/builders`, `pkg/service/coderepos`, and `pkg/service/daemons`, with emphasis on asynchronous task creation, queue payloads, and status transitions

**P1 recommendations**:

- [ ] Add handler tests for `pkg/server/handlers/webhooks`, covering request validation, authorization, service calls, and error responses
- [ ] Add service tests for `pkg/service/webhooks`, covering URL validation, quota checks, event enqueueing, log deletion, and log resend behavior
- [ ] Add targeted handler tests for builders, code repositories, daemons, users, repositories, namespaces, and tags where coverage is below 50%
- [ ] Add repository tests for `pkg/dal/repository/namespace` and `pkg/dal/repository/registry`, focusing on complex queries, soft-delete behavior, count updates, and size reconciliation

**P2 recommendations**:

- [ ] Add integration-style tests for `pkg/background/buildrunner/docker`, `pkg/background/daemon/gc`, and `pkg/background/daemon/scan/vulngrype` only after the service-layer contracts are covered
- [ ] Add targeted tests for `pkg/infra/workq`, `pkg/infra/workq/inmemory`, `pkg/infra/workq/redis`, `pkg/storage`, and `pkg/telemetry` when their adapters change
- [ ] Keep analytics, validators, systems, tokens, and password tests at their current priority unless related code changes increase risk

---

## Priority Recommendations

| Priority | Item                                                        | Impact                                                              |
| -------- | ----------------------------------------------------------- | ------------------------------------------------------------------- |
| P0 done  | Database connection pool configuration                      | Missing limits can exhaust or overload the database under high load |
| P0 done  | Push hot-path `SUM` queries                                 | Serializes concurrent pushes with many artifacts                    |
| P0       | Analytics database counter                                  | Row-lock contention under concurrent pushes                         |
| P1       | Rate limiting and overload protection                       | Prevents traffic spikes from taking down the service                |
| P1 done  | Insufficient cache system                                   | Reduces DB pressure and improves response time                      |
| P1 done  | Outdated Kubernetes deployment configuration                | HPA may fail on newer Kubernetes versions                           |
| P2       | Storage streaming and multipart upload                      | Improves large image push performance and reliability               |
| P2       | Weak observability                                          | Improves troubleshooting and capacity planning                      |
| P2 done  | Fine-grained background-task mutual exclusion               | Avoids duplicate task execution in multi-replica deployments        |
| P3 done  | Distributed lock reliability                                | Protects data consistency in edge cases                             |
| P3 done  | Graceful shutdown improvements                              | Prevents request loss during rolling updates                        |
| P3       | Security hardening                                          | Improves production compliance                                      |
| P4       | Code architecture optimization, including global DB removal | Improves testability and future multi-tenant support                |
| P4       | Audit and repository creation responsibility boundaries     | Reduces cross-layer business logic and improves maintainability     |
