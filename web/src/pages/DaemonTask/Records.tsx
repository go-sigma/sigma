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
import dayjs from 'dayjs';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Link, useLocation, useParams, useSearchParams } from 'react-router-dom';
import { Tooltip } from 'flowbite';

import Header from "../../components/Header";
import IMenu from "../../components/Menu";
import Notification from "../../components/Notification";
import Pagination from "../../components/Pagination";
import Settings from "../../Settings";
import { IGcArtifactRecordList, IHTTPError, IOrder } from "../../interfaces";

export default function ({ localServer }: { localServer: string }) {
  const location = useLocation();

  const { namespace, resource, runner_id } = useParams<{ namespace: string, resource: string, runner_id: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId = searchParams.get('namespace_id') == null ? 0 : parseInt(searchParams.get('namespace_id') || "");

  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const [refreshState, setRefreshState] = useState({});

  useEffect(() => {
    const timer = setInterval(() => {
      setRefreshState({});
    }, 5000);
    return () => {
      clearInterval(timer);
    };
  }, []);

  const [recordList, setRunnerList] = useState<IGcArtifactRecordList>({} as IGcArtifactRecordList);

  const fetchNamespace = () => {
    let url = localServer + `/api/v1/daemons/${resource}/${namespaceId}/runners/${runner_id}/records/?limit=${Settings.PageSize}&page=${page}`;
    if (sortName !== "") {
      url += `&sort=${sortName}&method=${sortOrder.toString()}`
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const recordList = response.data as IGcArtifactRecordList;
        setRunnerList(recordList);
        setTotal(recordList.total);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  useEffect(() => { fetchNamespace() }, [refreshState, page, sortOrder, sortName]);

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Namespace Daemon Task</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item={location.pathname == "/settings/daemon-tasks" ? "daemon-tasks" : "repositories"} namespace={namespace} />
        <div className="flex flex-col w-0 flex-1 overflow-visible">
          <main className="relative z-0 focus:outline-none" tabIndex={0}>
            <Header title={location.pathname == "/settings/daemon-tasks" ? "Setting - Daemon Task" : "Namespace - Daemon Task"}
              props={
                location.pathname == "/settings/daemon-tasks" ? null : (
                  <div className="flex space-x-8">
                    <Link
                      to={`/namespaces/${namespace}/repositories`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Repository list
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/namespace-users`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Users
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/webhooks`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Webhook
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/daemon-tasks?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize"
                    >
                      Daemon task
                    </Link>
                  </div>
                )
              } />
          </main>
          <div className="flex flex-1 overflow-visible">
            <div className="align-middle inline-block min-w-full border-gray-200">
              <table className="min-w-full flex-1 overflow-visible">
                <thead>
                  <tr className="border-gray-200">
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Digest</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Status</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Created At</span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {
                    recordList.items?.map(record => {
                      return (
                        <tr className="border-b" key={record.id}>
                          <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer">
                            {record.digest}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            <div className="flex items-center space-x-3 lg:pl-2">
                              <div className="flex">
                                <div>
                                  {record.status}
                                </div>
                                {
                                  record.status == "Failed" ? (
                                    <>
                                      <div id={`tooltip-message-${record.id}`} role="tooltip" className="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
                                        {record.message}
                                        <div className="tooltip-arrow" data-popper-arrow></div>
                                      </div>
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5 mx-auto text-red-600" id={`target-tooltip-${record.id}`}
                                        onClick={e => {
                                          let tooltip = new Tooltip(document.getElementById(`tooltip-message-${record.id}`),
                                            document.getElementById(`target-tooltip-${record.id}`), { triggerType: "click" });
                                          tooltip.show();
                                          e.preventDefault();
                                        }}
                                      >
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                      </svg>
                                    </>
                                  ) : null
                                }
                              </div>
                            </div>
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {record.created_at || "-"}
                          </td>
                        </tr>
                      );
                    })
                  }
                </tbody>
              </table>
            </div>
          </div>
          <div style={{ marginTop: "auto" }}>
            <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
          </div>
        </div>
      </div>
    </Fragment>
  )
}
