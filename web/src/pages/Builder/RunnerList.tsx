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
import dayjs from "dayjs";
import { useNavigate } from 'react-router-dom';
import { Fragment, useEffect, useState } from "react";
import { Dialog, Transition } from "@headlessui/react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { Link, useSearchParams, useParams } from "react-router-dom";

import Settings from "../../Settings";
import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Toast from "../../components/Notification";
import Pagination from "../../components/Pagination";
import OrderHeader from "../../components/OrderHeader";

import { IRepositoryItem, IHTTPError, IBuilderItem, IOrder, IBuilderRunnerItem, IBuilderRunnerList, IRunOrRerunRunnerResponse } from "../../interfaces";

export default function ({ localServer }: { localServer: string }) {
  const navigate = useNavigate();

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repository_id = parseInt(searchParams.get("repository_id") || "");
  const namespaceId = parseInt(searchParams.get("namespace_id") || "");

  const [builderObj, setBuilderObj] = useState<IBuilderItem>();
  const [repositoryObj, setRepositoryObj] = useState<IRepositoryItem>();

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const [branchText, setBranchText] = useState("");
  const [branchTextValid, setBranchTextValid] = useState(true);
  useEffect(() => { branchText != "" && setBranchTextValid(/^[a-zA-Z0-9_-]{1,64}$/.test(branchText)) }, [branchText]);

  useEffect(() => {
    axios.get(localServer + `/api/v1/namespaces/${namespaceId}/repositories/${repository_id}`).then(response => {
      if (response?.status === 200) {
        const r = response.data as IRepositoryItem;
        setRepositoryObj(r);
        if (r.builder !== undefined && r.builder !== null) {
          setBuilderObj(r.builder);
        }
        setBranchText(r.builder?.scm_branch || "");
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespace, repository_id]);

  const [runnerObjs, setRunnerObjs] = useState<IBuilderRunnerItem[]>()

  const [createRunnerModal, setCreateRunnerModal] = useState(false);
  const [tagTemplateText, setTagTemplateText] = useState("");
  const [tagTemplateTextValid, setTagTemplateTextValid] = useState(true);
  useEffect(() => {
    if (tagTemplateText == "") {
      return;
    }
    axios.get(localServer + `/api/v1/validators/tag?tag=${tagTemplateText}`).then(response => {
      if (response?.status === 204) {
        setTagTemplateTextValid(true);
      } else {
        setTagTemplateTextValid(false);
      }
    }).catch(error => {
      console.log(error);
      setTagTemplateTextValid(false);
    });
  }, [tagTemplateText]);

  const [descriptionText, setDescriptionText] = useState("");
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => { descriptionText != "" && setDescriptionTextValid(/^.{0,50}$/.test(descriptionText)) }, [descriptionText]);

  useEffect(() => {
    if (repositoryObj != undefined) {
      let tag = localStorage.getItem(`${repositoryObj?.name}-tag`);
      if (tag != null) {
        setTagTemplateText(tag);
      }
    }
  }, [repositoryObj]);

  const createRunner = () => {
    if (builderObj == undefined) {
      return;
    }
    if (!(tagTemplateTextValid && branchTextValid && descriptionTextValid)) {
      Toast({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    if (builderObj.source !== "Dockerfile" && branchText === "") {
      Toast({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    const data: { [key: string]: any } = {
      raw_tag: tagTemplateText,
    };
    if (builderObj.source !== "Dockerfile") {
      data["scm_branch"] = branchText;
    }
    if (descriptionText !== "") {
      data["description"] = descriptionText;
    }
    if (repositoryObj != undefined && repositoryObj?.name != "") {
      localStorage.setItem(`${repositoryObj?.name}-tag`, `${tagTemplateText}`);
    }
    axios.post(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj.id}/runners/run`, data).then(response => {
      if (response?.status === 201) {
        let data = response.data as IRunOrRerunRunnerResponse;
        navigate(`/namespaces/${namespace}/repository/runner-logs/${data.runner_id}?repository_id=${repositoryObj?.id}&namespace_id=${namespaceId}`);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const [refreshState, setRefreshState] = useState({});

  useEffect(() => {
    const timer = setInterval(() => {
      if (builderObj === undefined) {
        return;
      }
      setRefreshState({});
    }, 5000);
    return () => {
      clearInterval(timer);
    };
  }, [builderObj]);

  const [createdAtOrder, setCreatedAtOrder] = useState(IOrder.None);
  const [costOrder, setCostOrder] = useState(IOrder.None);
  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");

  const resetOrder = () => {
    setCostOrder(IOrder.None);
    setCreatedAtOrder(IOrder.None);
  }

  useEffect(() => {
    if (builderObj === undefined) {
      return;
    }
    let url = localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj.id}/runners/?limit=${Settings.PageSize}&page=${page}`
    if (sortName !== "" && sortOrder !== IOrder.None) {
      url += `&sort=${sortName}&method=${sortOrder.toString()}`
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const r = response.data as IBuilderRunnerList;
        setRunnerObjs(r.items);
        setTotal(r.total);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespace, repository_id, builderObj, refreshState, sortOrder, sortName]);

  return (
    <>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Runners</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen max-h-screen flex overflow-hidden bg-white">
        <Menu localServer={localServer} item="repositories" namespace={namespace} repository={repositoryObj?.name} selfClick={true} />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title="Repository"
              props={
                (
                  <div className="sm:flex sm:space-x-8">
                    <Link
                      to={`/namespaces/${namespace}/repository/summary?repository=${repositoryObj?.name}&repository_id=${repository_id}&namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Summary
                    </Link>
                    <span
                      className="z-10 inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize cursor-pointer"
                    >
                      Builder
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
                      onClick={() => { navigate(`/builders/setup/${repositoryObj?.builder?.id}?namespace=${namespace}&namespace_id=${repositoryObj?.namespace_id}&repository=${repositoryObj?.name}&repository_id=${repositoryObj?.id}&namespace_stick=true&repository_stick=true&back_to=/namespaces/${namespace}/repository/runners?repository=${repositoryObj?.name}%26repository_id=${repositoryObj?.id}%26namespace_id=${namespaceId}`); }}
                    >Update</button>
                  )
                }
                {
                  builderObj === undefined ? null : (
                    <button className="my-auto px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                      onClick={() => { setCreateRunnerModal(true) }}
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
                    Please <Link to={`/builders/setup?namespace=${namespace}&namespace_id=${repositoryObj?.namespace_id}&repository=${repositoryObj?.name}&repository_id=${repositoryObj?.id}&namespace_stick=true&repository_stick=true`} className="hover:underline-offset-1">configure</Link> the builder first.
                  </div>
                ) : (
                  <table className="min-w-full flex-1">
                    <thead>
                      <tr>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <span className="lg:pl-2">Tag</span>
                        </th>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <span className="lg:pl-2">Status</span>
                        </th>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <OrderHeader text={"Elapsed"}
                            orderStatus={costOrder} setOrder={e => {
                              resetOrder();
                              setCostOrder(e);
                              setSortOrder(e);
                              setSortName("duration");
                            }} />
                        </th>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <OrderHeader text={"Created at"}
                            orderStatus={createdAtOrder} setOrder={e => {
                              resetOrder();
                              setCreatedAtOrder(e);
                              setSortOrder(e);
                              setSortName("created_at");
                            }} />
                        </th>
                        <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          Action
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-100 max-h-max">
                      {
                        runnerObjs?.map(runnerObj => {
                          return (
                            <TableItem key={runnerObj.id} localServer={localServer} namespace={namespace || ""} repositoryObj={repositoryObj} runnerObj={runnerObj} />
                          );
                        })
                      }
                    </tbody>
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
        </div>
      </div >

      <Transition.Root show={createRunnerModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setCreateRunnerModal}>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </Transition.Child>

          <div className="fixed inset-0 z-10 overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
                  <label htmlFor="tagTemplate" className="block text-sm font-medium leading-6 text-gray-900">
                    <span className="text-red-600">*</span>Tag
                  </label>
                  <div className="relative mt-2 rounded-md shadow-sm">
                    <input
                      type="text"
                      id="tagTemplate"
                      name="tagTemplate"
                      placeholder="tag template"
                      className={(tagTemplateTextValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                      value={tagTemplateText}
                      onChange={e => {
                        setTagTemplateText(e.target.value);
                      }}
                    />
                    {
                      tagTemplateTextValid ? (
                        <div></div>
                      ) : (
                        <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                          </svg>
                        </div>
                      )
                    }
                  </div>
                  <p className="mt-1 text-xs text-red-600">
                    {
                      tagTemplateTextValid ? (
                        <span></span>
                      ) : (
                        <span>
                          Not a valid tag template, you can try 'main', '&#123;&#123;.ScmRef&#125;&#125;', '&#123;&#123;.ScmBrach&#125;&#125;'.
                        </span>
                      )
                    }
                  </p>
                  {
                    builderObj?.source !== "Dockerfile" ? (
                      <>
                        <label htmlFor="first-name" className="block text-sm font-medium leading-6 text-gray-900">
                          <span className="text-red-600">*</span>Branch
                        </label>
                        <div className="relative mt-2 rounded-md shadow-sm">
                          <input
                            type="text"
                            name="namespace"
                            placeholder="1-64 characters"
                            className={(branchTextValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                            value={branchText}
                            onChange={e => {
                              setBranchText(e.target.value);
                            }}
                          />
                          {
                            branchTextValid ? (
                              <div></div>
                            ) : (
                              <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                </svg>
                              </div>
                            )
                          }
                        </div>
                        <p className="mt-1 text-xs text-red-600">
                          {
                            branchTextValid ? (
                              <span></span>
                            ) : (
                              <span>
                                Not a valid branch, you can try 'main', 'master', 'dev'.
                              </span>
                            )
                          }
                        </p>
                      </>
                    ) : null
                  }

                  <label htmlFor="first-name" className="block text-sm font-medium text-gray-700">
                    Description
                  </label>
                  <div className="relative mt-2 rounded-md shadow-sm">
                    <textarea
                      name="description"
                      placeholder="30 characters"
                      className={(descriptionTextValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                      value={descriptionText}
                      onChange={e => setDescriptionText(e.target.value)}
                    />
                    {
                      descriptionTextValid ? (
                        <div></div>
                      ) : (
                        <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                          </svg>
                        </div>
                      )
                    }
                  </div>
                  <p className="mt-1 text-xs text-red-600">
                    {
                      descriptionTextValid ? (
                        <span></span>
                      ) : (
                        <span>
                          Not a valid description, max 50 characters.
                        </span>
                      )
                    }
                  </p>

                  <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                    <button
                      type="button"
                      className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                      onClick={() => createRunner()}
                    >
                      Create
                    </button>
                    <button
                      type="button"
                      className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                      onClick={() => setCreateRunnerModal(false)}
                    >
                      Cancel
                    </button>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>

    </>
  )
}

function TableItem({ localServer, namespace, repositoryObj, runnerObj }: { localServer: string, namespace: string, repositoryObj: IRepositoryItem | undefined, runnerObj: IBuilderRunnerItem }) {
  const navigate = useNavigate();

  let rerunAction = () => {
    axios.get(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repositoryObj?.id}/builders/${runnerObj.builder_id}/runners/${runnerObj.id}/rerun`).then(response => {
      if (response?.status === 200) {
        let data = response.data as IRunOrRerunRunnerResponse;
        navigate(`/namespaces/${namespace}/repository/runner-logs/${data.runner_id}?repository_id=${repositoryObj?.id}&namespace_id=${repositoryObj?.namespace_id}`, { replace: true });
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
    axios.get(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repositoryObj?.id}/builders/${runnerObj.builder_id}/runners/${runnerObj.id}/stop`).then(response => {
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
    <tr className="align-middle">
      <td className="px-6 py-4 w-5/6 whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer"
        onClick={() => {
          navigate(`/namespaces/${namespace}/repository/runner-logs/${runnerObj.id}?repository_id=${repositoryObj?.id}&namespace_id=${repositoryObj?.namespace_id}`);
        }}
      >
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {runnerObj.tag !== null ? runnerObj.tag : runnerObj.raw_tag}
              <span className="text-gray-500 font-normal ml-2">{runnerObj.description}</span>
            </span>
          </div>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {runnerObj.status}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {runnerObj.duration || "-"}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {dayjs().to(dayjs(runnerObj.created_at))}
      </td>
      {
        runnerObj?.status === "Success" || runnerObj?.status === "Failed" || runnerObj?.status === "Stopped" ? (
          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer hover:text-gray-700"
            onClick={() => { rerunAction() }}
          >
            Rerun
          </td>
        ) : (
          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer hover:text-gray-700"
            onClick={() => {
              if (runnerObj?.status === "Stopping") {
                return;
              }
              stopAction()
            }}
          >
            Stop
          </td>
        )
      }
    </tr>
  )
}
