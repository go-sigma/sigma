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
import { Fragment, useEffect, useState } from "react";
import { useNavigate } from 'react-router-dom';
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

  const [builderObj, setBuilderObj] = useState<IBuilderItem>();
  const [repositoryObj, setRepositoryObj] = useState<IRepositoryItem>();

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

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

  const [runnerObjs, setRunnerObjs] = useState<IBuilderRunnerItem[]>()

  useEffect(() => {
    if (builderObj === undefined) {
      return;
    }
    axios.get(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj.id}/runners/`).then(response => {
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
  }, [namespace, repository_id, builderObj])

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
  const [branchText, setBranchText] = useState("");
  const [branchTextValid, setBranchTextValid] = useState(true);
  useEffect(() => { branchText != "" && setBranchTextValid(/^[a-zA-Z0-9_-]{1,64}$/.test(branchText)) }, [branchText]);
  const [descriptionText, setDescriptionText] = useState("");
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => { descriptionText != "" && setDescriptionTextValid(/^.{0,50}$/.test(descriptionText)) }, [descriptionText]);

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
      tag: tagTemplateText,
    };
    if (builderObj.source !== "Dockerfile") {
      data["branch"] = branchText;
    }
    if (descriptionText !== "") {
      data["description"] = descriptionText;
    }
    axios.post(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj.id}/runners/run`, data).then(response => {
      if (response?.status === 201) {
        // TODO: redirect to log page
      } else {
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
                      onClick={() => { navigate(`/builders/setup/${repositoryObj?.builder?.id}?namespace=${namespace}&namespace_id=${repositoryObj?.namespace_id}&repository=${repositoryObj?.name}&repository_id=${repositoryObj?.id}&namespace_stick=true&repository_stick=true`); }}
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
                    <tbody className="bg-white divide-y divide-gray-100 max-h-max">
                      {
                        runnerObjs?.map(runnerObj => {
                          return (
                            <TableItem key={runnerObj.id} runnerObj={runnerObj} />
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

function TableItem({ runnerObj }: { runnerObj: IBuilderRunnerItem }) {
  console.log(runnerObj, runnerObj.tag);

  const navigate = useNavigate();
  return (
    <tr className="align-middle">
      <td className="px-6 py-4 w-5/6 whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer"
      // onClick={() => {
      //   window.open(repository.clone_url, "_blank");
      // }}
      >
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {runnerObj.tag}
              <span className="text-gray-500 font-normal ml-2">{runnerObj.description}</span>
            </span>
          </div>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {runnerObj.status}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {dayjs().to(dayjs(runnerObj.created_at))}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {dayjs().to(dayjs(runnerObj.created_at))}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer hover:text-gray-700"
        onClick={() => {
          navigate("/builders/setup")
        }}
      >
        Setup
      </td>
    </tr>
  )
}
