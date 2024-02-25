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

import Toast from 'react-hot-toast';
import axios from "axios";
import dayjs from 'dayjs';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Link, useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { Tooltip } from 'flowbite';

import Header from "../../components/Header";
import IMenu from "../../components/Menu";
import Notification from "../../components/Notification";
import Pagination from "../../components/Pagination";
import Settings from "../../Settings";
import { IGcArtifactRunnerList, IGcBlobRunnerList, IGcRepositoryRunnerList, IGcTagRunnerList, IHTTPError, IOrder } from "../../interfaces";

export default function ({ localServer }: { localServer: string }) {
  const location = useLocation();
  const navigate = useNavigate();

  const { namespace, resource } = useParams<{ namespace: string, resource: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId = searchParams.get('namespace_id') == null ? 0 : parseInt(searchParams.get('namespace_id') || "");

  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const [repositoryRunnerList, setRepositoryRunnerList] = useState<IGcRepositoryRunnerList>({} as IGcRepositoryRunnerList);
  const [tagRunnerList, setTagRunnerList] = useState<IGcTagRunnerList>({} as IGcTagRunnerList);
  const [artifactRunnerList, setArtifactRunnerList] = useState<IGcArtifactRunnerList>({} as IGcArtifactRunnerList);
  const [blobRunnerList, setBlobRunnerList] = useState<IGcBlobRunnerList>({} as IGcBlobRunnerList);

  const [refreshState, setRefreshState] = useState({});

  useEffect(() => {
    const timer = setInterval(() => {
      setRefreshState({});
    }, 5000);
    return () => {
      clearInterval(timer);
    };
  }, []);

  const fetchNamespace = () => {
    let url = localServer + `/api/v1/daemons/${resource}/${namespaceId}/runners/?limit=${Settings.PageSize}&page=${page}`;
    if (sortName !== "") {
      url += `&sort=${sortName}&method=${sortOrder.toString()}`
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        if (resource === "gc-repository") {
          const repositoryRunnerList = response.data as IGcRepositoryRunnerList;
          setRepositoryRunnerList(repositoryRunnerList);
          setTotal(repositoryRunnerList.total);
        } else if (resource === "gc-tag") {
          const tagRunnerList = response.data as IGcTagRunnerList;
          setTagRunnerList(tagRunnerList);
          setTotal(tagRunnerList.total);
        } else if (resource === "gc-artifact") {
          const artifactRunnerList = response.data as IGcArtifactRunnerList;
          setArtifactRunnerList(artifactRunnerList);
          setTotal(artifactRunnerList.total);
        } else if (resource === "gc-blob") {
          const blobRunnerList = response.data as IGcBlobRunnerList;
          setBlobRunnerList(blobRunnerList);
          setTotal(blobRunnerList.total);
        }
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

  const createGcRunner = () => {
    axios.post(localServer + `/api/v1/daemons/${resource}/${namespaceId}/runners/`, {}).then(response => {
      if (response?.status === 201) {
        let msg = "";
        if (resource === "gc-repository") {
          msg = "Garbage collect empty repository task will run in seconds";
        } else if (resource === "gc-tag") {
          msg = "Garbage collect tag task will run in seconds";
        } else if (resource === "gc-artifact") {
          msg = "Garbage collect artifact task will run in seconds";
        } else if (resource === "gc-blob") {
          msg = "Garbage collect blob task will run in seconds";
        }
        Toast.success(msg);
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - {location.pathname.startsWith("/settings") ? "Daemon Task Runner" : "Namespace Daemon Task"}</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item={location.pathname.startsWith("/settings") ? "daemon-tasks" : "repositories"} namespace={namespace} />
        <div className="flex flex-col w-0 flex-1 overflow-visible">
          <main className="relative z-0 focus:outline-none" tabIndex={0}>
            <Header title={location.pathname.startsWith("/settings") ? "Setting - Daemon Task" : "Namespace - Daemon Task"}
              props={
                location.pathname.startsWith("/settings") ? null : (
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
            <div className="pt-2 pb-2 flex justify-between">
              <div className="pr-2 pl-2">
              </div>
              <div className="pr-2 pl-2 flex flex-col">
                <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                  onClick={e => createGcRunner()}
                >Run</button>
              </div>
            </div>
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className="align-middle inline-block min-w-full border-gray-200">
              <table className="min-w-full flex-1">
                <thead>
                  <tr className="border-gray-200">
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Status</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Success</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Failed</span>
                    </th>
                    <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Elapsed</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Started At</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Ended At</span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {
                    resource === "gc-repository" && repositoryRunnerList.items?.map(runner => {
                      return (
                        <tr className="border-b" key={runner.id}>
                          <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                            onClick={e => {
                              if (location.pathname.startsWith("/settings")) {
                                navigate(`/settings/daemon-tasks/${resource}/${runner.id}/records?namespace_id=${namespaceId}`);
                              } else {
                                navigate(`/namespaces/${namespace}/daemon-tasks/${resource}/${runner.id}/records?namespace_id=${namespaceId}`);
                              }
                            }}
                          >
                            <div className="flex items-center space-x-3 lg:pl-2">
                              <div className="flex">
                                <div>
                                  {runner.status}
                                </div>
                                {
                                  runner.status == "Failed" ? (
                                    <>
                                      <div id={`tooltip-message-${runner.id}`} role="tooltip" className="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
                                        {runner.message}
                                        <div className="tooltip-arrow" data-popper-arrow></div>
                                      </div>
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5 mx-auto text-red-600" id={`target-tooltip-${runner.id}`}
                                        onClick={e => {
                                          let tooltip = new Tooltip(document.getElementById(`tooltip-message-${runner.id}`),
                                            document.getElementById(`target-tooltip-${runner.id}`), { triggerType: "click" });
                                          tooltip.show();
                                          e.stopPropagation();
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
                            {runner.success_count || 0}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.failed_count || 0}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.duration || "-"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.started_at || "-"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.ended_at || "-"}
                          </td>
                        </tr>
                      );
                    })
                  }
                  {
                    resource === "gc-tag" && tagRunnerList.items?.map(runner => {
                      return (
                        <tr className="border-b" key={runner.id}>
                          <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                            onClick={e => {
                              if (location.pathname.startsWith("/settings")) {
                                navigate(`/settings/daemon-tasks/${resource}/${runner.id}/records?namespace_id=${namespaceId}`);
                              } else {
                                navigate(`/namespaces/${namespace}/daemon-tasks/${resource}/${runner.id}/records?namespace_id=${namespaceId}`);
                              }
                            }}
                          >
                            <div className="flex items-center space-x-3 lg:pl-2">
                              <div className="flex">
                                <div>
                                  {runner.status}
                                </div>
                                {
                                  runner.status == "Failed" ? (
                                    <>
                                      <div id={`tooltip-message-${runner.id}`} role="tooltip" className="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
                                        {runner.message}
                                        <div className="tooltip-arrow" data-popper-arrow></div>
                                      </div>
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5 mx-auto text-red-600" id={`target-tooltip-${runner.id}`}
                                        onClick={e => {
                                          let tooltip = new Tooltip(document.getElementById(`tooltip-message-${runner.id}`),
                                            document.getElementById(`target-tooltip-${runner.id}`), { triggerType: "click" });
                                          tooltip.show();
                                          e.stopPropagation();
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
                            {runner.success_count || 0}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.failed_count || 0}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.duration || "-"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.started_at || "-"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.ended_at || "-"}
                          </td>
                        </tr>
                      );
                    })
                  }
                  {
                    resource === "gc-artifact" && artifactRunnerList.items?.map(runner => {
                      return (
                        <tr className="border-b" key={runner.id}>
                          <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                            onClick={e => {
                              if (location.pathname.startsWith("/settings")) {
                                navigate(`/settings/daemon-tasks/${resource}/${runner.id}/records?namespace_id=${namespaceId}`);
                              } else {
                                navigate(`/namespaces/${namespace}/daemon-tasks/${resource}/${runner.id}/records?namespace_id=${namespaceId}`);
                              }
                            }}
                          >
                            <div className="flex items-center space-x-3 lg:pl-2">
                              <div className="flex">
                                <div>
                                  {runner.status}
                                </div>
                                {
                                  runner.status == "Failed" ? (
                                    <>
                                      <div id={`tooltip-message-${runner.id}`} role="tooltip" className="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
                                        {runner.message}
                                        <div className="tooltip-arrow" data-popper-arrow></div>
                                      </div>
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5 mx-auto text-red-600" id={`target-tooltip-${runner.id}`}
                                        onClick={e => {
                                          let tooltip = new Tooltip(document.getElementById(`tooltip-message-${runner.id}`),
                                            document.getElementById(`target-tooltip-${runner.id}`), { triggerType: "click" });
                                          tooltip.show();
                                          e.stopPropagation();
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
                            {runner.success_count || 0}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.failed_count || 0}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.duration || "-"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.started_at || "-"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.ended_at || "-"}
                          </td>
                        </tr>
                      );
                    })
                  }
                  {
                    resource === "gc-blob" && blobRunnerList.items?.map(runner => {
                      return (
                        <tr className="border-b" key={runner.id}>
                          <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                            onClick={e => {
                              if (location.pathname.startsWith("/settings")) {
                                navigate(`/settings/daemon-tasks/${resource}/${runner.id}/records?namespace_id=${namespaceId}`);
                              } else {
                                navigate(`/namespaces/${namespace}/daemon-tasks/${resource}/${runner.id}/records?namespace_id=${namespaceId}`);
                              }
                            }}
                          >
                            <div className="flex items-center space-x-3 lg:pl-2">
                              <div className="flex">
                                <div>
                                  {runner.status}
                                </div>
                                {
                                  runner.status == "Failed" ? (
                                    <>
                                      <div id={`tooltip-message-${runner.id}`} role="tooltip" className="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
                                        {runner.message}
                                        <div className="tooltip-arrow" data-popper-arrow></div>
                                      </div>
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5 mx-auto text-red-600" id={`target-tooltip-${runner.id}`}
                                        onClick={e => {
                                          let tooltip = new Tooltip(document.getElementById(`tooltip-message-${runner.id}`),
                                            document.getElementById(`target-tooltip-${runner.id}`), { triggerType: "click" });
                                          tooltip.show();
                                          e.stopPropagation();
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
                            {runner.success_count || 0}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.failed_count || 0}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.duration || "-"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.started_at || "-"}
                          </td>
                          <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                            {runner.ended_at || "-"}
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
