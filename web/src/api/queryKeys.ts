/**
 * Copyright 2026 sigma
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { QueryParams } from "./types";

export const queryKeys = {
  users: {
    self: ["users", "self"] as const,
    list: (params: QueryParams) => ["users", "list", params] as const,
  },
  systems: {
    config: ["systems", "config"] as const,
    endpoint: ["systems", "endpoint"] as const,
    version: ["systems", "version"] as const,
  },
  namespaces: {
    hot: ["namespaces", "hot"] as const,
    detail: (id: number | string) => ["namespaces", "detail", id] as const,
    list: (params: QueryParams) => ["namespaces", "list", params] as const,
  },
  repositories: {
    detail: (namespaceID: number | string, id: number | string) => ["repositories", "detail", namespaceID, id] as const,
    list: (namespaceID: number | string, params: QueryParams) => ["repositories", "list", namespaceID, params] as const,
  },
  tags: {
    list: (namespaceID: number | string, repositoryID: number | string, params: QueryParams) => ["tags", "list", namespaceID, repositoryID, params] as const,
  },
  webhooks: {
    detail: (id: number | string) => ["webhooks", "detail", id] as const,
    list: (params: QueryParams) => ["webhooks", "list", params] as const,
    logs: (id: number | string, params: QueryParams) => ["webhooks", "logs", id, params] as const,
  },
  daemonTasks: {
    rule: (resource: string, namespaceID: number | string) => ["daemon-tasks", "rule", resource, namespaceID] as const,
    latestRunner: (resource: string, namespaceID: number | string) => ["daemon-tasks", "latest-runner", resource, namespaceID] as const,
    runners: (resource: string, namespaceID: number | string, params: QueryParams) => ["daemon-tasks", "runners", resource, namespaceID, params] as const,
    records: (resource: string, runnerID: number | string, params: QueryParams) => ["daemon-tasks", "records", resource, runnerID, params] as const,
  },
  builders: {
    detail: (namespaceID: number | string, repositoryID: number | string, id: number | string) => ["builders", "detail", namespaceID, repositoryID, id] as const,
    runners: (namespaceID: number | string, repositoryID: number | string, builderID: number | string, params: QueryParams) => ["builders", "runners", namespaceID, repositoryID, builderID, params] as const,
  },
  codeRepositories: {
    providers: ["code-repositories", "providers"] as const,
    owners: (provider: string) => ["code-repositories", "owners", provider] as const,
    list: (provider: string, params: QueryParams) => ["code-repositories", "list", provider, params] as const,
  },
} as const;
