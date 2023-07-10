/**
 * Copyright 2023 XImager
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
import { useParams, Link } from 'react-router-dom';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';

import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Pagination from "../../components/Pagination";
import Settings from "../../Settings";

import TableItem from "../Repository/TableItem";
import "./index.css";

import { IRepository, IRepositoryList, IHTTPError } from "../../interfaces";

export default function Repository({ localServer }: { localServer: string }) {
  const [repositoryList, setRepositoryList] = useState<IRepositoryList>({} as IRepositoryList);
  const [refresh, setRefresh] = useState({});
  const [last, setLast] = useState(0);
  const [searchRepository, setSearchRepository] = useState("");
  const [total, setTotal] = useState(0);

  const { namespace } = useParams<{ namespace: string }>();

  useEffect(() => {
    let url = localServer + `/api/v1/namespaces/${namespace}/repositories/?limit=${Settings.PageSize}&last=${last}`;
    if (searchRepository !== "") {
      url += `&name=${searchRepository}`;
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const repositoryList = response.data as IRepositoryList;
        setRepositoryList(repositoryList);
        setTotal(repositoryList.total);
      }
    });
  }, [refresh, last]);

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>XImager - Repository</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <Menu item="Repository" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="flex-1 relative z-0 focus:outline-none" tabIndex={0}>
            <Header title="Repository" props={
              (
                <div className="pl-6 sm:flex sm:space-x-8 h-12">
                  <Link
                    to={`/namespaces/${namespace}/repositories`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                  >
                    Repository list
                  </Link>
                  <Link
                    to="#"
                    className="inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize"
                  >
                    Users
                  </Link>
                  <Link
                    to={`/namespaces/${namespace}/namespace-webhooks`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                  >
                    Webhook
                  </Link>
                  <Link
                    to={`/namespaces/${namespace}/namespace-daemon-tasks`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                  >
                    Daemon task
                  </Link>
                </div>
              )
            } />
            <div>
              <div className="align-middle inline-block min-w-full border-b border-gray-200">
                <table className="min-w-full">
                  <thead>
                    <tr className="border-gray-200">
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Repository</span>
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        Size
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        Tag count
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        Created at
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        Updated at
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        Action
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-100">
                    {
                      repositoryList.items?.map((repository, index) => {
                        return (
                          <TableItem key={index} index={index} namespace={namespace || ""} repository={repository} />
                        );
                      })
                    }
                  </tbody>
                </table>
              </div>
            </div>
          </main>
          <Pagination limit={Settings.PageSize} last={last} setLast={setLast} total={total} />
        </div>
      </div>
    </Fragment >
  )
}
