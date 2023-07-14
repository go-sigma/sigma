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
import dayjs from 'dayjs';
import { Tooltip } from 'flowbite';
import { useParams } from 'react-router-dom';
import { useCopyToClipboard } from 'react-use';
import { useSearchParams } from 'react-router-dom';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import relativeTime from 'dayjs/plugin/relativeTime';

import Settings from "../../Settings";
import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Toast from "../../components/Notification";
import Pagination from "../../components/Pagination";

import "./index.css";
import TableItem from "./TableItem";

import { ITag, ITagList, IHTTPError } from "../../interfaces";

export default function Tag({ localServer }: { localServer: string }) {
  const [tagList, setTagList] = useState<ITagList>({} as ITagList);
  const [refresh, setRefresh] = useState({});
  const [page, setPage] = useState(1);
  const [searchTag, setSearchTag] = useState("");
  const [total, setTotal] = useState(0);

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repository = searchParams.get('repository');

  const [, copyToClipboard] = useCopyToClipboard();

  const fetchTags = () => {
    let url = localServer + `/api/v1/namespaces/${namespace}/tags/?repository=${repository}&limit=${Settings.PageSize}&page=${page}`;
    if (searchTag !== "") {
      url += `&name=${searchTag}`;
    }
    axios.get(url).then(response => {
      if (response.status === 200) {
        const tagList = response.data as ITagList;
        setTagList(tagList);
        setTotal(tagList.total);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  useEffect(fetchTags, [refresh, page]);

  const imageDomain = () => {
    if (localServer.startsWith("http://")) {
      return localServer.substring(7);
    } else if (localServer.startsWith("https://")) {
      return localServer.substring(8)
    } else {
      return localServer;
    }
  }

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>XImager - Tag</title>
        </Helmet>
      </HelmetProvider>
      <div
        id={"tooltip-top-content"}
        role="tooltip"
        className="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
        Copied!
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <Menu item="Tag" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden max-h-screen">
          <main className="">
            <Header title="Tag" />
            <div className="pt-2 pb-2 flex">
              <div className="pr-2 pl-2">
                <div className="flex gap-4">
                  <div className="relative mt-2 flex items-center">
                    <label
                      htmlFor="tagSearch"
                      className="absolute -top-2 left-2 inline-block bg-white px-1 text-xs font-medium text-gray-900"
                    >
                      Tag
                    </label>
                    <input
                      type="text"
                      id="tagSearch"
                      placeholder="search tag"
                      value={searchTag}
                      onChange={e => { setSearchTag(e.target.value); }}
                      onKeyDown={e => { if (e.key == "Enter") { fetchTags() } }}
                      className="block w-full h-10 rounded-md border-0 py-1.5 pr-14 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    />
                    <div className="absolute inset-y-0 right-0 flex py-1.5 pr-1.5">
                      <kbd className="inline-flex items-center rounded border border-gray-200 px-1 font-sans text-xs text-gray-400">
                        enter
                      </kbd>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </main>
          <div className="flex-1 overflow-y-auto">
            <div className="flex flex-col border-b border-gray-200">
              {
                tagList.items?.map((tag, index) => {
                  return (
                    <div key={tag.id} className="p-4 border-t border-gray-200 hover:shadow-md last:hover:shadow-none">
                      {/* first row begin */}
                      <div className="flex">
                        <div className="flex-1">
                          <span className="font-semibold text-gray-600">
                            {tag.name}
                          </span>
                        </div>
                        <div>
                          <code className="block text-xs bg-gray-700 p-2 text-gray-200 cursor-pointer rounded-md w-96 text-ellipsis whitespace-nowrap overflow-hidden"
                            id={"tooltip-top-btn-" + index}
                            onClick={e => {
                              copyToClipboard(`docker pull ${imageDomain()}/${repository}:${tag.name}`);
                              let tooltip = new Tooltip(document.getElementById("tooltip-top-content"),
                                document.getElementById("tooltip-top-btn-" + index.toString()), { triggerType: "click" });
                              tooltip.show();
                              setTimeout(() => { tooltip.hide() }, 1500)
                            }}
                          >docker pull {imageDomain()}/{repository}:{tag.name}</code>
                        </div>
                      </div>
                      {/* first row end */}

                      {/* second row begin */}
                      <div className="text-sm text-gray-600">
                        Last pushed <span className="font-semibold">{dayjs().to(dayjs(tag.pushed_at))}</span>
                      </div>
                      {/* second row end */}

                      {/* third row begin */}
                      <table className="min-w-full">
                        <thead>
                          <tr className="">
                            <th className="pt-5 pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Digest
                            </th>
                            <th className="pt-5 pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Os/Arch
                            </th>
                            <th className="pt-5 pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Size
                            </th>
                            <th className="pt-5 pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Last pull
                            </th>
                            <th className="pt-5 pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Pull Times
                            </th>
                            <th className="pt-5 pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Command
                            </th>
                            <th className="pt-5 pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Vulnerabilities
                            </th>
                          </tr>
                        </thead>
                        <TableItem localServer={localServer} namespace={namespace || ""} repository={repository || ""} artifactDigest={tag.digest} artifact={tag.raw} />
                      </table>
                      {/* third row end */}
                    </div>
                  );
                })
              }
            </div>
          </div>
          <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
        </div>
      </div>
    </Fragment>
  )
}
