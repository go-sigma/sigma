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
import { useParams } from "react-router-dom";

import Toast from "../../components/Notification";
import { IHTTPError, IUserSelf, IUserLoginResponse } from "../../interfaces";

import "./index.css";

export default function ({ localServer }: { localServer: string }) {
  const { provider } = useParams<{ provider: string }>();
  const searchParams = new URLSearchParams(location.search);
  const code = searchParams.get("code");

  const [requestDone, setRequestDone] = useState(false);
  const [success, setSuccess] = useState(true);

  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");

  useEffect(() => {
    if (code == "") {
      return;
    }

    let url = localServer + `/api/v1/oauth2/${provider}/callback?code=${code}`;
    axios.get(url).then(response => {
      if (response.status === 200) {
        let resp = response.data as IUserLoginResponse;
        localStorage.setItem("token", resp.token);
        localStorage.setItem("refresh_token", resp.refresh_token);
        setUsername(resp.username);
        setEmail(resp.email);
        setSuccess(true);
        setTimeout(() => { setRequestDone(true); }, 500)
      } else {
        const token = localStorage.getItem("token");
        if (token !== "") {
          axios.get(localServer + "/api/v1/users/self").then(response => {
            if (response.status === 200) {
              const user = response.data as IUserSelf;
              setSuccess(true);
              setUsername(user.username);
              setEmail(user.email);
              setTimeout(() => { setRequestDone(true); }, 500)
            } else {
              const errorcode = response.data as IHTTPError;
              Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
              setSuccess(false);
              setTimeout(() => { setRequestDone(true); }, 500)
            }
          }).catch(error => {
            const errorcode = error.response.data as IHTTPError;
            Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
            setSuccess(false);
            setTimeout(() => { setRequestDone(true); }, 500)
          });
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
          setSuccess(false);
          setTimeout(() => { setRequestDone(true); }, 500)
        }
      }
    }).catch(error => {
      const token = localStorage.getItem("token");
      if (token !== "") {
        axios.get(localServer + "/api/v1/users/self").then(response => {
          if (response.status === 200) {
            const user = response.data as IUserSelf;
            setSuccess(true);
            setUsername(user.username);
            setEmail(user.email);
            setTimeout(() => { setRequestDone(true); }, 500)
          } else {
            const errorcode = response.data as IHTTPError;
            Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
            setSuccess(false);
            setTimeout(() => { setRequestDone(true); }, 500)
          }
        }).catch(error => {
          const errorcode = error.response.data as IHTTPError;
          Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
          setSuccess(false);
          setTimeout(() => { setRequestDone(true); }, 500)
        });
      } else {
        const errorcode = error.response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
        setSuccess(false);
        setTimeout(() => { setRequestDone(true); }, 5000)
      }
    });
  }, [code]);

  const selfUpdate = () => {
    let url = localServer + "/api/v1/users/self";
    axios.put(url, {
      username: username,
      email: email,
    }).then(response => {
      if (response.status === 202) {
        window.location.assign("/");
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  return (
    <div className="bg-white min-h-screen flex items-center">
      <div className="relative isolate w-full">
        <div
          className="absolute inset-x-0 -top-40 -z-10 transform-gpu overflow-hidden blur-3xl sm:-top-80"
          aria-hidden="true"
        >
          <div
            className="relative left-[calc(50%-11rem)] aspect-[1155/678] w-[36.125rem] -translate-x-1/2 rotate-[30deg] bg-gradient-to-tr from-[#ff80b5] to-[#9089fc] opacity-30 sm:left-[calc(50%-30rem)] sm:w-[72.1875rem]"
            style={{
              clipPath:
                'polygon(74.1% 44.1%, 100% 61.6%, 97.5% 26.9%, 85.5% 0.1%, 80.7% 2%, 72.5% 32.5%, 60.2% 62.4%, 52.4% 68.1%, 47.5% 58.3%, 45.2% 34.5%, 27.5% 76.7%, 0.1% 64.9%, 17.9% 100%, 27.6% 76.8%, 76.1% 97.7%, 74.1% 44.1%)',
            }}
          />
        </div>
        {
          requestDone ? (
            <div className="py-24">
              <div className="mx-auto max-w-7xl px-6 lg:px-8">
                <div className="mx-auto max-w-2xl text-center">
                  <h1 className="text-5xl font-bold tracking-tight text-gray-800">
                    {
                      success ? (
                        <span>Login succeed</span>
                      ) : (
                        <span>Login failed</span>
                      )
                    }
                  </h1>
                  <div className="mt-10 flex items-center justify-center gap-x-6">
                    {
                      success ? (
                        <div className="bg-white px-6 py-12 shadow sm:rounded-lg sm:px-12 min-w-[400px]">
                          <div className="space-y-6 text-left" >
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
                                  onChange={e => {
                                    setUsername(e.target.value);
                                  }}
                                />
                              </div>
                              <div className="h-1">
                                <p className="text-red-600 text-sm">xxx</p>
                              </div>
                            </div>

                            <div>
                              <label htmlFor="email" className="block text-sm font-medium leading-6 text-gray-900">
                                Email
                              </label>
                              <div className="mt-2">
                                <input
                                  id="email"
                                  name="email"
                                  type="email"
                                  required
                                  value={email}
                                  className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                  onChange={(e) => {
                                    setEmail(e.target.value);
                                  }}
                                />
                              </div>
                              <div className="h-1">
                                <p className="text-red-600 text-sm">xxx</p>
                              </div>
                            </div>

                            <div>
                              <button
                                className="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                onClick={selfUpdate}
                              >
                                Submit
                              </button>
                            </div>
                          </div>
                        </div>
                      ) : (
                        <button
                          className="rounded-md bg-indigo-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                          onClick={() => {
                            window.location.assign("/#/login");
                          }}
                        >
                          Back to login
                        </button>
                      )
                    }
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div className="lds-spinner"><div></div><div></div><div></div><div></div><div></div><div></div><div></div><div></div><div></div><div></div><div></div><div></div></div>
          )
        }
        <div
          className="absolute inset-x-0 top-[calc(100%-13rem)] -z-10 transform-gpu overflow-hidden blur-3xl sm:top-[calc(100%-30rem)]"
          aria-hidden="true"
        >
          <div
            className="relative left-[calc(50%+3rem)] aspect-[1155/678] w-[36.125rem] -translate-x-1/2 bg-gradient-to-tr from-[#ff80b5] to-[#9089fc] opacity-30 sm:left-[calc(50%+36rem)] sm:w-[72.1875rem]"
            style={{
              clipPath:
                'polygon(74.1% 44.1%, 100% 61.6%, 97.5% 26.9%, 85.5% 0.1%, 80.7% 2%, 72.5% 32.5%, 60.2% 62.4%, 52.4% 68.1%, 47.5% 58.3%, 45.2% 34.5%, 27.5% 76.7%, 0.1% 64.9%, 17.9% 100%, 27.6% 76.8%, 76.1% 97.7%, 74.1% 44.1%)',
            }}
          />
        </div>
      </div >
    </div >
  )
}
