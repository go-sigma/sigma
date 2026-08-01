# High Availability and High Load Optimization TODO

> Architecture analysis based on the current codebase as of 2026-07-13.
> Re-audited on 2026-08-01 against the current codebase and git history.
> Items are grouped by status: **未完成 / 进行中 / 已完成**，按优先级排序。

---

## 未完成

### 1. Missing Rate Limiting and Overload Protection

**Current state**: middleware packages include `authn`, `authz`, `audit`, `bodylimit`, `etag`, `extractor`, `healthz`, and `metrics`

Already completed:

- [x] Configure full `http.Server` timeouts: `ReadTimeout`, `WriteTimeout`, `IdleTimeout`
- [x] Add `MaxBytesReader` request body limits (`http.bodyLimit`, split by regular API / manifest / blob upload)
- [x] Support `storage.redirect` to redirect blob pull traffic to object storage

Remaining (the core of the item):

- [ ] Add rate-limiting middleware by IP, user, or namespace (token bucket or sliding window)
- [ ] Add circuit breakers (e.g. `sony/gobreaker`) to protect DB and Redis calls, with graceful degradation

**Why top priority**: the only remaining item that directly prevents traffic spikes from taking the service down, which is the primary goal of this document.

### 2. Weak Observability

**Current state**: [`pkg/server/middlewares/metrics/metrics.go`](file:///Users/tosone/code/github/sigma/pkg/server/middlewares/metrics/metrics.go) only exposes `http_requests_total` and `http_request_duration_seconds`.

- [ ] Add business-level Prometheus metrics: blob upload/download throughput and size distributions, manifest push/pull latency, GC task duration and cleanup volume, storage usage, workqueue backlog
- [ ] Add daemon and cronjob task metrics for success, failure, and duration
- [ ] Implement log sampling through `slog` with `NewSamplingHandler` or a custom handler
- [ ] Confirm and document the existing `/metrics` endpoint (or add one)
- [ ] Define key alert rules for error rate, P99 latency, storage usage, and queue backlog

**Why high priority**: fully open, and an HA system cannot be operated or capacity-planned without business metrics and alerting; it is the enabler for validating all other work in this document.

### 3. Security Hardening

**Current state**: multiple security risks remain

Completed:

- [x] Support injecting database, Redis, and object storage credentials through Kubernetes Secrets and overriding YAML configuration through environment variables
- [x] Remove `multiStatements=true` from MySQL DSNs ([`pkg/dal/dal.go`](file:///Users/tosone/code/github/sigma/pkg/dal/dal.go)) and use single-statement execution — goose splits migration files per statement, and no business code executes multi-statement SQL

Remaining:

- [ ] Add login failure rate limiting, for example 5 attempts per minute per IP (also complements item 1)

**Why high priority**: login brute-force protection is cheap to add and closes an active attack surface; removing `multiStatements=true` reduces SQL injection risk.

---

## 进行中

### 4. Missing Fine-Grained Distributed Mutual Exclusion for Some Background Tasks

**Result**:

- Workqueue uses channels in single-instance mode and Redis MQ in multi-instance mode; consumers never claim the same message twice
- GC uses rule-level locks, builder uses builder-ID-level locks, analytics flush uses hour-level locks; audit, size, and vulnerability tasks use their own logical task keys
- Lock acquisition failure skips the current task quickly; builder also uses DB conditional updates to claim expired tasks; vulnerability tasks recover stale `Doing` tasks after a 6-hour grace period
- Multi-replica Helm deployments use Redis workqueue + Redis locker together; task operations stay idempotent with DB atomic status transitions as the final fallback

**Completed items**:

- [x] Add Redis-based fine-grained distributed locks only where task deduplication is required
- [x] Name lock keys by task type and resource ID, avoiding daemon-level global locks
- [x] Skip the current task quickly when lock acquisition fails and continue processing the next task
- [x] Keep task operations idempotent and use database atomic status transitions as the final fallback where needed

---

## 已完成

### 5. Database Connection Pool Configuration

**Result**: [`pkg/dal/dal.go`](file:///Users/tosone/code/github/sigma/pkg/dal/dal.go)

- MySQL and PostgreSQL now configure `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, and `SetConnMaxIdleTime` instead of using unlimited driver defaults
- All pool parameters are exposed in `config.yaml` so each environment can tune them

### 6. Full Aggregation Queries on the Push Hot Path

**Result**: [`pkg/dal/repository/registry/artifact.go`](file:///Users/tosone/code/github/sigma/pkg/dal/repository/registry/artifact.go)

- Repository/namespace size is updated atomically inside the push transaction with `UPDATE repository SET size = size + ?` (O(1) instead of O(N))
- Quota checks use the incremental counters instead of real-time `SELECT SUM(...)`

### 7. Database Hot Spots in the Analytics Backend

**Result**: [`pkg/infra/counter/`](file:///Users/tosone/code/github/sigma/pkg/infra/counter/), [`pkg/background/daemon/analytics/flush.go`](file:///Users/tosone/code/github/sigma/pkg/background/daemon/analytics/flush.go), [`pkg/service/analytics/analytics.go`](file:///Users/tosone/code/github/sigma/pkg/service/analytics/analytics.go)

- The `database` counter backend has been removed; only `inmemory` and `redis` are registered
- `counterBackend` defaults to `inmemory` in `conf/config.yaml`; the inmemory counter aggregates by hour with mutex-protected maps and flushes dirty hours to the DB in `Flush`
- Inmemory and redis backends both flush during graceful shutdown via `graceful.RunAtShutdown`
- Production deployments can switch to `redis` for shared multi-replica aggregation

**Follow-ups (small)**:

- [ ] Restore the current-hour counters from the database on startup so a rolling restart does not lose the in-flight hour
- [ ] Wire the `analytics.reconcileCron` config value to a reconcile job (defined in config but not yet scheduled)

### 8. Insufficient Cache System

**Result**: [`pkg/infra/cache/`](file:///Users/tosone/code/github/sigma/pkg/infra/cache/), [`pkg/dal/repository/namespace/namespace_cache.go`](file:///Users/tosone/code/github/sigma/pkg/dal/repository/namespace/namespace_cache.go), [`pkg/dal/repository/registry/repository_cache.go`](file:///Users/tosone/code/github/sigma/pkg/dal/repository/registry/repository_cache.go), [`pkg/app/cache_prewarm.go`](file:///Users/tosone/code/github/sigma/pkg/app/cache_prewarm.go)

- `Get` merges concurrent fetches for the same missed key with `singleflight`; in-memory and Redis caches both support TTL
- Short-lived negative caching prevents deterministic misses (e.g. `gorm.ErrRecordNotFound`) from hitting the DB
- `cache_requests_total` and `cache_fetch_duration_seconds` expose low-cardinality metrics by backend, prefix, and result
- Namespace/repository metadata is cached through DI-injected DAL cached decorators; transaction-scoped `New...Repository(tx)` returns the raw implementation to avoid caching uncommitted data
- Create/update/delete/size paths invalidate cache after successful writes; startup prewarming covers recently updated namespaces/repositories up to `cache.prewarm.limit`
- Multi-replica production deployments should use Redis cache so invalidation is shared; single-node dev can use inmemory

### 9. Distributed Lock Reliability

**Result**: [`pkg/infra/lock/`](file:///Users/tosone/code/github/sigma/pkg/infra/lock/)

- Redis locks use `github.com/go-redsync/redsync/v4` with a renewal watchdog on `AcquireWithRenew`
- Lock retry uses exponential backoff through the redsync retry delay function
- Metrics cover lock wait time, hold time, and renewal results
- Multi-node Redlock/Raft quorum was evaluated and intentionally not introduced; re-evaluate only if stronger multi-Redis consistency is required

### 10. Missing Streaming and Multipart Uploads in Storage

**Result**: [`pkg/storage/storage.go`](file:///Users/tosone/code/github/sigma/pkg/storage/storage.go), drivers in [`pkg/storage/`](file:///Users/tosone/code/github/sigma/pkg/storage/), [`pkg/service/distribution/upload/upload.go`](file:///Users/tosone/code/github/sigma/pkg/service/distribution/upload/upload.go)

- The `StorageDriver` interface now exposes a multipart abstraction: `CreateUploadID`, `UploadPart`, `CommitUpload`, `AbortUpload`
- All four drivers (filesystem, S3, COS, OSS) implement multipart upload; the blob upload path uses `UploadPart`/`CommitUpload` with part-number tracking
- Multipart copy constants (`MultipartCopyThresholdSize=64MB`, chunk size, max concurrency) support large object copies

**Follow-ups (small)**:

- [ ] Expose upload progress tracking through a `StreamingWriter`/progress callback API (OCI `PATCH` already provides resumable semantics)

### 11. Incomplete Graceful Shutdown and Startup Checks

**Result**: [`pkg/graceful/graceful.go`](file:///Users/tosone/code/github/sigma/pkg/graceful/graceful.go), [`pkg/app/readiness.go`](file:///Users/tosone/code/github/sigma/pkg/app/readiness.go), [`pkg/app/drain.go`](file:///Users/tosone/code/github/sigma/pkg/app/drain.go)

- One shared 30-second `context.WithTimeout` budget for both `httpServer.Shutdown(ctx)` and `graceful.Shutdown(ctx)`
- `Shutdown(ctx)` is idempotent (`sync.Once`), panic-safe (`defer recover()` with structured slog fields), and returns `context.DeadlineExceeded` on timeout
- `/readyz` checks DB ping, Redis ping, and storage `Ping` (implemented by all four drivers via probe object write/delete); startup validation retries dependencies with 2-second backoff up to 60 seconds
- Pre-stop draining: `SetDraining(true)` makes `/readyz` return 503, waits `drainDelay` (default 5s), then shuts down; manifests add `lifecycle.preStop` as a fallback and `terminationGracePeriodSeconds: 45`
- Tests follow project conventions (`require`, `t.Context()`, idempotency and panic-recovery cases)

### 12. Outdated Kubernetes Deployment Configuration

**Result**: [`deploy/sigma/templates/`](file:///Users/tosone/code/github/sigma/deploy/sigma/templates/)

- HPA upgraded to `autoscaling/v2`; PodDisruptionBudget added
- Default requests/limits configured for server, worker, distribution, and web
- Go components use `/readyz` for readinessProbe (dependency checks) and `/healthz` for livenessProbe (process liveness); `terminationGracePeriodSeconds` added

### 13. Code Architecture Optimization

**Result**: [`pkg/dal/dal.go`](file:///Users/tosone/code/github/sigma/pkg/dal/dal.go), [`pkg/service/repositories/repositories.go`](file:///Users/tosone/code/github/sigma/pkg/service/repositories/repositories.go), [`pkg/service/distribution/manifest/manifest.go`](file:///Users/tosone/code/github/sigma/pkg/service/distribution/manifest/manifest.go)

- Global `DB *gorm.DB` removed; database connections are injected through dig (`pkg/dal/dig.go`)
- Commented-out dead code removed
- Audit moved to Gin middleware recording request context (user, path, method) instead of hand-written audit calls in services/repositories
- Auto-create orchestration (namespace/repository, namespace member, webhook) moved from the registry DAL into the repository and manifest services (commit `29b101e9`); the DAL repository now only handles repository table CRUD

### 14. Unit Test Coverage Priorities

**Result**: coverage added across the runtime-critical paths:

- ✅ P0: `pkg/service/distribution/manifest` — push/pull flows, digest validation, tag handling, delete, error mapping, referrers (commit `a7abbb63`)
- ✅ P0: `pkg/service/distribution/blob` / `upload` — blob reads, uploads, mounts, cleanup, storage errors
- ✅ P0: `pkg/server/handlers/distribution/base`, `blob`, `manifest`, `upload` — HTTP methods, headers, status codes, range behavior, distribution-spec errors (commit `72107f6f`)
- ✅ P0: `pkg/service/repositories`, `namespaces`, `users` — permissions, quotas, member changes, login flows, state updates
- ✅ P0: `pkg/service/builders`, `coderepos`, `daemons` — task creation, queue payloads, status transitions
- ✅ P1: `pkg/server/handlers/webhooks` + `pkg/service/webhooks` — validation, authorization, event enqueueing, log deletion/resend
- ✅ P1: `pkg/dal/repository/namespace` and `registry` — complex queries, soft delete, count updates, size reconciliation
- ✅ P2: `pkg/background/daemon/gc`, `pkg/background/daemon/scan/vulngrype`, `pkg/infra/workq` (+redis) — background executor and queue adapters

**Leftovers**: keep `analytics`, `validators`, `systems`, `tokens`, `password`, `pkg/storage`, and `pkg/telemetry` tests current when their adapters change; extend integration-style tests for `pkg/background/buildrunner/docker` and any new storage driver behavior.
