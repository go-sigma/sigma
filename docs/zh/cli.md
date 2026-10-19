---
title: CLI 与 AI Agent
---

# CLI 与 AI Agent

Sigma 提供两种从终端或 AI agent 操作远程服务端的方式。两者都调用同一套 HTTP
API，因此 web 界面能做的事都可以脚本化。

| 方式 | 适用场景 |
| --- | --- |
| `sigma-cli` | 人、CI 流水线、可执行命令的 agent |
| `sigma-api-operator` skill | 需要访问 CLI 尚未覆盖的 API 的 AI agent |

Sigma 不再内置 MCP server。无法执行命令的客户端应使用轻量 HTTP 客户端调用
OpenAPI 文档描述的 REST API。

## 构建 CLI

```bash
make build-cli
./bin/sigma-cli --help
```

## 登录

`sigma-cli` 将 Basic Auth 凭据保存在 `~/.sigma/cli-credentials.json`，权限为
`0600`。可通过 `--credentials-file` 或 `SIGMA_CREDENTIALS_FILE` 修改位置。

```bash
./bin/sigma-cli login \
  --server https://sigma.example.com \
  --username sigma \
  --password 'Admin@123'
```

也可以不落盘，直接用全局参数或环境变量提供凭据：

```bash
export SIGMA_SERVER=https://sigma.example.com
export SIGMA_USERNAME=sigma
export SIGMA_PASSWORD='Admin@123'
```

全局参数：

| 参数 | 环境变量 | 默认值 |
| --- | --- | --- |
| `--server` | `SIGMA_SERVER` | `http://localhost:3000` |
| `--credentials-file` | `SIGMA_CREDENTIALS_FILE` | `~/.sigma/cli-credentials.json` |
| `--timeout` | – | `30s` |
| `--debug` | – | `false` |

## 命令参考

### 命名空间

```bash
sigma-cli namespace list --page 1 --limit 10
sigma-cli namespace get <namespace_id>
sigma-cli namespace create library --visibility public --description "基础镜像"
sigma-cli namespace update <namespace_id> --size-limit 107374182400
sigma-cli namespace delete <namespace_id>
```

### 仓库

```bash
sigma-cli repo list <namespace_id>
sigma-cli repo get <namespace_id> <repository_id>
sigma-cli repo create <namespace_id> library/alpine --visibility public
sigma-cli repo update <namespace_id> <repository_id> --description "Alpine 镜像"
sigma-cli repo delete <namespace_id> <repository_id>
```

### Tag

```bash
sigma-cli tag list <namespace_id> <repository_id> --type image
sigma-cli tag get <namespace_id> <repository_id> <tag_id>
sigma-cli tag delete <namespace_id> <repository_id> <tag_id>
```

### 制品

制品相关命令使用命名空间和仓库的**名称**，而不是 ID。

```bash
sigma-cli artifact list <namespace_name> <repository_name>
sigma-cli artifact get <namespace_name> <repository_name> <digest>
sigma-cli artifact delete <namespace_name> <repository_name> <digest>
```

### 成员

```bash
sigma-cli member list <namespace_id>
sigma-cli member add <namespace_id> <user_id> --role reader
sigma-cli member update <namespace_id> <user_id> --role admin
sigma-cli member delete <namespace_id> <user_id>
```

### 用户

```bash
sigma-cli user list --without-admin
sigma-cli user create --username alice --password '...' --email alice@example.com --role user
sigma-cli user update <user_id> --role admin --status enabled
```

所有命令成功时输出 JSON，失败时输出格式化后的错误响应体，因此可以直接和
`jq` 组合使用：

```bash
sigma-cli namespace list --limit 100 | jq -r '.items[].name'
```

## AI Agent

`tools/skill/sigma-api-operator` 为 Claude Code、Copilot CLI 等 CLI agent 打包了
一个 skill：

```text
tools/skill/sigma-api-operator/SKILL.md
tools/skill/sigma-api-operator/swagger.yaml
```

该 skill 要求 agent 在发起请求前先加载 OpenAPI 文档。对于 CLI 尚未覆盖的资源
——builder、daemon、webhook、validator、代码仓库和 token——必须通过 REST API
调用。

修改 handler 后刷新打包的接口文档：

```bash
make swagen
```

## 安全建议

- 破坏性命令（`delete`、凭据与权限变更）必须由 agent 明确确认后执行。
- 远程服务端优先使用 HTTPS；Basic Auth 凭据会随每个请求发送。
- 不要打印密码、token、私钥或 `Authorization` 头。
- 返回非 2xx 响应体时，先脱敏再展示给用户。
