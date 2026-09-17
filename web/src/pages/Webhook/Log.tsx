/**
 * Copyright 2024 sigma
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
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle } from "@/components/ui/dialog";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Link, useParams, useSearchParams, useLocation } from 'react-router-dom';

import Header from "../../components/Header";
import IMenu from "../../components/Menu";
import Notification from "../../components/Notification";
import Pagination from "../../components/Pagination";
import Settings from "../../Settings";
import { IHTTPError, INamespaceItem, IOrder, IUserSelf, IWebhookLogItem, IWebhookLogList } from "../../interfaces";
import OrderHeader from "../../components/OrderHeader";
import { EllipsisVerticalIcon, ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { NamespaceRole, UserRole } from "../../interfaces/enums";

export default function ({ localServer }: { localServer: string }) {
  const location = useLocation();
  const { namespace, webhook_id } = useParams<{ namespace: string, webhook_id: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId = searchParams.get('namespace_id');
  const [namespaceObj, setNamespaceObj] = useState<INamespaceItem>({} as INamespaceItem);
  const webhookId = parseInt(webhook_id || "0");

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

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

  const [refresh, setRefresh] = useState({});
  const [createdAtOrder, setCreatedAtOrder] = useState(IOrder.None);
  const [updatedAtOrder, setUpdatedAtOrder] = useState(IOrder.None);
  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");
  const [webhookLogList, setWebhookLogList] = useState<IWebhookLogList>({} as IWebhookLogList);

  const resetOrder = () => {
    setCreatedAtOrder(IOrder.None);
    setUpdatedAtOrder(IOrder.None);
  }

  const [fetchWebhookSuccess, setFetchWebhookSuccess] = useState(false);

  useEffect(() => {
    let url = localServer + `/api/v1/webhooks/${webhookId}/logs/?limit=${Settings.PageSize}&page=${page}`;
    if (sortName !== "") {
      url += `&sort=${sortName}&method=${sortOrder.toString()}`
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const webhookLogList = response.data as IWebhookLogList;
        setWebhookLogList(webhookLogList);
        setTotal(webhookLogList.total);
        setFetchWebhookSuccess(true);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
        setFetchWebhookSuccess(false);
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      setFetchWebhookSuccess(false);
    });
  }, [refresh, page, sortOrder, sortName]);

  const [webhookPingModal, setWebhookPingModal] = useState(false);

  const webhookPing = () => {
    axios.get(`${localServer}/api/v1/webhooks/${webhookId}/ping`).then(response => {
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

  useEffect(() => {
    const interval = setInterval(() => {
      if (fetchWebhookSuccess) {
        setRefresh({});
      }
    }, 5000);
    return () => {
      clearInterval(interval);
    };
  }, [fetchWebhookSuccess]);

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
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700 w-55">
        Less than 10, unit is second.
      </div>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item={location.pathname.startsWith("/settings") ? "webhooks" : "repositories"} />
        <div className="flex flex-col flex-1 max-h-screen">
          <main className="relative z-0 focus:outline-none" tabIndex={0}>
            <Header title="Webhook" props={
              location.pathname.startsWith("/settings") ? null : (
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
            <div className="pt-1 pb-1 flex justify-between items-center min-h-15">
              <div className="px-4">
                <div className="flex gap-4">
                  <div className="relative mt-2 flex items-center">
                  </div>
                </div>
              </div>
              <div className="px-4 flex flex-col">
                <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                  onClick={() => { setWebhookPingModal(true) }}
                >Ping</button>
              </div>
            </div>
          </main>
          <div className="flex-1 flex overflow-y-auto">
            <div className="align-middle inline-block min-w-full border-b border-gray-200">
              <table className="min-w-full flex-1">
                <thead>
                  <tr>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Event</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Action</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Status</span>
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
                    webhookLogList.items?.map((webhookLog, index) => {
                      return (
                        <TableItem key={webhookLog.id} index={index} userObj={userObj} namespaceObj={namespaceObj} localServer={localServer} webhookLogObj={webhookLog} setRefresh={setRefresh} />
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
      <Dialog open={webhookPingModal} onOpenChange={setWebhookPingModal}>
        <DialogContent className="sm:max-w-lg">
          <div className="flex items-start gap-4">
            <div className="mx-auto flex size-10 shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0">
              <ExclamationTriangleIcon className="size-6 text-red-600" aria-hidden="true" />
            </div>
            <div className="text-center sm:text-left">
              <DialogTitle>Send webhook ping event</DialogTitle>
              <DialogDescription className="mt-2">
                Are you sure you want to send the webhook ping event?
              </DialogDescription>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setWebhookPingModal(false)}>
              Cancel
            </Button>
            <Button onClick={() => { setWebhookPingModal(false); webhookPing(); }}>
              Send
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Fragment >
  );
}

function TableItem({ localServer, index, userObj, namespaceObj, webhookLogObj, setRefresh }: { localServer: string, index: number, userObj: IUserSelf, namespaceObj: INamespaceItem, webhookLogObj: IWebhookLogItem, setRefresh: (param: any) => void }) {
  const canManageWebhookLog = userObj.role == UserRole.Admin || userObj.role == UserRole.Root || (namespaceObj.role != undefined && (namespaceObj.role == NamespaceRole.Admin || namespaceObj.role == NamespaceRole.Manager));
  const [webhookLogResendModal, setWebhookLogResendModal] = useState(false);
  const [webhookLogDeleteModal, setWebhookLogDeleteModal] = useState(false);

  const webhookLogResend = () => {
    axios.get(`${localServer}/api/v1/webhooks/${webhookLogObj.id}/logs/${webhookLogObj.id}/resend`).then(response => {
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

  const webhookLogDelete = () => {
    axios.delete(`${localServer}/api/v1/webhooks/${webhookLogObj.id}/logs/${webhookLogObj.id}`).then(response => {
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

  const [drawerShow, setDrawerShow] = useState(false);

  return (
    <tr className="align-middle">
      <td className="px-6 py-4 w-5/6 whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer"
        onClick={() => {
          setDrawerShow(true);
        }}
      >
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {webhookLogObj.resource_type}
            </span>
          </div>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-center cursor-pointer">
        {webhookLogObj.action}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-center cursor-pointer">
        {webhookLogObj.status_code}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {dayjs.utc(webhookLogObj.created_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {dayjs.utc(webhookLogObj.updated_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}
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
            <DropdownMenuItem disabled={!canManageWebhookLog} onClick={() => setWebhookLogResendModal(true)}>
              Resend
            </DropdownMenuItem>
            <DropdownMenuItem variant="destructive" disabled={!canManageWebhookLog} onClick={() => setWebhookLogDeleteModal(true)}>
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </td>
      <td className="absolute hidden" onClick={e => { e.preventDefault() }}>
        <Dialog open={webhookLogResendModal} onOpenChange={setWebhookLogResendModal}>
          <DialogContent className="sm:max-w-lg">
            <div className="flex items-start gap-4">
              <div className="mx-auto flex size-10 shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0">
                <ExclamationTriangleIcon className="size-6 text-red-600" aria-hidden="true" />
              </div>
              <div className="text-center sm:text-left">
                <DialogTitle>Resend webhook event</DialogTitle>
                <DialogDescription className="mt-2">
                  Are you sure you want to resend the webhook event{" "}
                  <span className="text-foreground font-medium capitalize">{webhookLogObj.resource_type}</span>
                </DialogDescription>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setWebhookLogResendModal(false)}>
                Cancel
              </Button>
              <Button variant="destructive" onClick={() => { setWebhookLogResendModal(false); webhookLogResend() }}>
                Resend
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </td>
      <td className="absolute hidden" onClick={e => { e.preventDefault() }}>
        <Dialog open={webhookLogDeleteModal} onOpenChange={setWebhookLogDeleteModal}>
          <DialogContent className="sm:max-w-lg">
            <div className="flex items-start gap-4">
              <div className="mx-auto flex size-10 shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0">
                <ExclamationTriangleIcon className="size-6 text-red-600" aria-hidden="true" />
              </div>
              <div className="text-center sm:text-left">
                <DialogTitle>Delete webhook log</DialogTitle>
                <DialogDescription className="mt-2">
                  Are you sure you want to delete the webhook event{" "}
                  <span className="text-foreground font-medium capitalize">{webhookLogObj.resource_type}</span>
                </DialogDescription>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setWebhookLogDeleteModal(false)}>
                Cancel
              </Button>
              <Button variant="destructive" onClick={() => { setWebhookLogDeleteModal(false); webhookLogDelete() }}>
                Delete
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </td>
      <td className="absolute" onClick={e => { e.preventDefault() }}>
        <div className={`fixed inset-0 z-30 bg-black/20 transition-opacity ${drawerShow ? "opacity-100" : "pointer-events-none opacity-0"}`} onClick={() => setDrawerShow(false)} />
        <div id={`drawer-${index}`} className={`fixed top-0 right-0 z-40 h-screen p-4 overflow-y-auto bg-white w-200 shadow-xl transition-transform dark:bg-gray-800 ${drawerShow ? "translate-x-0" : "translate-x-full"}`} aria-labelledby="drawer-right-label">
          <button
            type="button"
            className="absolute right-4 top-4 rounded-md text-gray-400 hover:text-gray-600"
            onClick={() => setDrawerShow(false)}
          >
            <span className="sr-only">Close drawer</span>
            ×
          </button>
          <h5 id="drawer-right-label" className="items-center pb-4 pr-8 text-base font-semibold text-gray-500 dark:text-gray-400 border-b">
            Request headers
          </h5>
          <kbd className="text-gray-600 whitespace-pre-wrap text-sm py-4 block border-b">
            {
              Object.entries(JSON.parse(webhookLogObj.req_header))
                .map(([k, v]) => `${k}: ${v}`)
                .join('\n')
            }
          </kbd>
          <h5 id="drawer-right-label" className="items-center py-4 text-base font-semibold text-gray-500 dark:text-gray-400 border-b">
            Request body
          </h5>
          <kbd className="text-gray-600 whitespace-pre-wrap text-sm py-4 block border-b">
            {
              JSON.stringify(JSON.parse(webhookLogObj.req_body), null, 2)
            }
          </kbd>
          <h5 id="drawer-right-label" className="items-center py-4 text-base font-semibold text-gray-500 dark:text-gray-400 border-b">
            Response headers
          </h5>
          <kbd className="text-gray-600 whitespace-pre-wrap text-sm py-4 block border-b">
            {
              Object.entries(JSON.parse(webhookLogObj.resp_header))
                .map(([k, v]) => `${k}: ${v}`)
                .join('\n')
            }
          </kbd>
          <h5 id="drawer-right-label" className="items-center py-4 text-base font-semibold text-gray-500 dark:text-gray-400 border-b">
            Response body
          </h5>
          <kbd className="text-gray-600 whitespace-pre-wrap text-sm mt-4 block">
            {
              webhookLogObj.resp_body
            }
          </kbd>
        </div>
      </td>
    </tr>
  );
}
