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

import { AxiosError } from 'axios';

import { IHTTPError } from "../interfaces";

export class ApiError extends Error {
  status?: number;
  code?: number;
  title: string;
  description: string;

  constructor({ status, code, title, description }: { status?: number; code?: number; title: string; description: string }) {
    super(description);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.title = title;
    this.description = description;
  }
}

export function toApiError(error: unknown): ApiError {
  if (error instanceof ApiError) {
    return error;
  }
  if (error instanceof AxiosError) {
    const data = error.response?.data as Partial<IHTTPError> | undefined;
    return new ApiError({
      status: error.response?.status,
      code: data?.code,
      title: data?.title || "Request failed",
      description: data?.description || error.message,
    });
  }
  if (error instanceof Error) {
    return new ApiError({
      title: "Request failed",
      description: error.message,
    });
  }
  return new ApiError({
    title: "Request failed",
    description: "unknown request error",
  });
}
