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
import { useClickAway } from 'react-use';
import { useEffect, useRef, useState } from "react";
import { Link, useSearchParams, useNavigate } from 'react-router-dom';

import Toast from "../../components/Notification";
import { INamespace, INamespaceList, IHTTPError, IUserSelf, IEndpoint } from "../../interfaces";

export default function ({ localServer, item, namespace, repository, tag }: { localServer: string, item: string, namespace?: string, repository?: string, tag?: string }) {
  const [showProfileMenu, setShowProfileMenu] = useState(false);
  const [menuActive, setMenuActive] = useState(item === "" ? "home" : item);
  const navigate = useNavigate();

  const [searchParams] = useSearchParams();
  const isAnonymous = (searchParams.get('anonymous') || "false") === "true";

  const ref = useRef<HTMLDivElement>() as React.MutableRefObject<HTMLDivElement>;
  useClickAway(ref, () => {
    if (showProfileMenu) {
      setShowProfileMenu(!showProfileMenu);
    }
  });

  const [username, setUsername] = useState("");

  // get user info
  useEffect(() => {
    if (!isAnonymous) {
      axios.get(localServer + "/api/v1/users/self").then(response => {
        if (response.status === 200) {
          const user = response.data as IUserSelf;
          setUsername(user.username);
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
        }
      }).catch(error => {
        const errorcode = error.response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      });
    }
  }, [])

  const [hotNamespaceTotal, setHotNamespaceTotal] = useState(0);
  const [hotNamespaceList, setHotNamespaceList] = useState<INamespace[]>([]);

  // get hot namespace
  useEffect(() => {
    if (!isAnonymous) {
      axios.get(localServer + "/api/v1/namespaces/hot").then(response => {
        if (response.status === 200) {
          const namespaces = response.data as INamespaceList;
          setHotNamespaceTotal(namespaces.total);
          setHotNamespaceList(namespaces.items);
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
        }
      }).catch(error => {
        const errorcode = error.response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      });
    }
  }, [])

  const logout = () => {
    if (!isAnonymous) {
      let tokens: string[] = [localStorage.getItem("token") || "", localStorage.getItem("refresh_token") || ""];
      axios.post(localServer + "/api/v1/users/logout", {
        tokens: tokens
      }).then(response => {
        localStorage.removeItem("token");
        localStorage.removeItem("refresh_token");
        navigate("/login");
      }).catch(error => {
        const errorcode = error.response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      });
    }
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

  return (
    <div className="flex flex-shrink-0">
      <div className="flex flex-col w-64 border-r border-gray-200 pt-5 pb-4 bg-white">
        <div className="flex items-center flex-shrink-0 px-6 cursor-pointer" onClick={e => {
          if (isAnonymous) {
            navigate("/namespaces");
          } else {
            navigate("/home");
          }
        }}>
          <img className="h-10 w-auto" src="/title.svg" alt="Workflow" />
        </div>
        <div className="h-0 flex-1 flex flex-col overflow-y-auto">
          {
            !isAnonymous && (
              <div className="px-3 mt-6 relative inline-block text-left" ref={ref}>
                <div>
                  <button type="button" className="group w-full bg-gray-100 rounded-md px-3.5 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-100 focus:ring-purple-500" onClick={() => { setShowProfileMenu(!showProfileMenu) }}>
                    <span className="flex w-full justify-between items-center">
                      <span className="flex min-w-0 items-center justify-between space-x-3">
                        <span className="flex min-w-0 flex-col">
                          <span className="text-gray-700 text-sm font-medium truncate">{username}</span>
                        </span>
                      </span>
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-gray-500" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" >
                        <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 15L12 18.75 15.75 15m-7.5-6L12 5.25 15.75 9" />
                      </svg>
                    </span>
                  </button>
                </div>
                <div className={showProfileMenu ? "" : "hidden"}>
                  <div className="z-10 mx-3 origin-top absolute right-0 left-0 mt-1 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 divide-y divide-gray-200">
                    <div className="py-1">
                      <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">View profile</div>
                      <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">Settings</div>
                      <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">Notifications</div>
                    </div>
                    <div className="py-1">
                      <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">Get desktop app</div>
                      <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">Support</div>
                    </div>
                    <div className="py-1">
                      <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900" onClick={logout}>Logout</div>
                    </div>
                  </div>
                </div>
              </div>
            )
          }
          <nav className="px-3 mt-6">
            <div className="space-y-1">
              {
                !isAnonymous && (
                  <Link to={`/home`} className={`text-gray-700 group flex items-center px-2 py-2 text-sm font-medium rounded-md ${menuActive === "home" ? "bg-gray-100" : "hover:bg-gray-50 text-gray-700"}`} onClick={e => {
                    setMenuActive("home");
                    item === "home" && e.preventDefault();
                  }}>
                    <span className="text-gray-400 mr-3 h-6 w-6">
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" />
                      </svg>
                    </span>
                    Home
                  </Link>
                )
              }
              <Link to={`/namespaces`} className={`text-gray-700 group flex items-center px-2 py-2 text-sm font-medium rounded-md ${menuActive === "namespaces" ? "bg-gray-100" : "hover:bg-gray-50 text-gray-700"}`} onClick={e => {
                setMenuActive("namespaces");
                item === "namespaces" && e.preventDefault();
              }}>
                <span className="text-gray-400 mr-3 h-6 w-6">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 5.25h16.5m-16.5 4.5h16.5m-16.5 4.5h16.5m-16.5 4.5h16.5" />
                  </svg>
                </span>
                Namespaces
              </Link>
              {
                (item === "repositories" || item === "tags" || item === "artifacts") && (
                  <Link to={`/namespaces/${namespace}/repositories`} className={`text-gray-700 group flex items-center px-4 py-2 text-sm font-medium rounded-md ${menuActive === "repositories" ? "bg-gray-100" : "hover:bg-gray-50 text-gray-700"}`} onClick={e => {
                    setMenuActive("repositories");
                    item === "repositories" && e.preventDefault();
                  }}>
                    <span className="text-gray-400 mr-3 h-6 w-6">
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z" />
                      </svg>
                    </span>
                    Repositories
                  </Link>
                )
              }
              {
                (item === "tags" || item === "artifacts") && (
                  <Link to={`/namespaces/${namespace}/repository/tags?repository=${repository}`} className={`text-gray-700 group flex items-center px-4 py-2 text-sm font-medium rounded-md ${menuActive === "tags" ? "bg-gray-100" : "hover:bg-gray-50 text-gray-700"}`} onClick={e => {
                    setMenuActive("tags");
                    item === "tags" && e.preventDefault();
                  }}>
                    <span className="text-gray-400 mr-3 h-6 w-6">
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z" />
                        <path strokeLinecap="round" strokeLinejoin="round" d="M6 6h.008v.008H6V6z" />
                      </svg>
                    </span>
                    Tags
                  </Link>
                )
              }
              {
                item === "artifacts" && (
                  <Link to={`/artifacts`} className={`text-gray-700 group flex items-center px-4 py-2 text-sm font-medium rounded-md ${menuActive === "namespaces" ? "bg-gray-100" : "hover:bg-gray-50 text-gray-700"}`} onClick={e => {
                    setMenuActive("namespaces");
                    item === "artifacts" && e.preventDefault();
                  }}>
                    <span className="text-gray-400 mr-3 h-6 w-6">
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0l4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0l-5.571 3-5.571-3" />
                      </svg>
                    </span>
                    Artifacts
                  </Link>
                )
              }
              <Link to={`/coderepos`} className={`text-gray-700 group flex items-center px-2 py-2 text-sm font-medium rounded-md ${menuActive === "coderepos" ? "bg-gray-100" : "hover:bg-gray-50 text-gray-700"}`} onClick={e => {
                setMenuActive("namespaces");
                item === "namespaces" && e.preventDefault();
              }}>
                <span className="text-gray-400 mr-3 h-6 w-6">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3m-19.5 0a4.5 4.5 0 01.9-2.7L5.737 5.1a3.375 3.375 0 012.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 01.9 2.7m0 0a3 3 0 01-3 3m0 3h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008zm-3 6h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008z" />
                  </svg>
                </span>
                Code Repository
              </Link>
              {
                !isAnonymous && (
                  <Link to={`/settings`} className={`text-gray-700 group flex items-center px-2 py-2 text-sm font-medium rounded-md ${menuActive === "settings" ? "bg-gray-100" : "hover:bg-gray-50 text-gray-700"}`} onClick={e => {
                    setMenuActive("settings");
                    item === "settings" && e.preventDefault();
                  }}>
                    <span className="text-gray-400 mr-3 h-6 w-6">
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M10.343 3.94c.09-.542.56-.94 1.11-.94h1.093c.55 0 1.02.398 1.11.94l.149.894c.07.424.384.764.78.93.398.164.855.142 1.205-.108l.737-.527a1.125 1.125 0 011.45.12l.773.774c.39.389.44 1.002.12 1.45l-.527.737c-.25.35-.272.806-.107 1.204.165.397.505.71.93.78l.893.15c.543.09.94.56.94 1.109v1.094c0 .55-.397 1.02-.94 1.11l-.893.149c-.425.07-.765.383-.93.78-.165.398-.143.854.107 1.204l.527.738c.32.447.269 1.06-.12 1.45l-.774.773a1.125 1.125 0 01-1.449.12l-.738-.527c-.35-.25-.806-.272-1.203-.107-.397.165-.71.505-.781.929l-.149.894c-.09.542-.56.94-1.11.94h-1.094c-.55 0-1.019-.398-1.11-.94l-.148-.894c-.071-.424-.384-.764-.781-.93-.398-.164-.854-.142-1.204.108l-.738.527c-.447.32-1.06.269-1.45-.12l-.773-.774a1.125 1.125 0 01-.12-1.45l.527-.737c.25-.35.273-.806.108-1.204-.165-.397-.505-.71-.93-.78l-.894-.15c-.542-.09-.94-.56-.94-1.109v-1.094c0-.55.398-1.02.94-1.11l.894-.149c.424-.07.765-.383.93-.78.165-.398.143-.854-.107-1.204l-.527-.738a1.125 1.125 0 01.12-1.45l.773-.773a1.125 1.125 0 011.45-.12l.737.527c.35.25.807.272 1.204.107.397-.165.71-.505.78-.929l.15-.894z" />
                        <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      </svg>
                    </span>
                    Setting
                  </Link>
                )
              }
            </div>
            <div className="mt-8">
              {
                !(isAnonymous || hotNamespaceTotal === 0) && (
                  <div>
                    <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider" id="teams-headline">
                      Hot namespace
                    </h3>
                    <div className="mt-1 space-y-1" role="group" aria-labelledby="teams-headline">
                      {
                        hotNamespaceList.map((ns: INamespace, index: number) => {
                          return (
                            <Link
                              key={index}
                              to={`/namespaces/${ns.name}/repositories`}
                              className="group flex items-center px-3 py-2 text-sm font-medium text-gray-700 rounded-md hover:text-gray-900 hover:bg-gray-50 cursor-pointer"
                              onClick={e => {
                                item === "repositories" && ns.name === namespace && e.preventDefault();
                              }}
                            >
                              <span className="w-2.5 h-2.5 mr-4 bg-indigo-500 rounded-full" aria-hidden="true"></span>
                              <span className="truncate">
                                {ns.name}
                              </span>
                            </Link>
                          )
                        })
                      }
                    </div>
                  </div>
                )
              }
            </div>
          </nav>
        </div>
        <div className="flex justify-center">
          <a type="button" className="rounded-md bg-white px-20 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50" href={`${endpoint}/swagger/index.html`}>
            <span className="text-gray-700">API Docs</span>
          </a>
        </div>
      </div>
    </div >
  );
}
