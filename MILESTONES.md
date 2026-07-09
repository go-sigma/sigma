# Sigma Milestones

This document tracks the next two major milestones for Sigma: `1.4.0` and `2.0.0`.

`1.4.0` focuses on turning the existing work into a reliable production deployment experience. `2.0.0` focuses on larger architecture changes and intentionally breaking improvements.

---

## 1.4.0

### Positioning

`1.4.0` is an incremental release for production deployment stability. The goal is to make Sigma easier to install, upgrade, and run in common production setups such as Kubernetes, multiple replicas, object storage, Redis, and PostgreSQL or MySQL.

This release should avoid breaking API or configuration changes where possible. Any required configuration change must have clear defaults and migration notes.

### Goals

- Improve the usability and production defaults of the Helm chart
- Replace the bundled object storage from SeaweedFS to rs3
- Keep publishing the standalone `sigma-builder` image and make Helm reference it by default
- Consolidate Secrets, ConfigMaps, probes, PDBs, HPAs, and resource defaults
- Complete basic high-load optimizations to reduce database and cache hot spots
- Remove deprecated code and stale commented-out code to reduce maintenance cost

### Scope

- Helm chart
  - Use rs3 as the default bundled object storage
  - Inject rs3 credentials through Kubernetes Secrets
  - Initialize the rs3 bucket through the data directory layout
  - Configure reasonable resources, PDBs, HPAs, readiness probes, liveness probes, and startup probes for server, worker, distribution, and web components
  - Do not import the builder image through a post job by default, so installation does not depend on an image copy task

- Image build
  - Do not embed a builder archive into the main image, avoiding larger images and a more complex build chain
  - Keep publishing the standalone `sigma-builder` image
  - Make Helm use a configurable external builder image by default, for example `ghcr.io/go-sigma/sigma-builder:<tag>`
  - Document how users can import `sigma-builder` into their target registry for offline or internal-network deployments
  - Keep offline or internal-network import as a documentation-driven registry operation, not as a built-in Sigma tool command

- High-load foundation
  - Make MySQL and PostgreSQL connection pool parameters configurable
  - Use incremental counters and reconciliation for repository and namespace sizes, avoiding full `SUM` queries on the push hot path
  - Support singleflight, TTL, negative caching, active invalidation, and prewarming for namespace and repository metadata cache
  - Support Redis workqueue and Redis locker for background-task deduplication in multi-replica deployments

- Operations and security
  - Inject database, Redis, and object storage credentials through Secrets
  - Make `/readyz` cover DB, Redis, and storage dependency checks
  - Make graceful shutdown and pre-stop draining configurable
  - Remove commented-out dead code and obsolete deployment templates

### Out of Scope

- No API v2 design
- No removal of existing core configuration fields
- No rewrite of the storage abstraction
- No new external control plane or operator
- No hard requirement for all users to migrate to Redis cache, workqueue, or locker, although production documentation should recommend Redis-backed modes

### Acceptance Criteria

- The Helm chart installs successfully with default values, and server, worker, distribution, web, PostgreSQL, Redis, and rs3 all start normally
- The Kubernetes builder uses the independently published `sigma-builder` image by default
- `helm template` and `helm lint` pass
- Local Docker builds and GitHub Actions builds continue to publish both `sigma` and `sigma-builder` images independently
- Key sqlite3, MySQL, and PostgreSQL tests pass
- Documentation covers bundled rs3, external S3, builder image configuration, manual builder import, Secret injection, and recommended production settings

### Release Checklist

- Update `deploy/sigma/Chart.yaml` `appVersion`
- Update configuration documentation and Helm values documentation
- Update the changelog
- Document data migration considerations from bundled SeaweedFS to rs3
- State clearly that the builder image is published as a standalone image by default, and offline import is a manual flow

---

## 2.0.0

### Positioning

`2.0.0` is an architecture upgrade release. The goal is to provide clearer boundaries, stronger extensibility, and a more stable long-term foundation for larger registry deployments.

