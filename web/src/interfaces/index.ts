/**
 * Copyright 2023 sigma
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

export interface INotification {
  level?: string;
  title: string;
  message: string;
  autoClose?: number;
}

export interface IHTTPError {
  code: number;
  title: string;
  description: string;
}

export interface IUserItem {
  id: number;
  username: string;
  email: string;
  status: string;
  last_login: string;
  namespace_count: number;
  namespace_limit: number;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface IUserList {
  items: IUserItem[];
  total: number;
}

export interface IUserSelf {
  id: number;
  username: string;
  email: string;
}

export interface IOauth2ClientID {
  client_id: string;
}

export interface IUserLoginResponse {
  id: number;
  username: string;
  email: string;
  token: string;
  refresh_token: string;
}

export interface INamespaceItem {
  id: number;
  name: string;
  description: string;
  size: number;
  size_limit: number;
  repository_count: number;
  repository_limit: number;
  tag_count: number;
  tag_limit: number;
  visibility: string;
  created_at: string;
  updated_at: string;
}

export interface INamespaceList {
  items: INamespaceItem[];
  total: number;
}

export interface IRepositoryItem {
  id: number;
  namespace_id: number;
  name: string;
  description: string;
  overview: string;
  visibility: string;
  tag_count: number;
  tag_limit: number;
  size_limit: number;
  size: number;
  created_at: string;
  updated_at: string;

  builder?: IBuilderItem;
}

export interface IRepositoryList {
  items: IRepositoryItem[];
  total: number;
}

export interface IArtifactList {
  items: IArtifact[];
  total: number;
}

export interface IArtifact {
  id: number;
  digest: string;
  media_type: string;
  raw: string;
  config_media_type: string;
  config_raw: string;
  type: string;
  size: number;
  blob_size: number;
  last_pull: string;
  pushed_at: string;
  pull_times: number;
  vulnerability: string;
  sbom: string;
  created_at: string;
  updated_at: string;
}

export interface ITag {
  id: number;
  name: string;
  artifact: IArtifact;
  artifacts: IArtifact[];
  pushed_at: string;
  created_at: string;
  updated_at: string;
}

export interface ITagList {
  items: ITag[];
  total: number;
}

export enum IOrder {
  Asc = "asc",
  Desc = "desc",
  None = "none",
};

export interface ISizeWithUnit {
  unit: string;
  size: number;
}

export interface IVuln {
  critical: string;
  high: string;
  medium: string;
  low: string;
}

export interface IDistro {
  name: string;
  version: string;
}

export interface ISbom {
  distro: IDistro;
  os: string;
  architecture: string;
}

export interface IImageConfig {
  architecture: string;
  os: string;
}

export interface IEndpoint {
  endpoint: string;
}

export interface ICodeRepositoryOwnerItem {
  id: number;
  owner_id: string;
  owner: string;
  is_org: boolean;
  created_at: string;
  updated_at: string;
}

export interface ICodeRepositoryOwnerList {
  items: ICodeRepositoryOwnerItem[];
  total: number;
}

export interface ICodeRepositoryItem {
  id: number;
  repository_id: string;
  name: string;
  owner_id: string;
  owner: string;
  is_org: boolean;
  clone_url: string;
  ssh_url: string;
  oci_repo_count: number;
  created_at: string;
  updated_at: string;
}

export interface ICodeRepositoryList {
  items: ICodeRepositoryItem[];
  total: number;
}

export interface ICodeRepositoryBranchItem {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface ICodeRepositoryBranchList {
  items: ICodeRepositoryBranchItem[];
  total: number;
}

export interface ICodeRepositoryProviderItem {
  provider: string;
}

export interface ICodeRepositoryProviderList {
  items: ICodeRepositoryProviderItem[];
  total: number;
}

export interface ICodeRepositoryUser3rdParty {
  id: number;
  account_id: string;
  cr_last_update_timestamp: string;
  cr_last_update_status: string;
  cr_last_update_message: string;
  created_at: string;
  updated_at: string;
}

export interface IBuilderItem {
  id: number;
  repository_id: number;
  source: string;

  code_repository_id?: number;

  dockerfile?: string;

  scm_repository?: string;
  scm_credential_type?: string;
  scm_ssh_key?: string;
  scm_token?: string;
  scm_username?: string;
  scm_password?: string;

  scm_branch?: string;

  scm_depth?: number;
  scm_submodule?: boolean;

  cron_rule?: string;
  cron_branch?: string;
  cron_tag_template?: string;

  webhook_branch_name?: string;
  webhook_branch_tag_template?: string;
  webhook_tag_tag_template?: string;

  buildkit_insecure_registry?: string[];
  buildkit_context?: string;
  buildkit_dockerfile?: string;
  buildkit_platforms: string[];
}

export interface IBuilderRunnerList {
  items: IBuilderRunnerItem[];
  total: number;
}

export interface IBuilderRunnerItem {
  id: number;
  builder_id: number;
  status: string;
  tag?: string;
  raw_tag: string;
  description?: string;
  scm_branch?: string;

  started_at?: string;
  ended_at?: string;
  raw_duration?: number;
  duration?: string;

  created_at: string;
  updated_at: string;
}

export interface IRunOrRerunRunnerResponse {
  runner_id: number;
}

export interface IVersion {
  version: string;
  git_hash: string;
  build_date: string;
}
