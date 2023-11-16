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

import dayjs from "dayjs";
import axios from "axios";
import { useParams, Link } from 'react-router-dom';
import { Fragment, useEffect, useState } from "react";
import { Dialog, Transition, Listbox } from "@headlessui/react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { ChevronUpDownIcon, CheckIcon } from '@heroicons/react/20/solid';

import Regex from "../../utils/regex";
import Settings from "../../Settings";
import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Toast from "../../components/Notification";
import Pagination from "../../components/Pagination";
import OrderHeader from "../../components/OrderHeader";
import QuotaSimple from "../../components/QuotaSimple";
import { IHTTPError, IOrder, IUserItem, IUserList } from "../../interfaces";

const supportRoles = [
  { id: 1, name: 'Admin' },
  { id: 2, name: 'User' },
];

const supportStatus = [
  { id: 1, name: 'Active' },
  { id: 2, name: 'Inactive' },
];

export default function ({ localServer }: { localServer: string }) {
  const [searchUsername, setSearchUsername] = useState("");
  const [createUserModal, setCreateUserModal] = useState(false);

  const [usernameText, setUsernameText] = useState("");
  const [usernameTextValid, setUsernameTextValid] = useState(true);
  useEffect(() => {
    if (usernameText.length > 0) {
      setUsernameTextValid(Regex.Username.test(usernameText))
    }
  }, [usernameText]);
  const [passwordText, setPasswordText] = useState("");
  const [passwordTextValid, setPasswordTextValid] = useState(true);
  useEffect(() => {
    if (passwordText.length > 0) {
      axios.post(localServer + `/api/v1/validators/password`, {
        password: passwordText,
      }).then(response => {
        if (response?.status === 204) {
          setPasswordTextValid(true);
        } else {
          setPasswordTextValid(false);
        }
      }).catch(error => {
        console.log(error);
        setPasswordTextValid(false);
      });
    }
  }, [passwordText]);

  const [createdAtOrder, setCreatedAtOrder] = useState(IOrder.None);
  const [lastLoginOrder, setLastLoginOrder] = useState(IOrder.None);
  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");

  const resetOrder = () => {
    setLastLoginOrder(IOrder.None);
    setCreatedAtOrder(IOrder.None);
  }

  const [userList, setUserList] = useState<IUserList>({} as IUserList);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [refresh, setRefresh] = useState({});

  const [namespaceCountLimit, setNamespaceCountLimit] = useState<string | number>(0);
  const [namespaceCountLimitValid, setNamespaceCountLimitValid] = useState(true);
  useEffect(() => { setNamespaceCountLimitValid(Number.isInteger(namespaceCountLimit) && parseInt(namespaceCountLimit.toString()) >= 0) }, [namespaceCountLimit]);
  const [emailInput, setEmailInput] = useState("");
  const [emailInputValid, setEmailInputValid] = useState(true);
  useEffect(() => { if (emailInput.length > 0) { setEmailInputValid(Regex.Email.test(emailInput)); } }, [emailInput]);

  useEffect(() => {
    let url = `${localServer}/api/v1/users/?page=${page}`;
    if (searchUsername !== "") {
      url += `&name=${searchUsername}`;
    }
    if (sortName !== "") {
      url += `&sort=${sortName}&method=${sortOrder.toString()}`
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const r = response.data as IUserList;
        setUserList(r);
        setTotal(r.total);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [refresh]);

  const [role, setRole] = useState("User");

  const createUser = () => {
    if (!(usernameTextValid && passwordTextValid)) {
      Toast({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    axios.post(localServer + `/api/v1/users/`, {
      username: usernameText,
      password: passwordText,
      email: emailInput,
      namespace_limit: namespaceCountLimit,
      role: role,
    }).then(response => {
      if (response?.status === 201) {
        Toast({ level: "success", title: "Success", message: "Create user success" });
        setCreateUserModal(false);
        setRefresh({});
        setUsernameText("");
        setPasswordText("");
        setEmailInput("");
        setNamespaceCountLimit(0);
        setRole("User");
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
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Repositories</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <Menu localServer={localServer} item="users" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title="Setting - Users" />
            <div className="pt-2 pb-2 flex justify-between">
              <div className="pr-2 pl-2 flex gap-2">
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
                      value={searchUsername}
                      onChange={e => { setSearchUsername(e.target.value); }}
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
                {/* <div>
                  <button className="block mt-2 px-2 py-1.5 h-10 text-gray-500 border-0 rounded-md shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
                    </svg>
                  </button>
                </div> */}
              </div>
              <div className="pr-2 pl-2">
                <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                  onClick={() => { setCreateUserModal(true) }}
                >Create</button>
              </div>
            </div>
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className="border-t w-full">
              <table className="min-w-full flex-1">
                <thead>
                  <tr>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Username</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Namespace</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Status</span>
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <OrderHeader text={"Last Login"}
                        orderStatus={lastLoginOrder} setOrder={e => {
                          resetOrder();
                          setLastLoginOrder(e);
                          setSortOrder(e);
                          setSortName("last_login");
                          setRefresh({});
                        }} />
                    </th>
                    <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <OrderHeader text={"Created at"}
                        orderStatus={createdAtOrder} setOrder={e => {
                          resetOrder();
                          setCreatedAtOrder(e);
                          setSortOrder(e);
                          setSortName("created_at");
                          setRefresh({});
                        }} />
                    </th>
                    <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      Action
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-100 max-h-max">
                  {
                    userList.items?.map(userObj => {
                      return (
                        <TableItem key={userObj.id} localServer={localServer} user={userObj} setRefresh={setRefresh} />
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
      <Transition.Root show={createUserModal} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setCreateUserModal}>
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
                  <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900">
                    <span className="text-red-600">*</span>Username
                  </label>
                  <div className="relative mt-2 rounded-md shadow-sm">
                    <input
                      type="text"
                      id="usernameText"
                      name="usernameText"
                      placeholder="username"
                      className={(usernameTextValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                      value={usernameText}
                      onChange={e => {
                        setUsernameText(e.target.value);
                      }}
                    />
                    {
                      usernameTextValid ? null : (
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
                      usernameTextValid ? null : (
                        <span>
                          Not a valid username, you can try 'hello', 'test'.
                        </span>
                      )
                    }
                  </p>
                  <label htmlFor="passwordText" className="block text-sm font-medium text-gray-700">
                    <span className="text-red-600">*</span>Password
                  </label>
                  <div className="relative mt-2 rounded-md shadow-sm">
                    <input
                      type="password"
                      id="passwordText"
                      name="passwordText"
                      placeholder="password"
                      className={(passwordTextValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                      value={passwordText}
                      onChange={e => {
                        setPasswordText(e.target.value);
                      }}
                    />
                    {
                      passwordTextValid ? (
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
                      passwordTextValid ? null : (
                        <span>
                          Password is invalid, please try 'Admin@123'.
                        </span>
                      )
                    }
                  </p>

                  <label htmlFor="emailInput" className="block text-sm font-medium text-gray-700">
                    <span className="text-red-600">*</span>Email
                  </label>
                  <div className="relative mt-2 rounded-md shadow-sm">
                    <input
                      type="text"
                      id="emailInput"
                      name="emailInput"
                      placeholder="email"
                      className={(emailInputValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                      value={emailInput}
                      onChange={e => {
                        setEmailInput(e.target.value);
                      }}
                    />
                    {
                      emailInputValid ? (
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
                      emailInputValid ? null : (
                        <span>
                          Password is invalid, please try 'Admin@123'.
                        </span>
                      )
                    }
                  </p>

                  <label htmlFor="namespace_count_limit" className="block text-sm font-medium text-gray-700">
                    Namespace count limit
                  </label>
                  <div className="relative mt-2 rounded-md shadow-sm">
                    <input
                      type="number"
                      id="namespace_count_limit"
                      name="namespace_count_limit"
                      placeholder="0 means no limit"
                      className={(namespaceCountLimitValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                      value={namespaceCountLimit}
                      onChange={e => setNamespaceCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                    />
                    {
                      namespaceCountLimitValid ? null : (
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
                      namespaceCountLimitValid ? null : (
                        <span>
                          Not a valid namespace count limit, should be non-negative integer.
                        </span>
                      )
                    }
                  </p>

                  <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                    Role
                  </label>
                  <div className="mt-2 flex flex-row items-center h-[36px]">
                    <Listbox
                      value={role}
                      onChange={(source: string) => {
                        setRole(source);
                      }}>
                      <div className="relative mt-1 w-full">
                        <Listbox.Button
                          className="cursor-pointer relative w-full rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm h-[36px]"
                        >
                          <span className="block truncate">{role}</span>
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
                              supportRoles.map(source => (
                                <Listbox.Option key={source.name} value={source.name} className={({ active }) =>
                                  `relative cursor-pointer select-none py-1.5 pl-10 pr-4 ${active ? 'bg-gray-100 text-gray-900' : 'text-gray-900'
                                  }`
                                }>
                                  {({ selected }) => (
                                    <>
                                      <span
                                        className={`block truncate ${selected ? 'font-medium' : 'font-normal'}`}
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

                  <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                    <button
                      type="button"
                      className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                      onClick={() => createUser()}
                    >
                      Create
                    </button>
                    <button
                      type="button"
                      className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                      onClick={() => setCreateUserModal(false)}
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
  );
}

function TableItem({ localServer, user, setRefresh }: { localServer: string, user: IUserItem, setRefresh: (param: any) => void }) {
  const [status, setStatus] = useState(user.status === "" ? "Active" : user.status);
  const [role, setRole] = useState(user.role === "" ? "Normal" : user.role);
  const [usernameText, setUsernameText] = useState(user.username);
  const [usernameTextValid, setUsernameTextValid] = useState(true);
  useEffect(() => {
    if (usernameText.length > 0) {
      setUsernameTextValid(Regex.Username.test(usernameText))
    }
  }, [usernameText]);
  const [passwordText, setPasswordText] = useState("");
  const [passwordTextValid, setPasswordTextValid] = useState(true);
  useEffect(() => {
    if (passwordText.length > 0) {
      axios.post(localServer + `/api/v1/validators/password`, {
        password: passwordText,
      }).then(response => {
        if (response?.status === 204) {
          setPasswordTextValid(true);
        } else {
          setPasswordTextValid(false);
        }
      }).catch(error => {
        console.log(error);
        setPasswordTextValid(false);
      });
    }
  }, [passwordText]);

  const [namespaceCountLimit, setNamespaceCountLimit] = useState<string | number>(user.namespace_limit);
  const [namespaceCountLimitValid, setNamespaceCountLimitValid] = useState(true);
  useEffect(() => { setNamespaceCountLimitValid(Number.isInteger(namespaceCountLimit) && parseInt(namespaceCountLimit.toString()) >= 0) }, [namespaceCountLimit]);
  const [emailInput, setEmailInput] = useState(user.email);
  const [emailInputValid, setEmailInputValid] = useState(true);
  useEffect(() => { if (emailInput.length > 0) { setEmailInputValid(Regex.Email.test(emailInput)); } }, [emailInput]);

  const [updateUserModal, setUpdateUserModal] = useState(false);

  const updateUser = () => {
    const data: { [key: string]: any } = {
      email: emailInput,
      username: usernameText,
      status: status,
      namespace_limit: namespaceCountLimit,
    };
    if (passwordText.length != 0) {
      data["password"] = passwordText;
    }
    axios.put(localServer + `/api/v1/users/${user.id}`, data).then(response => {
      if (response?.status === 204) {
        Toast({ level: "success", title: "Success", message: "Update user info success" });
        setUpdateUserModal(false);
        setRefresh({});
        setPasswordText("");
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
    <tr className="align-middle">
      <td className="px-6 py-4 w-5/6 whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer"
      // onClick={() => {
      //   navigate(`/namespaces/${namespace.name}/repositories`);
      // }}
      >
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {user.username}
              <span className="text-gray-500 font-normal ml-4">{user.email}</span>
            </span>
          </div>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        <QuotaSimple current={user.namespace_count} limit={user.namespace_limit} />
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {user.status}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs().to(dayjs(user.last_login))}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs().to(dayjs(user.created_at))}
      </td>
      <td className="text-center px-4 py-4 whitespace-nowrap text-sm text-gray-500 cursor-pointer hover:text-gray-700"
        onClick={e => {
          setUpdateUserModal(true);
        }}>
        Update
      </td>
      <td>
        <Transition.Root show={updateUserModal} as={Fragment}>
          <Dialog as="div" className="relative z-10" onClose={setUpdateUserModal}>
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
                    <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900">
                      <span className="text-red-600">*</span>Username
                    </label>
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <input
                        type="text"
                        id="usernameText"
                        name="usernameText"
                        placeholder="username"
                        className={(usernameTextValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                        value={usernameText}
                        onChange={e => {
                          setUsernameText(e.target.value);
                        }}
                      />
                      {
                        usernameTextValid ? null : (
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
                        usernameTextValid ? null : (
                          <span>
                            Not a valid username, you can try 'hello', 'test'.
                          </span>
                        )
                      }
                    </p>
                    <label htmlFor="passwordText" className="block text-sm font-medium text-gray-700">
                      Password
                    </label>
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <input
                        type="password"
                        id="passwordText"
                        name="passwordText"
                        placeholder="password"
                        className={(passwordTextValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                        value={passwordText}
                        onChange={e => {
                          setPasswordText(e.target.value);
                        }}
                      />
                      {
                        passwordTextValid ? (
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
                        passwordTextValid ? null : (
                          <span>
                            Not a valid description, max 50 characters.
                          </span>
                        )
                      }
                    </p>

                    <label htmlFor="emailInput" className="block text-sm font-medium text-gray-700">
                      <span className="text-red-600">*</span>Email
                    </label>
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <input
                        type="text"
                        id="emailInput"
                        name="emailInput"
                        placeholder="email"
                        className={(emailInputValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                        value={emailInput}
                        onChange={e => {
                          setEmailInput(e.target.value);
                        }}
                      />
                      {
                        emailInputValid ? null : (
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
                        emailInputValid ? null : (
                          <span>
                            Not a valid description, max 50 characters.
                          </span>
                        )
                      }
                    </p>

                    <label htmlFor="namespace_count_limit" className="block text-sm font-medium text-gray-700">
                      Namespace count limit
                    </label>
                    <div className="relative mt-2 rounded-md shadow-sm">
                      <input
                        type="number"
                        id="namespace_count_limit"
                        name="namespace_count_limit"
                        placeholder="0 means no limit"
                        className={(namespaceCountLimitValid ? "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6" : "block w-full rounded-md border-0 py-1.5 pr-10 text-red-900 ring-1 ring-inset ring-red-300 placeholder:text-red-300 focus:ring-2 focus:ring-inset focus:ring-red-500 sm:text-sm sm:leading-6")}
                        value={namespaceCountLimit}
                        onChange={e => setNamespaceCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                      />
                      {
                        namespaceCountLimitValid ? null : (
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
                        namespaceCountLimitValid ? null : (
                          <span>
                            Not a valid namespace count limit, should be non-negative integer.
                          </span>
                        )
                      }
                    </p>

                    <div className="grid grid-cols-2 gap-4">
                      <div className="col-span-1">
                        <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                          Role
                        </label>
                        <div className="mt-2 flex flex-row items-center h-[36px]">
                          <Listbox
                            value={role}
                            onChange={(source: string) => {
                              setRole(source);
                            }}>
                            <div className="relative mt-1 w-full">
                              <Listbox.Button
                                className="cursor-pointer relative w-full rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm h-[36px]"
                              >
                                <span className="block truncate">{role}</span>
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
                                    supportRoles.map(source => (
                                      <Listbox.Option key={source.name} value={source.name} className={({ active }) =>
                                        `relative cursor-pointer select-none py-1.5 pl-10 pr-4 ${active ? 'bg-gray-100 text-gray-900' : 'text-gray-900'
                                        }`
                                      }>
                                        {({ selected }) => (
                                          <>
                                            <span
                                              className={`block truncate ${selected ? 'font-medium' : 'font-normal'}`}
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
                      <div className="col-span-1">
                        <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                          Status
                        </label>
                        <div className="mt-2 flex flex-row items-center h-[36px]">
                          <Listbox
                            value={status}
                            onChange={(source: string) => {
                              setStatus(source);
                            }}>
                            <div className="relative mt-1 w-full">
                              <Listbox.Button
                                className="cursor-pointer relative w-full rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm h-[36px]"
                              >
                                <span className="block truncate">{status}</span>
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
                                    supportStatus.map(source => (
                                      <Listbox.Option key={source.name} value={source.name} className={({ active }) =>
                                        `relative cursor-pointer select-none py-1.5 pl-10 pr-4 ${active ? 'bg-gray-100 text-gray-900' : 'text-gray-900'
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
                    </div>

                    <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                      <button
                        type="button"
                        className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                        onClick={() => updateUser()}
                      >
                        Update
                      </button>
                      <button
                        type="button"
                        className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                        onClick={() => setUpdateUserModal(false)}
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
