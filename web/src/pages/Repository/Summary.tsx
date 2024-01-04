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

import axios from "axios";
import gfm from "@bytemd/plugin-gfm";
import { useEffect, useState } from "react";
import { Editor, Viewer } from "@bytemd/react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { Link, useSearchParams, useParams } from "react-router-dom";

import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Toast from "../../components/Notification";
import { IRepositoryItem, IHTTPError, ISystemConfig } from "../../interfaces";

import "./index.css";

export default function ({ localServer }: { localServer: string }) {
  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repository_id = parseInt(searchParams.get("repository_id") || "");
  const namespaceId = parseInt(searchParams.get("namespace_id") || "");

  const [repositoryObj, setRepositoryObj] = useState<IRepositoryItem>({} as IRepositoryItem);

  const [overview, setOverview] = useState("");
  const [overviewValid, setOverviewValid] = useState(true);
  useEffect(() => { setOverviewValid(overview?.length < 100000) }, [overview]);

  useEffect(() => {
    axios.get(localServer + `/api/v1/namespaces/${namespaceId}/repositories/${repository_id}`).then(response => {
      if (response.status === 200) {
        const r = response.data as IRepositoryItem;
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
    axios.put(localServer + `/api/v1/namespaces/${namespaceId}/repositories/${repository_id}`, {
      overview: overview,
    } as IRepositoryItem, {}).then(response => {
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

  const [gotConfig, setGotConfig] = useState(false);
  const [config, setConfig] = useState<ISystemConfig>({
    daemon: {
      builder: false
    }
  } as ISystemConfig);

  useEffect(() => {
    axios.get(localServer + "/api/v1/systems/config").then(response => {
      if (response.status === 200) {
        const config = response.data as ISystemConfig;
        setConfig(config);
        setGotConfig(true);
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
    <>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Repository Summary</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen max-h-screen flex overflow-hidden bg-white">
        <Menu localServer={localServer} item="repositories" namespace={namespace} namespace_id={namespaceId.toString()} />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title="Repository"
              props={
                gotConfig && (
                  <div className="sm:flex sm:space-x-8">
                    <span
                      className="z-10 inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize cursor-pointer"
                    >
                      Summary
                    </span>
                    {
                      config.daemon.builder && (
                        <Link
                          to={`/namespaces/${namespace}/repository/runners?repository=${repositoryObj.name}&repository_id=${repository_id}&namespace_id=${namespaceId}`}
                          className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                        >
                          Builder
                        </Link>
                      )
                    }
                    <Link
                      to={`/namespaces/${namespace}/repository/tags?repository=${repositoryObj.name}&repository_id=${repository_id}&namespace_id=${namespaceId}`}
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
