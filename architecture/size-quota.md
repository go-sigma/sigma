# Size and Count Quotas (quota) Design

> Status: Final (design document)
> Last updated: 2026-11-04
> Scope: `namespace.size_limit`, `repository.size_limit`, `namespace.tag_limit`,
> `repository.tag_limit`, `namespace.repository_limit`, and the global `storage.capacity`.
> This document covers measurement semantics, measurement timing, where limits are enforced,
> delete rollback, and reconciliation only; it does not redesign the UI.
> Implementation status is tracked by the corresponding PRs and code.

---

## 1. Summary

1. **`size` must be measured in blob bytes and deduplicated per blob.**
   Today it uses `artifact.blobs_size` (the sum of the client-declared `descriptor.size`
   values in the manifest), which is neither deduplicated nor server-measured.
2. **Blobs and namespaces/repositories are many-to-many, so `blobs` must not get a
   `namespace_id` column.** `blobs` is globally unique by `digest`, so one digest is one row;
   a single ownership column would turn the relation into one-to-many and break correct
   accounting for cross-namespace sharing and mounts.
3. **Add one join table, `blob_attributions`**, that explicitly records "scope X is charged
   for blob Y":
   - written idempotently on manifest push, with the insert delta summed by `batch_id`;
   - `namespace.size` / `repository.size` still use the existing columns and atomic increments;
   - blob push performs a per-scope best-effort admission check and participates in the global
     hard cap in §5.8.
4. **The join table also removes push × push double counting** (primary-key idempotency).
   The delete × push under-count risk is avoided by "no synchronous subtraction on delete +
   mark dirty + reconcile"; phase 1 does not need row locks.
5. **`tag_limit` and `repository_limit` are not enforced at all today**: the fields are only
   read/written by the API, and the three corresponding error-code helpers
   (`GenDSErrCodeResourceCountQuotaExceed*`) are unused.
6. **The size rollback path is incomplete today** (API delete repository and GC delete
   repository neither subtract nor mark dirty), which leaves the recorded size permanently
   inflated and unrecoverable by reconciliation.
7. **Constraint: do not add a `0002` SQL migration.** The join table and other schema changes
   go directly into `pkg/dal/migrations/{sqlite3,turso,mysql,postgresql}/0001_initialize.sql`
   for all four dialects. There is no historical data, so no legacy-DB compatibility or backfill
   is required.
8. **`storage.capacity` becomes a global hard limit**: the total physical blob bytes across all
   namespaces (globally deduplicated) must not exceed the capacity. For example, with
   capacity=100M, two namespaces each with a 100M limit still share a 100M total blob budget.

---

## 2. Current State (Code Facts)

### 2.1 Data Fields

| Table | Fields | Notes |
| --- | --- | --- |
| `namespaces` | `size`, `size_limit`, `size_dirty`, `tag_limit`, `tag_count`, `repository_limit`, `repository_count` | `pkg/dal/models/namespace.go` |
| `repositories` | `size`, `size_limit`, `size_dirty`, `tag_limit`, `tag_count` | `pkg/dal/models/repository.go` |
| `artifacts` | `size` (manifest bytes), `blobs_size` | `pkg/dal/models/artifact.go` |
| `blobs` | `size` (measured at upload), globally unique by `digest` | `pkg/dal/models/blob.go` |

`0` means unlimited; `size_dirty=true` means the row is waiting for reconciliation.

> Background: the current "atomic increments + background reconciliation" implementation came
> from `.trae/documents/atomic-size-counter-reconcile.md`. That change was driven by
> **performance** (replacing `SELECT SUM(...)` on the push hot path with an O(1) `UPDATE`), and
> it simply reused `artifact.blobs_size` without redefining what `size` should measure.
> This document fixes that measurement definition.

### 2.2 Where Size Is Written Today

| Trigger | Code | What is counted |
| --- | --- | --- |
| Normal manifest push | `pkg/service/distribution/manifest/manifest.go:405-411`, `:563-580` | `artifact.blobs_size` |
| Index/manifest list push | `manifest.go:664-681` | child manifest `descriptor.size` (JSON size, not layers) |
| Delete manifest by digest | `manifest.go:923-947` | subtracts `artifact.blobs_size` |
| GC delete artifact | `pkg/background/daemon/gc/gc_artifact.go:158-187` | subtracts `Σ artifact.blobs_size` |
| Blob upload (`PutUpload` / single-shot `PostUpload`) | `pkg/service/distribution/upload/upload.go:130-141`, `:320-326` | **not counted** |
| Proxy cache fill | `pkg/service/distribution/blob/blob.go:177-193` | **not counted** |
| Delete blob (API / GC) | `blob.go:79-104`, `pkg/background/daemon/gc/gc_blob.go:141-181` | **no rollback** |
| API delete repository / namespace | `pkg/service/repositories/repositories.go:221-257`, `pkg/service/namespaces/namespace.go:234-265` | **no rollback, no dirty mark** |
| GC delete empty repository | `pkg/background/daemon/gc/gc_repository.go:122-155` | **no namespace size rollback** |

Where `blobs_size` comes from:

```go
// manifest.go:405-411
var blobsSize int64
digests := make([]string, 0, len(manifest.References()))
for _, reference := range manifest.References() {
    digests = append(digests, reference.Digest.String())
    blobsSize += reference.Size          // value declared by the client in the manifest
}
digests = dedupeDigests(digests)         // dedupe only applies to digests, not to blobsSize
```

### 2.3 Reconciliation

`pkg/background/cronjob/size/size.go` recomputes with `GetNamespaceSize` /
`GetRepositorySize` (`pkg/dal/repository/registry/artifact.go:474-498`), but those functions
also use `SUM(artifacts.blobs_size)` — **the same wrong definition used to repair the wrong
definition**.

`pkg/background/daemon/pushed/artifact.go:106-135` also has a "recompute and overwrite
namespace/repository size after push" path using the same functions. It registers a consumer
for `DaemonArtifactPushed`, but nothing in the repository `Produce`s that event (dead code), so
it does not overwrite the incremental values today. If it is ever enabled, it must be switched
to the new definition.

### 2.4 Count Limits

- `tagRepository.Create` (`pkg/dal/repository/registry/tag.go:88-113`) does a
  find-then-insert/update and performs no limit check (it is also not an atomic upsert;
  concurrent creation of the same new tag hits the unique index).
- The comments at `manifest.go:557` and `:658` say "maybe got exceed tag quota error", but
  nothing downstream returns that error.
