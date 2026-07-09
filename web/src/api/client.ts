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

import axios, { AxiosError, AxiosHeaders, AxiosRequestConfig } from 'axios';

import { apiBaseURL } from "../app/config";
import { toApiError } from "./errors";
import { QueryParams } from "./types";

let navigateToLogin: (() => void) | undefined;

export const apiClient = axios.create({
  baseURL: apiBaseURL,
});

export function setApiNavigationHandlers(handlers: { navigateToLogin: () => void }) {
  navigateToLogin = handlers.navigateToLogin;
}

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  config.headers = AxiosHeaders.from(config.headers);
  if (config.headers.get("Authorization") === undefined && token !== null) {
    config.headers.set("Authorization", `Bearer ${token}`);
  }
  return config;
});

apiClient.interceptors.response.use(response => {
  return response;
}, error => {
  if (error instanceof AxiosError && error.response?.status === 401) {
    const url = error.response.config?.url || "";
    if (!url.endsWith("/api/v1/users/login")) {
      navigateToLogin?.();
    }
  }
  return Promise.reject(toApiError(error));
});

export async function requestData<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await apiClient.request<T>(config);
  return response.data;
}

export function cleanParams(params: QueryParams): QueryParams {
  return Object.fromEntries(
    Object.entries(params).filter(([, value]) => value !== undefined && value !== null && value !== "")
  );
}
