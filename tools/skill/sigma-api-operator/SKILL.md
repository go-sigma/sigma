---
name: "sigma-api-operator"
description: "Operates Sigma resources through sigma-cli, falling back to the Sigma OpenAPI spec. Invoke when external CLI agents need to inspect or manage a Sigma server."
---

# Sigma API Operator

Use this skill when an external CLI agent, such as Claude Code, Copilot CLI, or a
similar automation tool, needs to inspect or manage Sigma resources.

There are two ways to talk to Sigma. Prefer the CLI; use the HTTP API only for
resources the CLI does not wrap.

| Path | Use for |
| --- | --- |
| `sigma-cli` | namespaces, repositories, tags, artifacts, members, users |
| HTTP API + `swagger.yaml` | builders, daemons, webhooks, validators, code repositories, tokens |

This skill is distributed with `swagger.yaml`, which describes the Sigma REST API
surface available to the agent.

## Preferred: sigma-cli

Prefer `sigma-cli` whenever it covers the operation. It handles authentication
from a credentials file, prints JSON, and keeps the agent away from hand-built
URLs.

```bash
sigma-cli login --server "${SIGMA_SERVER}" --username "${SIGMA_USERNAME}" --password "${SIGMA_PASSWORD}"
```

Commands accept IDs (and, for artifacts, names) as positional arguments:

```bash
sigma-cli namespace list --page 1 --limit 10
sigma-cli namespace get <namespace_id>
sigma-cli namespace create library --visibility public
sigma-cli repo list <namespace_id>
sigma-cli repo create <namespace_id> library/alpine
sigma-cli tag list <namespace_id> <repository_id>
sigma-cli artifact list <namespace_name> <repository_name>
sigma-cli member add <namespace_id> <user_id> --role reader
sigma-cli user list --without-admin
```

Every command prints JSON on success and a formatted error body on failure, so
output composes with `jq`:

```bash
sigma-cli namespace list --limit 100 | jq -r '.items[].name'
```

Run `sigma-cli <resource> --help` to confirm the flags of a subcommand before
using it. Do not guess flags.

If a resource is not listed in `sigma-cli --help`, it is not implemented in the
CLI yet. Use the HTTP API workflow below instead.

## API Specification

Before calling Sigma APIs, load the packaged OpenAPI document:

```text
swagger.yaml
```

Use the OpenAPI document as the source of truth for:

- Available paths and HTTP methods.
- Path, query, and body parameters.
- Request and response JSON schemas.
- Required fields and enum values.

## Finding APIs in `swagger.yaml`

`swagger.yaml` is large. Do not read the whole file into the model context.
Inspect it with line-oriented search tools and then read only the matching
operation block.

The file is organized in this order:

1. `basePath`: the API prefix, currently `/api/v1`.
2. `definitions`: reusable request and response schemas.
3. `paths`: REST endpoints grouped by path.
4. `securityDefinitions`: authentication scheme metadata.

Start from `paths`, not `definitions`, when looking for an operation:

```bash
rg -n "^paths:|^  /" swagger.yaml
```

Search by resource prefix when the target resource is known:

```bash
rg -n "^  /(namespaces|users|systems|tokens|validators|webhooks|coderepos|daemons)" swagger.yaml
rg -n "^  /namespaces/.*repositories|^  /namespaces/.*tags|builders" swagger.yaml
```

Search by natural-language action when the target path is unclear:

```bash
rg -n "summary: .*namespace|summary: .*repository|summary: .*tag|summary: .*webhook" swagger.yaml
```

After finding a path line, read a bounded range around it instead of loading the
whole file:

```bash
sed -n '3592,3680p' swagger.yaml
```

Within an operation block, inspect these fields:

- The HTTP method under the path, such as `get`, `post`, `put`, or `delete`.
- `summary` for the human-readable action.
- `tags` for the resource domain.
- `parameters` for path, query, and body inputs.
- `responses` for success and error payloads.
- `security` to confirm Basic Auth is required.

If a body parameter or response uses `$ref`, jump to the schema definition:

```bash
rg -n "^  api.PostNamespaceRequest:" swagger.yaml
sed -n '<start>,<end>p' swagger.yaml
```

Common resource prefixes:

