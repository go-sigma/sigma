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
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle } from "@/components/ui/dialog";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Link, useParams, useSearchParams, useLocation, useNavigate } from 'react-router-dom';
import { Tooltip } from '../../utils';
import dayjs from "dayjs";

import Header from "../../components/Header";
import IMenu from "../../components/Menu";
import Notification from "../../components/Notification";
import Pagination from "../../components/Pagination";
import Settings from "../../Settings";
import { IHTTPError, INamespaceItem, IOrder, IUserSelf, IWebhookItem, IWebhookList } from "../../interfaces";
import OrderHeader from "../../components/OrderHeader";
import { EllipsisVerticalIcon, ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { NamespaceRole, UserRole } from "../../interfaces/enums";
import { useTranslation } from "../../i18n/useTranslation";

export default function ({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
  const location = useLocation();
  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId = searchParams.get('namespace_id');
  const [namespaceObj, setNamespaceObj] = useState<INamespaceItem>({} as INamespaceItem);

  useEffect(() => {
    if (location.pathname.startsWith("/settings")) {
      return;
    }
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

  const [userObj, setUserObj] = useState<IUserSelf>({} as IUserSelf);

  useEffect(() => {
    axios.get(localServer + "/api/v1/users/self").then(response => {
      if (response.status === 200) {
        const user = response.data as IUserSelf;
        setUserObj(user);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, []);

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const [enable, setEnable] = useState(true);
  const [eventNamespace, setEventNamespace] = useState(true);
  const [eventRepository, setEventRepository] = useState(true);
  const [eventTag, setEventTag] = useState(true);
  const [eventMember, setEventMember] = useState(true);
  const [eventArtifact, setEventArtifact] = useState(true);
  const [eventDaemonTaskGc, setEventDaemonTaskGc] = useState(true);

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
  const [createdAtOrder, setCreatedAtOrder] = useState(IOrder.None);
  const [updatedAtOrder, setUpdatedAtOrder] = useState(IOrder.None);
  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");
  const [webhookList, setWebhookList] = useState<IWebhookList>({} as IWebhookList);

  const resetOrder = () => {
    setCreatedAtOrder(IOrder.None);
    setUpdatedAtOrder(IOrder.None);
  }

  const fetchWebhook = () => {
    let url = localServer + `/api/v1/webhooks/?limit=${Settings.PageSize}&page=${page}`;
    if (sortName !== "") {
      url += `&sort=${sortName}&method=${sortOrder.toString()}`;
    }
    if (namespaceId != null && namespaceId != "0") {
      url += `&namespace_id=${namespaceId}`;
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
      event_daemon_task_gc: eventDaemonTaskGc,
    };
    if (secret != undefined && secret.length != 0) {
      data["secret"] = secret;
    }
    if (showSslVerify) {
      data["ssl_verify"] = sslVerify;
    }

    let u = `${localServer}/api/v1/webhooks/`;
    if (location.pathname.startsWith("/settings")) {
      data["event_namespace"] = eventNamespace;
      data["namespace_id"] = parseInt(namespaceId || "");
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
        setCreateWebhookModal(false);
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
          <title>{t("header.webhook")}</title>
        </Helmet>
      </HelmetProvider>
      <div
        id="tooltip-webhook-retry-duration"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700 w-55">
        Less than 10, unit is second.
      </div>
      <div className="min-h-screen flex overflow-hidden bg-white dark:bg-gray-950">
        <IMenu localServer={localServer} item={location.pathname.startsWith("/settings") ? "webhooks" : "repositories"} />
        <div className="flex flex-col flex-1 max-h-screen">
          <main className="relative z-0 focus:outline-none" tabIndex={0}>
            <Header title={t("header.webhook")} props={
              location.pathname.startsWith("/settings") ? null : (
                <div className="flex space-x-8">
                  <Link
                    to={`/namespaces/${namespace}/namespace-summary?namespace_id=${namespaceId}`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                  >
                    {t("common.summary")}
                  </Link>
                  <Link
                    to={`/namespaces/${namespace}/repositories?namespace_id=${namespaceId}`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                  >
                    {t("common.repositoryList")}
                  </Link>
                  <Link
                    to={`/namespaces/${namespace}/members?namespace_id=${namespaceId}`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                  >
                    {t("common.members")}
                  </Link>
                  <Link
                    to={`/namespaces/${namespace}/daemon-tasks?namespace_id=${namespaceId}`}
                    className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                  >
                    {t("common.daemonTask")}
                  </Link>
                  <Link
                    to="#"
                    className="inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize"
                  >
                    {t("common.webhook")}
                  </Link>
                </div>
              )
            } />
            <div className="pt-1 pb-1 flex justify-between items-center min-h-15">
              <div className="px-4">
                <div className="flex gap-4">
                  <div className="relative mt-2 flex items-center">
                  </div>
                </div>
              </div>
              <div className="px-4 flex flex-col">
                <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                  onClick={() => { setCreateWebhookModal(true) }}
                >{t("common.create")}</button>
              </div>
            </div>
          </main>
          <div className="flex-1 flex overflow-y-auto">
            <div className="align-middle inline-block min-w-full border-b border-gray-200">
              <table className="min-w-full flex-1">
                <thead>
                  <tr>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">URL</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Enable</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">SSL Verify</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <OrderHeader text={"Created at"} orderStatus={createdAtOrder} setOrder={(e) => {
                        resetOrder();
                        setCreatedAtOrder(e);
                        setSortOrder(e);
                        setSortName("created_at");
                      }} />
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <OrderHeader text={"Updated at"} orderStatus={updatedAtOrder} setOrder={(e) => {
                        resetOrder();
                        setUpdatedAtOrder(e);
                        setSortOrder(e);
                        setSortName("updated_at");
                      }} />
                    </th>
                    <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      Action
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-100 max-h-max">
                  {
                    webhookList.items?.map((webhook, index) => {
                      return (
                        <TableItem key={webhook.id} index={index} userObj={userObj} namespaceObj={namespaceObj} localServer={localServer} webhookObj={webhook} setRefresh={setRefresh} />
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
      <Dialog open={createWebhookModal} onOpenChange={setCreateWebhookModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogTitle className="border-b pb-4">Create webhook</DialogTitle>
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
                      <div className="relative w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:start-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
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
                  <div className="relative w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:start-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                </label>
              </div>
            </div>
            <div className="mt-4 flex flex-row gap-4">
              {
                location.pathname.startsWith("/settings") ? (
                  <div className="flex items-center">
                    <input id="event-namespace" type="checkbox"
                      checked={eventNamespace}
                      onChange={e => setEventNamespace(!eventNamespace)}
                      className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                    <label htmlFor="event-namespace" className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300">Namespace Event</label>
                  </div>
                ) : null
              }
              <div className="flex items-center">
                <input id="event-member"
                  checked={eventMember}
                  onChange={e => setEventMember(!eventMember)}
                  type="checkbox" className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                <label htmlFor="event-member" className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300">Member Event</label>
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
                <input id="event-daemon-task-gc" type="checkbox"
                  checked={eventDaemonTaskGc}
                  onChange={e => setEventDaemonTaskGc(!eventDaemonTaskGc)}
                  className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                <label htmlFor="event-daemon-task-gc" className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300">Gc Event</label>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateWebhookModal(false)}>Cancel</Button>
            <Button onClick={() => createWebhook()}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Fragment >
  )
}

function TableItem({ localServer, index, userObj, namespaceObj, webhookObj, setRefresh }: { localServer: string, index: number, userObj: IUserSelf, namespaceObj: INamespaceItem, webhookObj: IWebhookItem, setRefresh: (param: any) => void }) {
  const canManageWebhook = userObj.role == UserRole.Admin || userObj.role == UserRole.Root || (namespaceObj.role != undefined && (namespaceObj.role == NamespaceRole.Admin || namespaceObj.role == NamespaceRole.Manager));
  const location = useLocation();
  const navigate = useNavigate();

  const [deleteWebhookModal, setDeleteWebhookModal] = useState(false);
  const [updateWebhookModal, setUpdateWebhookModal] = useState(false);

  const deleteWebhook = () => {
    axios.delete(`${localServer}/api/v1/webhooks/${webhookObj.id}`).then(response => {
      if (response.status === 204) {
        setRefresh({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const [enable, setEnable] = useState(webhookObj.enable);
  const [eventNamespace, setEventNamespace] = useState(webhookObj.event_namespace);
  const [eventRepository, setEventRepository] = useState(webhookObj.event_repository);
  const [eventTag, setEventTag] = useState(webhookObj.event_tag);
  const [eventMember, setEventMember] = useState(webhookObj.event_member);
  const [eventArtifact, setEventArtifact] = useState(webhookObj.event_artifact);

  const [retryTimes, setRetryTimes] = useState<string | number>(webhookObj.retry_times);
  const [retryTimesValid, setRetryTimesValid] = useState(true);
  useEffect(() => { setRetryTimesValid(Number.isInteger(retryTimes) && parseInt(retryTimes.toString()) >= 1 && parseInt(retryTimes.toString()) <= 5) }, [retryTimes]);
  const [retryDuration, setRetryDuration] = useState<string | number>(webhookObj.retry_duration);
  const [retryDurationValid, setRetryDurationValid] = useState(true);
  useEffect(() => { setRetryDurationValid(Number.isInteger(retryDuration) && parseInt(retryDuration.toString()) >= 0 && parseInt(retryDuration.toString()) <= 10) }, [retryDuration]);

  const [showSslVerify, setShowSslVerify] = useState(webhookObj.ssl_verify);

  const [sslVerify, setSslVerify] = useState(true);
  const [secret, setSecret] = useState<string | undefined>(webhookObj.secret);
  const [secretValid, setSecretValid] = useState(true);
  useEffect(() => { if (secret != undefined && secret.length >= 0 && secret.length <= 63) { setSecretValid(true); } }, [secret]);
  const [url, setUrl] = useState<string>(webhookObj.url);
  const [urlValid, setUrlValid] = useState(true);
  useEffect(() => { url != "" && setUrlValid((url.startsWith("http://") || url.startsWith("https://")) && /(https:\/\/www\.|http:\/\/www\.|https:\/\/|http:\/\/)?[a-zA-Z]{2,}(\.[a-zA-Z]{2,})(\.[a-zA-Z]{2,})?\/[a-zA-Z0-9]{2,}|((https:\/\/www\.|http:\/\/www\.|https:\/\/|http:\/\/)?[a-zA-Z]{2,}(\.[a-zA-Z]{2,})(\.[a-zA-Z]{2,})?)|(https:\/\/www\.|http:\/\/www\.|https:\/\/|http:\/\/)?[a-zA-Z0-9]{2,}\.[a-zA-Z0-9]{2,}\.[a-zA-Z0-9]{2,}(\.[a-zA-Z0-9]{2,})?/.test(url) && url.length <= 128) }, [url]);

  useEffect(() => {
    if (url.startsWith("https://")) {
      setShowSslVerify(true);
    } else {
      setShowSslVerify(false);
    }
  }, [url]);

  const updateWebhook = () => {
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
    if (namespaceObj.id != 0) {
      data["event_namespace"] = eventNamespace
      u += `?namespace_id=${namespaceObj.id}`
    }
    axios.put(`${localServer}/api/v1/webhooks/${webhookObj.id}`, data).then(response => {
      if (response.status === 204) {
        setRefresh({});
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
    <tr className="align-middle">
      <td className="px-6 py-4 w-5/6 whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer"
        onClick={() => {
          if (location.pathname.startsWith("/settings")) {
            navigate(`/settings/webhooks/${webhookObj.id}`);
          } else {
            navigate(`/namespaces/${namespaceObj.name}/webhooks/${webhookObj.id}?namespace_id=${namespaceObj.id}`);
          }
          // navigate(`/namespaces/${namespaceObj.name}/webhooks/${webhookObj.id}?namespace_id=${namespaceObj.id}`);
        }}
      >
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {webhookObj.url}
            </span>
          </div>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-center cursor-pointer">
        {webhookObj.enable ? "Active" : "Inactive"}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-center cursor-pointer">
        {webhookObj.ssl_verify ? "Enable" : "Disable"}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {dayjs.utc(webhookObj.created_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {dayjs.utc(webhookObj.updated_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}
      </td>
      <td className="pr-3 whitespace-nowrap text-center" onClick={e => {
        e.stopPropagation();
      }}>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button variant="ghost" size="icon-sm" className="text-gray-500">
                <span className="sr-only">Open options</span>
                <EllipsisVerticalIcon className="size-5" aria-hidden="true" />
              </Button>
            }
          />
          <DropdownMenuContent align="end" className="w-28">
            <DropdownMenuItem disabled={!canManageWebhook} onClick={() => setUpdateWebhookModal(true)}>
              Update
            </DropdownMenuItem>
            <DropdownMenuItem variant="destructive" disabled={!canManageWebhook} onClick={() => setDeleteWebhookModal(true)}>
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </td>
      <td className="absolute hidden" onClick={e => { e.preventDefault() }}>
        <Dialog open={updateWebhookModal} onOpenChange={setUpdateWebhookModal}>
          <DialogContent className="sm:max-w-lg">
            <DialogTitle className="border-b pb-4">Update webhook</DialogTitle>
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
                        <div className="relative w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:start-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
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
                    <div className="relative w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:start-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
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
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setUpdateWebhookModal(false)}>Cancel</Button>
              <Button onClick={() => updateWebhook()}>Save</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </td>
      <td className="absolute hidden" onClick={e => { e.preventDefault() }}>
        <Dialog open={deleteWebhookModal} onOpenChange={setDeleteWebhookModal}>
          <DialogContent className="sm:max-w-lg">
            <div className="flex items-start gap-4">
              <div className="mx-auto flex size-10 shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0">
                <ExclamationTriangleIcon className="size-6 text-red-600" aria-hidden="true" />
              </div>
              <div className="text-center sm:text-left">
                <DialogTitle>Delete webhook</DialogTitle>
                <DialogDescription className="mt-2">
                  Are you sure you want to delete the webhook{" "}
                  <span className="text-foreground font-medium">{webhookObj.url}</span>
                </DialogDescription>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDeleteWebhookModal(false)}>
                Cancel
              </Button>
              <Button variant="destructive" onClick={() => { setDeleteWebhookModal(false); deleteWebhook(); }}>
                Delete
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </td>
    </tr>
  );
}
