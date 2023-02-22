export interface INamespace {
  id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface INamespaceList {
  items: INamespace[];
  total: number;
}

export interface INotification {
  title: string;
  message: string;
}

export interface IHTTPError {
  code: number;
  title: string;
  message: string;
}
