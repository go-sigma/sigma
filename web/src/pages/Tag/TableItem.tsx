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
import humanFormat from "human-format";
import { useState, Fragment } from "react";
import { useCopyToClipboard } from 'react-use';
import { useNavigate } from 'react-router-dom';
import relativeTime from 'dayjs/plugin/relativeTime';
import { Dialog, Menu, Transition } from "@headlessui/react";
import { EllipsisVerticalIcon } from "@heroicons/react/20/solid";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";

import Toast from "../../components/Notification";
import { ITag, IHTTPError } from "../../interfaces";

dayjs.extend(relativeTime);

export default function TableItem({ localServer, index, namespace, repository, tag, setRefresh }: { localServer: string, index: number, namespace: string, repository: string, tag: ITag, setRefresh: (param: any) => void }) {
  const navigate = useNavigate();

  const [deleteTagModal, setDeleteTagModal] = useState(false);
  const [copyCommandModal, setCopyCommandModal] = useState(false);
  const [, copyToClipboard] = useCopyToClipboard();

  const [copyStatus1, setCopyStatus1] = useState(false);
  const [copyStatus2, setCopyStatus2] = useState(false);
  const [copyStatus3, setCopyStatus3] = useState(false);
  const [copyStatus4, setCopyStatus4] = useState(false);

  const imageDomain = () => {
    if (localServer.startsWith("http://")) {
      return localServer.substring(7);
    } else if (localServer.startsWith("https://")) {
      return localServer.substring(8)
    } else {
      return localServer;
    }
  }

  const deleteTag = () => {
    axios.delete(localServer + `/api/v1/namespaces/${namespace}/tags/${tag.id}?repository=${repository}`).then(response => {
      if (response.status === 204) {
        setRefresh({});
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
    <tr className="">
      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer">
        <div className="flex items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            {tag.name}
          </div>
        </div>
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-xs text-gray-500 text-right">
        {tag.digest}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {humanFormat(tag.size, { scale: "binary", unit: "B" })}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {tag.pull_times === 0 ? "Never pulled" : dayjs().to(dayjs(tag.last_pull))}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {tag.pull_times}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs().to(dayjs(tag.pushed_at))}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right"
        onClick={e => { setCopyCommandModal(true) }}
      >
        <div className="mx-auto w-5 h-5 cursor-pointer">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 3.75H6.912a2.25 2.25 0 00-2.15 1.588L2.35 13.177a2.25 2.25 0 00-.1.661V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18v-4.162c0-.224-.034-.447-.1-.661L19.24 5.338a2.25 2.25 0 00-2.15-1.588H15M2.25 13.5h3.86a2.25 2.25 0 012.012 1.244l.256.512a2.25 2.25 0 002.013 1.244h3.218a2.25 2.25 0 002.013-1.244l.256-.512a2.25 2.25 0 012.013-1.244h3.859M12 3v8.25m0 0l-3-3m3 3l3-3" />
          </svg>
        </div>
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
                      (active ? 'bg-gray-50' : '') + ' block px-3 py-1 text-sm leading-6 text-gray-900 hover:text-white hover:bg-red-600 cursor-pointer'
                    }
                    onClick={e => { setDeleteTagModal(true) }}
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
        <Transition.Root show={deleteTagModal} as={Fragment}>
          <Dialog as="div" className="relative z-10" onClose={setDeleteTagModal}>
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
                          Delete tag
                        </Dialog.Title>
                        <div className="mt-2">
                          <p className="text-sm text-gray-500">
                            Are you sure you want to delete the tag <span className="text-black font-medium">{imageDomain()}/{repository}:{tag.name}</span>
                          </p>
                        </div>
                      </div>
                    </div>
                    <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:ml-3 sm:w-auto"
                        onClick={e => { setDeleteTagModal(false); deleteTag(); }}
                      >
                        Delete
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                        onClick={() => setDeleteTagModal(false)}
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
        <Transition.Root show={copyCommandModal} as={Fragment}>
          <Dialog as="div" className="relative z-10" onClose={setCopyCommandModal}>
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

            <div className="fixed inset-0 z-10 overflow-y-auto pt-40">
              <div className="flex min-h-full justify-center p-4 text-center items-start">
                <Transition.Child
                  as={Fragment}
                  enter="ease-out duration-300"
                  enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                  enterTo="opacity-100 translate-y-0 sm:scale-100"
                  leave="ease-in duration-500"
                  leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                  leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                >
                  <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:p-6">
                    <div className="sm:flex sm:items-start">
                      <div className="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                        <div className="m-2">
                          <code className="text-xs">
                            <span className={copyStatus1 ? "cursor-pointer text-green-500" : "cursor-pointer"}
                              onClick={e => {
                                setCopyStatus1(true);
                                copyToClipboard(`docker pull ${imageDomain()}/${repository}:${tag.name}`);
                                setTimeout(() => { setCopyStatus1(false) }, 1000);
                                Toast({ level: "info", title: "Copied", message: "You have copied the text to your clipboard", autoClose: 800 });
                                setCopyCommandModal(false);
                              }}
                            >docker pull {imageDomain()}/{repository}:{tag.name}</span>
                            <br />
                            <span className={copyStatus2 ? "cursor-pointer text-green-500" : "cursor-pointer"}
                              onClick={e => {
                                setCopyStatus2(true);
                                copyToClipboard(`podman pull ${imageDomain()}/${repository}:${tag.name}`);
                                setTimeout(() => { setCopyStatus2(false) }, 1000);
                                Toast({ level: "info", title: "Copied", message: "You have copied the text to your clipboard", autoClose: 800 });
                                setCopyCommandModal(false);
                              }}
                            >podman pull {imageDomain()}/{repository}:{tag.name}</span>
                            <br />
                            <span className={copyStatus3 ? "cursor-pointer text-green-500" : "cursor-pointer"}
                              onClick={e => {
                                setCopyStatus3(true);
                                copyToClipboard(`podman pull ${imageDomain()}/${repository}@${tag.digest}`);
                                setTimeout(() => { setCopyStatus3(false) }, 1000);
                                Toast({ level: "info", title: "Copied", message: "You have copied the text to your clipboard", autoClose: 800 });
                                setCopyCommandModal(false);
                              }}
                            >docker pull {imageDomain()}/{repository}@{tag.digest}</span>
                            <br />
                            <span className={copyStatus4 ? "cursor-pointer text-green-500" : "cursor-pointer"}
                              onClick={e => {
                                setCopyStatus4(true);
                                copyToClipboard(`podman pull ${imageDomain()}/${repository}@${tag.digest}`);
                                setTimeout(() => { setCopyStatus4(false) }, 1000);
                                Toast({ level: "info", title: "Copied", message: "You have copied the text to your clipboard", autoClose: 800 });
                                setCopyCommandModal(false);
                              }}
                            >podman pull {imageDomain()}/{repository}@{tag.digest}</span>
                          </code>
                        </div>
                      </div>
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
