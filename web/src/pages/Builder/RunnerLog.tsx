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

import '@xterm/xterm/css/xterm.css';

import axios from "axios";
import { AttachAddon } from '@xterm/addon-attach';
import { FitAddon } from "@xterm/addon-fit";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { Terminal } from "@xterm/xterm";
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import Header from "../../components/Header";
import Menu from "../../components/Menu";
import Toast from "../../components/Notification";
import { IBuilderItem, IBuilderRunnerItem, IEndpoint, IHTTPError, IRepositoryItem, IRunOrRerunRunnerResponse } from "../../interfaces";

let term = new Terminal({
  fontFamily: 'Menlo, Monaco, "Courier New", monospace',
  fontSize: 14,
  disableStdin: true,
  cursorInactiveStyle: 'none',
  cursorStyle: 'bar',
  convertEol: true,
});


export default function ({ localServer }: { localServer: string }) {
  const navigate = useNavigate();

  const { namespace, runner_id } = useParams<{ namespace: string, runner_id: string }>();
  const [runnerId, setRunnerId] = useState(0);
  const [searchParams] = useSearchParams();
  const repository_id = parseInt(searchParams.get("repository_id") || "");
  const namespaceId = parseInt(searchParams.get("namespace_id") || "");
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
    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(document.getElementById("terminal") || new HTMLElement);
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
    axios.get(localServer + `/api/v1/namespaces/${namespaceId}/repositories/${repository_id}`).then(response => {
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
      // const fitAddon = new FitAddon();
      // term.loadAddon(fitAddon);
      // fitAddon.fit();
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
        navigate(`/namespaces/${namespace}/repository/runner-logs/${data.runner_id}?repository_id=${repository_id}&namespace_id=${repositoryObj?.namespace_id}`, { replace: true });
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
        <Menu localServer={localServer} item="repositories" namespace={namespace} repository={repositoryObj?.name} selfClick={true} />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title="Repository"
              props={
                (
                  <div className="flex space-x-8">
                    <Link
                      to={`/namespaces/${namespace}/repository/summary?repository=${repositoryObj?.name}&repository_id=${repository_id}&namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Summary
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/repository/runners?repository=${repositoryObj?.name}&repository_id=${repository_id}&namespace_id=${namespaceId}`}
                      className="z-10 inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize cursor-pointer"
                    >
                      Builder
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/repository/tags?repository=${repositoryObj?.name}&repository_id=${repository_id}&namespace_id=${namespaceId}`}
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
                  runnerObj?.status === "Success" || runnerObj?.status === "Failed" || runnerObj?.status === "Stopped" ? (
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
              <div className='pr-2 pl-2 flex gap-1'>
                <div className='text-gray-600 px-2 py-2 h-10'>
                  Elapsed: {runnerObj?.status != "Failed" && runnerObj?.status != "Success" && runnerObj?.status != "Stopped" ? Math.floor((Date.now() - (runnerObj?.started_at || Date.now())) / 1000) : Math.floor((runnerObj.raw_duration || 0) / 1000)}s
                </div>
              </div>
            </div>
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className="w-full h-full pt-2 pl-3 bg-black">
              <div id="terminal" className="h-full"></div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
