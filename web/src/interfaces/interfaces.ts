/**
 * The MIT License (MIT)
 *
 * Copyright © 2023 Tosone <i@tosone.cn>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

export interface INotification {
  level?: string;
  title: string;
  message: string;
}

export interface IHTTPError {
  code: number;
  title: string;
  message: string;
}

export interface INamespace {
  id: number;
  name: string;
  description: string;
  artifact_count: number;
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
  artifact_count: number;
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
  created_at: string;
  updated_at: string;
}

export interface ITagList {
  items: ITag[];
  total: number;
}