| Resource area     | Path or search hint                                                                                          |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| System            | `/systems/config`, `/systems/endpoint`, `/systems/version`                                                   |
| Login and users   | `/users/`, `/users/login`, `/users/logout`                                                                   |
| Tokens            | `/tokens`                                                                                                    |
| Namespaces        | `/namespaces/`, `/namespaces/{namespace_id}`                                                                 |
| Namespace members | `/namespaces/{namespace_id}/members`                                                                         |
| Repositories      | `/namespaces/{namespace_id}/repositories`                                                                    |
| Tags              | `/namespaces/{namespace_id}/tags`                                                                            |
| Builders          | `builders` under namespace and repository paths                                                              |
| Daemons           | `/daemons/gc-*`                                                                                              |
| Webhooks          | `/webhooks/` and `/webhooks/{webhook_id}/logs`                                                               |
| Validators        | `/validators/cron`, `/validators/password`, `/validators/reference`, `/validators/regexp`, `/validators/tag` |
| Code repositories | `/coderepos/{provider}`                                                                                      |

## Connection Settings

Use these environment variables when available:

```text
SIGMA_SERVER=https://sigma.example.com
SIGMA_USERNAME=sigma
SIGMA_PASSWORD=...
```

Normalize `SIGMA_SERVER` by trimming a trailing slash before building request
URLs.

`sigma-cli` reads these on its own; for raw HTTP calls, send Basic Auth on each
request:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  "${SIGMA_SERVER}/api/v1/..."
```

Do not print passwords, tokens, private keys, access keys, secret keys, or raw
Authorization headers.

## HTTP Operation Workflow

Use this only when `sigma-cli` has no matching command.

1. Read `swagger.yaml`.
2. Find the operation that matches the user's resource and action.
3. Build the request URL from `SIGMA_SERVER`, `/api/v1`, and the OpenAPI path.
4. Fill path parameters, query parameters, and JSON request bodies from the
   OpenAPI operation.
5. Ask for explicit confirmation before destructive operations such as delete,
   privilege changes, credential changes, or bulk updates.
6. Return concise results with resource names, IDs, status codes, and relevant
   error bodies.

## Request Rules

- Use the method, path, query parameters, path parameters, and JSON body defined
  by the OpenAPI document.
- Set `Accept: application/json` for API calls.
- Set `Content-Type: application/json` when sending a JSON request body.
- Preserve UUID, digest, namespace, repository, tag, and artifact values exactly
  as provided by the user or returned by Sigma.
- For paginated list APIs, include explicit `page` and `limit` values unless the
  user asks for server defaults.
- For update APIs, send only fields requested by the user or required by the
  OpenAPI schema.

## Common Resource Workflows

The resources below are **not** implemented in `sigma-cli`, so call them over
HTTP. Always confirm the exact path, method, and schema in `swagger.yaml` before
executing an example.

Garbage collection rules are namespace scoped. Read the current rule:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  "${SIGMA_SERVER}/api/v1/daemons/gc-artifact/${NAMESPACE_ID}/"
```

Update the rule, then start a run:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  --header "Content-Type: application/json" \
  --request PUT \
  --data '{"cron_rule":"0 2 * * *","cron_enabled":true,"retention_day":7}' \
  "${SIGMA_SERVER}/api/v1/daemons/gc-artifact/${NAMESPACE_ID}/"

curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Content-Type: application/json" \
  --request POST \
  --data '{}' \
  "${SIGMA_SERVER}/api/v1/daemons/gc-artifact/${NAMESPACE_ID}/runners/"
```

List webhooks and their delivery logs:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  "${SIGMA_SERVER}/api/v1/webhooks/?page=1&limit=10"

curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  "${SIGMA_SERVER}/api/v1/webhooks/${WEBHOOK_ID}/logs/?page=1&limit=10"
```

Validate a value before saving it. Validators only check input; they do not read
server state, and they return `204 No Content` on success:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Content-Type: application/json" \
  --request POST \
  --data '{"cron":"0 2 * * *"}' \
  "${SIGMA_SERVER}/api/v1/validators/cron"
```

List code repositories for a provider such as `github` or `gitlab`:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  "${SIGMA_SERVER}/api/v1/coderepos/${PROVIDER}?page=1&limit=10"
```

Read the current user's access token metadata:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  "${SIGMA_SERVER}/api/v1/tokens"
```

## Safety Rules

- Never rely on stale API knowledge if `swagger.yaml` can be generated or read.
- Never write credentials to files unless the user explicitly asks for it.
- Do not run destructive API calls from ambiguous natural-language requests.
- Prefer HTTPS for remote Sigma servers.
- Surface non-2xx response bodies to the user after redacting sensitive fields.
