# Size / 数量配额（quota）设计

> 状态：已定稿（设计文档）
> 最后更新：2026-11-04
> 范围：`namespace.size_limit`、`repository.size_limit`、`namespace.tag_limit`、`repository.tag_limit`、
> `namespace.repository_limit`、全局 `storage.capacity`。
> 本文只讨论“计量口径、计量时机、限流位置、删除回退、对账”，不做 UI 改版。
> 实施进度以对应 PR / 代码为准。

---

## 1. 结论

1. **`size` 应该按 blob 字节算，并且按 blob 去重。**
   现在用的是 `artifact.blobs_size`（manifest 里客户端声明的 `descriptor.size` 之和），
   既不去重，也不是服务端实测值。
2. **blob 和 namespace / repository 是多对多关系，不能在 `blobs` 表上加 `namespace_id`。**
   `blobs` 按 `digest` 全局唯一，一个 digest 只有一行；加单个归属列会把它变成一对多，
   跨 namespace 共享 / 挂载就无法正确计费。
3. **推荐加一张中间表 `blob_attributions`**，把“某个 scope 占用了某个 blob”显式记录下来：
   - manifest push 时幂等写入，按 `batch_id` 汇总本次新增字节；
   - `namespace.size` / `repository.size` 继续用现有字段做原子增量；
   - blob push 做 per-scope best-effort 准入，并参与 §5.8 的全局容量硬限制。
4. **中间表顺带解决 push × push 的并发重复计量**（主键幂等）。
   delete × push 的少计风险用“删除不做同步减法 + 置 dirty + 对账”规避，第一期不需要行锁。
5. 现状里 **`tag_limit` 和 `repository_limit` 完全没有生效**：字段只在 API 里读写，
   三个对应的错误码函数（`GenDSErrCodeResourceCountQuotaExceed*`）无人调用。
6. 现状里 size 的**删除回退路径不完整**（API 删除 repository、GC 删除 repository 都不回退
   也不标记 dirty），会出现“永久虚高且对账也修不回来”。
7. **本次约束：不新增 `0002` SQL 迁移。** 中间表和其他 schema 改动直接并入
   `pkg/dal/migrations/{sqlite3,turso,mysql,postgresql}/0001_initialize.sql`，四种方言同步修改。
   当前没有历史数据，不考虑老库兼容和回填。
8. **`storage.capacity` 要做全局硬限制**：所有 namespace 的 blob 物理总量（全局去重后）
   不能超过 capacity。例如 capacity=100M，两个 namespace 各设 100M，总 blob 也只能到 100M。

---

## 2. 现状（代码事实）

### 2.1 数据字段

| 表 | 字段 | 说明 |
| --- | --- | --- |
| `namespaces` | `size`, `size_limit`, `size_dirty`, `tag_limit`, `tag_count`, `repository_limit`, `repository_count` | `pkg/dal/models/namespace.go` |
| `repositories` | `size`, `size_limit`, `size_dirty`, `tag_limit`, `tag_count` | `pkg/dal/models/repository.go` |
| `artifacts` | `size`（manifest 本身字节）, `blobs_size` | `pkg/dal/models/artifact.go` |
| `blobs` | `size`（上传时实测）, 全局按 `digest` 唯一 | `pkg/dal/models/blob.go` |

`0` 表示不限制；`size_dirty=true` 表示等待对账。

> 历史背景：当前这套“原子增量 + 后台对账”的实现来自
> `.trae/documents/atomic-size-counter-reconcile.md`。那次改动的目标是**性能**
> （把 push 热路径的 `SELECT SUM(...)` 换成 O(1) 的 `UPDATE`），
> 计量口径直接沿用了已有的 `artifact.blobs_size`，并没有重新定义“size 到底该计什么”。
> 本文要解决的就是这个口径问题。

### 2.2 计量写入点（当前）

| 时机 | 代码 | 计的是什么 |
| --- | --- | --- |
| 普通 manifest push | `pkg/service/distribution/manifest/manifest.go:405-411`、`:563-580` | `artifact.blobs_size` |
| index/manifest list push | `manifest.go:664-681` | 子 manifest 的 `descriptor.size`（JSON 大小，不是 layer） |
| 按 digest 删除 manifest | `manifest.go:923-947` | 回退 `artifact.blobs_size` |
| GC 删除 artifact | `pkg/background/daemon/gc/gc_artifact.go:158-187` | 回退 `Σ artifact.blobs_size` |
| blob 上传（`PutUpload` / 单次 `PostUpload`） | `pkg/service/distribution/upload/upload.go:130-141`、`:320-326` | **不计量** |
| proxy 回源缓存 blob | `pkg/service/distribution/blob/blob.go:177-193` | **不计量** |
| 删除 blob（API / GC） | `blob.go:79-104`、`pkg/background/daemon/gc/gc_blob.go:141-181` | **不回退** |
| API 删除 repository / namespace | `pkg/service/repositories/repositories.go:221-257`、`pkg/service/namespaces/namespace.go:234-265` | **不回退、不置 dirty** |
| GC 删除空 repository | `pkg/background/daemon/gc/gc_repository.go:122-155` | **不回退 namespace size** |

`blobs_size` 的来源：

```go
// manifest.go:405-411
var blobsSize int64
digests := make([]string, 0, len(manifest.References()))
for _, reference := range manifest.References() {
    digests = append(digests, reference.Digest.String())
    blobsSize += reference.Size          // manifest 里客户端声明的值
}
digests = dedupeDigests(digests)         // 去重只作用于 digests，没有作用于 blobsSize
```

### 2.3 对账

`pkg/background/cronjob/size/size.go` 用 `GetNamespaceSize` / `GetRepositorySize`
（`pkg/dal/repository/registry/artifact.go:474-498`）重算，而这两个函数同样是
`SUM(artifacts.blobs_size)`，即**用同一个错误口径修错误口径**。

