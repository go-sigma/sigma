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
import { Dialog, Transition } from "@headlessui/react";
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import { Tooltip } from 'flowbite';

import Header from "../../components/Header";
import IMenu from "../../components/Menu";
import Notification from "../../components/Notification";
import Pagination from "../../components/Pagination";
import Settings from "../../Settings";
import { IHTTPError, INamespaceItem, IOrder, IWebhookList } from "../../interfaces";

export default function Repository({ localServer }: { localServer: string }) {
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

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const [enable, setEnable] = useState(true);
  const [eventNamespace, setEventNamespace] = useState(true);
  const [eventRepository, setEventRepository] = useState(true);
  const [eventTag, setEventTag] = useState(true);
  const [eventMember, setEventMember] = useState(true);
  const [eventArtifact, setEventArtifact] = useState(false);

  const [retryTimes, setRetryTimes] = useState<string | number>(1);
  const [retryTimesValid, setRetryTimesValid] = useState(true);
  useEffect(() => { setRetryTimesValid(Number.isInteger(retryTimes) && parseInt(retryTimes.toString()) >= 1 && parseInt(retryTimes.toString()) <= 5) }, [retryTimes]);
  const [retryDuration, setRetryDuration] = useState<string | number>(3);
  const [retryDurationValid, setRetryDurationValid] = useState(true);
  useEffect(() => { setRetryDurationValid(Number.isInteger(retryDuration) && parseInt(retryDuration.toString()) >= 0 && parseInt(retryDuration.toString()) <= 10) }, [retryDuration]);

  const [showSslVerify, setShowSslVerify] = useState(false);

  const [sslVerify, setSslVerify] = useState(true);
  const [secret, setSecret] = useState<string | undefined>();
  const [secretValid, setSecretValid] = useState(true);
  useEffect(() => { if (secret != undefined && secret.length >= 0 && secret.length <= 63) { setSecretValid(true); } }, [secret]);
  const [url, setUrl] = useState<string>("");
  const [urlValid, setUrlValid] = useState(true);
  useEffect(() => { url != "" && setUrlValid((url.startsWith("http://") || url.startsWith("https://")) && /(https:\/\/www\.|http:\/\/www\.|https:\/\/|http:\/\/)?[a-zA-Z]{2,}(\.[a-zA-Z]{2,})(\.[a-zA-Z]{2,})?\/[a-zA-Z0-9]{2,}|((https:\/\/www\.|http:\/\/www\.|https:\/\/|http:\/\/)?[a-zA-Z]{2,}(\.[a-zA-Z]{2,})(\.[a-zA-Z]{2,})?)|(https:\/\/www\.|http:\/\/www\.|https:\/\/|http:\/\/)?[a-zA-Z0-9]{2,}\.[a-zA-Z0-9]{2,}\.[a-zA-Z0-9]{2,}(\.[a-zA-Z0-9]{2,})?/.test(url) && url.length <= 128) }, [url]);

  const [createWebhookModal, setCreateWebhookModal] = useState(false);

  useEffect(() => {
    if (url.startsWith("https://")) {
      setShowSslVerify(true);
    } else {
      setShowSslVerify(false);
    }
  }, [url]);

  const [refresh, setRefresh] = useState({});
  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");
  const [webhookList, setWebhookList] = useState<IWebhookList>({} as IWebhookList);

  const fetchWebhook = () => {
    let url = localServer + `/api/v1/webhooks/?limit=${Settings.PageSize}&page=${page}`;
    if (sortName !== "") {
      url += `&sort=${sortName}&method=${sortOrder.toString()}`
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const webhookList = response.data as IWebhookList;
        setWebhookList(webhookList);
        setTotal(webhookList.total);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  useEffect(() => { fetchWebhook() }, [refresh, page, sortOrder, sortName]);

  const createWebhook = () => {
    if (url === "") {
      setUrlValid(false);
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    if (!(retryTimesValid && retryDurationValid && secretValid && urlValid)) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    const data: { [key: string]: any } = {
      enable: enable,
      url: url,
      retry_times: retryTimes,
      retry_duration: retryDuration,
      event_repository: eventRepository,
      event_tag: eventTag,
      event_artifact: eventArtifact,
      event_member: eventMember,
    };
    if (secret != undefined && secret.length != 0) {
      data["secret"] = secret;
    }
    if (showSslVerify) {
      data["ssl_verify"] = sslVerify;
    }

    let u = `${localServer}/api/v1/webhooks/`;
    if (namespaceId != null) {
      data["event_namespace"] = eventNamespace
      u += `?namespace_id=${namespaceId}`
    }
    axios.post(u, data, {}).then(response => {
      if (response.status === 201) {
        setRefresh({});
        setUrl("");
        setUrlValid(true);
        setSecret("");
        setSecretValid(true);
        setRetryTimes(1);
        setRetryDuration(5);
        setEnable(true);
        setEventNamespace(true);
        setEventRepository(true);
        setEventTag(true);
        setEventArtifact(false);
        setEventMember(true);
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
          <title>sigma - Namespace Webhook</title>
        </Helmet>
      </HelmetProvider>
      <div
        id="tooltip-webhook-retry-duration"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700 w-[220px]">
        Less than 10, unit is second.
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item="Repository" />
        <div className="flex flex-col flex-1 max-h-screen">
          <main className="relative z-0 focus:outline-none" tabIndex={0}>
            <Header title="Webhook" props={
              (
                <div className="flex space-x-8">
                  <Link
                    to={`/namespaces/${namespace}/namespace-summary?namespace_id=${namespaceId}`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                  >
                    Summary
                  </Link>
                  <Link
                    to={`/namespaces/${namespace}/repositories?namespace_id=${namespaceId}`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
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
                    to="#"
                    className="inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize"
                  >
                    Webhook
                  </Link>
                </div>
              )
            } />
            <div className="pt-1 pb-1 flex justify-between items-center min-h-[60px]">
              <div className="pr-2 pl-2">
                <div className="flex gap-4">
                  <div className="relative mt-2 flex items-center">
                  </div>
                </div>
              </div>
              <div className="pr-2 pl-2 flex flex-col">
                <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                  onClick={() => { setCreateWebhookModal(true) }}
                >Create</button>
              </div>
            </div>
          </main>
          <div className="flex-1 flex overflow-y-auto">
            <div className="align-middle inline-block min-w-full border-b border-gray-200">
              <table className="min-w-full flex-1">
                <thead>
                  <tr>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Namespace</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      {/* <OrderHeader text={"Size"} orderStatus={sizeOrder} setOrder={(e) => {
                        resetOrder();
                        setSizeOrder(e);
                        setSortOrder(e);
                        setSortName("size");
                      }} /> */}
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      {/* <OrderHeader text={"Repository count"} orderStatus={repositoryCountOrder} setOrder={(e) => {
                        resetOrder();
                        setRepositoryOrder(e);
                        setSortOrder(e);
                        setSortName("repository_count");
                      }} /> */}
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      {/* <OrderHeader text={"Tag count"} orderStatus={tagCountOrder} setOrder={(e) => {
                        resetOrder();
                        setTagCountOrder(e);
                        setSortOrder(e);
                        setSortName("tag_count");
                      }} /> */}
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      Visibility
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      {/* <OrderHeader text={"Created at"} orderStatus={createdAtOrder} setOrder={(e) => {
                        resetOrder();
                        setCreatedAtOrder(e);
                        setSortOrder(e);
                        setSortName("created_at");
                      }} /> */}
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      {/* <OrderHeader text={"Updated at"} orderStatus={updatedAtOrder} setOrder={(e) => {
                        resetOrder();
                        setUpdatedAtOrder(e);
                        setSortOrder(e);
                        setSortName("updated_at");
                      }} /> */}
                    </th>
                    <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      Action
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-100 max-h-max">
                  {/* {
                    namespaceList.items?.map((namespace, index) => {
                      return (
                        <TableItem key={namespace.id} index={index} user={userObj} namespace={namespace} localServer={localServer} setRefresh={setRefresh} />
                      );
                    })
                  } */}
                </tbody>
              </table>
            </div>
          </div>
          <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
        </div>
      </div>
      <Transition.Root show={createWebhookModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setCreateWebhookModal}>
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
                <Dialog.Panel className="relative transform rounded-lg bg-white px-6 pb-4 text-left shadow-xl transition-all">
                  <Dialog.Title
                    as="h3"
                    className="text-lg font-medium leading-6 text-gray-900 border-b pt-4 pb-4"
                  >
                    Create webhook
                  </Dialog.Title>
                  <div className="flex flex-col gap-0 mt-4">
                    <div className="grid grid-cols-12 gap-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">URL</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-10">
                        <input
                          type="text"
                          name="description"
                          placeholder="128 characters"
                          className={(urlValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                          value={url}
                          onChange={e => setUrl(e.target.value)}
                        />
                      </div>
                    </div>
                    <div className="grid grid-cols-12 gap-4 mt-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="leading-6 ">Secret</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-10">
                        <input
                          type="text"
                          name="description"
                          placeholder="max 63 characters"
                          className={(secretValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                          value={secret}
                          onChange={e => setSecret(e.target.value)}
                        />
                      </div>
                    </div>
                    {
                      showSslVerify ? (
                        <div className="grid grid-cols-12 gap-4 mt-4">
                          <div className="col-span-2 flex flex-row">
                            <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                              <div className="flex">
                                <span className="leading-6 ">SSL Verify</span>
                                <span>:</span>
                              </div>
                            </label>
                          </div>
                          <div className="col-span-10 flex flex-row">
                            <label className="inline-flex items-center cursor-pointer">
                              <input type="checkbox"
                                checked={sslVerify}
                                onChange={e => setSslVerify(!sslVerify)}
                                className="sr-only peer" />
                              <div className="relative w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                            </label>
                          </div>
                        </div>
                      ) : null
                    }
                    <div className="grid grid-cols-12 gap-4 mt-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="leading-6 ">Retry Times</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4 flex flex-row">
                        <input
                          type="text"
                          name="description"
                          placeholder="1 <= times <= 5"
                          className={(retryTimesValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                          value={retryTimes}
                          onChange={e => setRetryTimes(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                        />
                      </div>
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="leading-6 ">Retry Duration</span>
                            <div className="flex flex-row cursor-pointer"
                              id="gcRepositoryRetentionDaysHelp"
                              onClick={e => {
                                let tooltip = new Tooltip(document.getElementById("tooltip-gc-repository-retention-days"),
                                  document.getElementById("gcRepositoryRetentionDaysHelp"), { triggerType: "click" });
                                tooltip.show();
                              }}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 block my-auto ml-0.5">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                              </svg>
                            </div>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4 flex flex-row">
                        <input
                          type="text"
                          name="description"
                          placeholder="less than 10"
                          className={(retryDurationValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                          value={retryDuration}
                          onChange={e => setRetryDuration(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                        />
                      </div>
                    </div>
                    <div className="grid grid-cols-12 gap-4 mt-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="leading-6 ">Enable</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-10 flex flex-row">
                        <label className="inline-flex items-center cursor-pointer">
                          <input type="checkbox"
                            checked={enable}
                            onChange={e => setEnable(!enable)}
                            className="sr-only peer" />
                          <div className="relative w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                        </label>
                      </div>
                    </div>
                    <div className="mt-4 flex flex-row gap-4">
                      <div className="flex items-center">
                        <input id="event-namespace" type="checkbox"
                          checked={eventNamespace}
                          onChange={e => setEventNamespace(!eventNamespace)}
                          className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                        <label htmlFor="event-namespace" className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300">Namespace Event</label>
                      </div>
                      <div className="flex items-center">
                        <input id="event-repository" type="checkbox"
                          checked={eventRepository}
                          onChange={e => setEventRepository(!eventRepository)}
                          className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                        <label htmlFor="event-repository" className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300">Repository Event</label>
                      </div>
                      <div className="flex items-center">
                        <input id="event-tag" type="checkbox"
                          checked={eventTag}
                          onChange={e => setEventTag(!eventTag)}
                          className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                        <label htmlFor="event-tag" className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300">Tag Event</label>
                      </div>
                      <div className="flex items-center">
                        <input id="event-artifact" type="checkbox"
                          checked={eventArtifact}
                          onChange={e => setEventArtifact(!eventArtifact)}
                          className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                        <label htmlFor="event-artifact" className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300">Artifact Event</label>
                      </div>
                      <div className="flex items-center">
                        <input id="event-member"
                          checked={eventMember}
                          onChange={e => setEventMember(!eventMember)}
                          type="checkbox" className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                        <label htmlFor="event-member" className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300">Member Event</label>
                      </div>
                    </div>
                    <div className="flex flex-row-reverse mt-4 pt-4 border-t">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={e => createWebhook()}
                      >
                        Create
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={e => { setCreateWebhookModal(false) }}
                      >
                        Cancel
                      </button>
                    </div>
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

function TableItem() {
  return (
    <tr className="align-middle">
      <td className="px-6 py-4 w-5/6 whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer"
        onClick={() => {
          // navigate(`/namespaces/${namespace.name}/repositories?namespace_id=${namespace.id}`);
        }}
      >
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {/* {namespace.name}
              <span className="text-gray-500 font-normal ml-4">{namespace.description}</span> */}
            </span>
          </div>
        </div>
      </td>
    </tr>
  );
}