This release may introduce breaking changes, but every breaking change must include migration guidance and a compatible replacement path.

### Goals

- Reshape selected DAL, service, storage, and analytics boundaries
- Build a more complete operations foundation with observability, rate limiting, circuit breaking, and capacity controls
- Support streaming and multipart uploads for large blobs
- Remove default implementations or compatibility paths that do not fit high-load production use
- Prepare the foundation for future API, plugin, operator, and multi-tenant improvements

### Scope

- Storage architecture
  - Support multipart upload for S3, COS, and OSS drivers
  - Automatically switch large blobs to a multipart upload strategy
  - Introduce a `StreamingWriter` or `MultipartUploader` abstraction
  - Evaluate resumable uploads and upload progress tracking
  - Clarify boundaries between redirect, proxy, cache, and backend object storage

- Analytics and counters
  - Remove the database counter backend
  - Keep the inmemory and Redis counter backends
  - Recommend the Redis backend for production, with batched flush and recovery behavior
  - Unify the counting semantics for namespaces, repositories, artifacts, and tags

- Code architecture
  - Split the auto-create repository flow
  - Keep DAL repositories responsible only for table-level CRUD and queries
  - Move namespace, repository, member, and webhook orchestration to the service layer
  - Reduce direct dependency on the default global `query.Q` entrypoint and progressively standardize query context through DI

- Observability
  - Add blob upload/download throughput metrics, size distributions, and manifest push/pull latency metrics
  - Add metrics for GC, scan, builder, webhook, analytics, and other background tasks
  - Add workqueue backlog, retry, and dead-letter metrics
  - Provide recommended Prometheus alert rules
  - Support log sampling to avoid log amplification under high load

- Security and resilience
  - Add login failure rate limiting
  - Remove `multiStatements=true` from MySQL DSNs
  - Evaluate where CSRF protection applies
  - Define a JWT key rotation strategy
  - Add rate limiting and circuit breaking to protect DB, Redis, object storage, and external registries

- Deployment and operations
  - Define three deployment modes: single-node, lightweight production, and highly available production
  - Provide a configuration migration guide from 1.x to 2.0
  - Evaluate whether an operator or controller is needed for deployment, upgrades, backup, and migration

### Potential Breaking Changes

- Remove the database analytics counter backend
- Rename or change defaults for selected configuration fields
- Change repository auto-create behavior boundaries
- Change the storage interface
- Change selected Helm values defaults for bundled middleware
- Change builder image publishing and import behavior

### Out of Scope

- No guarantee of full compatibility with all internal 1.x service and DAL behavior
- No guarantee to keep all historical configuration aliases
- No commitment to ship a complete operator in the first 2.0 release
- No attempt to fully normalize every third-party registry and cloud-provider difference behind one abstraction

### Acceptance Criteria

- A migration guide from 1.x to 2.0 is available
- Every breaking change has a clear explanation and replacement path
- Large blob pushes support multipart upload on object storage backends
- Analytics no longer depends on synchronous database hot-row updates under high-concurrency push workloads
- Key-path metrics, alerts, and log sampling are usable for production troubleshooting
- Service and DAL boundaries are clear, and auto-create repository orchestration no longer lives in DAL repositories

### Release Checklist

- Publish `2.0.0-rc.1` for migration validation
- Provide upgrade notes for the Helm chart, configuration files, database migrations, and API behavior
- Provide a minimal migration path for `1.4.x` users
- List removed capabilities and replacement options

---

## Version Relationship

- `1.4.0`: production deployment consolidation, prioritizing the current work while preserving compatibility
- `2.0.0`: architecture evolution, handling long-term changes that may require breaking adjustments

The quality of `1.4.0` directly affects the migration cost of `2.0.0`. Deployment, image build, cache, lock, probe, Secret, and documentation work that can be completed without breaking compatibility should be prioritized for `1.4.0`.