另外 `pkg/background/daemon/pushed/artifact.go:106-135` 也有一段“push 后全量重算并覆盖
namespace/repository size”的逻辑，同样使用这两个函数。它注册的消费者是
`DaemonArtifactPushed`，但仓库里没有任何地方 `Produce` 这个事件（属于未启用的死代码），
所以当前不会覆盖增量值；如果后续启用，必须同步改成新口径。

### 2.4 数量限制

- `tagRepository.Create`（`pkg/dal/repository/registry/tag.go:88-113`）是“先查后插/更新”，
  没有任何 limit 判断（也不是原子 upsert，并发创建同名新 tag 会撞唯一索引）。
- `manifest.go:557`、`:658` 的注释写着 “maybe got exceed tag quota error”，但下游不会返回该错误。
- `CreateRepository`（`pkg/service/repositories/repositories.go:74-127`）和
  `ensureRepository`（`manifest.go:240-312`）都不检查 `namespace.repository_limit`。
- `tag_count` / `repository_count` 由 tag 和 repository 的写入路径维护，列表页直接读取存储的计数字段。

### 2.5 现有错误码

`pkg/server/errcode/distribution.go:257-330` 已有且可复用：

- `GenDSErrCodeResourceSizeQuotaExceedNamespace`
- `GenDSErrCodeResourceSizeQuotaExceedRepository`
- `GenDSErrCodeResourceCountQuotaExceedRepository`（未使用）
- `GenDSErrCodeResourceCountQuotaExceedNamespaceRepository`（未使用）
- `GenDSErrCodeResourceCountQuotaExceedNamespaceTag`（未使用）

---

## 3. 问题清单

### P0（会导致配额失真或形同虚设）

1. **同一个 blob 被重复计量**
   - 多个 image 共享 base layer：每个 artifact 各自计一次。
   - 同一个 manifest 里重复引用同一个 digest：`blobsSize` 不去重。
   - index 的 `blobs_size` 是子 manifest JSON 大小，和子 artifact 的 layer 口径不一致。
2. **使用客户端声明的 `descriptor.size`，不是服务端实测值**
   - 客户端可以在 manifest 里少报 size，从而绕过配额。
   - 正确值在 `blobs.size`（上传时按实际字节数写入）。
3. **blob 上传完全没有配额约束**
   - 只上传 blob、不推 manifest，可以无限占用磁盘；`namespace.size` 仍是 0。
   - manifest 被配额拒绝时，blob 和 manifest 对象都已经落盘，成为孤儿（只能等 GC）。
4. **`tag_limit` / `repository_limit` 未实现**（见 2.4）。
5. **删除路径不回退或不对账**
   - API 删除 repository：namespace 的 size 不减，且不置 `size_dirty`，对账永远不修。
   - GC 删除 repository：同上。
   - 删除 tag：不影响 size（artifact 还在），符合预期。
   - 删除 blob：在新口径下由中间表清理，见 5.4。

### P1（正确性/并发/运维）

1. **并发计量竞态**
   - push × push（两个 artifact 共享 blob）：用中间表的主键幂等后可以彻底解决。
   - delete × push：可能少计，需要行锁或“删除不做同步减法”（见 5.6）。
2. **对账口径错误**：`SUM(artifacts.blobs_size)` 无法修复上述重复计量。
3. **limit 修改逻辑有问题（已定规则：只能调大）**
   - `namespace.UpdateNamespace`（`pkg/service/namespaces/namespace.go:182-184`）用
     `old > new` 判断，把 `0`（不限制）当成最小值：`100 → 0`（改成不限制）被拒绝，
     `0 → 100`（从无限制改成有限制）被允许，两种情况都判反了；正确规则见 5.7.1。
   - `repository.UpdateRepository` 完全没有校验，需要补同样的“只能调大”规则。
4. **软删 artifact 与 blob 关联残留**
   - `artifactRepository.DeleteByID` 是软删，`artifact_blobs` 行不清理。
   - `BlobRepository.FindAssociateWithArtifact`（`blob.go:178-203`）的接口注释写着
     “still associated with live artifacts”，但实现未过滤 `artifacts.deleted_at`，
     导致“曾经被引用过”的 blob 永远无法被 GC（物理泄漏）。
5. **manifest 先落盘、后校验配额**
    - `manifest.go:399` 先 `Storage.Upload`，`:565` 才可能因配额失败返回；
      失败后 manifest 对象残留。
6. **并发写同一 digest 可能 500**
    - `RepoBlob.Create` 直接 insert，依赖 `(digest, deleted_at)` 唯一索引；
      并发时第二个请求会拿到唯一键冲突而不是幂等成功。
    - `artifactRepository.Create` 同理：`isNewArtifact` 是在事务外判断的，
      并发 push 同一 digest 时第二个事务会撞 `(repository_id, digest, deleted_at)`
      唯一索引并返回 `UNKNOWN`，而不是幂等成功。

### P2（可后续）

1. **manifest 引用的 blob 不存在时不报 `BLOB_UNKNOWN`**
    - `FindByDigests`（`blob.go:210-213`）静默返回存在的子集，size 仍按声明值计入。
2. `tag_count` / `repository_count` 是维护中的计数器；quota 校验和列表页直接读取，不再实时 `COUNT(*)`。
3. `storage.capacity`（`pkg/config/configuration.go:570-573`）只在首页展示，不校验。

---

## 4. 目标语义

定义（`0` = 不限制）：

- `repository.size` = 该 repository 下**未删除 artifact 引用到的 distinct blob 的 `blob.size` 之和**。
- `namespace.size` = 该 namespace 下**任意 repository 的未删除 artifact 引用到的 distinct blob 的 `blob.size` 之和**。
- 同一个 blob 被多个 namespace 引用时，**每个 namespace 各计一次**（配额是“逻辑用量”，不是物理占用）。
- 全局物理占用继续看首页的 `storage.used`（`repository.SumBlobSize`，已经是 distinct 口径）。
- manifest JSON 本身（`artifact.size`）不计入 size 配额（字节很小，可后续再议）。
- index / manifest list 的 references 是 manifest 而不是 blob，因此**不产生 size 增量**；
  它引用的子 artifact 已经各自计过 layer。
