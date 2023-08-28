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
import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Helmet, HelmetProvider } from "react-helmet-async";

import Toast from "../../components/Notification";
import { IHTTPError, IUserLoginResponse, IOauth2ClientID, IEndpoint } from "../../interfaces";

export default function Login({ localServer }: { localServer: string }) {
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const login = (username: string, password: string) => {
    let url = localServer + `/api/v1/users/login`;
    axios.post(url, { "username": username, "password": password })
      .then(response => {
        const resp = response.data as IUserLoginResponse;
        localStorage.setItem("token", resp.token);
        localStorage.setItem("refresh_token", resp.refresh_token);
        navigate("/");
      }).catch(err => {
        console.log(err)
      })
  }

  const [endpoint, setEndpoint] = useState("");

  useEffect(() => {
    let url = localServer + `/api/v1/systems/endpoint`;
    axios.get(url).then(response => {
      if (response.status === 200) {
        let e = response.data as IEndpoint;
        setEndpoint(e.endpoint);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [])

  useEffect(() => { // if user have already login redirect to home
    axios.get(localServer + "/api/v1/users/self").then(response => {
      if (response !== undefined && response.status === 200) {
        navigate("/");
      }
    }).catch(error => {
      console.log("get user self info failed:", error);
    });
  }, []);

  return (
    <>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Login</title>
        </Helmet>
      </HelmetProvider>
      <div className="flex min-h-full flex-1 flex-col justify-center py-12 sm:px-6 lg:px-8">
        <div className="mt-10 sm:mx-auto sm:w-full sm:max-w-[480px]">
          <div className="bg-white px-6 py-12 shadow sm:rounded-lg sm:px-12">
            <div className="space-y-6" >
              <div>
                <label htmlFor="username" className="block text-sm font-medium leading-6 text-gray-900">
                  Username
                </label>
                <div className="mt-2">
                  <input
                    id="username"
                    name="username"
                    type="text"
                    value={username}
                    required
                    className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    onChange={(e) => {
                      setUsername(e.target.value);
                    }}
                  />
                </div>
                <div className="h-1">
                  <p className="text-red-600 text-sm">xxx</p>
                </div>
              </div>

              <div>
                <label htmlFor="password" className="block text-sm font-medium leading-6 text-gray-900">
                  Password
                </label>
                <div className="mt-2">
                  <input
                    id="password"
                    name="password"
                    type="password"
                    required
                    value={password}
                    className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    onChange={(e) => {
                      setPassword(e.target.value);
                    }}
                  />
                </div>
                <div className="h-1">
                  <p className="text-red-600 text-sm">xxx</p>
                </div>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <input
                    id="remember-me"
                    name="remember-me"
                    type="checkbox"
                    className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                  />
                  <label htmlFor="remember-me" className="ml-3 block text-sm leading-6 text-gray-900">
                    Remember me
                  </label>
                </div>

                <div className="text-sm leading-6">
                  <a href="#" className="font-semibold text-indigo-600 hover:text-indigo-500">
                    Forgot password?
                  </a>
                </div>
              </div>

              <div>
                <button
                  className="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                  onClick={() => {
                    login(username, password);
                  }}
                >
                  Sign in
                </button>
              </div>
            </div>

            <div>
              <div className="relative mt-10">
                <div className="absolute inset-0 flex items-center" aria-hidden="true">
                  <div className="w-full border-t border-gray-200" />
                </div>
                <div className="relative flex justify-center text-sm font-medium leading-6">
                  <span className="bg-white px-6 text-gray-900">Or continue with</span>
                </div>
              </div>
              <div className="mt-6 grid grid-cols-1 gap-4">
                <button className="flex w-full items-center justify-center gap-3 rounded-md bg-[#24292F] px-3 py-1.5 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#24292F] border-gray-800">Anonymous</button>
              </div>
              <GitHubButton localServer={localServer} endpoint={endpoint} />
              <GitLabButton localServer={localServer} endpoint={endpoint} />
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

function GitHubButton({ localServer, endpoint }: { localServer: string, endpoint: string }) {
  const [clientID, setClientID] = useState("");

  useEffect(() => {
    axios.get(`${localServer}/api/v1/oauth2/github/client_id`).then(response => {
      if (response.status == 200) {
        const data = response.data as IOauth2ClientID;
        setClientID(data.client_id);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, []);

  return (
    <div className="mt-6 grid grid-cols-1 gap-4">
      <a
        href={`https://github.com/login/oauth/authorize?client_id=${clientID}&redirect_uri=${encodeURIComponent(`${endpoint}/api/v1/oauth2/github/redirect_callback?endpoint=${encodeURIComponent(location.protocol + "//" + location.host)}`)}&scope=repo`}
        className="flex w-full items-center justify-center gap-3 rounded-md bg-[#24292F] px-3 py-1.5 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#24292F]"
      >
        <svg className="h-5 w-5" aria-hidden="true" fill="currentColor" viewBox="0 0 20 20">
          <path
            fillRule="evenodd"
            d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z"
            clipRule="evenodd"
          />
        </svg>
        <span className="text-sm font-semibold leading-6">GitHub</span>
      </a>
    </div>
  );
}


function GitLabButton({ localServer, endpoint }: { localServer: string, endpoint: string }) {
  const [clientID, setClientID] = useState("");

  useEffect(() => {
    axios.get(`${localServer}/api/v1/oauth2/gitlab/client_id`).then(response => {
      if (response.status == 200) {
        const data = response.data as IOauth2ClientID;
        setClientID(data.client_id);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, []);

  return (
    <div className="mt-6 grid grid-cols-1 gap-4">
      <a
        href={`https://gitlab.com/oauth/authorize?client_id=${clientID}&redirect_uri=${encodeURIComponent(`${endpoint}/api/v1/oauth2/gitlab/redirect_callback?endpoint=${encodeURIComponent(location.protocol + "//" + location.host)}`)}&response_type=code&scope=read_repository+read_user+api+read_api`}
        className="flex w-full items-center justify-center gap-3 rounded-md  bg-red-600 px-3 py-1.5 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#24292F]"
      >
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="90 90 210 210" className="h-5 w-5"><g><path className="fill-[#e24329]" d="M282.83,170.73l-.27-.69-26.14-68.22a6.81,6.81,0,0,0-2.69-3.24,7,7,0,0,0-8,.43,7,7,0,0,0-2.32,3.52l-17.65,54H154.29l-17.65-54A6.86,6.86,0,0,0,134.32,99a7,7,0,0,0-8-.43,6.87,6.87,0,0,0-2.69,3.24L97.44,170l-.26.69a48.54,48.54,0,0,0,16.1,56.1l.09.07.24.17,39.82,29.82,19.7,14.91,12,9.06a8.07,8.07,0,0,0,9.76,0l12-9.06,19.7-14.91,40.06-30,.1-.08A48.56,48.56,0,0,0,282.83,170.73Z" /><path className="fill-[#fc6d26]" d="M282.83,170.73l-.27-.69a88.3,88.3,0,0,0-35.15,15.8L190,229.25c19.55,14.79,36.57,27.64,36.57,27.64l40.06-30,.1-.08A48.56,48.56,0,0,0,282.83,170.73Z" /><path className="fill-[#fca326]" d="M153.43,256.89l19.7,14.91,12,9.06a8.07,8.07,0,0,0,9.76,0l12-9.06,19.7-14.91S209.55,244,190,229.25C170.45,244,153.43,256.89,153.43,256.89Z" /><path className="fill-[#fc6d26]" d="M132.58,185.84A88.19,88.19,0,0,0,97.44,170l-.26.69a48.54,48.54,0,0,0,16.1,56.1l.09.07.24.17,39.82,29.82s17-12.85,36.57-27.64Z" /></g></svg>
        <span className="text-sm font-semibold leading-6">GitLab</span>
      </a>
    </div>
  );
}
