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
import { useNavigate } from 'react-router-dom';
import { useEffect, useState, Fragment } from "react";
import relativeTime from 'dayjs/plugin/relativeTime';
import { Dialog, Menu, Transition } from '@headlessui/react';
import { EllipsisVerticalIcon } from '@heroicons/react/20/solid';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline'

import calcUnit from "../../utils/calcUnit";
import Toast from "../../components/Notification";
import { INamespace, IHTTPError } from "../../interfaces";

import Quota from "../../components/Quota";
import QuotaSimple from "../../components/QuotaSimple";

dayjs.extend(relativeTime);

export default function TableItem({ localServer, index, namespace, setRefresh }: { localServer: string, index: number, namespace: INamespace, setRefresh: (param: any) => void }) {
  const navigate = useNavigate();

  const [updateNamespaceModal, setUpdateNamespaceModal] = useState(false);
  const [deleteNamespaceModal, setDeleteNamespaceModal] = useState(false);

  const [namespaceText, setNamespaceText] = useState(namespace.name);
  const [namespaceTextValid, setNamespaceTextValid] = useState(true);
  useEffect(() => { namespaceText != "" && setNamespaceTextValid(/^[a-z][0-9a-z-]{0,20}$/.test(namespaceText)) }, [namespaceText])
  const [descriptionText, setDescriptionText] = useState(namespace.description);
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => { descriptionText != "" && setDescriptionTextValid(/^.{0,30}$/.test(descriptionText)) }, [descriptionText]);
  const [repositoryCountLimit, setRepositoryCountLimit] = useState<string | number>(namespace.repository_limit);
  const [repositoryCountLimitValid, setRepositoryCountLimitValid] = useState(true);
  useEffect(() => { setRepositoryCountLimitValid(Number.isInteger(repositoryCountLimit) && parseInt(repositoryCountLimit.toString()) >= 0) }, [repositoryCountLimit])
  const [tagCountLimit, setTagCountLimit] = useState<string | number>(namespace.tag_limit);
  const [tagCountLimitValid, setTagCountLimitValid] = useState(true);
  useEffect(() => { setTagCountLimitValid(Number.isInteger(tagCountLimit) && parseInt(tagCountLimit.toString()) >= 0) }, [tagCountLimit])
  let calcUnitObj = calcUnit(namespace.size_limit);
  const [realSizeLimit, setRealSizeLimit] = useState(0);
  const [sizeLimit, setSizeLimit] = useState<string | number>(calcUnitObj.size);
  const [sizeLimitValid, setSizeLimitValid] = useState(true);
  const [sizeLimitUnit, setSizeLimitUnit] = useState(calcUnitObj.unit);
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
  }, [sizeLimit, sizeLimitUnit])
  const [namespaceVisibility, setNamespaceVisibility] = useState("private");

  const updateNamespace = () => {
    setUpdateNamespaceModal(false);
    axios.put(localServer + `/api/v1/namespaces/${namespace.id}`, {
      description: descriptionText,
      size_limit: realSizeLimit,
      repository_limit: repositoryCountLimit,
      tag_limit: tagCountLimit,
      visibility: namespaceVisibility,
    } as INamespace, {}).then(response => {
      if (response.status === 204) {
        setNamespaceText("");
        setDescriptionText("");
        setNamespaceVisibility("private")
        setRepositoryCountLimit(0);
        setTagCountLimit(0);
        setSizeLimit(0);
        setRefresh({});
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }

  const deleteNamespace = () => {
    axios.delete(localServer + `/api/v1/namespaces/${namespace.id}`).then(response => {
      if (response.status === 204) {
        setRefresh({});
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }

  return (
    <tr className="align-middle">
      <td className="px-6 py-4 w-5/6 whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer"
        onClick={() => {
          navigate(`/namespaces/${namespace.name}/repositories`);
        }}
      >
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {namespace.name}
              <span className="text-gray-500 font-normal ml-4">{namespace.description}</span>
            </span>
          </div>
        </div>
      </td>
      <td className="px-6 py-4 w-1/6 whitespace-nowrap text-gray-500">
        <Quota current={namespace.size} limit={namespace.size_limit} />
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        <QuotaSimple current={namespace.repository_count} limit={namespace.repository_limit} />
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        <QuotaSimple current={namespace.tag_count} limit={namespace.tag_limit} />
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        <span className='capitalize'>{namespace.visibility}</span>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs().to(dayjs(namespace.created_at))}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs().to(dayjs(namespace.updated_at))}
      </td>
      <td className="pr-3 whitespace-nowrap text-center" onClick={e => {
        e.stopPropagation();
      }}>
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
                      ' block px-3 py-1 text-sm leading-6 text-gray-900 cursor-pointer'
                    }
                    onClick={e => {
                      setUpdateNamespaceModal(true);
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
                      (active ? 'bg-gray-50' : '') + ' block px-3 py-1 text-sm leading-6 text-gray-900 hover:text-white hover:bg-red-600 cursor-pointer'
                    }
                    onClick={e => {
                      setDeleteNamespaceModal(true);
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
      <td className="absolute hidden" onClick={e => { e.preventDefault() }}>
        <Transition.Root show={updateNamespaceModal} as={Fragment}>
          <Dialog as="div" className="relative z-10" onClose={setUpdateNamespaceModal}>
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-600"
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
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <input
                        type="text"
                        name="namespace"
                        placeholder="2-20 lowercase characters"
                        className={(namespaceTextValid ? "disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-500 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-500 block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                        value={namespaceText}
                        onChange={e => {
                          setNamespaceText(e.target.value);
                        }}
                        disabled
                      />
                      {
                        namespaceTextValid ? (
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
                        namespaceTextValid ? (
                          <span></span>
                        ) : (
                          <span>
                            Not a valid namespace name, 2-20 lowercase characters.
                          </span>
                        )
                      }
                    </p>
                    <label htmlFor="first-name" className="block text-sm font-medium text-gray-700">
                      Description
                    </label>
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <input
                        type="text"
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
                            Not a valid description, max 30 characters.
                          </span>
                        )
                      }
                    </p>
                    <label htmlFor="namespace_visibility" className="block text-sm font-medium text-gray-700">
                      Visibility
                    </label>
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <select
                        id="namespace_visibility"
                        name="namespace_visibility"
                        className="mt-2 block w-full rounded-md border-0 py-1.5 pl-3 pr-10 text-gray-900 ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        value={namespaceVisibility}
                        onChange={e => { setNamespaceVisibility(e.target.value) }}
                      >
                        <option value="private">Private</option>
                        <option value="public">Public</option>
                      </select>
                    </div>
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
                    <div className="grid grid-cols-2 gap-4">
                      <div className="col-span-1">
                        <label htmlFor="repository_count_limit" className="block text-sm font-medium text-gray-700">
                          Repository count limit
                        </label>
                        <div className="relative mt-2 rounded-md shadow-sm">
                          <input
                            type="number"
                            id="repository_count_limit"
                            name="repository_count_limit"
                            placeholder="0 means no limit"
                            className={(repositoryCountLimitValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                            value={repositoryCountLimit}
                            onChange={e => setRepositoryCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                          />
                          {
                            repositoryCountLimitValid ? (
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
                            repositoryCountLimitValid ? (
                              <span></span>
                            ) : (
                              <span>
                                Not a valid repository count limit, should be non-negative integer.
                              </span>
                            )
                          }
                        </p>
                      </div>
                      <div className="col-span-1">
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
                      </div>
                    </div>
                    <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={() => updateNamespace()}
                      >
                        Update
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={() => setUpdateNamespaceModal(false)}
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
      <td className="absolute hidden" onClick={e => { e.preventDefault() }}>
        <Transition.Root show={deleteNamespaceModal} as={Fragment}>
          <Dialog as="div" className="relative z-10" onClose={setDeleteNamespaceModal}>
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
                          Delete namespace
                        </Dialog.Title>
                        <div className="mt-2">
                          <p className="text-sm text-gray-500">
                            Are you sure you want to delete the namespace <span className="text-black font-medium">{namespace.name}</span>
                          </p>
                        </div>
                      </div>
                    </div>
                    <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:ml-3 sm:w-auto"
                        onClick={e => { setDeleteNamespaceModal(false); deleteNamespace(); }}
                      >
                        Delete
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                        onClick={() => setDeleteNamespaceModal(false)}
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
    </tr >
  );
}