- 超限错误统一走 403 `DENIED`，复用现有 `GenDSErrCodeResourceSizeQuotaExceed*`。

举个栗子：同一 namespace 下有两个镜像 A、B，共享一个 100 MB 的 base layer，各有一个
10 MB 的独有 layer。

- 现状（按 artifact 的 `blobs_size` 累加）：100 + 10 + 100 + 10 = 220 MB。
- 目标（按 distinct blob）：100 + 10 + 10 = 120 MB。
- tag 本身不占空间：给 A 打 10 个 tag，size 仍然是 110 MB。

### 4.1 为什么是“中间表”，而不是在 blobs 上加列

blob 和 namespace / repository 是**多对多**：

- 同一个 digest 全库只有一行（`blobs_unique_with_digest UNIQUE (digest, deleted_at)`）；
- 它可以被多个 repository 引用，也可以被多个 namespace 引用；
- 在 `blobs` 上加单个 `namespace_id` / `repository_id` 会把它变成一对多，
  第二个引用方无法记录，跨 namespace 挂载也无法正确计费。

所以归属信息必须放在关系表里：

```text
blob_attributions(scope_type, scope_id, blob_id, size, batch_id)
  scope_type = 'namespace' | 'repository'
  (scope_type, scope_id, blob_id) 为主键
```

这样：

- `SUM(size) WHERE scope_type = ? AND scope_id = ?` 就是该 scope 的去重后大小；
- 主键 `(scope_type, scope_id, blob_id)` 天然幂等，解决 push × push 的重复计量；
- 删除 scope 内最后一个引用时，删掉对应行即可回退。

> 如果更看重可读性，也可以拆成 `namespace_blobs` / `repository_blobs` 两张表，
> 逻辑完全等价；本文用一张 `blob_attributions` 说明。

---

## 5. 推荐方案：中间表 + 原子增量

### 5.1 表结构

一张多态中间表，直接加进 `0001_initialize.sql`（四种方言同步，不新增 0002）。
**完整的分方言建表 SQL（可直接粘贴）见 7.1。**

设计要点：

- `size` 冗余存储：`blobs.size` 在 blob 生命周期内不变，冗余后
  `SUM(size)` 只需要扫中间表，不再 join `blobs`。
- `batch_id`：每次 push 生成一个 UUIDv7，用来识别“本次事务真正插入了哪些行”。
  冲突跳过的行保留旧的 `batch_id`，不会计入本次增量。
- `blob_attributions_idx_blob`：blob 被 GC / 删除时按 `blob_id` 清理归属。
- `scope_type` 用字符串便于排查；Go 侧定义常量，避免拼写漂移。
- 不建外键（与现有表保持一致，项目本来就不用 FK）。

### 5.2 manifest push 改造

`pkg/service/distribution/manifest/manifest.go`：

1. 先校验所有 reference 对应的 blob 都存在（否则 `BLOB_UNKNOWN`），并用
   `FindByDigests` 返回的 `blob.size` 计算“本 artifact 内去重后的展示用 `BlobsSize`”。
2. 在现有 `query.Q.Transaction` 内，插入 artifact / tag / m2m 之后，对
   repository 和 namespace 两个 scope 分别做（**不要只在 `isNewArtifact` 时做**，
   每次都幂等执行，顺便自愈历史缺失的归属行）：

   ```text
   batchID := uuid.NewV7String()

   INSERT INTO blob_attributions (scope_type, scope_id, blob_id, size, batch_id, created_at)
   SELECT ?, ?, b.id, b.size, ?, ?
   FROM blobs b
   WHERE b.id IN (:blob_ids) AND b.deleted_at = 0
   ON CONFLICT (scope_type, scope_id, blob_id) DO NOTHING;
   -- MySQL 用 INSERT IGNORE 或等价写法

   delta := SELECT COALESCE(SUM(size), 0)
            FROM blob_attributions
            WHERE scope_type = ? AND scope_id = ? AND batch_id = ?;
   ```

3. 先 `IncrementSize(scope, delta)`（已有原子条件更新，
   `namespace.go:218-243`、`repository.go:322-347`），失败即返回 403 并回滚；
   `delta` 为 0 时不需要调用。
4. 删除 `IncrementSize(..., artifactObj.BlobsSize)` 的调用（`:563-580`、`:664-681`）。
5. index 分支：index 的 references 是 manifest 而不是 blob，不产生归属行，增量自然为 0，
   不要再用子 manifest 的 `descriptor.size` 计入 size。
6. `recordNamespaceSizeDelta`（analytics）改用同一个 `delta`，保证图表和配额口径一致。
7. 把 `Storage.Upload(manifestPath, ...)` 移到配额检查/事务成功之后；
   若保持现状，至少失败时 `Storage.Delete(manifestPath)`。

为什么用 `batch_id` 而不是 `RowsAffected`：

- 批量 `INSERT ... ON CONFLICT DO NOTHING` 的 `RowsAffected` 只能拿到行数，拿不到“本次插入行的
  `size` 之和”；逐行插入再累加会产生 N 次往返。
- `batch_id` 让“本次插入的行”可以直接用一条 `SUM` 查出来，且天然幂等。
- 并发的第二个事务插入冲突、`SUM` 为 0，不会重复加。

### 5.3 blob push 准入（best-effort）

`pkg/service/distribution/upload/upload.go`：

- `PutUpload`：在“digest 已存在 → 直接成功返回”的分支之后（`:222-234`），
  用 `sizeBefore + contentLength` 作为新增字节，读取 repository / namespace 的
  `size + size_limit`，若会超限直接返回 403 `GenDSErrCodeResourceSizeQuotaExceed*`，
  在 `CommitUpload` 之前拒绝（避免白传）。