- Neither `CreateRepository` (`pkg/service/repositories/repositories.go:74-127`) nor
  `ensureRepository` (`manifest.go:240-312`) checks `namespace.repository_limit`.
- `tag_count` / `repository_count` are maintained by the tag and repository write
  paths, and the list pages read the stored counters directly.

### 2.5 Existing Error Codes

`pkg/server/errcode/distribution.go:257-330` already defines these and they can be reused:

- `GenDSErrCodeResourceSizeQuotaExceedNamespace`
- `GenDSErrCodeResourceSizeQuotaExceedRepository`
- `GenDSErrCodeResourceCountQuotaExceedRepository` (unused)
- `GenDSErrCodeResourceCountQuotaExceedNamespaceRepository` (unused)
- `GenDSErrCodeResourceCountQuotaExceedNamespaceTag` (unused)

---

## 3. Problem List

### P0 (quota is wrong or effectively absent)

1. **The same blob is counted multiple times**
   - Multiple images share a base layer: each artifact counts it separately.
   - The same digest appears twice in one manifest: `blobsSize` does not deduplicate.
   - An index's `blobs_size` is the JSON size of its child manifests, which is inconsistent
     with the layer-based definition used by the child artifacts.
2. **It uses the client-declared `descriptor.size`, not the server-measured value**
   - A client can under-report size in the manifest and bypass the quota.
   - The correct value is `blobs.size`, measured from the actual bytes at upload time.
3. **Blob upload has no quota constraint at all**
   - Uploading blobs without pushing a manifest can consume unbounded disk; `namespace.size`
     stays 0.
   - When a manifest is rejected by the quota, the blobs and the manifest object are already
     stored, becoming orphans (only GC can reclaim them).
4. **`tag_limit` / `repository_limit` are not implemented** (see 2.4).
5. **Delete paths do not roll back or reconcile**
   - API delete repository: namespace size is not reduced and `size_dirty` is not set, so
     reconciliation never repairs it.
   - GC delete repository: same.
   - Delete tag: does not affect size (the artifact remains), which is expected.
   - Delete blob: cleaned up through the join table under the new model; see 5.4.

### P1 (correctness / concurrency / operations)

1. **Concurrent measurement races**
   - push × push (two artifacts sharing a blob): fully solved by the join table's primary-key
     idempotency.
   - delete × push: may under-count; requires row locks or "no synchronous subtraction on
     delete" (see 5.6).
2. **Wrong reconciliation definition**: `SUM(artifacts.blobs_size)` cannot repair the duplicate
   counting above.
3. **`limit` update logic is wrong (rule decided: only increase)**
   - `namespace.UpdateNamespace` (`pkg/service/namespaces/namespace.go:182-184`) uses
     `old > new` and treats `0` (unlimited) as the smallest value: `100 → 0` (switch to
     unlimited) is rejected, and `0 → 100` (switch from unlimited to a finite limit) is
     allowed. Both cases are inverted; the correct rule is in 5.7.1.
   - `repository.UpdateRepository` has no validation at all; add the same "only increase" rule.
4. **Soft-deleted artifacts leave blob associations behind**
   - `artifactRepository.DeleteByID` is a soft delete; `artifact_blobs` rows are not cleaned up.
   - The interface comment on `BlobRepository.FindAssociateWithArtifact` (`blob.go:178-203`)
     says "still associated with live artifacts", but the implementation does not filter
     `artifacts.deleted_at`, so a blob that was ever referenced can never be GC'd (physical
     leak).
5. **Manifest is stored before the quota is checked**
    - `manifest.go:399` calls `Storage.Upload` first; `:565` may later fail on quota, leaving
      the manifest object behind.
6. **Concurrent writes of the same digest can return 500**
    - `RepoBlob.Create` inserts directly and relies on the `(digest, deleted_at)` unique index;
      a concurrent second request gets a unique-key violation instead of an idempotent success.
    - The same applies to `artifactRepository.Create`: `isNewArtifact` is decided outside the
      transaction, so a concurrent push of the same digest hits the
      `(repository_id, digest, deleted_at)` unique index and returns `UNKNOWN` instead of
      succeeding idempotently.

### P2 (can be deferred)

1. **A missing referenced blob does not return `BLOB_UNKNOWN`**
    - `FindByDigests` (`blob.go:210-213`) silently returns the subset that exists, while size is
      still counted from the declared values.
2. `tag_count` / `repository_count` are maintained counters; quota checks and list
    pages read them directly instead of running a live `COUNT(*)`.
3. `storage.capacity` (`pkg/config/configuration.go:570-573`) is only displayed on the home
    page; it is not enforced.

---

## 4. Target Semantics

Definitions (`0` = unlimited):

- `repository.size` = the sum of `blob.size` over the **distinct blobs referenced by
  non-deleted artifacts in that repository**.
- `namespace.size` = the sum of `blob.size` over the **distinct blobs referenced by non-deleted
  artifacts in any repository of that namespace**.
- When the same blob is referenced by multiple namespaces, **each namespace counts it once**
  (the quota is a "logical usage", not physical occupancy).
- Global physical usage continues to use the home page's `storage.used`
  (`repository.SumBlobSize`, already distinct).
- The manifest JSON itself (`artifact.size`) is not counted toward the size quota (it is tiny;
  can be revisited later).
- An index / manifest list's references are manifests, not blobs, so it **produces no size
  increment**; the child artifacts already count their own layers.
- Per-scope quota violations return 403 `DENIED`, reusing the existing
  `GenDSErrCodeResourceSizeQuotaExceed*`.

Example: two images A and B in the same namespace share a 100 MB base layer and each has a
10 MB unique layer.

