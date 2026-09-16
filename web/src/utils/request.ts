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

import axios from 'axios';
import { NavigateFunction } from 'react-router-dom';

import { setApiNavigationHandlers } from '../api/client';
import { IUserLoginResponse } from '../interfaces';

let REFRESH_TOKEN_INTERVAL: ReturnType<typeof setInterval> | null;

const REFRESH_TOKEN_INTERVAL_TIMEOUT = 60 * 1000; // 10s

/** Install the axios request/response interceptors used across the app. */
export const setupAxiosInterceptor = (navigate: NavigateFunction) => {
  setApiNavigationHandlers({
    navigateToLogin: () => navigate('/login'),
  });
  axios.interceptors.response.clear()
  axios.interceptors.request.clear();
  axios.interceptors.response.use(response => {
    return response;
  }, error => {
    if (error?.response?.status === 401) {
      if (error?.response?.config?.url?.endsWith("/api/v1/users/login")) {
        return Promise.resolve(error?.response);
      } else {
        navigate('/login');
      }
    } else if (error?.response?.status === 500) {
      return Promise.reject(error);
    } else {
      return Promise.resolve(error?.response);
    }
  });
  axios.interceptors.request.use((config: any) => {
    const token = localStorage.getItem('token');
    if (config.headers.Authorization === undefined || config.headers.Authorization === null) {
      if (token === null) {
        if (!config.url.endsWith("/api/v1/users/login") && location.hash != "#/login") {
          navigate('/login');
          return Promise.reject(new Error('request has been banned by axios interceptor'));
        }
      } else {
        config.headers.Authorization = "Bearer " + token;
      }
    }
    return config;
  });
}

/** Exchange the refresh token for a new access token. */
export function refreshToken(
  localServer: string,
  onFailed: () => void
) {
  if (localStorage.getItem('refresh_token') == null) {
    return;
  }
  let headers: { [key: string]: any } = {
    "Authorization": "Bearer " + localStorage.getItem('refresh_token'),
  };

  let url = localServer + `/api/v1/users/login`;
  axios.post(url, {}, {
    headers: headers,
  })
    .then(response => {
      if (response?.status === 200) {
        const resp = response.data as IUserLoginResponse;
        localStorage.setItem("token", resp.token);
        localStorage.setItem("refresh_token", resp.refresh_token);
        localStorage.setItem("username", resp.username);
        localStorage.setItem("email", resp.email);
      } else {
        onFailed()
      }
    }).catch(err => {
      onFailed()
    })
}

/** Start the periodic access token refresh. */
export function setupAutoRefreshToken(
  localServer: string,
  onFailed: () => void
) {
  if (REFRESH_TOKEN_INTERVAL) return;
  REFRESH_TOKEN_INTERVAL = REFRESH_TOKEN_INTERVAL = setInterval(() => {
    refreshToken(localServer, onFailed);
  }, REFRESH_TOKEN_INTERVAL_TIMEOUT);
}

/** Stop the periodic access token refresh. */
export function teardownAutoRefreshToken() {
  if (!REFRESH_TOKEN_INTERVAL) return;
  clearInterval(REFRESH_TOKEN_INTERVAL);
  REFRESH_TOKEN_INTERVAL = null;
}
