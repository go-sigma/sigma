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

export interface INamespace {
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
  items: INamespace[];
  total: number;
}

export interface IRepository {
  id: number;
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
}

export interface IRepositoryList {
  items: IRepository[];
  total: number;
}

export interface IArtifactList {
  items: IArtifact[];
  total: number;
}

export interface IArtifact {
  id: number;
  digest: string;
  raw: string;
  config_raw: string;
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
