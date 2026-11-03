# Home 仪表盘：按用户范围展示的设计方案

> 状态：设计稿（未实现）
> 目标读者：后端 / 前端开发
> 相关页面：`web/src/pages/Home/`

## 1. 背景

当前首页 `Home` 用的是写死的样例数据（`web/src/pages/Home/mock.ts`），而且语义上被当成了"全站概览"。实际上首页应该是**当前登录用户的工作台**：

- 普通用户只应该看到**自己**的活动、**自己有管理权限**的命名空间、**自己配置的**构建记录；
- 管理员（`Root` / `Admin`）才可以看全站视角；
- 普通用户不存在"全站磁盘总量"这个概念，所以不应该出现 `14.2% of 100 GiB used` 这种占比。

本方案定义每张卡片在两种身份下的数据口径、后端需要的改造（配置、表结构、查询、接口）以及前端如何消费。

## 2. 术语与角色

| 概念 | 定义 | 代码位置 |
|---|---|---|
| 管理员 | `users.role in ('Root', 'Admin')` | `pkg/api/enums` → `UserRoleRoot` / `UserRoleAdmin` |
| 普通用户 | `users.role = 'User'` | `UserRoleUser` |
| 命名空间成员角色 | `NamespaceAdmin` / `namespace_manager` / `NamespaceReader` | `namespace_members.role` |
| 管理权限 | 成员角色为 `NamespaceAdmin` 或 `namespace_manager`；`NamespaceReader` 只读，**不算**。管理员对所有命名空间都有管理权限 | `pkg/authz/permission.go` → `CheckRole` |
| 作用域 scope | `global`（全站，仅管理员）/ `self`（当前用户） | 本方案新增概念 |

> 为什么要区分 `NamespaceReader`：只读成员能看仓库，但不能推镜像、不能改构建配置。如果首页把他们管理的"只读命名空间"统计进去，会出现"我什么都没做但首页数字在涨"的误导。

## 3. 目标行为总览

| 卡片 | 管理员（scope=global） | 普通用户（scope=self） |
|---|---|---|
| Storage used | 全站 `SUM(namespaces.size)` / 配置容量 `storage.capacity`，显示百分比和进度条 | 我管理的命名空间 `SUM(size)`，**不显示百分比/进度条** |
| 命名空间 / 仓库 / Tag 计数 | 全站 | 我管理的命名空间范围 |
| Namespace usage | 全站命名空间，按占用倒序 | 我管理的命名空间，按占用倒序 |
| Push activity（热力图） | 全站推送 | 我自己的推送 |
| Pulls and pushes（趋势） | 全站推送 + 拉取 | 我自己的推送 + 拉取 |
| Builds（构建折线图） | 全站构建记录 | 我配置的 builder（或我触发的构建）产生的记录 |

### 3.1 为什么普通用户没有存储占比

`14.2 / 100 GiB` 的分子是"全站已用"，分母是"全站容量"。普通用户既看不到全站，也不应该知道全站容量，所以：

- 个人视角下 `storage.capacity` 返回 `0`，前端据此隐藏进度条和百分比，只展示"已用 3.2 GiB"；
- 如果确实想给个人视角一个分母，只能用"我管理的命名空间配额之和"（`SUM(size_limit)`，`0` 表示不限），但这是**多个独立配额的和**，语义上不是"总空间"，建议默认不做。可以作为后续可选项。

## 4. 逐卡片实现方案

### 4.1 Storage used

**数据源**：`namespaces.size` 是推送/删除时增量维护的缓存列，不需要扫描 `blobs`。

管理员：

```sql
SELECT COALESCE(SUM(size), 0) AS used
FROM namespaces
WHERE deleted_at = 0;
```

普通用户：

```sql
SELECT COALESCE(SUM(n.size), 0) AS used
FROM namespaces n
JOIN namespace_members m
  ON m.namespace_id = n.id
 AND m.user_id = ?
 AND m.deleted_at = 0
WHERE n.deleted_at = 0
  AND m.role IN ('namespace_manager', 'NamespaceAdmin');
```

**容量**来自新增配置项（见 §5.1）：