- Today (sum of each artifact's `blobs_size`): 100 + 10 + 100 + 10 = 220 MB.
- Target (distinct blobs): 100 + 10 + 10 = 120 MB.
- Tags take no space: adding 10 tags to A keeps its size at 110 MB.

### 4.1 Why a Join Table Instead of a Column on `blobs`

Blobs and namespaces/repositories are **many-to-many**:

- One digest is one row globally (`blobs_unique_with_digest UNIQUE (digest, deleted_at)`).
- It can be referenced by multiple repositories and by multiple namespaces.
- A single `namespace_id` / `repository_id` column on `blobs` would turn the relation into
  one-to-many: the second referrer cannot be recorded, and cross-namespace mounts cannot be
  charged correctly.

Ownership must therefore live in a relation table:

```text
blob_attributions(scope_type, scope_id, blob_id, size, batch_id)
  scope_type = 'namespace' | 'repository'
  primary key (scope_type, scope_id, blob_id)
```

Then:

- `SUM(size) WHERE scope_type = ? AND scope_id = ?` is the deduplicated size of that scope;
- the primary key `(scope_type, scope_id, blob_id)` is naturally idempotent and solves push ×
  push double counting;
- when the last reference in a scope is removed, deleting the row rolls the size back.

> If readability matters more, this can also be split into `namespace_blobs` /
> `repository_blobs`; the logic is identical. This document uses one `blob_attributions` table.

---

## 5. Recommended Design: Join Table + Atomic Increments

### 5.1 Table Structure

One polymorphic join table, added directly to `0001_initialize.sql` (all four dialects,
no `0002`). **The full per-dialect DDL is in 7.1 and can be pasted as-is.**

Design points:

- `size` is denormalized: `blobs.size` is immutable for the life of a blob, so storing it here
  lets `SUM(size)` scan only the join table without joining `blobs`.
- `batch_id`: each push generates a UUIDv7 identifying "which rows this transaction actually
  inserted". Conflicting rows keep their old `batch_id` and are excluded from the increment.
- `blob_attributions_idx_blob`: lets blob GC / deletion clean up attributions by `blob_id`.
- `scope_type` is a string for easier debugging; define Go constants to avoid typos.
- No foreign keys (consistent with the existing schema, which does not use FKs).

### 5.2 Manifest Push Changes

`pkg/service/distribution/manifest/manifest.go`:

1. First verify that every referenced blob exists (otherwise `BLOB_UNKNOWN`), and compute the
   display-only `BlobsSize` from the deduplicated `blob.size` values returned by
   `FindByDigests`.
2. Inside the existing `query.Q.Transaction`, after inserting the artifact / tag / m2m rows, do
   the following for both the repository scope and the namespace scope (**do it on every push,
   not only when `isNewArtifact`**, so it also heals missing attribution rows):

   ```text
   batchID := uuid.NewV7String()

   INSERT INTO blob_attributions (scope_type, scope_id, blob_id, size, batch_id, created_at)
   SELECT ?, ?, b.id, b.size, ?, ?
   FROM blobs b
   WHERE b.id IN (:blob_ids) AND b.deleted_at = 0
   ON CONFLICT (scope_type, scope_id, blob_id) DO NOTHING;
   -- MySQL uses INSERT IGNORE or an equivalent

   delta := SELECT COALESCE(SUM(size), 0)
            FROM blob_attributions
            WHERE scope_type = ? AND scope_id = ? AND batch_id = ?;
   ```

3. Call `IncrementSize(scope, delta)` (existing atomic conditional update,
   `namespace.go:218-243`, `repository.go:322-347`); on failure return 403 and roll back.
   Skip the call when `delta` is 0.
4. Remove the `IncrementSize(..., artifactObj.BlobsSize)` calls (`:563-580`, `:664-681`).
5. Index branch: an index's references are manifests, not blobs, so it produces no attribution
   rows and the increment is naturally 0. Do not count the child manifests' `descriptor.size`
   toward size anymore.
6. Make `recordNamespaceSizeDelta` (analytics) use the same `delta`, so charts and quota use
   the same definition.
7. Move `Storage.Upload(manifestPath, ...)` after the quota check / successful transaction; if
   it stays as-is, at least call `Storage.Delete(manifestPath)` on failure.

Why `batch_id` instead of `RowsAffected`:

- A batched `INSERT ... ON CONFLICT DO NOTHING` only reports a row count, not "the sum of the
  `size` values of the rows inserted this time"; inserting row by row to accumulate would cost
  N round trips.
- `batch_id` makes "rows inserted this time" retrievable with a single `SUM`, and is naturally
  idempotent.
- A concurrent second transaction conflicts and gets `SUM = 0`, so it does not add twice.

### 5.3 Blob Push Admission (Best-Effort)

`pkg/service/distribution/upload/upload.go`:

- `PutUpload`: after the "digest already exists → return success" branch (`:222-234`), use
  `sizeBefore + contentLength` as the new bytes, read the repository / namespace
  `size + size_limit`, and return 403 `GenDSErrCodeResourceSizeQuotaExceed*` if it would
  exceed the limit, before `CommitUpload` (so no bytes are committed).
- Single-shot `PostUpload` (`:97-163`): do the same `Content-Length` pre-check before writing;
  skip when the length is unavailable (best-effort).
- **Unify how `blob.size` is written**: `PutUpload` currently uses
  `sizeBefore + contentLength` (`:320-326`); change it to the measured
  `sizeBefore + sizeUploaded`. Single-shot `PostUpload` already uses `counter.Count()`, keep it.
- This check is only a per-scope early stop, not exact measurement:
  - concurrent uploads can still temporarily exceed the limit (acceptable);
  - a blob uploaded without a manifest does not write `blob_attributions`, is not counted in
    `namespace.size`, and is reclaimed by GC;
  - the exact per-scope enforcement happens at manifest push.
- **Decided: per-scope attribution is written only on manifest push**, not on blob push.
  Therefore orphan blobs are constrained only by the global cap in §5.8 and are not counted in
  namespace/repository size.

### 5.4 Delete / Rollback (Default: Soft Delete + Dirty + Reconciliation)

Phase 1 uniformly uses "no synchronous subtraction on delete":

**After soft-deleting an artifact / repository, only set the corresponding scope's
`size_dirty = true` and immediately trigger one reconciliation for that scope**; the
reconciliation deletes the extra `blob_attributions` rows and overwrites `size` with
`SUM(size)`.

Benefits:

- any concurrent interleaving (delete × push, delete × delete) cannot under-count; at worst the
  quota is temporarily high (the safe direction);
- the delete path needs no locks and no complex "unreachable" query inside the transaction.

| Path | Change |
| --- | --- |
| `DeleteManifest` by digest | Soft-delete the artifact in the transaction; after commit, mark namespace / repository dirty and trigger a scope reconciliation |
| `DeleteManifest` by tag | unchanged (the artifact and its bytes remain) |
| GC artifact | soft-delete, mark dirty, trigger reconciliation (can be batched) |
| API delete repository | delete that repository's attribution rows (`scope_type='repository'`); mark namespace dirty and trigger reconciliation |
| GC delete repository | same as above (currently not handled at all) |
| Delete blob (API / GC) | delete all attribution rows by `blob_id`; also release the global counter in §5.8 |
| Delete namespace | delete that namespace's attribution rows; its repository attribution rows can be left to repository deletion / GC |

> The default reconciliation interval is 6h, so a delete / rollback must **immediately trigger**
> a reconciliation for the scope; otherwise users will see the quota stay occupied for a long
> time. The periodic pass is only a safety net.

#### Future optimization (not in this phase): synchronous exact delete

If "quota is released immediately after delete commits" is required later, add a namespace row
lock to the delete transaction and subtract synchronously with the SQL below (this serializes
with push; see 5.6):

```sql
SELECT ab.blob_id, ab.size
FROM blob_attributions ab
WHERE ab.scope_type = 'namespace'
  AND ab.scope_id = :namespace_id
  AND ab.blob_id IN (:blob_ids)
  AND NOT EXISTS (
    SELECT 1
    FROM artifact_blobs x
    JOIN artifacts a ON a.id = x.artifact_id AND a.deleted_at = 0
    WHERE x.blob_id = ab.blob_id AND a.namespace_id = :namespace_id
  );
```

For the repository variant, replace `a.namespace_id` with `a.repository_id` and `scope_type`
with `repository`.

Execution order: soft-delete artifact → query the unreachable rows and sum them → delete those
rows → `DecrementSize(scope, sum)`. This path must be mutually exclusive with push, otherwise
delete × push can under-count (see 5.6).

> Edge case: deleting a child manifest referenced by an index. `artifact_artifacts` only
> represents the index → child artifact relation and does not participate in `artifact_blobs`
> reachability. If the child manifest may be deleted, its attribution rows are removed and the
> size is released early while the index still references it (a broken index).
> **Decided: `DeleteManifest` must reject deleting an artifact referenced by an index using
> `IsArtifactAssociatedWithArtifact` (implemented but currently unused), returning 409
> `Conflict`.**

### 5.5 Reconciliation

The join table is the direct source of truth for size, but to guard against "an artifact
references a blob yet the attribution row is missing" (historical data, exceptional paths),
reconciliation still repairs based on the "reachable blob set":

1. Process only namespaces / repositories with `size_dirty = true`.
2. **Insert missing rows**: find missing `(scope, blob)` pairs with the reachable query and
   `INSERT ... ON CONFLICT DO NOTHING`.
3. **Delete extra rows**: delete `blob_attributions` rows that are no longer reachable in that
   scope.
4. `size = SELECT SUM(size) FROM blob_attributions WHERE scope_type=? AND scope_id=?`.
5. Clear `size_dirty`.

Reachable queries (background reconciliation only; never on the push hot path):

```sql
-- blobs reachable from a namespace (distinct)
SELECT b.id, b.size
FROM repositories r
JOIN artifacts a ON a.repository_id = r.id AND a.deleted_at = 0
JOIN artifact_blobs ab ON ab.artifact_id = a.id
JOIN blobs b ON b.id = ab.blob_id AND b.deleted_at = 0
WHERE r.namespace_id = :namespace_id AND r.deleted_at = 0
GROUP BY b.id, b.size;

-- blobs reachable from a repository (distinct)
SELECT b.id, b.size
FROM artifacts a
JOIN artifact_blobs ab ON ab.artifact_id = a.id
JOIN blobs b ON b.id = ab.blob_id AND b.deleted_at = 0
WHERE a.repository_id = :repository_id AND a.deleted_at = 0
GROUP BY b.id, b.size;
```

- The batching + distributed lock structure in `pkg/background/cronjob/size/size.go` stays. For each
  dirty scope the runner calls `BlobAttributionRepository.SyncAndSum` (repair rows, then sum); the
  legacy `ArtifactRepository.GetNamespaceSize` / `GetRepositorySize` are no longer used by the
  reconciler.
- Every path that changes the reachable set sets `size_dirty = true`: manifest push / delete,
  GC artifact, API / GC delete repository, delete blob.
- Global capacity reconciliation:
  `quota_usages('global') = SUM(blobs.size) WHERE blobs.deleted_at = 0`; it can share the same
  cronjob.
- Besides the periodic pass, delete / rollback must **immediately trigger one scope
  reconciliation** (for example, push the scope onto a lightweight work queue, or run it
  synchronously); otherwise the quota only recovers at the next 6h cycle.
- Reconciliation is O(number of reachable blobs in the scope) but only runs for dirty rows, so
  it does not slow down push.

### 5.6 Concurrency Correctness

| Race | Result | Needs a fix? |
| --- | --- | --- |
| push × push (two artifacts sharing a blob) | join-table primary key is idempotent, **no double count** | solved |
| delete × push | with the default "no synchronous subtraction", **no under-count**; only matters for synchronous exact delete | solved by default |
| delete × delete | join-table rows are deleted once, **no under-count** | solved |

Why push × push is automatically exact:

- two transactions insert the same `(scope_type, scope_id, blob_id)`, and the primary key allows
  only one to succeed;
- the second transaction's `INSERT ... ON CONFLICT DO NOTHING` skips the conflict, so its
  `batch_id` sum is 0;
- "both transactions add once" cannot happen.

When delete × push can under-count (only with synchronous exact delete):

- the delete transaction decides "this blob is no longer reachable" and prepares to delete the
  attribution row;
- a concurrent push inserts a new artifact referencing the same blob, but its reference is not
  visible in the delete transaction's snapshot;
- the delete transaction deletes the row and calls `DecrementSize`, while the push transaction
  does not add because the row already existed;
- the artifact still references the blob, but size is under-counted.

Handling:

1. **Default: no synchronous subtraction on delete** (decided, simplest).
   Delete only soft-deletes + sets `size_dirty=true` + immediately triggers reconciliation for
   the scope; reconciliation deletes extra attribution rows and overwrites size. No concurrent
   interleaving can under-count; at worst size is temporarily high.
2. **Future optimization (not this phase): synchronous exact delete.** The delete transaction
   must also take a scope row lock so it is mutually exclusive with push:
   lock the `namespaces` row as the first statement in the transaction (`SELECT ... FOR UPDATE`),
   and have push take the same lock. Lock order is fixed: namespace → repository.
   Note that MySQL REPEATABLE READ establishes its snapshot at the first plain SELECT in the
   transaction, so the lock must come before any consistent read. GORM's SQLite dialect silently
   ignores `FOR UPDATE` (`gorm.io/driver/sqlite@v1.6.0/sqlite.go:126`), so either add
   `_txlock=immediate` to the DSN or accept "retry after busy"; Turso uses `BEGIN` with snapshot
   isolation, so treat conflicts as retryable errors.

> Push needs no lock because of primary-key idempotency; phase 1 deletes go through
> reconciliation, so do not introduce row-lock serialization.

### 5.7 Count Limits (tag / repository)

First decide "counters" vs "live COUNT":

- the fields already exist (`tag_count`, `repository_count`) and are now maintained as
  O(1) counters: tag and repository writes update the owning rows in the same transaction,
  and quota checks / list pages read the counters directly.

#### tag limit (repository level + namespace level)

Before creating a tag in `putManifestManifest` / `putManifestIndex` (only when
`isNewTag == true`; retags do not count):

1. Lock the parent row: `SELECT ... FOR UPDATE` (MySQL/PostgreSQL; SQLite/Turso serialize writes
   anyway).
2. `COUNT(*) FROM tags WHERE repository_id = ? AND deleted_at = 0` and compare with
   `repositories.tag_limit`.
3. `COUNT(*)` across the namespace (join `repositories`) and compare with
   `namespaces.tag_limit`.
4. On overflow return `GenDSErrCodeResourceCountQuotaExceedRepository` /
   `GenDSErrCodeResourceCountQuotaExceedNamespaceTag`; roll back the transaction.

#### repository limit

In `CreateRepository` (`pkg/service/repositories/repositories.go:74`) and
`ensureRepository` (`manifest.go:240`, auto-created on first push):
lock the namespace row → `COUNT(*) FROM repositories WHERE namespace_id = ? AND deleted_at = 0`
→ compare with `namespace.repository_limit` → on overflow return
`GenDSErrCodeResourceCountQuotaExceedNamespaceRepository`.

#### If switched to counters

- push: `UPDATE ... SET tag_count = tag_count + 1 WHERE id = ? AND (tag_limit = 0 OR tag_count + 1 <= tag_limit)`
  (`RowsAffected == 0` means overflow), once for the namespace and once for the repository in the
  same transaction.
- delete: `DeleteByName`, `DeleteByArtifactID`, GC tag, GC repository, and API delete repository
  must all decrement; add a count reconciliation job (reuse the `size` job framework).
- Pros: O(1) on the push hot path; cons: many delete paths, easy to miss, requires strict
  reconciliation.

### 5.7.1 `size_limit` Can Only Be Increased, Never Decreased

Semantics: `0` = unlimited (treated as infinity). Rules:

| Old value | New value | Decision |
| --- | --- | --- |
| `0` (unlimited) | non-zero | unlimited → finite: **decrease, reject** |
| non-zero | `0` (unlimited) | finite → unlimited: **increase, allow** |
| non-zero | non-zero and new < old | **decrease, reject** |
| non-zero | non-zero and new >= old | increase / unchanged, allow |
| `0` | `0` | unchanged, allow |

The current `UpdateNamespace` (`pkg/service/namespaces/namespace.go:182-184`) uses `old > new`
and treats `0` as the smallest value, which inverts both `0`-related cases above. It must be
changed to the rules above, and `UpdateRepository`
(`pkg/service/repositories/repositories.go:198-204`) must get the same validation.

Error message: `size_limit can only be increased; 0 means unlimited`.

### 5.8 Global Physical Cap `storage.capacity`

`storage.capacity` (`pkg/config/configuration.go:570-573`) is currently only displayed on the
home page and not enforced. Requirement: **the total physical blob bytes across all namespaces
(globally deduplicated) must not exceed the capacity.** For example, with capacity=100M, two
namespaces each with a 100M limit still share a 100M total blob budget. Orphan blobs have no
separate per-scope limit, but they are physical bytes and still occupy global capacity.

#### Measurement

- `used = SUM(blobs.size) WHERE blobs.deleted_at = 0`: includes orphan blobs and proxy-cached
  blobs.
- Excludes manifest JSON (`artifact.size`), consistent with the per-scope size definition.
- `capacity = 0` means unlimited.

#### Enforcement Points

- **blob commit**: when dedupe misses (a new blob row will really be created), perform an atomic
  reservation before moving the temporary object to its final path; on failure abort the upload
  and return 507.
- dedupe hit: no new bytes, no check.
- **pre-check**: use `Content-Length` at the start of the upload for a non-authoritative check,
  so obvious overflow fails early and fewer bytes are transferred; the commit-time atomic
  reservation remains authoritative.
- **proxy cache fill**: also counts toward global capacity; when full, **reject the proxy fill
  and return 507**.
- manifest push does not check global capacity again (blobs were already counted).

#### Counter

Do not run `SUM(blobs.size)` on every upload (O(N)). Use a single-row counter table
`quota_usages` (the second table, schema goes into 0001):

```sql
CREATE TABLE quota_usages (
  scope_type varchar(16) NOT NULL,
  scope_id   varchar(36) NOT NULL DEFAULT '',
  size       bigint NOT NULL DEFAULT 0,
  updated_at bigint NOT NULL,
  PRIMARY KEY (scope_type, scope_id)
);
INSERT INTO quota_usages (scope_type, scope_id, size, updated_at) VALUES ('global', '', 0, 0);
```

Reserve (atomic conditional UPDATE):

```sql
UPDATE quota_usages
SET size = size + :delta, updated_at = :now
WHERE scope_type = 'global' AND scope_id = ''
  AND (:capacity = 0 OR size + :delta <= :capacity);
-- RowsAffected == 0 -> capacity exceeded
```

Release (when the storage object is actually deleted): `size = MAX(size - :delta, 0)`.
Reconcile: `size = SUM(blobs.size) WHERE deleted_at = 0`.

> Align the release with the actual object deletion: subtract only after GC's
> `deleteBlobObject` successfully deletes the storage object. Note that the current API
> `DeleteBlob` only soft-deletes the DB row and does not delete the storage object, while GC
> only looks for blobs with `deleted_at = 0`; this leaks objects and excludes them from capacity.
> **Fix it in this phase** (delete the object directly, or enqueue a GC deletion task, and
> release the global counter after the object is deleted).

Why not a `global` row in `blob_attributions`: the join table is keyed by "reachable blobs per
scope", and orphan blobs are not in it; global capacity must count physical blobs. The two
definitions differ, so a separate counter table is clearer.

#### Concurrency

- The conditional UPDATE serializes on a single row, so two concurrent uploads cannot both break
  the capacity.
- Reserve before Move; if Move fails, compensate (subtract back). A crash may leave a high
  value, which reconciliation corrects.

#### Errors

Add `GenDSErrCodeStorageCapacityExceed(current, capacity, increase)` returning
**507 `Insufficient Storage`**. Per-namespace/repository size violations still use 403 `DENIED`;
do not mix them.

---

## 6. Alternatives

### Option A: no join table, live distinct join

- Compute the increment on push with the "reachable query + NOT EXISTS" from 5.5, and use the
  same query for reconciliation.
- Pros: no schema change.
- Cons: complex SQL (must `GROUP BY` / `DISTINCT`, otherwise it double counts), and every push
  spans `repositories / artifacts / artifact_blobs / blobs`; concurrent double counting can only
  be mitigated with row locks or reconciliation.
- Suitable only when "no schema changes and very low push volume"; not recommended.

### Option B: add ownership columns to `blobs` (not recommended)

- Add `namespace_id` / `repository_id` to `blobs` (first uploader).
- Pros: O(1), no join, simplest SQL.
- Cons (decisive):
  - blobs and namespaces/repositories are many-to-many; a single column breaks sharing;
  - with mounts / cross-namespace sharing, the consumer is free and the uploader is charged
    forever;
  - deleting an artifact does not release quota (must wait for blob GC);
  - historical blobs need a backfill.
- Even though it is "the simplest SQL", it is not adopted.

---

## 7. Schema Changes and Data Repair

### 7.1 Put All Schema Changes in 0001

- Two new tables in this phase:
  - `blob_attributions`: per-scope blob ownership;
  - `quota_usages`: global physical usage counter (single row `('global','')`).
  Both go directly into
  `pkg/dal/migrations/{sqlite3,turso,mysql,postgresql}/0001_initialize.sql`; the four files stay
  equivalent and **no `0002_*.sql` is added**.
  Do **not** add the `artifacts(namespace_id, deleted_at)` index in this phase; the existing
  `repositories` / `artifacts` indexes are sufficient. Add it to 0001 later if benchmarks need it.
- Follow the existing 0001 style for dialect differences: backticks for MySQL/SQLite/Turso,
  double quotes for PostgreSQL; boolean columns are `tinyint(1)` / `integer` / `boolean`; keep
  the existing naming style for keys and indexes.
- **There is no historical data**, so there is no "old databases will not rerun 0001"
  compatibility concern; a fresh database simply runs 0001.
  The initial `('global','')` row of `quota_usages` is inserted in 0001 as well.

#### 7.1.1 DDL (paste directly into 0001)

Place these after the `artifact_blobs` creation statements (Up section), and add the matching
`DROP TABLE IF EXISTS` to the Down section.

**MySQL** (`pkg/dal/migrations/mysql/0001_initialize.sql`):

```sql
CREATE TABLE IF NOT EXISTS `blob_attributions` (
  `scope_type` varchar(16) NOT NULL,
  `scope_id` varchar(36) NOT NULL,
  `blob_id` varchar(36) NOT NULL,
  `size` bigint NOT NULL DEFAULT 0,
  `batch_id` varchar(36) NOT NULL DEFAULT '',
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`scope_type`, `scope_id`, `blob_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE INDEX `blob_attributions_idx_batch` ON `blob_attributions` (`batch_id`);

CREATE INDEX `blob_attributions_idx_blob` ON `blob_attributions` (`blob_id`);
```

**PostgreSQL** (`pkg/dal/migrations/postgresql/0001_initialize.sql`):

```sql
CREATE TABLE IF NOT EXISTS "blob_attributions" (
  "scope_type" varchar(16) NOT NULL,
  "scope_id" varchar(36) NOT NULL,
  "blob_id" varchar(36) NOT NULL,
  "size" bigint NOT NULL DEFAULT 0,
  "batch_id" varchar(36) NOT NULL DEFAULT '',
  "created_at" bigint NOT NULL,
  PRIMARY KEY ("scope_type", "scope_id", "blob_id")
);

CREATE INDEX "blob_attributions_idx_batch" ON "blob_attributions" ("batch_id");

CREATE INDEX "blob_attributions_idx_blob" ON "blob_attributions" ("blob_id");
```

**SQLite3** (`pkg/dal/migrations/sqlite3/0001_initialize.sql`):

```sql
CREATE TABLE IF NOT EXISTS `blob_attributions` (
  `scope_type` varchar(16) NOT NULL,
  `scope_id` varchar(36) NOT NULL,
  `blob_id` varchar(36) NOT NULL,
  `size` integer NOT NULL DEFAULT 0,
  `batch_id` varchar(36) NOT NULL DEFAULT '',
  `created_at` integer NOT NULL,
  PRIMARY KEY (`scope_type`, `scope_id`, `blob_id`)
);

CREATE INDEX `blob_attributions_idx_batch` ON `blob_attributions` (`batch_id`);

CREATE INDEX `blob_attributions_idx_blob` ON `blob_attributions` (`blob_id`);
```

**Turso** (`pkg/dal/migrations/turso/0001_initialize.sql`, identical to SQLite3):

```sql
CREATE TABLE IF NOT EXISTS `blob_attributions` (
  `scope_type` varchar(16) NOT NULL,
  `scope_id` varchar(36) NOT NULL,
  `blob_id` varchar(36) NOT NULL,
  `size` integer NOT NULL DEFAULT 0,
  `batch_id` varchar(36) NOT NULL DEFAULT '',
  `created_at` integer NOT NULL,
  PRIMARY KEY (`scope_type`, `scope_id`, `blob_id`)
);

CREATE INDEX `blob_attributions_idx_batch` ON `blob_attributions` (`batch_id`);

CREATE INDEX `blob_attributions_idx_blob` ON `blob_attributions` (`blob_id`);
```

**`quota_usages` (global physical usage counter, single row `('global','')`)**

MySQL:

```sql
CREATE TABLE IF NOT EXISTS `quota_usages` (
  `scope_type` varchar(16) NOT NULL,
  `scope_id` varchar(36) NOT NULL DEFAULT '',
  `size` bigint NOT NULL DEFAULT 0,
  `updated_at` bigint NOT NULL,
  PRIMARY KEY (`scope_type`, `scope_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

INSERT INTO `quota_usages` (`scope_type`, `scope_id`, `size`, `updated_at`)
VALUES ('global', '', 0, 0);
```

PostgreSQL:

```sql
CREATE TABLE IF NOT EXISTS "quota_usages" (
  "scope_type" varchar(16) NOT NULL,
  "scope_id" varchar(36) NOT NULL DEFAULT '',
  "size" bigint NOT NULL DEFAULT 0,
  "updated_at" bigint NOT NULL,
  PRIMARY KEY ("scope_type", "scope_id")
);

INSERT INTO "quota_usages" ("scope_type", "scope_id", "size", "updated_at")
VALUES ('global', '', 0, 0);
```

SQLite3 / Turso:

```sql
CREATE TABLE IF NOT EXISTS `quota_usages` (
  `scope_type` varchar(16) NOT NULL,
  `scope_id` varchar(36) NOT NULL DEFAULT '',
  `size` integer NOT NULL DEFAULT 0,
  `updated_at` integer NOT NULL,
  PRIMARY KEY (`scope_type`, `scope_id`)
);

INSERT INTO `quota_usages` (`scope_type`, `scope_id`, `size`, `updated_at`)
VALUES ('global', '', 0, 0);
```

**Down section** (add to all four, quoting per dialect):

```sql
DROP TABLE IF EXISTS `quota_usages`;
DROP TABLE IF EXISTS `blob_attributions`;
-- PostgreSQL uses double quotes: DROP TABLE IF EXISTS "quota_usages"; ...
```

Companion GORM model (`pkg/dal/models/blob_attribution.go`):

```go
// BlobAttribution records which blob is charged to which scope.
type BlobAttribution struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`

	ScopeType string `gorm:"primaryKey;size:16"`  // namespace / repository
	ScopeID   string `gorm:"primaryKey;size:36"`
	BlobID    string `gorm:"primaryKey;size:36"`
	Size      int64  `gorm:"default:0"`
	BatchID   string `gorm:"size:36;default:''"` // index is created in 0001; not relying on AutoMigrate
}
```

Global counter model (`pkg/dal/models/quota_usage.go`):

```go
// QuotaUsage is the aggregated usage per scope; the global row uses scope_id = ''.
type QuotaUsage struct {
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`

	ScopeType string `gorm:"primaryKey;size:16"` // global / namespace / repository (only global is used for now)
	ScopeID   string `gorm:"primaryKey;size:36"`
	Size      int64  `gorm:"default:0"`
}
```

After adding the models run `make gormgen` to regenerate `pkg/dal/query` (do not hand-edit
`*.gen.go`), then implement the CRUD in `pkg/dal/repository/registry/`.

### 7.2 Data Repair (No Historical Data, So No Backfill)

There is no historical data, so **no backfill is needed**:

- a fresh database starts at 0, and blobs write attributions at manifest push;
- the global counter `quota_usages('global')` starts at 0 and changes when blobs are created /
  deleted.

If drift appears later, reconciliation already repairs it (insert missing, delete extra,
recompute `size` and global `used`). A manual `sigma tools quota-recompute --scope=...` entry
point can be kept, but it is not required in this phase.

---

## 8. Test Plan

### DAL (`pkg/dal/repository/registry`)

- Two artifacts sharing a layer: `SUM(blob_attributions.size)` counts it once for the
  repository / namespace.
- Duplicate digests in one manifest: only one attribution row.
- Cross-repository (same namespace): one namespace row and one row per repository.
- Cross-namespace: one attribution row per namespace (logical usage).
- Soft-delete artifact + reconciliation: the attribution rows are deleted and size drops.
- Index referencing child manifests: no extra attribution rows.
- Concurrent inserts of the same `(scope, blob)`: only one transaction's `batch_id` sums the
  row; the other is 0.
- Reconciliation run twice: idempotent, no double counting.

### Service (`pkg/service/distribution/manifest`)

- Over-limit push returns 403 `DENIED` and rolls back artifact/tag/attribution rows/counters.
- Under-limit push succeeds; re-pushing the same digest (`isNewArtifact=false`) does not
  double-count.
- Retag (same tag, different digest) does not count a new tag; the new digest's attribution rows
  are added normally.
- tag limit hit (repository level / namespace level).
- repository limit hit (API creation + auto-creation on first push).
- Missing referenced blob returns `BLOB_UNKNOWN`.

### Blob Upload

- Dedupe hit (digest exists) succeeds directly, writes no attribution, and does not increase
  global `used`.
- A new blob increases global `used`; deleting the blob decreases it.
- A per-namespace over-limit upload is rejected and not persisted (or rolled back).
- A global capacity over-limit upload is rejected and the counter is not over-subtracted.
- Concurrent uploads of the same digest succeed idempotently (no 500).

### Global Capacity

- `capacity=100M`, two namespaces with `size_limit=100M` each: total blobs above 100M are
  rejected.
- Dedupe hit (second upload of the same digest) does not consume capacity twice.
- Move fails after reservation: `used` is rolled back; a crash leaves a high value that
  reconciliation corrects.

### Reconciliation

- Create drift manually (delete an attribution row, or corrupt `size`), mark dirty, run
  `size.go`, and assert it is repaired.
- After deleting a repository, namespace size drops correctly (or dirty is set and
  reconciliation fixes it).
- After reconciliation, `size` matches the reachable distinct query; global `used` matches
  `SUM(blobs.size)`.

### e2e

- Push two images sharing a base layer; namespace size ≈ the deduplicated size.
- Set a small `size_limit`; an over-limit push returns `DENIED`.
- Set `tag_limit=1`; pushing a second tag returns `DENIED`.
- Lower the global `capacity`; new blob pushes are rejected.

---

## 9. Delivery Plan

| PR | Content | Risk |
| --- | --- | --- |
| PR1 | `blob_attributions` table (in 0001) + idempotent push writes + batch increment + reconciliation definition + dirty flag | Medium, touches the core push path |
| PR2 | Per-scope blob push admission + global `storage.capacity` (`quota_usages` table, Content-Length pre-check, commit reserve/release, 507) + cleanup after manifest quota failure | Medium, touches storage paths |
| PR3 | Unify delete paths (DeleteManifest / GC artifact / API / GC repository, 409 for index references) + async scope reconciliation + `DeleteBlob` object cleanup | Medium, touches background jobs |
| PR4 | `tag_limit` / `repository_limit` (live COUNT + `FOR UPDATE`) + `size_limit` only increases | Low |
| PR5 | Docs + UI copy ("size is deduplicated logical usage"; home `storage.used` is physical usage / global cap) | Low |

After PR1, new data is measured with the new definition; there is no historical data, so no
backfill is needed.
Do PR2's global capacity together with per-scope admission to avoid a window where physical
space is unconstrained.
Delete paths follow 10.1 ("soft delete + dirty + async reconciliation") and do not introduce row
locks.

---

## 10. Design Decisions (Finalized)

### 10.1 Decided

#### Data model and measurement

1. **Do not add ownership columns to `blobs`** (blobs and namespaces/repositories are
   many-to-many).
2. **Add one polymorphic join table `blob_attributions`** covering both `namespace` and
   `repository` scopes; do not split it into two tables.
3. **Measure size as distinct `blob.size`**; stop using `artifact.blobs_size`.
4. **Cross-namespace shared blobs count once per namespace** (logical quota).
5. **An index / manifest list's references are manifests, not blobs, so it produces no
   attribution rows**; the child artifacts already count their own layers.
6. **Manifest JSON bytes are not counted toward size**; only `blobs.size` is counted.
7. **Per-scope attribution is written only on manifest push**, not on blob push.

#### Schema

1. **Schema changes go into `0001_initialize.sql`**, with no `0002`; add two tables,
   `blob_attributions` and `quota_usages`.
2. **No historical data**: no legacy-DB compatibility and no data backfill.

#### Limit rules

1. **`size_limit` can only be increased, never decreased**; `0` means unlimited (treated as
    infinity).
    - `0 → finite`: decrease → reject
    - `finite → 0`: increase → allow
    - `finite → smaller finite`: reject
    See 5.7.1 for the exact rules.
2. **`tag_count` / `repository_count` are maintained O(1) counters, updated in the same
    transaction as the tag/repository row write**.

#### Delete semantics

1. **Delete paths use "soft delete + dirty + async workqueue reconciliation" in phase 1**;
    no synchronous subtraction and no row locks.
2. **Deleting a child manifest referenced by an index is rejected**, returning 409 `Conflict`.
3. **Fix `DeleteBlob` object cleanup in this phase**: delete the storage object (or enqueue a
    GC deletion task) and release the global counter after the object is deleted.

#### Global capacity

1. **`storage.capacity` is a global hard limit**: the total physical blob bytes across all
    namespaces must not exceed capacity; orphan blobs count toward global physical capacity.
2. **Global capacity counts only `blobs.size`**, including orphans and proxy-cached blobs, and
    excluding manifest JSON.
3. **Global capacity overflow returns 507 `Insufficient Storage`**; per-namespace/repository
    size overflow continues to use 403 `DENIED`.
4. **`quota_usages` is a single-row table `('global','')`** for the global counter, with atomic
    reserve / release.
5. **The authoritative reservation happens at blob commit**; at the start of an upload, use
    `Content-Length` for a non-authoritative pre-check to fail early and transfer fewer bytes.
6. **Proxy cache fill counts toward global capacity**; when full, reject the proxy fill too
    (507).
7. **Proxy cache fill is not counted in namespace/repository per-scope size**.
8. **Reconciliation: 6h periodic + async scope reconciliation after delete / rollback**.

### 10.2 Deferred (Future Work)

- Counting manifest JSON (`artifact.size`) toward size.
- Writing per-scope attribution on blob push (counting orphan blobs in namespace/repository).
- Making delete synchronous and exact (namespace row lock).
- Moving global capacity to a Redis counter / stronger consistency and multi-replica support.
- A separate quota for orphan blobs.
- The `artifacts(namespace_id, deleted_at)` index (the existing `repositories` / `artifacts`
  indexes are sufficient; add it to 0001 later if benchmarks need it).

---

## 11. File Map

- Models: `pkg/dal/models/{namespace,repository,artifact,blob}.go` (add `blob_attribution.go`, `quota_usage.go`)
- Atomic quota updates: `pkg/dal/repository/namespace/namespace.go`, `pkg/dal/repository/registry/repository.go`
- Join-table / global-counter DAL: `pkg/dal/repository/registry/blob_attribution.go`, `pkg/dal/repository/registry/quota_usage.go` (new)
- Reachable queries / reconciliation: `pkg/dal/repository/registry/artifact.go`, `pkg/background/cronjob/size/size.go`
- Push / delete: `pkg/service/distribution/manifest/manifest.go`
- Blob upload: `pkg/service/distribution/upload/upload.go`
- Blob proxy cache / delete: `pkg/service/distribution/blob/blob.go`
- GC: `pkg/background/daemon/gc/{gc_artifact,gc_blob,gc_repository,gc_tag}.go`
- Global capacity: `pkg/config/configuration.go` (`storage.capacity`), `pkg/server/errcode/distribution.go` (new capacity error)
- Error codes: `pkg/server/errcode/distribution.go`
- Schema changes: edit `pkg/dal/migrations/{sqlite3,turso,mysql,postgresql}/0001_initialize.sql` directly (DDL in 7.1.1; no 0002)
- UI: `web/src/pages/{Namespace,Repository,Home}/index.tsx`, `web/src/components/Quota*`

---

## 12. Implementation Notes

This section records the choices of the first implementation pass; it is a snapshot, not a
guarantee of the current code.

- `blob_attributions` / `quota_usages` DDL is in all four `0001_initialize.sql` files (no `0002`).
- `BlobAttributionRepository` uses raw GORM `Table(...)` for the batch upsert and dynamic
  `WHERE`; `make gormgen` has been run, so `query.BlobAttribution` / `query.QuotaUsage` are
  also available.
- Manifest push writes attributions for both scopes and uses the `batch_id` sum as the increment.
- Reconciliation calls `SyncAndSum` for dirty scopes; the legacy `GetNamespaceSize` /
  `GetRepositorySize` read the attribution table.
- Delete paths mark `size_dirty` and trigger a 30s async goroutine (not a durable workqueue) to
  reconcile immediately.
- `DeleteBlob` deletes the storage object, then removes the row and releases global capacity in
  one DB transaction.
- Global capacity is reserved together with the blob row insert (before moving the object into
  place); failures delete the row and release. The size cronjob syncs `quota_usages('global')`
  from live blobs under the same row lock as reservations.
- Reconciliation locks the scope row first (a no-op dirty update, so it also works where
  `FOR UPDATE` is ignored), then repairs attributions and writes the size and dirty flag in one
  transaction. sqlite3 uses `_txlock=immediate`.
- Tag/repository count limits are enforced with `COUNT(*)` and row locks; `size_limit` can only
  increase.
- UI copy labels size as deduplicated and home storage as physical.

---
