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
