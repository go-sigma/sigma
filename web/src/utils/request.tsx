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

import axios from 'axios';
import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'

const instance = axios.create({
  timeout: 20000,
  withCredentials: false,
});

instance.interceptors.request.use((config: any) => {
  const token = localStorage.getItem('token');
  token && (config.headers.Authorization = "Bearer " + token);
  return config;
});

const AxiosInterceptor = ({ children }: { children: any }) => {
  const navigate = useNavigate();
  useEffect(() => {
    const interceptor = instance.interceptors.response.use(response => {
      return response;
    }, error => {
      if (error?.response?.status === 401) {
        navigate('/login');
      }
      return Promise.reject(error);
    });
    return () => instance.interceptors.response.eject(interceptor);
  }, [navigate]);
  return children;
}

export default instance;

export { AxiosInterceptor };
