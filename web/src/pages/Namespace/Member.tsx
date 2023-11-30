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
import Toast from 'react-hot-toast';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { useParams, Link, useSearchParams } from 'react-router-dom';
import { ChevronUpDownIcon, CheckIcon, EllipsisVerticalIcon } from '@heroicons/react/20/solid';
import { Dialog, Listbox, Menu, Transition, Combobox } from "@headlessui/react";

import Settings from "../../Settings";
import IMenu from "../../components/Menu";
import Header from "../../components/Header";
import Pagination from "../../components/Pagination";
import Notification from "../../components/Notification";
import { IHTTPError, INamespaceItem, INamespaceRoleItem as INamespaceMemberItem, INamespaceRoleList, IUserItem, IUserList } from "../../interfaces";

const namespaceRoles = [
  { id: 1, name: 'NamespaceAdmin' },
  { id: 2, name: 'NamespaceManager' },
  { id: 3, name: 'NamespaceReader' },
];

export default function Member({ localServer }: { localServer: string }) {
  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId = searchParams.get('namespace_id') == null ? 0 : parseInt(searchParams.get('namespace_id') || "");

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [refresh, setRefresh] = useState({});
  const [memberSearch, setUsernameSearch] = useState("");
  const [createUserNamespaceModal, setCreateUserNamespaceModal] = useState(false);
  const [namespaceObj, setNamespaceObj] = useState<INamespaceItem>({} as INamespaceItem);
  const [memberList, setMemberList] = useState<INamespaceRoleList>({} as INamespaceRoleList);

  useEffect(() => {
    if (namespaceId == 0) {
      return;
    }
    let url = `${localServer}/api/v1/namespaces/${namespaceId}`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const namespaceObj = response.data as INamespaceItem;
        setNamespaceObj(namespaceObj);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespaceId]);

  useEffect(() => {
    let url = `${localServer}/api/v1/namespaces/${namespaceId}/members/?limit=${Settings.AutoCompleteSize}`;
    if (memberSearch !== "") {
      url += `&name=${memberSearch}`;
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const namespaceRoleList = response.data as INamespaceRoleList;
        setMemberList(namespaceRoleList);
        setTotal(namespaceRoleList.total);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [refresh]);

  const [userSearch, setUserSearch] = useState('');
  const [userList, setUserList] = useState<IUserItem[]>();
  const [userSelectedValid, setUserSelectValid] = useState(true);
  const [userSelected, setUserSelected] = useState<IUserItem>({} as IUserItem);
  const [addNamespaceRoleRole, setAddNamespaceRoleRole] = useState("NamespaceReader");

  useEffect(() => {
    let url = `${localServer}/api/v1/users/?limit=${Settings.AutoCompleteSize}&without_admin=true`;
    if (userSearch !== "") {
      url += `&name=${userSearch}`;
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const namespaceList = response.data as IUserList;
        setUserList(namespaceList.items);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [userSearch]);

  useEffect(() => {
    if (userSelected.username !== undefined) {
      setUserSelectValid(true);
    }
  }, [userSelected]);

  const addMember = () => {
    if (userSelected.username === undefined) {
      setUserSelectValid(false);
      return;
    }
    let url = `${localServer}/api/v1/namespaces/${namespaceId}/members/`;
    axios.post(url, {
      user_id: userSelected.id,
      role: addNamespaceRoleRole,
    }).then(response => {
      if (response?.status === 201) {
        Toast.success("Add member to namespace success");
        setCreateUserNamespaceModal(false);
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
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Namespace Member</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item="Repository" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title="Namespace - Member"
              props={
                (
                  <div className="sm:flex sm:space-x-8">
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
                      className="z-10 inline-flex items-center border-b border-indigo-500 px-1 pt-1 text-sm font-medium text-gray-900 capitalize"
                      onClick={e => {
                        e.stopPropagation();
                        e.preventDefault();
                      }}
                    >
                      Members
                    </Link>
                    {/* <Link
                      to={`/namespaces/${namespace}/namespace-webhooks`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Webhook
                    </Link> */}
                    <Link
                      to={`/namespaces/${namespace}/daemon-tasks?namespace_id=${namespaceId}`}
                      className="inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize"
                    >
                      Daemon task
                    </Link>
                  </div>
                )
              }
            />
            <div className="pt-2 pb-2 flex justify-between">
              <div className="pr-2 pl-2">
                <div className="flex gap-4">
                  <div className="relative mt-2 flex items-center">
                    <label
                      htmlFor="usernameSearch"
                      className="absolute -top-2 left-2 inline-block bg-white px-1 text-xs font-medium text-gray-900"
                    >
                      Username
                    </label>
                    <input
                      type="text"
                      id="usernameSearch"
                      placeholder="search username"
                      value={memberSearch}
                      onChange={e => { setUsernameSearch(e.target.value); }}
                      onKeyDown={e => {
                        if (e.key == "Enter") {
                          setRefresh({});
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
                <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                  onClick={e => { setCreateUserNamespaceModal(true) }}
                >Add</button>
              </div>
            </div>
          </main>
          <div className="flex-1 flex overflow-y-auto">
            <div className="align-middle inline-block min-w-full border-b border-gray-200">
              <table className="min-w-full flex-1">
                <thead>
                  <tr>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Username</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Role</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Added</span>
                    </th>
                    <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Action</span>
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-100 max-h-max">
                  {
                    memberList.items?.map((member, index) => {
                      return (
                        <TableItem key={index} index={index} localServer={localServer} namespace={namespaceObj} userSelectedArg={{ username: member.username, id: member.user_id } as IUserItem} member={member} setRefresh={setRefresh} />
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
      <Transition.Root show={createUserNamespaceModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setCreateUserNamespaceModal}>
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
                    Add member
                  </Dialog.Title>
                  <div className="flex flex-col gap-0 mt-4">
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2 flex flex-row">
                        <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                          <div className="flex">
                            <span className="text-red-600">*</span>
                            <span className="leading-6 ">User</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4">
                        <Combobox value={userSelected}
                          onChange={(user: IUserItem) => {
                            setUserSelected(user);
                          }}>
                          <div className="relative mt-1">
                            <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                              <Combobox.Input
                                id="namespace"
                                className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                                displayValue={(user: IUserItem) => user.username}
                                onChange={event => {
                                  setUserSearch(event.target.value);
                                }}
                              />
                              <Combobox.Button className="absolute inset-y-0 right-0 flex items-center pr-2">
                                <ChevronUpDownIcon
                                  className="h-5 w-5 text-gray-400"
                                  aria-hidden="true"
                                />
                              </Combobox.Button>
                            </div>
                            <Transition
                              as={Fragment}
                              leave="transition ease-in duration-100"
                              leaveFrom="opacity-100"
                              leaveTo="opacity-0"
                              afterLeave={() => setUserSearch('')}
                            >
                              <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm z-10">
                                {
                                  userList?.length === 0 ? (
                                    <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                      Nothing found.
                                    </div>
                                  ) : (
                                    userList?.map(user => (
                                      <Combobox.Option
                                        key={user.id}
                                        className={({ active }) =>
                                          `relative cursor-pointer select-none py-2 pl-4 pr-4 ${active ? 'bg-gray-200 text-gray-800' : 'text-gray-900'
                                          }`
                                        }
                                        value={user}
                                      >
                                        <span className={`block truncate font-normal`}>
                                          {user.username}
                                        </span>
                                      </Combobox.Option>
                                    ))
                                  )
                                }
                              </Combobox.Options>
                            </Transition>
                          </div>
                        </Combobox>
                      </div>
                    </div>
                    <div className="grid grid-cols-6 gap-4">
                      <div className="col-span-2"></div>
                      <div className="col-span-4">
                        {
                          userSelectedValid ? null : (
                            <p className="mt-1 text-xs text-red-600">
                              <span>
                                Please select a user.
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
                            <span className="leading-6 ">Role</span>
                            <span>:</span>
                          </div>
                        </label>
                      </div>
                      <div className="col-span-4">
                        <Listbox
                          value={addNamespaceRoleRole}
                          onChange={(source: string) => {
                            setAddNamespaceRoleRole(source);
                          }}>
                          <div className="relative w-full">
                            <Listbox.Button
                              className={() => {
                                let cursor = ''
                                if ((searchParams.get('code_repository_stick') || '') === 'true') {
                                  cursor = 'cursor-not-allowed ';
                                } else {
                                  cursor = 'cursor-pointer ';
                                }
                                return cursor + "relative w-full rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm"
                              }}
                            >
                              <span className="block truncate">{addNamespaceRoleRole}</span>
                              <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                                <ChevronUpDownIcon
                                  className="h-5 w-5 text-gray-400"
                                  aria-hidden="true"
                                />
                              </span>
                            </Listbox.Button>
                            <Listbox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm z-10">
                              <Transition
                                leave="transition ease-in duration-100"
                                leaveFrom="opacity-100"
                                leaveTo="opacity-0"
                              >
                                {
                                  namespaceRoles.map(source => (
                                    <Listbox.Option key={source.name} value={source.name} className={({ active }) =>
                                      `relative cursor-pointer select-none py-2 pl-10 pr-4 ${active ? 'bg-gray-100 text-gray-900' : 'text-gray-900'
                                      }`
                                    }>
                                      {({ selected }) => (
                                        <>
                                          <span
                                            className={`block truncate ${selected ? 'font-medium' : 'font-normal'
                                              }`}
                                          >
                                            {source.name}
                                          </span>
                                          {
                                            selected ? (
                                              <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-gray-600">
                                                <CheckIcon className="h-5 w-5" aria-hidden="true" />
                                              </span>
                                            ) : null
                                          }
                                        </>
                                      )}
                                    </Listbox.Option>
                                  ))
                                }
                              </Transition>
                            </Listbox.Options>
                          </div>
                        </Listbox>
                      </div>
                    </div>
                    <div className="flex flex-row-reverse mt-5">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={e => addMember()}
                      >
                        Add
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={e => { setCreateUserNamespaceModal(false) }}
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
  );
}

function TableItem({ localServer, index, namespace, userSelectedArg, member, setRefresh }: { localServer: string, index: number, namespace: INamespaceItem, userSelectedArg: IUserItem, member: INamespaceMemberItem, setRefresh: (param: any) => void }) {
  const [updateUserNamespaceModal, setUpdateUserNamespaceModal] = useState(false);
  const [userSelected] = useState<IUserItem>(userSelectedArg);
  const [addNamespaceRoleRole, setAddNamespaceRoleRole] = useState(member.role);

  const deleteMember = () => {
    let url = `${localServer}/api/v1/namespaces/${namespace.id}/members/?user_id=${userSelected.id}`;
    axios.delete(url).then(response => {
      if (response?.status === 204) {
        Toast.success("Delete member success");
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

  const updateMember = () => {
    let url = `${localServer}/api/v1/namespaces/${namespace.id}/members/`;
    axios.put(url, {
      user_id: userSelected.id,
      role: addNamespaceRoleRole,
    }).then(response => {
      if (response?.status === 204) {
        Toast.success("Update member success");
        setUpdateUserNamespaceModal(false);
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
      <td className="px-6 py-4 w-5/6 whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer" >
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {member.username}
            </span>
          </div>
        </div>
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {member.role.substring(9)}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {member.updated_at}
      </td>
      <td className="pr-3 whitespace-nowrap" onClick={e => {
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
                    onClick={e => setUpdateUserNamespaceModal(true)}
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
                    onClick={e => deleteMember()}
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
        <Transition.Root show={updateUserNamespaceModal} as={Fragment}>
          <Dialog as="div" className="relative z-10" onClose={setUpdateUserNamespaceModal}>
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
                      Update member
                    </Dialog.Title>
                    <div className="flex flex-col gap-0 mt-4">
                      {/* <div className="grid grid-cols-6 gap-4">
                        <div className="col-span-2 flex flex-row">
                          <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                            <div className="flex">
                              <span className="text-red-600">*</span>
                              <span className="leading-6 ">User</span>
                              <span>:</span>
                            </div>
                          </label>
                        </div>
                        <div className="col-span-4">
                          <Combobox value={userSelected}
                            onChange={(user: IUserItem) => {
                              setUserSelected(user);
                            }}>
                            <div className="relative mt-1">
                              <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                                <Combobox.Input
                                  id="namespace"
                                  className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                                  displayValue={(user: IUserItem) => user.username}
                                  onChange={event => {
                                    setUserSearch(event.target.value);
                                  }}
                                />
                                <Combobox.Button className="absolute inset-y-0 right-0 flex items-center pr-2">
                                  <ChevronUpDownIcon
                                    className="h-5 w-5 text-gray-400"
                                    aria-hidden="true"
                                  />
                                </Combobox.Button>
                              </div>
                              <Transition
                                as={Fragment}
                                leave="transition ease-in duration-100"
                                leaveFrom="opacity-100"
                                leaveTo="opacity-0"
                                afterLeave={() => setUserSearch('')}
                              >
                                <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm z-10">
                                  {
                                    userList.length === 0 ? (
                                      <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                        Nothing found.
                                      </div>
                                    ) : (
                                      userList.map(user => (
                                        <Combobox.Option
                                          key={user.id}
                                          className={({ active }) =>
                                            `relative cursor-pointer select-none py-2 pl-4 pr-4 ${active ? 'bg-gray-200 text-gray-800' : 'text-gray-900'
                                            }`
                                          }
                                          value={user}
                                        >
                                          <span className={`block truncate font-normal`}>
                                            {user.username}
                                          </span>
                                        </Combobox.Option>
                                      ))
                                    )
                                  }
                                </Combobox.Options>
                              </Transition>
                            </div>
                          </Combobox>
                        </div>
                      </div> */}
                      {/* <div className="grid grid-cols-6 gap-4">
                        <div className="col-span-2"></div>
                        <div className="col-span-4">
                          {
                            userSelectedValid ? null : (
                              <p className="mt-1 text-xs text-red-600">
                                <span>
                                  Please select a user.
                                </span>
                              </p>
                            )
                          }
                        </div>
                      </div> */}
                      <div className="grid grid-cols-6 gap-4">
                        <div className="col-span-2 flex flex-row">
                          <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                            <div className="flex">
                              <span className="text-red-600">*</span>
                              <span className="leading-6 ">Role</span>
                              <span>:</span>
                            </div>
                          </label>
                        </div>
                        <div className="col-span-4">
                          <Listbox
                            value={addNamespaceRoleRole}
                            onChange={(source: string) => {
                              setAddNamespaceRoleRole(source);
                            }}>
                            <div className="relative w-full">
                              <Listbox.Button
                                className={() => {
                                  let cursor = 'cursor-pointer '
                                  return cursor + "relative w-full rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm min-w-[200px]"
                                }}
                              >
                                <span className="block truncate">{addNamespaceRoleRole}</span>
                                <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                                  <ChevronUpDownIcon
                                    className="h-5 w-5 text-gray-400"
                                    aria-hidden="true"
                                  />
                                </span>
                              </Listbox.Button>
                              <Listbox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm z-10">
                                <Transition
                                  leave="transition ease-in duration-100"
                                  leaveFrom="opacity-100"
                                  leaveTo="opacity-0"
                                >
                                  {
                                    namespaceRoles.map(source => (
                                      <Listbox.Option key={source.name} value={source.name} className={({ active }) =>
                                        `relative cursor-pointer select-none py-2 pl-10 pr-4 ${active ? 'bg-gray-100 text-gray-900' : 'text-gray-900'
                                        }`
                                      }>
                                        {({ selected }) => (
                                          <>
                                            <span
                                              className={`block truncate ${selected ? 'font-medium' : 'font-normal'
                                                }`}
                                            >
                                              {source.name}
                                            </span>
                                            {
                                              selected ? (
                                                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-gray-600">
                                                  <CheckIcon className="h-5 w-5" aria-hidden="true" />
                                                </span>
                                              ) : null
                                            }
                                          </>
                                        )}
                                      </Listbox.Option>
                                    ))
                                  }
                                </Transition>
                              </Listbox.Options>
                            </div>
                          </Listbox>
                        </div>
                      </div>
                      <div className="flex flex-row-reverse mt-5">
                        <button
                          type="button"
                          className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                          onClick={e => updateMember()}
                        >
                          Update
                        </button>
                        <button
                          type="button"
                          className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                          onClick={e => { setUpdateUserNamespaceModal(false) }}
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
      </td>
    </tr>
  )
}
