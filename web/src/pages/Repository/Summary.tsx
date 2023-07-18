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

import "bytemd/dist/index.css";
import "github-markdown-css/github-markdown-light.css";

import axios from "axios";
import gfm from "@bytemd/plugin-gfm";
import { useEffect, useState } from "react";
import { Editor, Viewer } from "@bytemd/react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { Link, useSearchParams, useParams } from "react-router-dom";

import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Toast from "../../components/Notification";
import { IRepository, IHTTPError } from "../../interfaces";

export default function ({ localServer }: { localServer: string }) {
  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repository_id = parseInt(searchParams.get("repository_id") || "");

  const [repositoryObj, setRepositoryObj] = useState<IRepository>({} as IRepository);

  const [overview, setOverview] = useState("");
  const [overviewValid, setOverviewValid] = useState(true);
  useEffect(() => { setOverviewValid(overview?.length < 100000) }, [overview]);

  useEffect(() => {
    axios.get(localServer + `/api/v1/namespaces/${namespace}/repositories/${repository_id}`).then(response => {
      if (response.status === 200) {
        const r = response.data as IRepository;
        setRepositoryObj(r);
        setOverview(r.overview);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespace, repository_id])

  const [editorState, setEditorState] = useState(false);

  const updateRepository = () => {
    if (!(overviewValid)) {
      Toast({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    axios.put(localServer + `/api/v1/namespaces/${namespace}/repositories/${repository_id}`, {
      overview: overview,
    } as IRepository, {}).then(response => {
      if (response.status === 204) {
        Toast({ level: "info", title: "Success", message: "update overview success" });
        setEditorState(false);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }

  return (
    <>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Repository Summary</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen max-h-screen flex overflow-hidden bg-white">
        <Menu item="Repository" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title="Repository"
              breadcrumb={
                (
                  <nav className="flex" aria-label="Breadcrumb">
                    <ol className="inline-flex items-center space-x-1 md:space-x-0">
                      <li className="inline-flex items-center">
                        <Link to={"/namespaces"} className="inline-flex items-center text-sm font-medium text-gray-700 hover:text-blue-600 dark:text-gray-400 dark:hover:text-white">
                          <svg className="w-3 h-3 mr-2.5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                            <path d="m19.707 9.293-2-2-7-7a1 1 0 0 0-1.414 0l-7 7-2 2a1 1 0 0 0 1.414 1.414L2 10.414V18a2 2 0 0 0 2 2h3a1 1 0 0 0 1-1v-4a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v4a1 1 0 0 0 1 1h3a2 2 0 0 0 2-2v-7.586l.293.293a1 1 0 0 0 1.414-1.414Z" />
                          </svg>
                        </Link>
                      </li>
                      <li className="inline-flex items-center">
                        <span className="inline-flex items-center text-sm font-medium text-gray-700 hover:text-blue-600 dark:text-gray-400 dark:hover:text-white">
                          <Link to={`/namespaces/${namespace}/repositories`} className="inline-flex items-center text-sm font-medium text-gray-700 hover:text-blue-600 dark:text-gray-400 dark:hover:text-white">
                            {namespace}
                          </Link>
                        </span>
                      </li>
                      <li>
                        <div className="flex items-center">
                          <span className="text-gray-500 text-sm ml-1">/</span>
                          <span className="ml-1 text-sm font-medium text-gray-500 dark:text-gray-400">
                            {repositoryObj?.name?.substring((namespace?.length || 0) + 1)}
                          </span>
                          <span className="text-gray-500 text-sm ml-1">/</span>
                        </div>
                      </li>
                    </ol>
                  </nav>
                )
              }
              props={
                (
                  <div className="sm:flex sm:space-x-8">
                    <span
                      className="z-10 inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize cursor-pointer"
                    >
                      Summary
                    </span>
                    <Link
                      to={`/namespaces/${namespace}/repository/tags?repository=${repositoryObj.name}&repository_id=${repository_id}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Tag list
                    </Link>
                  </div>
                )
              } />
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className={(editorState ? "" : "pt-2 px-4") + " min-w-full min-h-full editor relative"} >
              {
                editorState ? (
                  <span></span>
                ) : (
                  <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-gray-700 bg-gray-200 hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3 absolute right-4 top-2"
                    onClick={() => { setEditorState(true) }}
                  >Edit</button>
                )
              }
              {
                editorState ? (
                  <Editor
                    placeholder='Write summary here with markdown'
                    value={overview}
                    plugins={[gfm()]}
                    onChange={e => setOverview(e)}
                  />
                ) : overview?.length === 0 ? (
                  <span className="text-gray-600">No description</span>
                ) : (
                  <Viewer plugins={[gfm()]} value={overview} />
                )
              }
            </div>
          </div>
          {
            editorState ? (
              <div
                className="flex flex-2 items-center justify-between border-gray-200 px-4 py-3 sm:px-6 border-t-0 bg-slate-100"
                aria-label="Pagination"
              >
                <div>
                </div>
                <div className="flex flex-1 justify-between sm:justify-end">
                  <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-gray-700 bg-gray-200 hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                    onClick={e => setEditorState(false)}
                  >Cancel</button>
                  <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                    onClick={() => { updateRepository() }}
                  >Update</button>
                </div>
              </div>
            ) : (
              <div></div>
            )
          }
        </div>
      </div >
    </>
  )
}