- 单次 `PostUpload`（`:97-163`）：同样在写盘前用 `Content-Length` 预检；拿不到长度时跳过（best-effort）。
- **统一 blob.size 的写法**：`PutUpload` 现在用 `sizeBefore + contentLength`（`:320-326`），
  改成实测的 `sizeBefore + sizeUploaded`；单次 `PostUpload` 已经是 `counter.Count()`，保持即可。
- 该检查只是 per-scope 的“提前止损”，不是精确计量：
  - 并发上传仍可能短时超额（可接受）；
  - 上传但未推 manifest 的 blob 不写 `blob_attributions`，不计入 `namespace.size`，最终由 GC 回收；
  - 真正精确的 per-scope 拦截在 manifest push。
- **已确定：per-scope 归属只在 manifest push 写**，blob push 不写归属；
  因此孤儿 blob 只受 §5.8 的全局容量约束，不计入 namespace/repository size。

### 5.4 删除 / 回退改造（默认：软删 + 置 dirty + 对账）

第一期统一采用“删除不做同步减法”：

**软删 artifact / repository 后，只把对应 scope 置 `size_dirty = true`，并立即触发一次该 scope
的对账**；由对账删除多余的 `blob_attributions` 行，并把 `size` 覆盖为 `SUM(size)`。

这样做的好处：

- 任何并发交错（delete × push、delete × delete）都不会少计，最多是配额暂时偏高（安全方向）；
- 删除路径不需要锁，也不需要在事务里做复杂的“不可达”查询。

| 路径 | 改法 |
| --- | --- |
| `DeleteManifest` 按 digest | 事务内软删 artifact；事务提交后置 namespace / repository dirty，并触发一次 scope 对账 |
| `DeleteManifest` 按 tag | 不变（artifact 还在，字节也还在） |
| GC artifact | 软删后置 dirty + 触发对账（可按批聚合） |
| API 删除 repository | 删除该 repository 的归属行（`scope_type='repository'`）；namespace 置 dirty + 触发对账 |
| GC 删除 repository | 同上（当前完全不处理） |
| 删除 blob（API / GC） | 按 `blob_id` 删除所有归属行；同时释放 §5.8 的全局计数 |
| 删除 namespace | 删除该 namespace 的归属行；它的 repository 归属行可以留给 repository 删除或 GC 处理 |

> 对账间隔默认 6h，所以删除 / 回滚后必须**立即触发一次**该 scope 的对账，否则用户会看到
> 配额迟迟不释放。定时对账只作为兜底。

#### 后续优化（本期不做）：删除路径同步精确

如果以后要求“删除提交后 quota 立即恢复”，再给删除事务加 namespace 行锁，并按下面的 SQL
同步减量（会和 push 串行化，见 5.6）：

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

repository 版把 `a.namespace_id` 换成 `a.repository_id`、`scope_type` 换成 `repository`。

执行顺序：软删 artifact → 查不可达行并求和 → 删除这些行 → `DecrementSize(scope, sum)`。
这条路径需要和 push 互斥，否则可能出现 delete × push 少计（见 5.6）。

> 边界情况：删除被 index 引用的子 manifest。`artifact_artifacts` 只表示 index → 子 artifact
> 的关系，不参与 `artifact_blobs` 的可达性；如果允许删除子 manifest，它的归属行会被清掉，
> size 提前释放，而 index 仍然引用它（broken index）。**已确定：在 `DeleteManifest` 里用
> `IsArtifactAssociatedWithArtifact`（已实现但当前未调用）拒绝删除被 index 引用的 artifact，
> 返回 409 `Conflict`。**

### 5.5 对账

中间表是 size 的直接数据源，但为了防止“有 artifact 引用、却没有归属行”（历史数据、异常路径），
对账仍然按“可达 blob 集合”修复：

1. 只处理 `size_dirty = true` 的 namespace / repository。
2. **补缺失**：用可达查询找出缺少的 `(scope, blob)`，`INSERT ... ON CONFLICT DO NOTHING`。
3. **删多余**：删除 `blob_attributions` 中在该 scope 内已不可达的行。
4. `size = SELECT SUM(size) FROM blob_attributions WHERE scope_type=? AND scope_id=?`。
5. 清 `size_dirty`。

可达查询（仅后台对账使用，不会出现在 push 热路径）：

```sql
-- namespace 可达 blob（distinct）
SELECT b.id, b.size
FROM repositories r
JOIN artifacts a ON a.repository_id = r.id AND a.deleted_at = 0
JOIN artifact_blobs ab ON ab.artifact_id = a.id
JOIN blobs b ON b.id = ab.blob_id AND b.deleted_at = 0
WHERE r.namespace_id = :namespace_id AND r.deleted_at = 0
GROUP BY b.id, b.size;

-- repository 可达 blob（distinct）
SELECT b.id, b.size
FROM artifacts a
JOIN artifact_blobs ab ON ab.artifact_id = a.id
JOIN blobs b ON b.id = ab.blob_id AND b.deleted_at = 0
WHERE a.repository_id = :repository_id AND a.deleted_at = 0
GROUP BY b.id, b.size;
```

- `pkg/background/cronjob/size/size.go` 的批处理 + 分布式锁结构不用改；对每个 dirty scope
  改为调用 `BlobAttributionRepository.SyncAndSum`（先修复归属行，再求和）；
  `ArtifactRepository.GetNamespaceSize` / `GetRepositorySize` 不再被对账作业使用。
- 所有会改变可达集合的路径都置 `size_dirty = true`：
  manifest push / delete、GC artifact、API / GC 删除 repository、删除 blob。
- 全局容量对账：`quota_usages('global') = SUM(blobs.size) WHERE blobs.deleted_at = 0`，
  可以复用同一个 cronjob 一起跑。
- 除定时对账外，删除 / 回滚后要能**立即触发一次 scope 对账**（例如把 scope 放进一个
  轻量 workqueue，或直接同步执行一次）；否则配额要等下一个 6h 周期才恢复。
- 对账是 O(scope 内可达 blob 数)，但只对 dirty 行执行，不会拖慢 push。

