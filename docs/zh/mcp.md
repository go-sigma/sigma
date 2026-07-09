---
title: MCP Server
---

Sigma 可以在现有 `sigma server` 进程中内嵌 MCP server，用于 AI 客户端和自动化工具访问 Sigma 的管理能力。
MCP 默认关闭，需要显式开启。

## 暴露哪些能力

MCP server 只暴露管理面和元数据能力，不暴露 OCI Blob 上传/下载流量。

当前已支持的 tools：

- 系统：`system-version-get`、`system-endpoint-get`、`system-config-summary-get`
- Namespace：`namespace-list`、`namespace-get`、`namespace-create`、`namespace-update`、`namespace-delete`
- Namespace 成员：`namespace-member-list`、`namespace-member-add`、`namespace-member-update`、`namespace-member-delete`、`namespace-member-self-get`
- Repository：`repository-list`、`repository-get`、`repository-create`、`repository-update`、`repository-delete`
- Tag：`tag-list`、`tag-get`、`tag-delete`、`tag-manifest-raw-get`
- Artifact：`artifact-list`、`artifact-get`、`artifact-delete`

只读 resources：

- `sigma://system/version`
- `sigma://system/capabilities`

## 开启 MCP

在配置文件中设置 `mcp`：

```yaml
mcp:
  enabled: true
  path: /api/v1/mcp
  transport: streamable_http
  auditWrites: true
  toolTimeout: 30s
  maxRequestBody: 1048576
```

使用 Helm 部署时：

```bash
helm upgrade --install sigma ./deploy/sigma \
  --set config.mcp.enabled=true
```

MCP endpoint 复用现有 server service：

```text
https://<sigma-host>/api/v1/mcp
```

## 认证方式

MCP 使用 Sigma Basic Auth。每个 MCP 请求都必须携带：

```text
Authorization: Basic <base64(username:password)>
```

例如：

```bash
printf 'sigma:Admin@123' | base64
```

将编码后的值写入 MCP client 配置：

```json
{
  "mcpServers": {
    "sigma": {
      "url": "https://<sigma-host>/api/v1/mcp",
      "headers": {
        "Authorization": "Basic <base64(username:password)>"
      }
    }
  }
}
```

Basic Auth 会在每次请求中发送凭据，因此 MCP 必须通过 HTTPS 或具备 TLS 终止能力的网关暴露。

## 权限模型

MCP tool 会以认证后的 Sigma 用户身份执行，并复用 Sigma HTTP API 的资源权限模型：

- Root 和 admin 用户可以访问所有 MCP tools
- Namespace reader 可以读取 namespace 范围内的资源
- Namespace manager 可以管理 repository、tag 和 artifact
- Namespace admin 可以管理 namespace 元数据和成员

没有权限的调用会返回 tool error，不会执行对应操作。

## 审计和指标

当 `mcp.auditWrites` 开启时，写操作会记录到 audit 表。
审计字段使用：

- `method`: `MCP`
- `path`: MCP endpoint path
- `route`: MCP tool name

写入审计前会过滤敏感字段，例如 `password`、`token`、`privateKey`、`ak`、`sk`、`secret`、`authorization`。

MCP 指标复用现有 `/metrics` endpoint 暴露：

- `sigma_mcp_tool_calls_total`
- `sigma_mcp_tool_duration_seconds`
- `sigma_mcp_auth_failures_total`

## Tool 参数示例

列出 namespace 下的 repositories：

```json
{
  "namespace_id": "550e8400-e29b-41d4-a716-446655440000",
  "page": 1,
  "limit": 10
}
```

创建 repository：

```json
{
  "namespace_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "library/alpine",
  "description": "Alpine images"
}
```

获取 tag manifest：

```json
{
  "repository_id": "550e8400-e29b-41d4-a716-446655440000",
  "digest": "sha256:..."
}
```
