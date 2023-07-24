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
import { useParams, Link } from 'react-router-dom';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';

import Settings from "../../Settings";
import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Pagination from "../../components/Pagination";

import "./index.css";

export default function Repository({ localServer }: { localServer: string }) {
  const { namespace } = useParams<{ namespace: string }>();

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>XImager - Namespace Daemon Task</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <Menu localServer={localServer} item="Repository" />
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
                    to={`/namespaces/${namespace}/namespace-users`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
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
                    to="#"

                    className="inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize"
                  >
                    Daemon task
                  </Link>
                </div>
              )
            } />
            <div>
              <div className="align-middle inline-block min-w-full border-b border-gray-200">

              </div>
            </div>
          </main>
          <Pagination limit={Settings.PageSize} page={1} setPage={(e) => console.log(e)} total={10} />
        </div>
      </div>
    </Fragment >
  )
}