### 5.6 并发下的精确性

| 竞态 | 结果 | 是否需要修 |
| --- | --- | --- |
| push × push（两个 artifact 共享 blob） | 中间表主键幂等，**不会多计** | 已解决 |
| delete × push | 默认“删除不做同步减法”时**不会少计**；只有同步精确删除才需要注意 | 默认已解决 |
| delete × delete | 中间表行只会删一次，**不会少计** | 已解决 |

push × push 为什么自动精确：

- 两个事务插入同一 `(scope_type, scope_id, blob_id)`，主键保证只有一个成功；
- 第二个事务的 `INSERT ... ON CONFLICT DO NOTHING` 冲突跳过，`batch_id` 汇总为 0；
- 不会出现“两个事务都加一次”。

delete × push 什么时候会少计（仅同步精确删除时）：

- 删除事务判断“这个 blob 已不可达”，准备删归属行；
- 并发 push 插入新 artifact 引用同一个 blob，但它的引用在删除事务的快照里不可见；
- 删除事务删掉归属行并 `DecrementSize`，push 事务因为行已存在而没有加；
- 最终 artifact 还引用着 blob，但 size 少计了。

处理方式：

1. **默认：删除不做同步减法**（已确定，最省事）。
   删除只软删 + 置 `size_dirty=true` + 立即触发该 scope 的对账，
   由对账删除多余的归属行并覆盖 size。任何并发交错都不会少计，最坏是暂时偏高。
2. **后续优化（本期不做）：删除路径同步精确**。需要删除事务也取作用域行锁，和 push 互斥：
   在事务的第一条语句锁 `namespaces` 行（`SELECT ... FOR UPDATE`），push 事务也取同一把锁。
   锁顺序固定 namespace → repository。
   注意 MySQL REPEATABLE READ 的快照在事务内第一次普通 SELECT 时建立，
   锁必须在任何一致读之前；SQLite 的 GORM 方言会静默忽略 `FOR UPDATE`
   （`gorm.io/driver/sqlite@v1.6.0/sqlite.go:126`），需要 DSN 加 `_txlock=immediate`
   或接受“busy 后重试”；Turso 是 `BEGIN` 快照隔离，冲突按可重试错误处理。

> push 路径因为主键幂等，不需要加锁；第一期删除走对账即可，不要引入行锁串行化。

### 5.7 数量限制（tag / repository）

先明确“计数器”还是“实时 COUNT”：

- 字段已经存在（`tag_count`、`repository_count`），并且按 O(1) 计数器维护：
  tag / repository 的写入在同一事务内更新所属行，quota 校验和列表页直接读取计数器。

#### tag limit（仓库级 + namespace 级）

在 `putManifestManifest` / `putManifestIndex` 创建 tag 之前（仅 `isNewTag == true`，
retag 不计数）：

1. 锁住父行：`SELECT ... FOR UPDATE`（MySQL/PostgreSQL；SQLite/Turso 本身串行写）。
2. `COUNT(*) FROM tags WHERE repository_id = ? AND deleted_at = 0`，对比 `repositories.tag_limit`。
3. `COUNT(*)` 走 namespace（join `repositories`）对比 `namespaces.tag_limit`。
4. 超限返回 `GenDSErrCodeResourceCountQuotaExceedRepository` /
   `GenDSErrCodeResourceCountQuotaExceedNamespaceTag`；事务回滚。

#### repository limit

在 `CreateRepository`（`pkg/service/repositories/repositories.go:74`）和
`ensureRepository`（`manifest.go:240`，首次 push 自动建 repo）里：
锁 namespace 行 → `COUNT(*) FROM repositories WHERE namespace_id = ? AND deleted_at = 0`
→ 对比 `namespace.repository_limit` → 超限返回
`GenDSErrCodeResourceCountQuotaExceedNamespaceRepository`。

#### 如果改成计数器

- push：`UPDATE ... SET tag_count = tag_count + 1 WHERE id = ? AND (tag_limit = 0 OR tag_count + 1 <= tag_limit)`
  （`RowsAffected == 0` 即超限），namespace 和 repository 各一次，同一事务。
- 删除：`DeleteByName`、`DeleteByArtifactID`、GC tag、GC repository、API 删除 repository
  都要递减；建议加一个 count 对账作业（复用 `size` 作业的框架）。
- 优点：push 热路径 O(1)；缺点：删除路径多，容易漏，需要严格对账。

### 5.7.1 `size_limit` 只能调大，不能调小

口径：`0` = 不限制（视为无穷大）。规则：

| 旧值 | 新值 | 判定 |
| --- | --- | --- |
| `0`（不限制） | 非 0 | 从无限制变成有限制 → **调小，拒绝** |
| 非 0 | `0`（不限制） | 从有限制变成无限制 → **调大，允许** |
| 非 0 | 非 0 且新 < 旧 | **调小，拒绝** |
| 非 0 | 非 0 且新 >= 旧 | 调大 / 不变，允许 |
| `0` | `0` | 不变，允许 |

当前 `UpdateNamespace`（`pkg/service/namespaces/namespace.go:182-184`）用的是
`old > new`，把 `0` 当成最小值，正好把上面两个涉及 `0` 的场景判反了；
需要改成上面的规则，并在 `UpdateRepository`（`pkg/service/repositories/repositories.go:198-204`）
里补同样的校验。

错误信息用：`size_limit can only be increased; 0 means unlimited`。

### 5.8 全局物理上限 `storage.capacity`

`storage.capacity`（`pkg/config/configuration.go:570-573`）当前只在首页展示，不校验。
需求：**所有 namespace 的 blob 物理总量（全局去重后）不能超过 capacity**。
例如 capacity=100M，两个 namespace 各设 100M，总 blob 也只能到 100M。
孤儿 blob 不做单独的 per-scope 限制，但它们是物理字节，仍然占用全局容量。

#### 计量口径

