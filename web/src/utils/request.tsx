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
import { redirect } from "react-router-dom";

const instance = axios.create({
  timeout: 20000,
  withCredentials: true,
});

instance.interceptors.request.use(function (config: any) {
  const token = localStorage.getItem('token');
  token && (config.headers.Authorization = token);
  return config;
})

instance.interceptors.response.use(function (response) {
  return response;
}, function (error) {
  console.log(error);
  redirect("/login");
  return Promise.reject(error);
});

export default instance;
