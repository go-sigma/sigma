---
name: "sigma-api-operator"
description: "Operates Sigma resources through the current OpenAPI spec. Invoke when external CLI agents need to inspect or manage a Sigma server."
---

# Sigma API Operator

Use this skill when an external CLI agent, such as Claude Code, Copilot CLI, or a
similar automation tool, needs to inspect or manage Sigma resources through the
Sigma HTTP API.

This skill is distributed with `swagger.yaml`, which describes the Sigma REST API
surface available to the agent.

## API Specification

Before calling Sigma APIs, load the packaged OpenAPI document:

```text
swagger.yaml
```

When working from a newer Sigma source checkout, refresh the packaged document:

```bash
go tool swag init --output tools/skill/sigma-api-operator --outputTypes yaml
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
| Code repositories | `/coderepos/*` and `/{provider}/repos/coderepos/*`                                                           |

## Connection Settings

Use these environment variables when available:

```text
SIGMA_SERVER=https://sigma.example.com
SIGMA_USERNAME=sigma
SIGMA_PASSWORD=...
```

Normalize `SIGMA_SERVER` by trimming a trailing slash before building request
URLs.

Sigma management APIs use Basic Auth. Send Basic Auth on each request:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  "${SIGMA_SERVER}/api/v1/..."
```

Do not print passwords, tokens, private keys, access keys, secret keys, or raw
Authorization headers.

## REST Operation Workflow

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

Always confirm the exact path, method, and schema in `swagger.yaml` before
executing these examples.

List namespaces:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  "${SIGMA_SERVER}/api/v1/namespaces/?page=1&limit=10"
```

Create a namespace:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  --header "Content-Type: application/json" \
  --request POST \
  --data '{"name":"library","visibility":"public"}' \
  "${SIGMA_SERVER}/api/v1/namespaces/"
```

List repositories in a namespace:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  "${SIGMA_SERVER}/api/v1/namespaces/${NAMESPACE_ID}/repositories/?page=1&limit=10"
```

Create a repository:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  --header "Content-Type: application/json" \
  --request POST \
  --data '{"name":"library/alpine","description":"Alpine images"}' \
  "${SIGMA_SERVER}/api/v1/namespaces/${NAMESPACE_ID}/repositories/"
```

List tags:

```bash
curl --fail-with-body --silent --show-error \
  --user "${SIGMA_USERNAME}:${SIGMA_PASSWORD}" \
  --header "Accept: application/json" \
  "${SIGMA_SERVER}/api/v1/namespaces/${NAMESPACE_ID}/tags/?repository_id=${REPOSITORY_ID}&page=1&limit=10"
```

## Safety Rules

- Never rely on stale API knowledge if `swagger.yaml` can be generated or read.
- Never write credentials to files unless the user explicitly asks for it.
- Do not run destructive API calls from ambiguous natural-language requests.
- Prefer HTTPS for remote Sigma servers.
- Surface non-2xx response bodies to the user after redacting sensitive fields.