- `used = SUM(blobs.size) WHERE blobs.deleted_at = 0`：包含孤儿 blob、proxy 缓存 blob。
- 不含 manifest JSON（`artifact.size`），与 per-scope size 口径一致。
- `capacity = 0` 表示不限制。

#### 执行点

- **blob commit**：dedupe 未命中（确实要新增一个 blob）时，在把临时对象 Move 到最终路径
  之前做一次原子预留；失败则 abort 上传并返回 507。
- dedupe 命中：没有新增字节，不检查。
- **预检查**：上传开始时用 `Content-Length` 先做一次非权威检查，提前失败、少传字节；
  最终以 commit 时的原子预留为准。
- **proxy 回源缓存**：同样计入全局容量；容量满时**拒绝回源并返回 507**。
- manifest push 不重复检查全局容量（blob 已经计过）。

#### 计数器

不能每次上传都 `SUM(blobs.size)`（O(N)）。用单行计数表 `quota_usages`
（第二张表，schema 并入 0001，已确定）：

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

预留（原子条件 UPDATE）：

```sql
UPDATE quota_usages
SET size = size + :delta, updated_at = :now
WHERE scope_type = 'global' AND scope_id = ''
  AND (:capacity = 0 OR size + :delta <= :capacity);
-- RowsAffected == 0 → 容量不足
```

释放（storage 对象真正删除时）：`size = MAX(size - :delta, 0)`。
对账：`size = SUM(blobs.size) WHERE deleted_at = 0`。

> 释放时机要和“对象真正删除”对齐：GC 的 `deleteBlobObject` 删除 storage 对象成功后再减。
> 当前 API `DeleteBlob` 只软删 DB 行、不删 storage 对象，而 GC 只找 `deleted_at = 0`
> 的 blob，会泄漏对象且不计入容量；**本期一并修**（直接删对象或投递 GC 删除任务，
> 并在对象删除成功后释放全局计数）。

为什么不用 `blob_attributions` 的 global 行：中间表是按“scope 可达 blob”记的，
孤儿 blob 不在里面；全局容量要算物理 blob，两者口径不同，所以单独一个计数表更清晰。

#### 并发

- 条件 UPDATE 在单行上串行化，两个并发上传不会同时突破 capacity；
- 预留成功后再 Move，Move 失败要补偿（减回去）；进程崩溃留下的偏高由对账修正。

#### 错误

新增 `GenDSErrCodeStorageCapacityExceed(current, capacity, increase)`，返回
**507 `Insufficient Storage`**（已确定）。per-namespace/repository 的 size 超限仍用
403 `DENIED`，两者不要混用。

---

## 6. 其它备选方案

### 方案 A：不加中间表，实时 distinct join

- push 时用 5.5 的“可达查询 + NOT EXISTS”算增量，对账也用它。
- 优点：不改 schema。
- 缺点：SQL 复杂（必须 `GROUP BY` / `DISTINCT`，否则重复计数），push 时每次要跨
  `repositories / artifacts / artifact_blobs / blobs` 计算；并发重复计量只能靠行锁或对账兜底。
- 适合“暂时不想动 schema、且 push 量很小”的场景；不作为本次推荐。

### 方案 B：给 blob 加归属列（不推荐）

- `blobs` 增加 `namespace_id` / `repository_id`（首次上传者）。
- 优点：O(1)、无 join、SQL 最简单。
- 缺点（决定性）：
  - blob 与 namespace/repository 是多对多，加单列会破坏共享语义；
  - mount / 跨 namespace 共享时，使用方免费，上传方永久计费；
  - 删除 artifact 不释放配额（要等 blob GC）；
  - 历史 blob 需要 backfill。
- 因此即使它“SQL 最简单”，也不采用。

---

## 7. Schema 改动与数据修复

### 7.1 Schema 改动统一放 0001

- 本次新增两张表：
  - `blob_attributions`：per-scope 的 blob 归属；
  - `quota_usages`：全局物理用量计数器（单行 `('global','')`）。
  以上都直接改 `pkg/dal/migrations/{sqlite3,turso,mysql,postgresql}/0001_initialize.sql`，
  四份保持等价，**不新增 `0002_*.sql`**。
  `artifacts(namespace_id, deleted_at)` 索引本期**不加**，现有 `repositories` / `artifacts`
  索引够用，压测需要时再加进 0001。
- 方言差异按现有 0001 的写法处理：MySQL/SQLite/Turso 用反引号，PostgreSQL 用双引号；
  布尔列分别是 `tinyint(1)` / `integer` / `boolean`；主键和索引名保持现有命名风格。
- **当前没有历史数据**，不需要考虑“老库不重跑 0001”的兼容问题；全新库直接执行 0001 即可。
  `quota_usages` 的 `('global','')` 初始行也直接在 0001 里 `INSERT`。

#### 7.1.1 建表 SQL（直接粘进 0001）

放在 `artifact_blobs` 建表语句之后（Up 段），并在 Down 段加对应的
`DROP TABLE IF EXISTS`。

**MySQL**（`pkg/dal/migrations/mysql/0001_initialize.sql`）：

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

**PostgreSQL**（`pkg/dal/migrations/postgresql/0001_initialize.sql`）：

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

**SQLite3**（`pkg/dal/migrations/sqlite3/0001_initialize.sql`）：

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

**Turso**（`pkg/dal/migrations/turso/0001_initialize.sql`，与 SQLite3 完全一致）：

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

**`quota_usages`（全局物理用量计数，单行 `('global','')`）**

MySQL：

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

PostgreSQL：

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

SQLite3 / Turso：

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

**Down 段**（四份都加，引号按方言）：

```sql
DROP TABLE IF EXISTS `quota_usages`;
DROP TABLE IF EXISTS `blob_attributions`;
-- PostgreSQL 用双引号：DROP TABLE IF EXISTS "quota_usages"; ...
```

配套的 GORM model（`pkg/dal/models/blob_attribution.go`）：

