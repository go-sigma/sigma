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

export const setupAxiosInterceptor = (navigate: any) => {
  axios.interceptors.response.use(response => {
    return response;
  }, error => {
    if (error?.response?.status === 401) {
      navigate('/login');
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
        if (!config.url.endsWith("/api/v1/users/login")) {
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