```yaml
storage:
  # sigma 管理的存储总量（字节）。只用于管理员首页展示总用量占比，0 表示不限制/不展示。
  capacity: 109951162777600   # 100 TiB
```

返回给前端时：

- 管理员：`{"used": 12884901888, "capacity": 109951162777600}`
- 普通用户：`{"used": 3435973836, "capacity": 0}`

### 4.2 命名空间 / 仓库 / Tag 计数

直接复用 `namespaces` 上的三个缓存列（`repository_count`、`tag_count`、`size`），避免 `COUNT(*)` 扫大表。

管理员：

```sql
SELECT COUNT(*)                          AS namespaces,
       COALESCE(SUM(repository_count),0) AS repositories,
       COALESCE(SUM(tag_count), 0)       AS tags,
       COALESCE(SUM(size), 0)            AS used
FROM namespaces
WHERE deleted_at = 0;
```

普通用户：同一份 SQL 加上 §4.1 的 `namespace_members` 过滤。

### 4.3 Namespace usage（列表）

管理员：全站按占用倒序；普通用户：我管理的命名空间。

```sql
SELECT n.id, n.name, n.visibility, n.size, n.size_limit,
       n.repository_count, n.tag_count
FROM namespaces n
JOIN namespace_members m
  ON m.namespace_id = n.id AND m.user_id = ? AND m.deleted_at = 0
WHERE n.deleted_at = 0
  AND m.role IN ('namespace_manager', 'NamespaceAdmin')   -- 管理员时删掉这段
ORDER BY n.size DESC
LIMIT 10;
```

前端行点击仍然跳到 `/namespaces/{name}/repositories`（`useNamespaceId` 会按名字反查 id）。

### 4.4 Push activity（热力图）

现状：`GET /api/v1/users/{user_id}/activity/heatmap?days=365` 只统计**推送**，并且"本人或管理员"才能查。

改造后新增统一接口（见 §6.2）：

- `scope=self`：`user_activity_hourlies WHERE user_id = ?`（已有数据）
- `scope=global`：`namespace_activity_hourlies` 全表按小时求和（已有数据）

注意 `hour` 列是 `YYYYMMDDHH` 的整数（`pkg/service/analytics/analytics.go:346`），按天聚合用 `hour / 100`。

```sql
-- self
SELECT hour / 100 AS day, SUM(push_count) AS push
FROM user_activity_hourlies
WHERE user_id = ? AND hour >= ? AND hour <= ?
GROUP BY day ORDER BY day;

-- global（管理员）
SELECT hour / 100 AS day, SUM(push_count) AS push
FROM namespace_activity_hourlies
WHERE hour >= ? AND hour <= ?
GROUP BY day ORDER BY day;
```

### 4.5 Pulls and pushes（趋势）

**这是唯一一个必须改表的地方**：`user_activity_hourlies` 目前只有 `push_count`，没有拉取计数，`RecordPull` 也只写了命名空间维度。

- `scope=global`：`namespace_activity_hourlies` 的 `push_count` / `pull_count` 求和（已有数据）。
- `scope=self`：需要给 `user_activity_hourlies` 增加 `pull_count`，并在 `RecordPull` 里写用户维度（`PullEvent.UserID` 字段已经存在，只是当前被忽略）。

```sql
-- self
SELECT hour / 100 AS day, SUM(push_count) AS push, SUM(pull_count) AS pull
FROM user_activity_hourlies
WHERE user_id = ? AND hour >= ? AND hour <= ?
GROUP BY day ORDER BY day;
```

口径说明：按"**操作人**"统计，即"我拉取了哪些镜像"，而不是"别人拉取了我的镜像"。匿名拉取（`UserID` 为空）不计入个人维度。

### 4.6 Builds（构建折线图）

现状：`builder_runners` 只挂在 `builders` 上，`builders` 只挂在 `repositories` 上，**没有任何用户字段**。所以"我的构建"有两种口径：

**方案 A（不改表，先能跑）**：我管理的命名空间下的 builder 产生的记录。

