---
title: MCP Server
---

Sigma can expose an embedded MCP server for AI clients and automation tools.
The MCP endpoint runs inside the existing `sigma server` process and is disabled by default.

## What It Exposes

The MCP server focuses on registry management metadata and controlled operations.
It does not expose OCI blob upload/download traffic.

Current tools include:

- System: `system-version-get`, `system-endpoint-get`, `system-config-summary-get`
- Namespace: `namespace-list`, `namespace-get`, `namespace-create`, `namespace-update`, `namespace-delete`
- Namespace members: `namespace-member-list`, `namespace-member-add`, `namespace-member-update`, `namespace-member-delete`, `namespace-member-self-get`
- Repository: `repository-list`, `repository-get`, `repository-create`, `repository-update`, `repository-delete`
- Tag: `tag-list`, `tag-get`, `tag-delete`, `tag-manifest-raw-get`
- Artifact: `artifact-list`, `artifact-get`, `artifact-delete`

The read-only resources include:

- `sigma://system/version`
- `sigma://system/capabilities`

## Enable MCP

MCP is configured through the `mcp` section:

```yaml
mcp:
  enabled: true
  path: /api/v1/mcp
  transport: streamable_http
  auditWrites: true
  toolTimeout: 30s
  maxRequestBody: 1048576
```

When deploying with Helm:

```bash
helm upgrade --install sigma ./deploy/sigma \
  --set config.mcp.enabled=true
```

The endpoint is served by the existing server service:

```text
https://<sigma-host>/api/v1/mcp
```

## Authentication

MCP uses Sigma Basic Auth. Every MCP request must include:

```text
Authorization: Basic <base64(username:password)>
```

For example:

```bash
printf 'sigma:Admin@123' | base64
```

Use the encoded value in the MCP client configuration:

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

Because Basic Auth sends credentials with each request, expose MCP only over HTTPS or through a TLS-terminating gateway.

## Authorization

Tools run as the authenticated Sigma user.
Sigma applies the same resource authorization model used by the HTTP API:

- Root and admin users can access all MCP tools
- Namespace readers can read namespace-scoped resources
- Namespace managers can manage repositories, tags, and artifacts
- Namespace admins can manage namespace metadata and members

Unauthorized calls return a tool error and do not execute the requested operation.

## Auditing and Metrics

Write operations are recorded in the audit table when `mcp.auditWrites` is enabled.
Audit records use:

- `method`: `MCP`
- `path`: the MCP endpoint path
- `route`: the MCP tool name

Sensitive fields such as `password`, `token`, `privateKey`, `ak`, `sk`, `secret`, and `authorization` are redacted before audit storage.

Prometheus metrics are exposed through the existing `/metrics` endpoint:

- `sigma_mcp_tool_calls_total`
- `sigma_mcp_tool_duration_seconds`
- `sigma_mcp_auth_failures_total`

## Example Tool Arguments

List repositories in a namespace:

```json
{
  "namespace_id": "550e8400-e29b-41d4-a716-446655440000",
  "page": 1,
  "limit": 10
}
```

Create a repository:

```json
{
  "namespace_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "library/alpine",
  "description": "Alpine images"
}
```

Get a tag manifest:

```json
{
  "repository_id": "550e8400-e29b-41d4-a716-446655440000",
  "digest": "sha256:..."
}
```
