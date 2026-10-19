---
title: CLI & AI Agents
---

# CLI & AI Agents

Sigma exposes two ways to operate a remote server from a terminal or an AI
agent. Both talk to the same HTTP API, so anything you can do in the Sigma web
UI can be scripted.

| Path | Best for |
| --- | --- |
| `sigma-cli` | Humans, CI pipelines, and shell-capable agents |
| `sigma-api-operator` skill | AI agents that need to discover APIs beyond what the CLI covers |

There is no embedded MCP server. Agents that cannot execute commands should use
a thin HTTP client against the REST API described by the OpenAPI document.

## Build the CLI

```bash
make build-cli
./bin/sigma-cli --help
```

## Login

`sigma-cli` stores Basic Auth credentials in `~/.sigma/cli-credentials.json`
with `0600` permissions. Override the location with `--credentials-file` or
`SIGMA_CREDENTIALS_FILE`.

```bash
./bin/sigma-cli login \
  --server https://sigma.example.com \
  --username sigma \
  --password 'Admin@123'
```

Credentials can also be supplied without a credentials file through the
persistent flags or environment variables:

```bash
export SIGMA_SERVER=https://sigma.example.com
export SIGMA_USERNAME=sigma
export SIGMA_PASSWORD='Admin@123'
```

Global flags:

| Flag | Env | Default |
| --- | --- | --- |
| `--server` | `SIGMA_SERVER` | `http://localhost:3000` |
| `--credentials-file` | `SIGMA_CREDENTIALS_FILE` | `~/.sigma/cli-credentials.json` |
| `--timeout` | – | `30s` |
| `--debug` | – | `false` |

## Command Reference

### Namespace

```bash
sigma-cli namespace list --page 1 --limit 10
sigma-cli namespace get <namespace_id>
sigma-cli namespace create library --visibility public --description "Base images"
sigma-cli namespace update <namespace_id> --size-limit 107374182400
sigma-cli namespace delete <namespace_id>
```

### Repository

```bash
sigma-cli repo list <namespace_id>
sigma-cli repo get <namespace_id> <repository_id>
sigma-cli repo create <namespace_id> library/alpine --visibility public
sigma-cli repo update <namespace_id> <repository_id> --description "Alpine images"
sigma-cli repo delete <namespace_id> <repository_id>
```

### Tag

```bash
sigma-cli tag list <namespace_id> <repository_id> --type image
sigma-cli tag get <namespace_id> <repository_id> <tag_id>
sigma-cli tag delete <namespace_id> <repository_id> <tag_id>
```

### Artifact

Artifact commands take namespace and repository **names** rather than IDs.

```bash
sigma-cli artifact list <namespace_name> <repository_name>
sigma-cli artifact get <namespace_name> <repository_name> <digest>
sigma-cli artifact delete <namespace_name> <repository_name> <digest>
```

### Member

```bash
sigma-cli member list <namespace_id>
sigma-cli member add <namespace_id> <user_id> --role reader
sigma-cli member update <namespace_id> <user_id> --role admin
sigma-cli member delete <namespace_id> <user_id>
```

### User

```bash
sigma-cli user list --without-admin
sigma-cli user create --username alice --password '...' --email alice@example.com --role user
sigma-cli user update <user_id> --role admin --status enabled
```

All commands print JSON on success and a formatted error body on failure, so
output composes with `jq`:

```bash
sigma-cli namespace list --limit 100 | jq -r '.items[].name'
```

## AI Agents

`tools/skill/sigma-api-operator` packages a skill for CLI agents such as Claude
Code or Copilot CLI:

```text
tools/skill/sigma-api-operator/SKILL.md
tools/skill/sigma-api-operator/swagger.yaml
```

The skill instructs the agent to load the OpenAPI document before issuing
requests. This is required for resources the CLI does not wrap yet — builders,
daemons, webhooks, validators, code repositories, and tokens — which must be
called through the REST API.

Refresh the packaged specification after changing handlers:

```bash
make swagen
```

## Safety

- Destructive commands (`delete`, credential and privilege changes) require
  explicit confirmation by the agent.
- Prefer HTTPS for remote servers; Basic Auth credentials travel with every
  request.
- Do not print passwords, tokens, private keys, or `Authorization` headers.
- Redact sensitive fields before surfacing non-2xx response bodies.
