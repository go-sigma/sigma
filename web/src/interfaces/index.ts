/**
 * Copyright 2023 XImager
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

export interface IArtifact {
  id: number;
  digest: string;
  size: number;
  tag_count: number;
  tags: string[];
  created_at: string;
  updated_at: string;
}

export interface IArtifactList {
  items: IArtifact[];
  total: number;
}

export interface ITag {
  id: number;
  name: string;
  digest: string;
  size: number;
  last_pull: string;
  pull_times: number;
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
