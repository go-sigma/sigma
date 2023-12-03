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

import axios from "axios";

import { IUserLoginResponse } from "../interfaces";

let REFRESH_TOKEN_INTERVAL: ReturnType<typeof setInterval> | null;

const REFRESH_TOKEN_INTERVAL_TIMEOUT = 60 * 10 * 1000; // 10 mins

export function refreshToken(
  localServer: string,
  onFailed: () => void
) {
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
      } else {
        onFailed()
      }
    }).catch(err => {
      onFailed()
    })
}

export function setupAutoRefreshToken(
  localServer: string,
  onFailed: () => void
) {
  if (REFRESH_TOKEN_INTERVAL) return;
  REFRESH_TOKEN_INTERVAL = REFRESH_TOKEN_INTERVAL = setInterval(() => {
    refreshToken(localServer, onFailed);
  }, REFRESH_TOKEN_INTERVAL_TIMEOUT);
}

export function teardownAutoRefreshToken() {
  if (!REFRESH_TOKEN_INTERVAL) return;
  clearInterval(REFRESH_TOKEN_INTERVAL);
  REFRESH_TOKEN_INTERVAL = null;
}
