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
import { Tooltip } from 'flowbite';
import Toast from 'react-hot-toast';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Dialog, Menu, Transition } from "@headlessui/react";
import { EllipsisVerticalIcon } from "@heroicons/react/20/solid";
import { useParams, useSearchParams, Link } from 'react-router-dom';
import parser from 'cron-parser';
import dayjs from 'dayjs';

import Settings from "../../Settings";
import IMenu from "../../components/Menu";
import Header from "../../components/Header";
import { IGcRepositoryRule, IHTTPError } from "../../interfaces";
import Notification from "../../components/Notification";

export default function Repository({ localServer }: { localServer: string }) {
  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId = searchParams.get('namespace_id');

  const [gcRepositoryRuleExist, setGcRepositoryRuleExist] = useState(false);
  const [gcTagRuleExist, setGcTagRuleExist] = useState(false);
  const [gcArtifactRuleExist, setGcArtifactRuleExist] = useState(false);



  useEffect(() => {
    let url = `${localServer}/api/v1/daemons/gc-tag/${namespaceId}/`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        setGcTagRuleExist(true);
      } else if (response?.status === 404) {
        console.log("test");
        setGcTagRuleExist(false);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }, [])

  useEffect(() => {
    let url = `${localServer}/api/v1/daemons/gc-artifact/${namespaceId}/`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        setGcArtifactRuleExist(true);
      } else if (response?.status === 404) {
        console.log("test");
        setGcArtifactRuleExist(false);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }, []);

  const [gcRepositoryRuleConfigModal, setGcRepositoryRuleConfigModal] = useState(false);

  const [gcRepositoryRetentionDays, setGcRepositoryRetentionDays] = useState<string | number>(0);
  const [gcRepositoryRetentionDaysValid, setGcRepositoryRetentionDaysValid] = useState(true);
  useEffect(() => { setGcRepositoryRetentionDaysValid(Number.isInteger(gcRepositoryRetentionDays) && parseInt(gcRepositoryRetentionDays.toString()) >= 0 && parseInt(gcRepositoryRetentionDays.toString()) <= 180) }, [gcRepositoryRetentionDays]);

  const [gcRepositoryCronEnabled, setGcRepositoryCronEnabled] = useState(false);

  const [gcRepositoryCronRule, setGcRepositoryCronRule] = useState("");
  const [gcRepositoryCronRuleValid, setGcRepositoryCronRuleValid] = useState(true);
  const [gcRepositoryCronRuleNextRunAt, setGcRepositoryCronRuleNextRunAt] = useState("");

  useEffect(() => {
    let url = `${localServer}/api/v1/daemons/gc-repository/${namespaceId}/`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const gcRepositoryRule = response.data as IGcRepositoryRule;
        setGcRepositoryRuleExist(true);
        setGcRepositoryRetentionDays(gcRepositoryRule.retention_day);
        setGcRepositoryCronEnabled(gcRepositoryRule.cron_enabled);
        if (gcRepositoryRule.cron_enabled) {
          setGcRepositoryCronRule(gcRepositoryRule.cron_rule == undefined ? "" : gcRepositoryRule.cron_rule);
        }
      } else if (response?.status === 404) {
        console.log("test");
        setGcRepositoryRuleExist(false);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }, [])

  useEffect(() => {
    if (gcRepositoryCronRule.length > 0) {
      axios.post(localServer + `/api/v1/validators/cron`, {
        cron: gcRepositoryCronRule,
      }).then(response => {
        if (response?.status === 204) {
          setGcRepositoryCronRuleValid(true);
          let next = parser.parseExpression(gcRepositoryCronRule).next()
          setGcRepositoryCronRuleNextRunAt(`${dayjs(next.toDate()).format('YYYY-MM-DD HH:mm')}`);
        } else {
          setGcRepositoryCronRuleValid(false);
        }
      }).catch(error => {
        console.log(error);
        setGcRepositoryCronRuleValid(false);
      });
    }
  }, [gcRepositoryCronRule]);

  const createOrUpdateGcRepository = () => {
    if (!(gcRepositoryRetentionDaysValid && ((gcRepositoryCronEnabled && gcRepositoryCronRuleValid) || !gcRepositoryCronEnabled))) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    const data: { [key: string]: any } = {
      retention_day: gcRepositoryRetentionDays,
      cron_enabled: gcRepositoryCronEnabled,
    };
    if (gcRepositoryCronEnabled) {
      data["cron_rule"] = gcRepositoryCronRule;
    }
    axios.put(localServer + `/api/v1/daemons/gc-repository/${namespaceId}/`, data).then(response => {
      if (response?.status === 204) {
        Notification({ level: "success", title: "Success", message: "Create user success" });
        setGcRepositoryRuleConfigModal(false);
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
          <title>sigma - Namespace Daemon Task</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item="repositories" namespace={namespace} />
        <div className="flex flex-col w-0 flex-1 overflow-visible">
          <main className="relative z-0 focus:outline-none" tabIndex={0}>
            <Header title="Namespace - Daemon Task" props={
              (
                <div className="sm:flex sm:space-x-8">
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
            <div className="flex flex-1 overflow-visible">
              <div className="align-middle inline-block min-w-full border-gray-200">
                <table className="min-w-full flex-1 overflow-visible">
                  <thead>
                    <tr className="border-gray-200">
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Task</span>
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Status</span>
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Last trigger</span>
                      </th>
                      <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Next trigger</span>
                      </th>

                      <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                        <span className="lg:pl-2">Action</span>
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                      // onClick={() => {
                      //   navigate(`/namespaces/${namespace}/repository/tags?repository=${repository.name}&repository_id=${repository.id}`);
                      // }}
                      >
                        <div className="flex items-center space-x-3 lg:pl-2">
                          Garbage collect the empty repository
                        </div>
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        Running
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        2023-11-08 13:32:12
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        2023-11-08 13:32:12
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
                            <Menu.Items className={(1 > 10 ? "menu-action-top" : "mt-2") + " text-left absolute right-0 z-10 w-30 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"} >
                              <Menu.Item>
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900 cursor-pointer'
                                    }
                                    onClick={e => { setGcRepositoryRuleConfigModal(true); }}
                                  >
                                    {
                                      gcRepositoryRuleExist ? "Update" : "Configuration"
                                    }
                                  </div>
                                )}
                              </Menu.Item>
                              <Menu.Item
                                disabled={!gcRepositoryRuleExist}
                              >
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      (gcRepositoryRuleExist ? 'cursor-point' : 'cursor-not-allowed') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900'
                                    }
                                    onClick={e => {
                                      Toast.success('Task pushed into work queue');
                                    }}
                                  >
                                    Run
                                  </div>
                                )}
                              </Menu.Item>
                            </Menu.Items>
                          </Transition>
                        </Menu>
                      </td>
                    </tr>
                    <tr>
                      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                      // onClick={() => {
                      //   navigate(`/namespaces/${namespace}/repository/tags?repository=${repository.name}&repository_id=${repository.id}`);
                      // }}
                      >
                        <div className="flex items-center space-x-3 lg:pl-2">
                          Garbage collect the tag
                        </div>
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        Running
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        2023-11-08 13:32:12
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        2023-11-08 13:32:12
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
                            <Menu.Items className={(1 > 10 ? "menu-action-top" : "mt-2") + " text-left absolute right-0 z-10 w-30 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"} >
                              <Menu.Item>
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900 cursor-pointer'
                                    }
                                  // onClick={e => { setDeleteRepositoryModal(true) }}
                                  >
                                    {
                                      gcTagRuleExist ? "Update" : "Configuration"
                                    }
                                  </div>
                                )}
                              </Menu.Item>
                              <Menu.Item
                                disabled={!gcTagRuleExist}
                              >
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      (gcTagRuleExist ? 'cursor-point' : 'cursor-not-allowed') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900'
                                    }
                                  // onClick={e => { setUpdateRepositoryModal(true) }}
                                  >
                                    Run
                                  </div>
                                )}
                              </Menu.Item>
                            </Menu.Items>
                          </Transition>
                        </Menu>
                      </td>
                    </tr>
                    <tr>
                      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-normal text-gray-900 cursor-pointer"
                      // onClick={() => {
                      //   navigate(`/namespaces/${namespace}/repository/tags?repository=${repository.name}&repository_id=${repository.id}`);
                      // }}
                      >
                        <div className="flex items-center space-x-3 lg:pl-2">
                          Garbage collect the artifact
                        </div>
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        Running
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        2023-11-08 13:32:12
                      </td>
                      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                        2023-11-08 13:32:12
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
                            <Menu.Items className={(1 > 10 ? "menu-action-top" : "mt-2") + " text-left absolute right-0 z-10 w-30 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"} >
                              <Menu.Item>
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900 cursor-pointer'
                                    }
                                  // onClick={e => { setDeleteRepositoryModal(true) }}
                                  >
                                    {
                                      gcArtifactRuleExist ? "Update" : "Configuration"
                                    }
                                  </div>
                                )}
                              </Menu.Item>
                              <Menu.Item
                                disabled={!gcArtifactRuleExist}
                              >
                                {({ active }) => (
                                  <div
                                    className={
                                      (active ? 'bg-gray-100' : '') +
                                      (gcArtifactRuleExist ? 'cursor-point' : 'cursor-not-allowed') +
                                      ' block px-3 py-1 text-sm leading-6 text-gray-900'
                                    }
                                  // onClick={e => { setUpdateRepositoryModal(true) }}
                                  >
                                    Run
                                  </div>
                                )}
                              </Menu.Item>
                            </Menu.Items>
                          </Transition>
                        </Menu>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </main>
        </div>
      </div>
      <div
        id="tooltip-retention-days"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700 w-[350px]">
        Retention the empty repository for specific days,
        0 means delete immediately, available 0-180
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div
        id="tooltip-cron-rule"
        role="tooltip"
        className="absolute z-50 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
        '0 0 * * 6' means run at 00:00 every Saturday
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <Transition.Root show={gcRepositoryRuleConfigModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setGcRepositoryRuleConfigModal}>
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
                <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-6 pb-4 text-left shadow-xl transition-all">
                  <Dialog.Title
                    as="h3"
                    className="text-lg font-medium leading-6 text-gray-900 border-b pt-4 pb-4"
                  >
                    Garbage collect empty repository config
                  </Dialog.Title>
                  <div className="flex flex-col gap-0 mt-4">
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Retention Days</span>
                            <div className="flex flex-row cursor-pointer"
                              id="gcRepositoryRetentionDaysHelp"
                              onClick={e => {
                                let tooltip = new Tooltip(document.getElementById("tooltip-retention-days"),
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
                      <div className="col-span-4">
                        <div className="relative rounded-md shadow-sm">
                          <input
                            type="text"
                            id="namespace_count_limit"
                            name="namespace_count_limit"
                            placeholder="0 means no limit"
                            className={(gcRepositoryRetentionDaysValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                            value={gcRepositoryRetentionDays}
                            onChange={e => setGcRepositoryRetentionDays(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                          />
                          {
                            gcRepositoryRetentionDaysValid ? null : (
                              <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                </svg>
                              </div>
                            )
                          }
                        </div>

                      </div>
                    </div>
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2"></div>
                      <div className="col-span-4">
                        {
                          gcRepositoryRetentionDaysValid ? null : (
                            <p className="mt-1 text-xs text-red-600">
                              <span>
                                Not a valid retention days limit, available 0-180.
                              </span>
                            </p>
                          )
                        }
                      </div>
                    </div>
                    <div className="grid grid-cols-6 gap-4 mt-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">Cron Enabled</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4">
                        <div className="mt-0.5 flex flex-row items-center h-[36px]">
                          <label className="relative inline-flex items-center cursor-pointer">
                            <input type="checkbox" checked={gcRepositoryCronEnabled} className="sr-only peer" onChange={e => {
                              setGcRepositoryCronEnabled(e.target.checked);
                            }} />
                            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                          </label>
                        </div>
                      </div>
                    </div>
                    {
                      !gcRepositoryCronEnabled ? null : (
                        <>
                          <div className="grid grid-cols-6 gap-4 mt-4">
                            <div className="col-span-2 flex flex-row">
                              <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                                <div className="flex">
                                  <span className="text-red-600">*</span>
                                  <span className="leading-6 ">Cron Rule</span>
                                  <div className="flex flex-row cursor-pointer"
                                    id="gcRepositoryRuleHelp"
                                    onClick={e => {
                                      let tooltip = new Tooltip(document.getElementById("tooltip-cron-rule"),
                                        document.getElementById("gcRepositoryRuleHelp"), { triggerType: "click" });
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

                            <div className="col-span-4">
                              <div className="relative rounded-md shadow-sm">
                                <input
                                  type="text"
                                  id="gc_repository_cron_rule"
                                  name="gc_repository_cron_rule"
                                  placeholder="cron rule"
                                  className={(gcRepositoryCronRuleValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                                  value={gcRepositoryCronRule}
                                  onChange={e => setGcRepositoryCronRule(e.target.value)}
                                />
                                {
                                  gcRepositoryCronRuleValid ? null : (
                                    <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-5 w-5 text-red-500">
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                                      </svg>
                                    </div>
                                  )
                                }
                              </div>

                            </div>
                          </div>
                          <div className="grid grid-cols-6 gap-4">
                            <div className="col-span-2">
                            </div>
                            <div className="col-span-4">
                              {
                                !gcRepositoryCronRuleValid ? (
                                  <p className="mt-1 text-xs text-red-600">
                                    <span>
                                      Not a valid cron rule, you can try '0 0 * * 6'.
                                    </span>
                                  </p>
                                ) : gcRepositoryCronRule == "" ? null : (
                                  <p className="mt-1 text-xs text-gray-600">
                                    <span>
                                      Next run at '{gcRepositoryCronRuleNextRunAt}'.
                                    </span>
                                  </p>
                                )
                              }
                            </div>
                          </div>
                        </>
                      )
                    }
                    <div className="flex flex-row-reverse mt-5">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={e => createOrUpdateGcRepository()}
                      >
                        Create
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={e => { setGcRepositoryRuleConfigModal(false) }}
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
