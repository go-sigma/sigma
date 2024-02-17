/**
 * Copyright 2024 sigma
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

export enum NamespaceRole {
  Admin = "NamespaceAdmin",
  Manager = "NamespaceManager",
  Reader = "NamespaceReader",
}

export enum UserRole {
  Root = "Root",
  Admin = "Admin",
  User = "User",
  Anonymous = "Anonymous",
}

export enum WebhookResourceType {
  Ping = "ping",
  Namespace = "namespace",
  Repository = "repository",
  Tag = "tag",
  Artifact = "artifact",
  Member = "member"
}

export enum WebhookAction {
  Create = "create",
  Update = "update",
  Delete = "delete",
  Add = "add",
  Remove = "remove",
  Pull = "pull",
  Push = "push"
}
