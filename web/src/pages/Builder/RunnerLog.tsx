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

import 'xterm/css/xterm.css';
import './index.css';

import axios from "axios";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { AttachAddon } from 'xterm-addon-attach';
import { Helmet, HelmetProvider } from "react-helmet-async";
import { Link, useSearchParams, useParams } from "react-router-dom";

import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Toast from "../../components/Notification";

let term = new Terminal({
  fontFamily: 'Menlo, Monaco, "Courier New", monospace',
  fontSize: 14,
  disableStdin: true,
  cursorInactiveStyle: 'none',
  cursorStyle: 'bar',
  convertEol: true,
});

import { IRepositoryItem, IHTTPError, IBuilderItem, IEndpoint, IBuilderRunnerItem, IRunOrRerunRunnerResponse } from "../../interfaces";

export default function ({ localServer }: { localServer: string }) {
  const navigate = useNavigate();

  const { namespace, runner_id } = useParams<{ namespace: string, runner_id: string }>();
  const [runnerId, setRunnerId] = useState(0);
  const [searchParams] = useSearchParams();
  const repository_id = parseInt(searchParams.get("repository_id") || "");
  const [repositoryObj, setRepositoryObj] = useState<IRepositoryItem>();
  const [builderObj, setBuilderObj] = useState<IBuilderItem>();
  const [runnerObj, setRunnerObj] = useState<IBuilderRunnerItem>();
  const [runnerStatus, setRunnerStatus] = useState('');

  useEffect(() => {
    if (runner_id === undefined) {
      return;
    }
    setRunnerId(parseInt(runner_id));
  }, [runner_id]);

  useEffect(() => {
    term.open(document.getElementById("terminal") || new HTMLElement);
    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    fitAddon.fit();
  }, [])

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
  }, []);

  useEffect(() => {
    axios.get(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}`).then(response => {
      if (response?.status === 200) {
        const r = response.data as IRepositoryItem;
        setRepositoryObj(r);
        if (r.builder !== undefined && r.builder !== null) {
          setBuilderObj(r.builder);
        }
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespace, repository_id]);

  useEffect(() => {
    if (runnerStatus === '') {
      return;
    }
    if (repositoryObj === undefined || repositoryObj === null) {
      return;
    }
    let protocol = 'wss';
    if (endpoint.startsWith("https")) {
      protocol = 'wss';
    } else {
      protocol = 'ws';
    }
    let server = '';
    if (protocol.length === 3) {
      server = endpoint.substring(8);
    } else {
      server = endpoint.substring(7);
    }
    console.log(runnerStatus);
    if (runnerStatus === 'Pending') {
      term.clear();

      term.writeln('Pending...');
    } else if (runnerStatus === 'Building' || runnerStatus === 'Success' || runnerStatus === 'Failed') {
      term.clear();
      let ws = new WebSocket(`${protocol}://${server}/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj?.id}/runners/${runnerId}/log`);
      const attachAddon = new AttachAddon(ws);
      term.loadAddon(attachAddon);
    }
  }, [runnerStatus, repositoryObj, repository_id])

  const [refreshState, setRefreshState] = useState({});

  useEffect(() => {
    const timer = setInterval(() => {
      if (runnerObj === undefined) {
        return;
      }
      if (!(runnerObj?.status === 'Success' || runnerObj?.status === 'Failed')) {
        setRefreshState({});
      }
    }, 5000);
    return () => {
      clearInterval(timer);
    };
  }, [runnerObj]);

  useEffect(() => {
    if (builderObj === undefined || repositoryObj === undefined) {
      return;
    }
    axios.get(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj?.id}/runners/${runnerId}`).then(response => {
      if (response?.status === 200) {
        let data = response.data as IBuilderRunnerItem;
        setRunnerObj(data);
        setRunnerStatus(data.status);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [refreshState, builderObj, repositoryObj]);

  let rerunAction = () => {
    axios.get(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj?.id}/runners/${runnerId}/rerun`).then(response => {
      if (response?.status === 200) {
        let data = response.data as IRunOrRerunRunnerResponse;
        navigate(`/namespaces/${namespace}/repository/runner-logs/${data.runner_id}?repository_id=${repository_id}`, { replace: true });
        window.location.reload();
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  let stopAction = () => {
    axios.get(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj?.id}/runners/${runnerId}/stop`).then(response => {
      if (response?.status === 204) { } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  return (
    <>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Runner log</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen max-h-screen flex overflow-hidden bg-white">
        <Menu localServer={localServer} item="repositories" namespace={"namespace"} />
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
                    <Link
                      to={`/namespaces/${namespace}/repository/summary?repository=${repositoryObj?.name}&repository_id=${repository_id}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Summary
                    </Link>
                    <span
                      className="z-10 inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize cursor-pointer"
                    >
                      Runners
                    </span>
                    <Link
                      to={`/namespaces/${namespace}/repository/tags?repository=${repositoryObj?.name}&repository_id=${repository_id}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Tag list
                    </Link>
                  </div>
                )
              }
            />
            <div className="pt-2 pb-2 flex flex-row-reverse justify-between">
              <div className="pr-2 pl-2">
                {
                  runnerObj?.status === "Success" || runnerObj?.status === "Failed" ? (
                    <button className="my-auto px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                      onClick={() => { rerunAction() }}
                    >Rerun</button>
                  ) : (
                    <button className="my-auto px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                      onClick={() => { stopAction() }}
                    >Stop</button>
                  )
                }
              </div>
            </div>
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className="container-children w-full h-full">
              <div id="terminal" className="h-full"></div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