```go
// BlobAttribution records which blob is charged to which scope.
type BlobAttribution struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`

	ScopeType string `gorm:"primaryKey;size:16"`  // namespace / repository
	ScopeID   string `gorm:"primaryKey;size:36"`
	BlobID    string `gorm:"primaryKey;size:36"`
	Size      int64  `gorm:"default:0"`
	BatchID   string `gorm:"size:36;default:''"` // 索引在 0001 里建，不依赖 AutoMigrate
}
```

全局计数 model（`pkg/dal/models/quota_usage.go`）：

```go
// QuotaUsage is the aggregated usage per scope; the global row uses scope_id = ''.
type QuotaUsage struct {
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`

	ScopeType string `gorm:"primaryKey;size:16"` // global / namespace / repository（本期只用 global）
	ScopeID   string `gorm:"primaryKey;size:36"`
	Size      int64  `gorm:"default:0"`
}
```

加完 model 后执行 `make gormgen` 重新生成 `pkg/dal/query`（不要手改 `*.gen.go`），
再在 `pkg/dal/repository/registry/` 下实现增删查。

### 7.2 数据修复（没有历史数据，不需要回填）

当前没有历史数据，**不需要回填**：

- 全新库从 0 开始，blob 在 manifest push 时写归属；
- 全局计数 `quota_usages('global')` 从 0 开始，在 blob 创建 / 删除时增减。

以后如果出现漂移，对账已经能修复（补缺失、删多余、重算 `size` 和全局 `used`）；
也可以保留一个 `sigma tools quota-recompute --scope=...` 的手动修复入口，但不是本次必需。

---

## 8. 测试计划

### DAL（`pkg/dal/repository/registry`）

- 共享 layer 的两个 artifact：repository / namespace 的 `SUM(blob_attributions.size)` 只算一次。
- 同一个 manifest 内重复 digest：只插一行归属。
- 跨 repository（同 namespace）：namespace 只有一行归属，repository 各一行。
- 跨 namespace：各自一行归属（符合“逻辑用量”）。
- 软删 artifact + 触发对账后：对应归属行被删除，size 下降。
- index 引用子 manifest：不产生额外归属行。
- 并发插入同一 `(scope, blob)`：只有一个事务的 `batch_id` 汇总到该行，另一个为 0。
- 对账重复执行：幂等，不重复计数。

### service（`pkg/service/distribution/manifest`）

- 超限 push 返回 403 `DENIED`，artifact/tag/归属行/计数器均回滚。
- 未超限 push 成功；重复 push 同一 digest（`isNewArtifact=false`）不重复计量。
- retag（同名 tag 换 digest）不计新增 tag；新 digest 的归属行正常增加。
- tag limit 命中（repository 级 / namespace 级）。
- repository limit 命中（API 创建 + 首次 push 自动创建）。
- 引用的 blob 不存在时返回 `BLOB_UNKNOWN`。

### blob upload

- 去重命中（digest 已存在）直接成功、不写归属、不增加全局 `used`。
- 新 blob 成功时全局 `used` 增加；删除 blob 后 `used` 减少。
- per-namespace 超限上传被拒绝且不落盘（或落盘后回滚）。
- 全局容量超限上传被拒绝，且计数器不被多扣。
- 并发上传同一 digest 幂等成功（不允许 500）。

### 全局容量

- `capacity=100M`，两个 namespace 各 `size_limit=100M`：总 blob 超过 100M 时被拒绝。
- 去重命中（同一 digest 第二次上传）不重复占用全局容量。
- 预留后 Move 失败，`used` 回退；进程崩溃留下的偏高由对账修正。

### 对账

- 人为制造漂移（删掉一行归属、或把 `size` 改错），置 dirty，跑 `size.go`，断言修复。
- 删除 repository 后 namespace size 正确下降（或 dirty 被置位并被对账修正）。
- 对账后 `size` 与可达 distinct 查询一致；全局 `used` 与 `SUM(blobs.size)` 一致。

### e2e

- 连续 push 两个共享 base layer 的镜像，namespace size ≈ 去重后的大小。
- 把 `size_limit` 设为较小值，push 超出返回 `DENIED`。
- 设置 `tag_limit=1`，push 第二个 tag 返回 `DENIED`。
- 全局 `capacity` 调小后，新的 blob push 被拒绝。

---

## 9. 分期落地建议

| PR | 内容 | 风险 |
| --- | --- | --- |
| PR1 | `blob_attributions` 表（并入 0001）+ push 幂等写入 + batch 增量 + 对账口径 + dirty 标记 | 中，改核心 push 路径 |
| PR2 | blob push per-scope 准入 + 全局 `storage.capacity`（`quota_usages` 表、Content-Length 预检查、commit 预留/释放、507）+ manifest 配额失败后的对象清理 | 中，涉及存储路径 |
| PR3 | 删除路径统一（DeleteManifest / GC artifact / API / GC repository，index 引用返回 409）+ 异步 scope 对账 + `DeleteBlob` 对象清理 | 中，涉及后台任务 |
| PR4 | `tag_limit` / `repository_limit`（实时 COUNT + `FOR UPDATE`）+ `size_limit` 只能调大 | 低 |
| PR5 | 文档 + UI 文案（“size 为去重后的逻辑用量”，首页 `storage.used` 为物理占用/全局上限） | 低 |

PR1 完成后新数据就会按新口径计量；没有历史数据，不需要回填。
PR2 的全局容量和 per-scope 准入一起做，避免先上线一段“物理空间不受控”的窗口。
删除路径按 10.1 的“软删 + dirty + 异步对账”，不引入行锁。

---

## 10. 设计决策（已定稿）

### 10.1 已确定

#### 数据模型与口径

1. **不在 `blobs` 表加归属列**（blob 与 namespace/repository 是多对多）。
2. **新增一张多态中间表 `blob_attributions`**，同时承载 `namespace` 和 `repository`
   两种 scope，不拆成两张表。
3. **size 口径按 distinct `blob.size`**，不再使用 `artifact.blobs_size`。
4. **跨 namespace 共享 blob：每个 namespace 各算一次**（逻辑配额）。
5. **index / manifest list 的 references 是 manifest 而不是 blob，不产生归属行**；
   它引用的子 artifact 已经各自计过 layer。
6. **manifest JSON 字节不计入 size**，只算 `blobs.size`。
7. **per-scope 归属只在 manifest push 写**，blob push 不写归属。

#### Schema

1. **schema 改动并入 `0001_initialize.sql`**，不新增 `0002`；新增 `blob_attributions`
   和 `quota_usages` 两张表。
2. **没有历史数据**：不做老库兼容、不做数据回填。

#### limit 规则

1. **`size_limit` 只能调大，不能调小**；`0` 表示不限制（视为无穷大）。
    - `0 → 有限值`：调小 → 拒绝
    - `有限值 → 0`：调大 → 允许
    - 有限值 → 更小的有限值：拒绝
    具体判定见 5.7.1。
2. **`tag_count` / `repository_count` 采用 O(1) 计数器，与 tag/repository 行写入在同一事务内更新**。

#### 删除语义

1. **删除路径第一期采用“软删 + 置 dirty + 异步 workqueue 立即对账”**，
    不做同步减法、不加行锁。
2. **删除被 index 引用的子 manifest 拒绝**，返回 409 `Conflict`。
3. **API `DeleteBlob` 一并修复对象清理**：删除 storage 对象（或投递 GC 删除任务），
    对象删除成功后释放全局计数。

#### 全局容量

1. **`storage.capacity` 做全局硬限制**：所有 namespace 的 blob 物理总量不能超过 capacity；
    孤儿 blob 计入全局物理容量。
2. **全局容量只算 `blobs.size`**，包含孤儿和 proxy 缓存 blob，不含 manifest JSON。
3. **全局容量超限返回 507 `Insufficient Storage`**；per-namespace/repository 的 size 超限
    继续用 403 `DENIED`。
4. **`quota_usages` 单行表 `('global','')`** 做全局用量计数，原子预留 / 释放。
5. **全局容量在 blob commit 做权威预留**；上传开始时用 `Content-Length` 做一次非权威预检查，
    提前失败、少传字节。
6. **proxy 回源缓存计入全局容量**；容量满时也拒绝回源（507）。
7. **proxy 回源缓存不计入 namespace/repository 的 per-scope size**。
8. **对账：6h 定时 + 删除 / 回滚后异步触发一次 scope 对账**。

### 10.2 本期不做（后续演进）

- 把 manifest JSON（`artifact.size`）计入 size。
- blob push 时写 per-scope 归属（把孤儿 blob 计入 namespace/repository）。
- 删除路径改成同步精确（namespace 行锁）。
- 全局容量改用 Redis 计数 / 更强的一致性与多副本方案。
- 给孤儿 blob 做单独的 quota。
- `artifacts(namespace_id, deleted_at)` 索引（现有 `repositories` / `artifacts`
  索引已够用，压测需要时再加进 0001）。

---

## 11. 涉及文件速查

- 模型：`pkg/dal/models/{namespace,repository,artifact,blob}.go`（新增 `blob_attribution.go`、`quota_usage.go`）
- 配额原子更新：`pkg/dal/repository/namespace/namespace.go`、`pkg/dal/repository/registry/repository.go`
- 中间表 / 全局计数 DAL：`pkg/dal/repository/registry/blob_attribution.go`、`pkg/dal/repository/registry/quota_usage.go`（新增）
- 可达查询 / 对账：`pkg/dal/repository/registry/artifact.go`、`pkg/background/cronjob/size/size.go`
- push / 删除：`pkg/service/distribution/manifest/manifest.go`
- blob 上传：`pkg/service/distribution/upload/upload.go`
- blob 代理缓存 / 删除：`pkg/service/distribution/blob/blob.go`
- GC：`pkg/background/daemon/gc/{gc_artifact,gc_blob,gc_repository,gc_tag}.go`
- 全局容量：`pkg/config/configuration.go`（`storage.capacity`）、`pkg/server/errcode/distribution.go`（新增容量错误码）
- 错误码：`pkg/server/errcode/distribution.go`
- Schema 改动：直接改 `pkg/dal/migrations/{sqlite3,turso,mysql,postgresql}/0001_initialize.sql`（建表 SQL 见 7.1.1，不新增 0002）
- UI：`web/src/pages/{Namespace,Repository,Home}/index.tsx`、`web/src/components/Quota*`

---

## 12. 实现说明

本节记录第一轮实现的选择，是快照，不代表代码当前状态。

- `blob_attributions` / `quota_usages` 的 DDL 已并入四份 `0001_initialize.sql`（无 0002）。
- `BlobAttributionRepository` 的批量 upsert 和动态 `WHERE` 走原生 GORM `Table(...)`；
  `make gormgen` 已执行，`query.BlobAttribution` / `query.QuotaUsage` 也可用。
- manifest push 为 repository 和 namespace 两个 scope 写归属，并用 `batch_id` 求和作为增量。
- 对账对 dirty scope 调 `SyncAndSum`；legacy `GetNamespaceSize` / `GetRepositorySize`
  也改为读归属表。
- 删除路径置 `size_dirty`，并用 30s 的异步 goroutine（不是持久化 workqueue）立即对账。
- `DeleteBlob` 先删 storage 对象，再在同一个 DB 事务里删行并释放全局容量。
- 全局容量与 blob 行插入在同一个事务里预留（在 Move 之前）；失败时删行并释放。
  `size` cronjob 在与预留相同的行锁下从 live blob 同步 `quota_usages('global')`。
- 对账先锁 scope 行（用一次 no-op dirty update，兼容不支持 `FOR UPDATE` 的方言），
  再在同一个事务里修复归属、写回 size 并清 dirty；sqlite3 使用 `_txlock=immediate`。
- tag/repository 数量限制用 `COUNT(*)` + 行锁；`size_limit` 只能调大。
- UI 文案把 size 标为“去重后”，首页存储标为物理用量。
