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
import { useNavigate } from 'react-router-dom';
import { useEffect, useState } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { Link, useSearchParams, useParams } from "react-router-dom";

import Settings from "../../Settings";
import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Toast from "../../components/Notification";
import Pagination from "../../components/Pagination";
import OrderHeader from "../../components/OrderHeader";

import { IRepositoryItem, IHTTPError, IBuilderItem, IOrder } from "../../interfaces";

export default function ({ localServer }: { localServer: string }) {
  const navigate = useNavigate();

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repository_id = parseInt(searchParams.get("repository_id") || "");

  const [builderObj, setBuilderObj] = useState<IBuilderItem>();
  const [repositoryObj, setRepositoryObj] = useState<IRepositoryItem>();

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    axios.get(localServer + `/api/v1/namespaces/${namespace}/repositories/${repository_id}`).then(response => {
      if (response?.status === 200) {
        const r = response.data as IRepositoryItem;
        setRepositoryObj(r);
        if (r.builder !== undefined && r.builder !== null) {
          setBuilderObj(r.builder);
        }
      } else {
        console.log(response);
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespace, repository_id]);

  useEffect(() => {
    if (builderObj === undefined) {
      return;
    }
  }, [namespace, repository_id, builderObj])

  return (
    <>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Runners</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen max-h-screen flex overflow-hidden bg-white">
        <Menu localServer={localServer} item="repositories" namespace={namespace} />
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
              } />
            <div className="pt-2 pb-2 flex flex-row-reverse justify-between">
              <div className="pr-2 pl-2">
                {
                  builderObj === undefined ? (
                    <button className="my-auto px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                      onClick={() => { navigate(`/builders/setup?namespace=${namespace}&namespace_id=${repositoryObj?.namespace_id}&repository=${repositoryObj?.name}&repository_id=${repositoryObj?.id}&namespace_stick=true&repository_stick=true`); }}
                    >Configure</button>
                  ) : null
                }
                {
                  builderObj === undefined ? null : (
                    <button className="my-auto px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                    // onClick={() => { setCreateNamespaceModal(true) }}
                    >Update</button>
                  )
                }
                {
                  builderObj === undefined ? null : (
                    <button className="my-auto px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                    // onClick={() => { setCreateNamespaceModal(true) }}
                    >Build</button>
                  )
                }
              </div>
            </div>
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className="border-t w-full">
              {
                builderObj === undefined ? (
                  <div className="my-2 mx-2 text-gray-500 text-md">
                    Please <Link to="" className="hover:underline-offset-1">configure</Link> the builder first.
                  </div>
                ) : (
                  <table className="min-w-full flex-1">
                    <thead>
                      <tr>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <span className="lg:pl-2">Tag</span>
                        </th>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          Cost
                        </th>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <OrderHeader text={"Created at"}
                            orderStatus={IOrder.Asc} setOrder={(e) => {
                              // resetOrder();
                              // setCreatedAtOrder(e);
                              // setSortOrder(e);
                              // setSortName("created_at");
                            }} />
                        </th>
                        <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          Action
                        </th>
                      </tr>
                    </thead>
                    {/* <tbody className="bg-white divide-y divide-gray-100 max-h-max">
                    {
                      namespaceList.items?.map((namespace, index) => {
                        return (
                          <TableItem key={namespace.id} index={index} namespace={namespace} localServer={localServer} setRefresh={setRefresh} />
                        );
                      })
                    }
                  </tbody> */}
                  </table>
                )
              }
            </div>
          </div>
          {
            builderObj === undefined ? null : (
              <div style={{ marginTop: "auto" }}>
                <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
              </div>
            )
          }
          {/* <div
            className="flex flex-2 items-center justify-between border-gray-200 px-4 py-3 sm:px-6 border-t-0 bg-slate-100"
            aria-label="Pagination"
          >
            <div>
            </div>
            <div className="flex flex-1 justify-between sm:justify-end">
              <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-gray-700 bg-gray-200 hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
              // onClick={e => setEditorState(false)}
              >Cancel</button>
              <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
              // onClick={() => { updateRepository() }}
              >Update</button>
            </div>
          </div> */}
        </div>
      </div >
    </>
  )
}