```sql
SELECT COALESCE(r.started_at, r.created_at) / 86400000 AS day,
       SUM(CASE WHEN r.status = 'Success' THEN 1 ELSE 0 END) AS success,
       SUM(CASE WHEN r.status = 'Failed'  THEN 1 ELSE 0 END) AS failed
FROM builder_runners r
JOIN builders   b ON b.id = r.builder_id AND b.deleted_at = 0
JOIN repositories p ON p.id = b.repository_id AND p.deleted_at = 0
JOIN namespaces  n ON n.id = p.namespace_id AND n.deleted_at = 0
JOIN namespace_members m
  ON m.namespace_id = n.id AND m.user_id = ? AND m.deleted_at = 0
WHERE r.deleted_at = 0
  AND m.role IN ('namespace_manager', 'NamespaceAdmin')
  AND COALESCE(r.started_at, r.created_at) >= ?
GROUP BY day ORDER BY day;
```

**方案 B（推荐，语义准确）**：新增两个归属字段。

| 表 | 新字段 | 语义 |
|---|---|---|
| `builders` | `creator_id varchar(36)` | 谁配置的这个 builder |
| `builder_runners` | `triggered_by varchar(36)` | 这次构建是谁触发的（手动构建=触发人；cron/webhook 触发时为空） |

查询时用 `COALESCE(r.triggered_by, b.creator_id) = ?` 归属到用户，管理员去掉这个条件。

- 手动构建：`triggered_by` = 当前登录用户；
- `cron` / `webhook` 触发：没有"操作人"，`triggered_by` 留空，回退到 `creator_id`；
- 管理员视角：全站，不受限制。

`started_at` 是毫秒时间戳，`Pending` 状态的记录为空，所以聚合时用 `COALESCE(started_at, created_at)`。

性能：`builder_runners` 目前只有主键索引，需要补 `(builder_id, started_at)` 和 `(started_at)` 索引（见 §5.2）。

## 5. 后端改造清单

### 5.1 新增配置：存储总容量

文件：`pkg/config/configuration.go`

```go
// ConfigurationStorage
type ConfigurationStorage struct {
    RootDirectory string `yaml:"rootDirectory"`
    Redirect      bool   `yaml:"redirect"`
    // Capacity is the total storage sigma is allowed to use, in bytes.
    // 0 means unlimited; it is only used to render the usage ratio for administrators.
    Capacity int64 `yaml:"capacity"`
    Type     enums.StorageType `yaml:"type"`
    // ...
}
```

同步修改：

- `conf/config.yaml`、`conf/config-dev.yaml`、`conf/config-compose.yaml`、`conf/config-full.yaml`
- `docs/configuration.md`、`docs/zh/configuration.md`
- `deploy/sigma/values.yaml` + `deploy/sigma/templates/configmap.yaml`

> 备选位置是新增 `quota.capacity`。放在 `storage` 下的理由是它描述的就是存储后端的容量，且已有 `storage.redirect` 这类全局开关。

### 5.2 数据模型变更（迁移）

新增三个字段，一个索引。用 `make migration-create MIGRATION_NAME=home_dashboard_scope` 生成四种数据库的迁移，然后 `make gormgen` 重新生成 `pkg/dal/query`。

| 表 | 变更 | 原因 |
|---|---|---|
| `user_activity_hourlies` | `pull_count bigint NOT NULL DEFAULT 0` | 个人维度拉取统计 |
| `builders` | `creator_id varchar(36)` | 方案 B：构建归属 |
| `builder_runners` | `triggered_by varchar(36)` | 方案 B：构建归属 |
| `builder_runners` | `INDEX (builder_id, started_at)`、`INDEX (started_at)` | 折线图按时间范围聚合 |

同时更新 `pkg/dal/models/analytics.go`、`pkg/dal/models/builder.go`、mocks 与 spanner 测试数据（如果字段用于 GC fixture 需要同步）。

### 5.3 采集端改造

`pkg/service/analytics/analytics.go`：

