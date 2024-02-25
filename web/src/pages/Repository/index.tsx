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

import "./index.css";

import axios from "axios";
import dayjs from "dayjs";
import { Dialog, Menu, Transition } from "@headlessui/react";
import { EllipsisVerticalIcon } from "@heroicons/react/20/solid";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { useDebounce } from "react-use";

import Header from "../../components/Header";
import IMenu from "../../components/Menu";
import Notification from "../../components/Notification";
import OrderHeader from "../../components/OrderHeader";
import Pagination from "../../components/Pagination";
import Quota from "../../components/Quota";
import QuotaSimple from "../../components/QuotaSimple";
import Settings from "../../Settings";
import Toast from "../../components/Notification";
import calcUnit from "../../utils/calcUnit";
import { IHTTPError, INamespaceItem, IOrder, IRepositoryItem, IRepositoryList, IUserSelf } from "../../interfaces";
import { NamespaceRole, UserRole } from "../../interfaces/enums";

export default function ({ localServer }: { localServer: string }) {
  const [repositoryList, setRepositoryList] = useState<IRepositoryList>({} as IRepositoryList);
  const [refresh, setRefresh] = useState({});
  const [page, setPage] = useState(1);
  const [searchRepository, setSearchRepository] = useState("");
  const [total, setTotal] = useState(0);

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId = searchParams.get('namespace_id');
  const [namespaceObj, setNamespaceObj] = useState<INamespaceItem>({} as INamespaceItem);

  useEffect(() => {
    if (namespaceId == null || namespaceId == "") {
      return;
    }
    axios.get(`${localServer}/api/v1/namespaces/${namespaceId}`).then(response => {
      if (response.status == 200) {
        const namespaceObj = response.data as INamespaceItem;
        setNamespaceObj(namespaceObj);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }, []);

  const [repositoryText, setRepositoryText] = useState("");
  const [repositoryTextValid, setRepositoryTextValid] = useState(true);
  useEffect(() => { repositoryText != "" && setRepositoryTextValid(/^[a-z][0-9a-z-]{0,20}$/.test(repositoryText)) }, [repositoryText])
  const [descriptionText, setDescriptionText] = useState("");
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => { descriptionText != "" && setDescriptionTextValid(/^.{0,30}$/.test(descriptionText)) }, [descriptionText]);
  const [tagCountLimit, setTagCountLimit] = useState<string | number>(0);
  const [tagCountLimitValid, setTagCountLimitValid] = useState(true);
  useEffect(() => { setTagCountLimitValid(Number.isInteger(tagCountLimit) && parseInt(tagCountLimit.toString()) >= 0) }, [tagCountLimit])
  const [realSizeLimit, setRealSizeLimit] = useState(0);
  const [sizeLimit, setSizeLimit] = useState<string | number>(0);
  const [sizeLimitValid, setSizeLimitValid] = useState(true);
  const [sizeLimitUnit, setSizeLimitUnit] = useState("MiB");
  useEffect(() => { setSizeLimitValid(Number.isInteger(sizeLimit) && parseInt(sizeLimit.toString()) >= 0) }, [sizeLimit])
  useEffect(() => {
    let sl = 0;
    if (Number.isInteger(sizeLimit)) {
      sl = parseInt(sizeLimit.toString());
    }
    switch (sizeLimitUnit) {
      case "MiB":
        setRealSizeLimit(sl * 1 << 20);
        break;
      case "GiB":
        setRealSizeLimit(sl * 1 << 30);
        break;
      case "TiB":
        setRealSizeLimit(sl * 1 << 40);
        break;
    }
  }, [sizeLimit, sizeLimitUnit]);
  const [createRepositoryModal, setCreateRepositoryModal] = useState(false);

  const createRepository = () => {
    if (!(repositoryTextValid && descriptionTextValid && sizeLimitValid && tagCountLimitValid)) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    setCreateRepositoryModal(false);
    axios.post(localServer + `/api/v1/namespaces/${namespaceId}/repositories/`, {
      name: namespace + "/" + repositoryText,
      description: descriptionText,
      size_limit: realSizeLimit,
      tag_limit: tagCountLimit,
    } as IRepositoryItem, {}).then(response => {
      if (response.status === 201) {
        setRepositoryText("");
        setDescriptionText("");
        setTagCountLimit(0);
        setSizeLimit(0);
        setRefresh({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }

  const validateRepository = () => {
    if (repositoryText === "") {
      return;
    }
    let url = localServer + `/api/v1/validators/reference?reference=${namespace}/${repositoryText}`
    axios.get(url).then(response => {
      if (response?.status === 204) {
        setRepositoryTextValid(true);
      } else {
        setRepositoryTextValid(false);
      }
    }).catch(error => {
      setRepositoryTextValid(false);
    });
  }
  useDebounce(validateRepository, 300, [repositoryText]);

  const [sizeOrder, setSizeOrder] = useState(IOrder.None);
  const [tagCountOrder, setTagCountOrder] = useState(IOrder.None);
  const [createdAtOrder, setCreatedAtOrder] = useState(IOrder.None);
  const [updatedAtOrder, setUpdatedAtOrder] = useState(IOrder.None);
  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");

  const resetOrder = () => {
    setSizeOrder(IOrder.None);
    setTagCountOrder(IOrder.None);
    setCreatedAtOrder(IOrder.None);
    setUpdatedAtOrder(IOrder.None);
  }

  const fetchRepository = () => {
    let url = localServer + `/api/v1/namespaces/${namespaceId}/repositories/?limit=${Settings.PageSize}&page=${page}`;
    if (searchRepository !== "") {
      url += `&name=${searchRepository}`;
    }
    if (sortName !== "") {
      url += `&sort=${sortName}&method=${sortOrder.toString()}`
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const repositoryList = response.data as IRepositoryList;
        setRepositoryList(repositoryList);
        setTotal(repositoryList.total);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  useEffect(fetchRepository, [refresh, page]);

  const [userObj, setUserObj] = useState<IUserSelf>({} as IUserSelf);

  useEffect(() => {
    axios.get(localServer + "/api/v1/users/self").then(response => {
      if (response.status === 200) {
        const user = response.data as IUserSelf;
        setUserObj(user);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, []);

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Repositories</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item="repositories" namespace={namespace} namespace_id={namespaceId || ""} />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title="Repository"
              props={
                (
                  <div className="flex space-x-8">
                    <Link
                      to={`/namespaces/${namespace}/namespace-summary?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Summary
                    </Link>
                    <Link
                      to="#"
                      className="z-10 inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize"
                    >
                      Repository list
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/members?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Members
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/daemon-tasks?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Daemon task
                    </Link>
                    <Link
                      to={`/namespaces/${namespace}/webhooks?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Webhook
                    </Link>
                  </div>
                )
              }
            />
            <div className="pt-1 pb-1 flex justify-between items-center min-h-[60px]">
              <div className="pr-2 pl-2">
                <div className="flex gap-4">
                  <div className="relative mt-2 flex items-center">
                    <label
                      htmlFor="repositorySearch"
                      className="absolute -top-2 left-2 inline-block bg-white px-1 text-xs font-medium text-gray-900"
                    >
                      Repository
                    </label>
                    <input
                      type="text"
                      id="repositorySearch"
                      placeholder="search repository"
                      value={searchRepository}
                      onChange={e => { setSearchRepository(e.target.value); }}
                      onKeyDown={e => {
                        if (e.key == "Enter") {
                          fetchRepository()
                        }
                      }}
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
              <div className="pr-2 pl-2 flex flex-col">
                <button className={
                  (((userObj.role == UserRole.Admin || userObj.role == UserRole.Root || (namespaceObj.role != undefined && (namespaceObj.role == NamespaceRole.Admin || namespaceObj.role == NamespaceRole.Manager)))) ? ' cursor-pointer focus:ring-2 focus:ring-offset-2 ' : ' cursor-not-allowed ') +
                  "my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-purple-500 sm:order-1 sm:ml-3"}
                  onClick={() => {
                    ((userObj.role == UserRole.Admin || userObj.role == UserRole.Root || (namespaceObj.role != undefined && (namespaceObj.role == NamespaceRole.Admin || namespaceObj.role == NamespaceRole.Manager)))) && setCreateRepositoryModal(true);
                  }}
                >Create</button>
              </div>
            </div>
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className="align-middle inline-block min-w-full border-b border-gray-200">
              <table className="min-w-full flex-1">
                <thead>
                  <tr className="border-gray-200">
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Repository</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <OrderHeader text={"Size"} orderStatus={sizeOrder} setOrder={e => {
                        resetOrder();
                        setSizeOrder(e);
                        setSortOrder(e);
                        setSortName("size");
                      }} />
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <OrderHeader text={"Tag count"} orderStatus={tagCountOrder} setOrder={e => {
                        resetOrder();
                        setTagCountOrder(e);
                        setSortOrder(e);
                        setSortName("tag_count");
                      }} />
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <OrderHeader text={"Created at"} orderStatus={createdAtOrder} setOrder={e => {
                        resetOrder();
                        setCreatedAtOrder(e);
                        setSortOrder(e);
                        setSortName("created_at");
                      }} />
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <OrderHeader text={"Updated at"} orderStatus={updatedAtOrder} setOrder={e => {
                        resetOrder();
                        setUpdatedAtOrder(e);
                        setSortOrder(e);
                        setSortName("updated_at");
                      }} />
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
                        <TableItem key={index} localServer={localServer} index={index} user={userObj} namespace={namespaceObj} repository={repository} setRefresh={setRefresh} />
                      );
                    })
                  }
                </tbody>
              </table>
            </div>
          </div>
          <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
        </div>
      </div>
      <Transition.Root show={createRepositoryModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setCreateRepositoryModal}>
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
                  <label htmlFor="first-name" className="block text-sm font-medium leading-6 text-gray-900">
                    <span className="text-red-600">*</span>Name
                  </label>
                  <div className="relative mt-2 flex rounded-md shadow-sm">
                    <span className="inline-flex items-center rounded-l-md border border-r-0 border-gray-300 px-2 text-gray-500 sm:text-sm">
                      {namespace}/
                    </span>
                    <input
                      type="text"
                      name="namespace"
                      placeholder="2-20 lowercase characters"
                      className={(repositoryTextValid ? "block w-full rounded-r-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-r-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                      value={repositoryText}
                      onChange={e => {
                        setRepositoryText(e.target.value);
                      }}
                    />
                    {
                      repositoryTextValid ? (
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
                      repositoryTextValid ? (
                        <span></span>
                      ) : (
                        <span>
                          Not a valid repository name, please check again.
                        </span>
                      )
                    }
                  </p>
                  <label htmlFor="description" className="block text-sm font-medium text-gray-700 mt-1">
                    Description
                  </label>
                  <div className="relative mt-2 rounded-md shadow-sm">
                    <input
                      type="text"
                      name="description"
                      id="description"
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
                  <p className="mt-2 text-xs text-red-600">
                    {
                      descriptionTextValid ? (
                        <span></span>
                      ) : (
                        <span>
                          Not a valid description, max 30 characters.
                        </span>
                      )
                    }
                  </p>
                  <label htmlFor="size_limit" className="block text-sm font-medium text-gray-700 mt-2">
                    Size limit
                  </label>
                  <div className="relative mt-2 rounded-md shadow-sm">
                    <input
                      type="number"
                      id="size_limit"
                      name="size_limit"
                      placeholder="0 means no limit"
                      className={(sizeLimitValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                      value={sizeLimit}
                      onChange={e => setSizeLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                    />
                    <div className="absolute inset-y-0 right-0 flex items-center">
                      <label htmlFor="size_limit_unit" className="sr-only">
                        Size limit unit
                      </label>
                      <select
                        id="size_limit_unit"
                        name="size_limit_unit"
                        className="h-full rounded-md border-0 bg-transparent py-0 pl-2 pr-7 text-gray-500 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm"
                        value={sizeLimitUnit}
                        onChange={e => { setSizeLimitUnit(e.target.value) }}
                      >
                        <option value="MiB">MiB</option>
                        <option value="GiB">GiB</option>
                        <option value="TiB">TiB</option>
                      </select>
                    </div>
                  </div>
                  <p className="mt-1 text-xs text-red-600">
                    {
                      sizeLimitValid ? (
                        <span></span>
                      ) : (
                        <span>
                          Not a valid size limit, should be non-negative integer.
                        </span>
                      )
                    }
                  </p>
                  <label htmlFor="tag_count_limit" className="block text-sm font-medium text-gray-700">
                    Tag count limit
                  </label>
                  <div className="relative mt-2 rounded-md shadow-sm">
                    <input
                      type="number"
                      id="tag_count_limit"
                      name="tag_count_limit"
                      placeholder="0 means no limit"
                      className={(tagCountLimitValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                      value={tagCountLimit}
                      onChange={e => setTagCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                    />
                    {
                      tagCountLimitValid ? (
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
                      tagCountLimitValid ? (
                        <span></span>
                      ) : (
                        <span>
                          Not a valid tag count limit, should be non-negative integer.
                        </span>
                      )
                    }
                  </p>
                  <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                    <button
                      type="button"
                      className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                      onClick={() => createRepository()}
                    >
                      Create
                    </button>
                    <button
                      type="button"
                      className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                      onClick={() => setCreateRepositoryModal(false)}
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
    </Fragment >
  )
}

function TableItem({ localServer, index, user, namespace, repository, setRefresh }: { localServer: string, index: number, user: IUserSelf, namespace: INamespaceItem, repository: IRepositoryItem, setRefresh: (param: any) => void }) {
  const navigate = useNavigate();

  const [deleteRepositoryModal, setDeleteRepositoryModal] = useState(false);

  const [repositoryText, setRepositoryText] = useState(`${repository.name}`);
  const [repositoryTextValid, setRepositoryTextValid] = useState(true);
  const [descriptionText, setDescriptionText] = useState(repository.description);
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => { descriptionText != "" && setDescriptionTextValid(/^.{0,30}$/.test(descriptionText)) }, [descriptionText]);
  const [tagCountLimit, setTagCountLimit] = useState<string | number>(repository.tag_limit);
  const [tagCountLimitValid, setTagCountLimitValid] = useState(true);
  useEffect(() => { setTagCountLimitValid(Number.isInteger(tagCountLimit) && parseInt(tagCountLimit.toString()) >= 0) }, [tagCountLimit])
  const [realSizeLimit, setRealSizeLimit] = useState(repository.size_limit);
  let calcUnitObj = calcUnit(repository.size_limit);
  const [sizeLimit, setSizeLimit] = useState<string | number>(calcUnitObj.size);
  const [sizeLimitValid, setSizeLimitValid] = useState(true);
  const [sizeLimitUnit, setSizeLimitUnit] = useState(calcUnitObj.unit);

  useEffect(() => { setSizeLimitValid(Number.isInteger(sizeLimit) && parseInt(sizeLimit.toString()) >= 0) }, [sizeLimit]);
  useEffect(() => {
    let sl = 0;
    if (Number.isInteger(sizeLimit)) {
      sl = parseInt(sizeLimit.toString());
    }
    switch (sizeLimitUnit) {
      case "MiB":
        setRealSizeLimit(sl * 1 << 20);
        break;
      case "GiB":
        setRealSizeLimit(sl * 1 << 30);
        break;
      case "TiB":
        setRealSizeLimit(sl * 1 << 40);
        break;
    }
  }, [sizeLimit, sizeLimitUnit]);
  const [updateRepositoryModal, setUpdateRepositoryModal] = useState(false);

  const updateRepository = () => {
    if (!(descriptionTextValid && sizeLimitValid && tagCountLimitValid)) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    setUpdateRepositoryModal(false);
    axios.put(localServer + `/api/v1/namespaces/${namespace}/repositories/${repository.id}`, {
      description: descriptionText,
      size_limit: realSizeLimit,
      tag_limit: tagCountLimit,
    } as IRepositoryItem, {}).then(response => {
      if (response.status === 204) {
        setRefresh({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }

  const deleteRepository = () => {
    axios.delete(localServer + `/api/v1/namespaces/${namespace}/repositories/${repository.id}`).then(response => {
      if (response.status === 204) {
        setRefresh({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }

  return (
    <tr>
      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer"
        onClick={() => {
          navigate(`/namespaces/${namespace.name}/repository/tags?namespace_id=${repository.namespace_id}&repository=${repository.name}&repository_id=${repository.id}`);
        }}
      >
        <div className="flex items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            {repository.name}
            <span className="text-gray-500 font-normal ml-4">{repository.description}</span>
          </div>
        </div>
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        <Quota current={repository.size} limit={repository.size_limit} />
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        <QuotaSimple current={repository.tag_count} limit={repository.tag_limit} />
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs.utc(repository.created_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs.utc(repository.updated_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}
      </td>
      <td className="pr-3 whitespace-nowrap">
        <Menu as="div" className="relative flex-none" onClick={e => {
          e.stopPropagation();
        }}>
          <Menu.Button className="mx-auto -m-2.5 block p-2.5 text-gray-500 hover:text-gray-900 margin">
            <span className="sr-only">Open options</span>
            <EllipsisVerticalIcon className="h-5 w-5" aria-hidden="true" />
          </Menu.Button>
          <Transition
            as={Fragment}
            enter="transition ease-out duration-100"
            enterFrom="transform opacity-0 scale-95"
            enterTo="transform opacity-100 scale-100"
            leave="transition ease-in duration-75"
            leaveFrom="transform opacity-100 scale-100"
            leaveTo="transform opacity-0 scale-95"
          >
            <Menu.Items className={(index > 10 ? "menu-action-top" : "mt-2") + " text-left absolute right-0 z-10 w-20 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"} >
              <Menu.Item>
                {({ active }) => (
                  <div
                    className={
                      (active ? 'bg-gray-100' : '') +
                      (((user.role == UserRole.Admin || user.role == UserRole.Root || (namespace.role != undefined && (namespace.role == NamespaceRole.Admin || namespace.role == NamespaceRole.Manager)))) ? ' cursor-pointer ' : ' cursor-not-allowed ') +
                      ' block px-3 py-1 text-sm leading-6 text-gray-900 cursor-pointer'
                    }
                    onClick={e => {
                      ((user.role == UserRole.Admin || user.role == UserRole.Root || (namespace.role != undefined && (namespace.role == NamespaceRole.Admin || namespace.role == NamespaceRole.Manager)))) && setUpdateRepositoryModal(true);
                    }}
                  >
                    Update
                  </div>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <div
                    className={
                      (active ? 'bg-gray-50' : '') +
                      (((user.role == UserRole.Admin || user.role == UserRole.Root || (namespace.role != undefined && (namespace.role == NamespaceRole.Admin || namespace.role == NamespaceRole.Manager)))) ? ' cursor-pointer' : ' cursor-not-allowed') +
                      ' block px-3 py-1 text-sm leading-6 text-gray-900 hover:text-white hover:bg-red-600 cursor-pointer'
                    }
                    onClick={e => {
                      ((user.role == UserRole.Admin || user.role == UserRole.Root || (namespace.role != undefined && (namespace.role == NamespaceRole.Admin || namespace.role == NamespaceRole.Manager)))) && setDeleteRepositoryModal(true);
                    }}
                  >
                    Delete
                  </div>
                )}
              </Menu.Item>
            </Menu.Items>
          </Transition>
        </Menu>
      </td>
      <td>
        <Transition.Root show={updateRepositoryModal} as={Fragment}>
          <Dialog as="div" className="relative z-10" onClose={setUpdateRepositoryModal}>
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
                    <label htmlFor="first-name" className="block text-sm font-medium leading-6 text-gray-900">
                      Name
                    </label>
                    <div className="relative mt-2 flex rounded-md shadow-sm">
                      <input
                        type="text"
                        name="namespace"
                        placeholder="2-20 lowercase characters"
                        className={(repositoryTextValid ? "disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-500 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-r-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                        value={repositoryText}
                        disabled
                        onChange={e => {
                          setRepositoryText(e.target.value);
                        }}
                      />
                      {
                        repositoryTextValid ? (
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
                    <p className="mt-1 text-xs text-red-600" id="email-error">
                      {
                        repositoryTextValid ? (
                          <span></span>
                        ) : (
                          <span>
                            Not a valid repository name, please check again.
                          </span>
                        )
                      }
                    </p>
                    <label htmlFor="description" className="block text-sm font-medium text-gray-700 mt-1">
                      Description
                    </label>
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <input
                        type="text"
                        name="description"
                        id="description"
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
                    <p className="mt-1 text-xs text-red-600" id="email-error">
                      {
                        descriptionTextValid ? (
                          <span></span>
                        ) : (
                          <span>
                            Not a valid description, max 30 characters.
                          </span>
                        )
                      }
                    </p>
                    <label htmlFor="size_limit" className="block text-sm font-medium text-gray-700 mt-2">
                      Size limit
                    </label>
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <input
                        type="number"
                        id="size_limit"
                        name="size_limit"
                        placeholder="0 means no limit"
                        className={(sizeLimitValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                        value={sizeLimit}
                        onChange={e => setSizeLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                      />
                      <div className="absolute inset-y-0 right-0 flex items-center">
                        <label htmlFor="size_limit_unit" className="sr-only">
                          Size limit unit
                        </label>
                        <select
                          id="size_limit_unit"
                          name="size_limit_unit"
                          className="h-full rounded-md border-0 bg-transparent py-0 pl-2 pr-7 text-gray-500 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm"
                          value={sizeLimitUnit}
                          onChange={e => { setSizeLimitUnit(e.target.value) }}
                        >
                          <option value="MiB">MiB</option>
                          <option value="GiB">GiB</option>
                          <option value="TiB">TiB</option>
                        </select>
                      </div>
                    </div>
                    <p className="mt-1 text-xs text-red-600" id="email-error">
                      {
                        sizeLimitValid ? (
                          <span></span>
                        ) : (
                          <span>
                            Not a valid size limit, should be non-negative integer.
                          </span>
                        )
                      }
                    </p>
                    <label htmlFor="tag_count_limit" className="block text-sm font-medium text-gray-700">
                      Tag count limit
                    </label>
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <input
                        type="number"
                        id="tag_count_limit"
                        name="tag_count_limit"
                        placeholder="0 means no limit"
                        className={(tagCountLimitValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                        value={tagCountLimit}
                        onChange={e => setTagCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                      />
                      {
                        tagCountLimitValid ? (
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
                    <p className="mt-1 text-xs text-red-600" id="email-error">
                      {
                        tagCountLimitValid ? (
                          <span></span>
                        ) : (
                          <span>
                            Not a valid tag count limit, should be non-negative integer.
                          </span>
                        )
                      }
                    </p>
                    <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={() => updateRepository()}
                      >
                        Update
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={() => setUpdateRepositoryModal(false)}
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
      </td>
      <td>
        <Transition.Root show={deleteRepositoryModal} as={Fragment}>
          <Dialog as="div" className="relative z-10" onClose={setDeleteRepositoryModal}>
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
                  <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
                    <div className="sm:flex sm:items-start">
                      <div className="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10">
                        <ExclamationTriangleIcon className="h-6 w-6 text-red-600" aria-hidden="true" />
                      </div>
                      <div className="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                        <Dialog.Title as="h3" className="text-base font-semibold leading-6 text-gray-900">
                          Delete repository
                        </Dialog.Title>
                        <div className="mt-2">
                          <p className="text-sm text-gray-500">
                            Are you sure you want to delete the repository <span className="text-black font-medium">{repository.name}</span>
                          </p>
                        </div>
                      </div>
                    </div>
                    <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:ml-3 sm:w-auto"
                        onClick={e => { setDeleteRepositoryModal(false); deleteRepository(); }}
                      >
                        Delete
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                        onClick={() => setDeleteRepositoryModal(false)}
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
      </td>
    </tr>
  );
}
