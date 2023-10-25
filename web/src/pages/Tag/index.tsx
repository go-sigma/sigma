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
import { Tooltip } from 'flowbite';
import { useCopyToClipboard } from 'react-use';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Link, useSearchParams, useParams } from 'react-router-dom';

import Settings from "../../Settings";
import { trimHTTP } from "../../utils";
import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Toast from "../../components/Notification";
import Pagination from "../../components/Pagination";
import HelmSvg from "../../components/svg/helm";
import DockerSvg from "../../components/svg/docker";

import TableItem from "./TableItem";

import { ITagList, IHTTPError, IEndpoint } from "../../interfaces";

export default function Tag({ localServer }: { localServer: string }) {
  const [tagList, setTagList] = useState<ITagList>({} as ITagList);
  const [refresh, setRefresh] = useState({});
  const [page, setPage] = useState(1);
  const [searchTag, setSearchTag] = useState("");
  const [total, setTotal] = useState(0);

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repository = searchParams.get('repository');
  const repository_id = searchParams.get('repository_id');

  const [, copyToClipboard] = useCopyToClipboard();

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
  }, [namespace, repository])

  const fetchTags = () => {
    let url = localServer + `/api/v1/namespaces/${namespace}/tags/?repository=${repository}&limit=${Settings.PageSize}&page=${page}&type=image&type=imageIndex`;
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

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Tag</title>
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
        <Menu localServer={localServer} item="tags" namespace={namespace} repository={repository || ""} />
        <div className="flex flex-col w-0 flex-1 overflow-hidden max-h-screen">
          <main className="">
            <Header title="Tag"
              // breadcrumb={
              //   (
              //     <nav className="flex" aria-label="Breadcrumb">
              //       <ol className="inline-flex items-center space-x-1 md:space-x-0">
              //         <li className="inline-flex items-center">
              //           <Link to={""} className="inline-flex items-center text-sm font-medium text-gray-700 hover:text-blue-600 dark:text-gray-400 dark:hover:text-white">
              //             <svg className="w-3 h-3 mr-2.5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
              //               <path d="m19.707 9.293-2-2-7-7a1 1 0 0 0-1.414 0l-7 7-2 2a1 1 0 0 0 1.414 1.414L2 10.414V18a2 2 0 0 0 2 2h3a1 1 0 0 0 1-1v-4a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v4a1 1 0 0 0 1 1h3a2 2 0 0 0 2-2v-7.586l.293.293a1 1 0 0 0 1.414-1.414Z" />
              //             </svg>
              //           </Link>
              //         </li>
              //         <li className="inline-flex items-center">
              //           <span className="inline-flex items-center text-sm font-medium text-gray-700 hover:text-blue-600 dark:text-gray-400 dark:hover:text-white">
              //             <Link to={`/namespaces/${namespace}/repositories`} className="inline-flex items-center text-sm font-medium text-gray-700 hover:text-blue-600 dark:text-gray-400 dark:hover:text-white">
              //               {namespace}
              //             </Link>
              //           </span>
              //         </li>
              //         <li>
              //           <div className="flex items-center">
              //             <span className="text-gray-500 text-sm ml-1">/</span>
              //             <span className="ml-1 text-sm font-medium text-gray-500 dark:text-gray-400">
              //               {repository?.substring((namespace?.length || 0) + 1)}
              //             </span>
              //             <span className="text-gray-500 text-sm ml-1">/</span>
              //           </div>
              //         </li>
              //       </ol>
              //     </nav>
              //   )
              // }
              props={
                (
                  <div className="sm:flex sm:space-x-8">
                    <Link
                      to={`/namespaces/${namespace}/repository/summary?repository=${repository}&repository_id=${repository_id}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Summary
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/repository/runners?repository=${repository}&repository_id=${repository_id}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Runners
                    </Link>
                    <span
                      className="z-10 inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize cursor-pointer"
                    >
                      Tag list
                    </span>
                  </div>
                )
              }
            />
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
                  return tag.artifact.type === "cosign" ? null : (
                    <div key={tag.id} className="p-4 border-t border-gray-200 hover:shadow-md last:hover:shadow-none">
                      {/* first row begin */}
                      <div className="flex">
                        <div className="flex-1 flex gap-1">
                          {
                            tag.artifact.config_media_type === "application/vnd.cncf.helm.config.v1+json" ? (
                              <HelmSvg />
                            ) : tag.artifact.media_type === "application/vnd.oci.image.manifest.v1+json" ||
                              tag.artifact.media_type === "application/vnd.docker.distribution.manifest.v2+json" ||
                              tag.artifact.media_type === "application/vnd.docker.distribution.manifest.list.v2+json" ||
                              tag.artifact.media_type === "application/vnd.oci.image.index.v1+json" ? (
                              <DockerSvg />
                            ) : null
                          }
                          <span className="font-semibold text-gray-600 cursor-pointer"
                            id={"tooltip-tag-name-" + index}
                            onClick={e => {
                              copyToClipboard(`${tag.name}`);
                              let tooltip = new Tooltip(document.getElementById("tooltip-top-content"),
                                document.getElementById("tooltip-tag-name-" + index.toString()), { triggerType: "click" });
                              tooltip.show();
                            }}
                          >
                            {tag.name}
                          </span>
                        </div>
                        <div>
                          <code className="block text-xs bg-gray-700 p-2 text-gray-50 cursor-pointer rounded-md w-96 text-ellipsis whitespace-nowrap overflow-hidden"
                            id={"tooltip-top-btn-" + index}
                            onClick={e => {
                              let copyText = `docker pull ${trimHTTP(endpoint)}/${repository}:${tag.name}`;
                              if (tag.artifact.config_media_type === "application/vnd.cncf.helm.config.v1+json") {
                                copyText = `helm pull ${trimHTTP(endpoint)}/${repository} --version ${tag.name}`
                              }
                              copyToClipboard(copyText);
                              let tooltip = new Tooltip(document.getElementById("tooltip-top-content"),
                                document.getElementById("tooltip-top-btn-" + index.toString()), { triggerType: "click" });
                              tooltip.show();
                            }}
                          >
                            {
                              tag.artifact.config_media_type === "application/vnd.cncf.helm.config.v1+json" ? (
                                <>
                                  helm pull {trimHTTP(endpoint)}/{repository} --version {tag.name}
                                </>
                              ) : (
                                <>
                                  docker pull {trimHTTP(endpoint)}/{repository}:{tag.name}
                                </>
                              )
                            }
                          </code>
                        </div>
                      </div>
                      {/* first row end */}

                      {/* second row begin */}
                      <div className="text-xs text-gray-600">
                        Last pushed <span className="font-semibold">{dayjs().to(dayjs(tag.pushed_at))}</span>
                      </div>
                      <div className="mt-2 text-xs text-gray-600">
                        Pull times <span className="font-semibold">{tag.pull_times === undefined ? 0 : tag.pull_times}</span>
                      </div>
                      {/* second row end */}

                      {/* third row begin */}
                      <table className="mt-2 min-w-full">
                        <thead>
                          <tr className="">
                            <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Digest
                            </th>
                            <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Type
                            </th>
                            <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Distro
                            </th>
                            <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Os/Arch
                            </th>
                            <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Signing
                            </th>
                            <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Last pull
                            </th>
                            {/* <th className="pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Pull Times
                            </th> */}
                            <th className="pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Vulnerabilities
                            </th>
                            <th className="pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Size
                            </th>
                          </tr>
                        </thead>
                        <TableItem namespace={namespace || ""} repository={repository || ""} artifact={tag.artifact} artifacts={tag.artifacts} />
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