1. `RecordPull`：当 `event.UserID != ""` 时，同时写 `user_activity_hourlies` 的 `pull_count`；
2. `IncrUserPush` → 扩展成 `IncrUser(userID, hour, pushDelta, pullDelta)`（repository 层同名方法，`pkg/dal/repository/analytics/analytics.go`）；
3. `flushHour`：把 `pull_count` 一起落库；
4. 保留现有 `namespace_activity_hourlies` 的写入逻辑不变（管理员视角继续用）。

`pkg/background/build/coordinator.go`：创建 `BuilderRunner` 时写入 `triggered_by`（手动构建从 handler 的 context 取用户；cron/webhook 留空）。

### 5.4 新增仓储查询

新增文件：`pkg/dal/repository/overview/`（或放进 `analytics` 包，按域划分更清晰）。

需要的方法：

```go
type OverviewRepository interface {
    // 管理员：全站；普通用户：传入 managedNamespaceIDs 做过滤
    SumNamespaceUsage(ctx context.Context, namespaceIDs []string) (api.ResourceUsage, error)
    TopNamespaces(ctx context.Context, namespaceIDs []string, limit int) ([]*models.Namespace, error)
    // 活动
    ListUserActivity(ctx context.Context, userID string, startHour, endHour int64) ([]*models.UserActivityHourly, error)
    ListGlobalActivity(ctx context.Context, startHour, endHour int64) ([]*models.NamespaceActivityHourly, error)
    // 构建
    ListBuildActivity(ctx context.Context, userID string, since int64) ([]api.DailyBuildCount, error)
}
```

其中 `namespaceIDs` 为空切片表示"全站"（管理员），非空表示按 id 过滤；这样 service 层只负责算权限，SQL 只负责过滤，避免在 SQL 里拼角色判断。

### 5.5 新增 service / handler

- `pkg/service/overview/`：`Service` 接口 + 实现 + mocks，负责：
  1. 读当前用户，判断 `scope`（`Root`/`Admin` → `global`，否则 `self`）；
  2. 普通用户先查"我管理的命名空间 id 列表"（复用 `pkg/dal/repository/namespace`，新增方法 `ListManagedNamespaceIDs(ctx, userID)`）；
  3. 聚合各卡片数据；
  4. 管理员才填 `storage.capacity`。
- `pkg/server/handlers/overview/`：
  - `handler.go`：`Handler` 接口 + `handler struct { dig.In }` + `factory{}`（与 `pkg/server/handlers/systems` 一致）；
  - `overview.go`、`activity.go`、`builds.go` 三个 action 文件；
  - 路由注册在 `consts.APIV1` 下。

### 5.6 权限校验

| 接口 | 规则 |
|---|---|
| `GET /overview` | 登录即可，内部按角色决定 scope |
| `GET /activity/daily` | `scope=self` 固定当前用户；`scope=global` 要求 `Root`/`Admin`，否则 403 |
| `GET /builds/daily` | 同上 |
| `GET /namespaces/top` | 复用 `Authorizer.Namespace(...)` 逐个校验太重，改为"只返回我能管理的"，天然安全 |

统一用 `errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized/Forbidden, ...)` 返回。

### 5.7 性能与缓存

- 计数全部走 `namespaces` 的缓存列，不扫 `artifacts` / `tags`；
- 首页接口结果放 Redis（`cache` 已有实现）：key `sigma:overview:{userID}`，TTL 30s；管理员 key 为 `sigma:overview:global`；
- 构建折线图数据量小（30 天），一次 SQL 即可；
- 热力图 365 天 ≈ 8760 行/用户，SQL 用 `GROUP BY hour/100`，返回 365 个点。

## 6. 接口契约

### 6.1 `GET /api/v1/overview`

```json
{
  "scope": "self",
  "storage": { "used": 3435973836, "capacity": 0 },
  "counts": { "namespaces": 3, "repositories": 12, "tags": 40 },
  "activity": {
    "push": 88,
    "pull": 1024,
    "pushDelta": 12.4,
    "pullDelta": -3.1
  },
  "namespaceUsage": [
    {
      "id": "01a1...",
      "name": "library",
      "visibility": "public",
      "repositories": 8,
      "tags": 96,
      "size": 5430000000,
      "sizeLimit": 21474836480
    }
  ]
}
```

- `scope` 决定前端渲染：`self` 时忽略 `storage.capacity`，隐藏进度条；
- `pushDelta` / `pullDelta` 是与上一个等长周期比较的百分比，后端算好，避免前端再拼一次请求；
- `namespaceUsage` 最多 10 条（管理员取全站 Top 10）。

### 6.2 `GET /api/v1/activity/daily?days=365&scope=self`

```json
{ "items": [ { "date": "2026-11-03", "push": 5, "pull": 120 } ] }
```

`scope=global` 仅管理员可用。这个接口可以同时替代现有 `/users/{id}/activity/heatmap`（保留旧接口一个版本做兼容，内部转调新实现）。

### 6.3 `GET /api/v1/builds/daily?days=30&scope=self`

```json
{
  "items": [ { "date": "2026-11-03", "success": 7, "failed": 1 } ],
  "summary": { "success": 112, "failed": 11, "successRate": 91.1, "fastest": 120, "slowest": 389 }
}
```

## 7. 前端改造

文件：`web/src/pages/Home/`

1. 删除 `mock.ts`，改为请求 `/api/v1/overview`、`/api/v1/activity/daily`、`/api/v1/builds/daily`；
2. `Storage used`：
   - `storage.capacity > 0` 才渲染进度条和百分比；
   - 否则只显示 `used`，并把提示文案换成"已用存储"；
3. `Namespace usage`：数据来自 `namespaceUsage`，行点击跳转保持不变；
4. `Push activity` / `Pulls and pushes`：`scope` 只影响数据来源，图表本身不变；
5. `Builds`：数据来自 `/builds/daily`，保留 7/30 天切换（切换时只重取这个接口）；
6. 空状态：
   - 没有任何命名空间 → 首页展示"创建第一个命名空间"引导；
   - 没有构建记录 → `Builds` 卡片显示"还没有构建记录"；
7. i18n：新增 `home.scope.global` / `home.scope.self` 之类的标识，让用户知道看到的是"我的"还是"全站"。

## 8. 实施顺序

| 阶段 | 内容 | 交付物 |
|---|---|---|
| P0 | 配置项 `storage.capacity` + 文档 | 配置生效，`/systems/config` 可读 |
| P1 | `user_activity_hourlies.pull_count` + 采集端改造 | 个人维度拉取数据开始积累 |
| P2 | `overview` 仓储 / service / handler + `/overview` | 首页卡片可接真实数据（管理员视角先通） |
| P3 | `/activity/daily`、`/builds/daily` | 两张图表可接真实数据 |
| P4 | `builders.creator_id` / `builder_runners.triggered_by` + 归属逻辑 | "我的构建"语义准确 |
| P5 | 前端接接口、删 mock、空状态、文案 | 首页上线 |

> P1 越早做越好：拉取数据只能向前积累，历史数据无法回填（除非从审计日志重放）。

## 9. 测试计划

- 仓储层：`pkg/dal/repository/overview` 的四种数据库测试（sqlite3 / turso / mysql / postgresql），覆盖"管理员=全站、普通用户=管理范围、只读成员被排除"。
- Service 层：mock repository，覆盖 `scope` 判定、`capacity` 只对管理员返回。
- Handler 层：`testkit.NewGin()` + dig 容器，覆盖 403（普通用户请求 `scope=global`）。
- 前端：无自动化测试基建，手工用例——管理员/普通用户各登录一次，逐个卡片核对数字与数据库 `SELECT` 一致。

## 10. 待确认问题

1. **个人视角的"我的活动"**：按"我执行的操作"还是"发生在我命名空间上的操作"？本文按前者，后者更适合团队负责人。
2. **个人存储是否要分母**：用 `SUM(size_limit)` 作为分母，还是完全不显示百分比？（本文默认不显示）
3. **构建归属**：是否接受新增 `creator_id` / `triggered_by` 两个字段？如果不接受，只能退化成"我管理命名空间下的构建"。
4. **匿名拉取**：个人视角是否统计匿名拉取（无法归属到人）？默认不统计。
