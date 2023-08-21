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

import axios from 'axios';
import { useParams, Link } from 'react-router-dom';
import { Fragment, useEffect, useState } from 'react';
import { Listbox, Transition } from '@headlessui/react';
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { ChevronUpDownIcon } from '@heroicons/react/20/solid';

import Settings from '../../Settings';
import Menu from '../../components/Menu';
import Header from '../../components/Header';
import Pagination from '../../components/Pagination';
import OrderHeader from '../../components/OrderHeader';

import { INamespace, INamespaceList, IHTTPError, IOrder } from '../../interfaces';

const people = [
  { id: 1, name: 'Wade Cooper' },
  { id: 2, name: 'Arlene Mccoy' },
  { id: 3, name: 'Devon Webb' },
  { id: 4, name: 'Tom Cook' },
  { id: 5, name: 'Tanya Fox' },
  { id: 6, name: 'Hellen Schmidt' },
  { id: 7, name: 'Caroline Schultz' },
  { id: 8, name: 'Mason Heaney' },
  { id: 9, name: 'Claudie Smitham' },
  { id: 10, name: 'Emil Schaefer' },
]

function classNames(...classes: Array<string | boolean>) {
  return classes.filter(Boolean).join(' ')
}

export default function ({ localServer }: { localServer: string }) {
  const { provider } = useParams<{ provider: string }>();

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const [searchCodeRepo, setSearchCodeRepo] = useState("");

  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");

  const [organization, setOrganization] = useState(people[3])

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Code Repositories</title>
        </Helmet>
      </HelmetProvider>
      <div id="tooltip-hover" role="tooltip" className="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
        <div>
          Last synced <span className='text-gray-300'>last month</span>
        </div>
        <div>
          Failed, connect failed
        </div>
        <div className="tooltip-arrow" data-popper-arrow></div>
      </div>
      <div className="min-h-screen flex overflow-hidden bg-white min-w-[1600px]">
        <Menu localServer={localServer} item="coderepos" />
        <div className="flex flex-col flex-1 max-h-screen">
          {/* part 1 begin */}
          <main className="relative focus:outline-none" tabIndex={0}>
            <Header title="Code Repository" />
            <div className="pt-2 pb-2 flex justify-between">
              <div className="pr-2 pl-2 flex-1">
                <div className="flex">
                  <Listbox value={organization} onChange={setOrganization}>
                    {({ open }) => (
                      <>
                        <div className="relative mt-2 w-40 mr-2">
                          <label
                            htmlFor="codeRepositorySearch"
                            className="absolute -top-2 left-2 inline-block bg-white px-1 text-xs font-medium text-gray-900"
                          >
                            Organization
                          </label>
                          <Listbox.Button
                            id="codeRepositorySearch"
                            className="block h-10 rounded-md border-0 py-1.5 pr-5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 w-full pl-3 text-left"
                          >
                            <span className="block truncate">{organization.name}</span>
                            <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                              <ChevronUpDownIcon className="h-5 w-5 text-gray-400" aria-hidden="true" />
                            </span>
                          </Listbox.Button>
                          <Transition
                            show={open}
                            as={Fragment}
                            leave="transition ease-in duration-100"
                            leaveFrom="opacity-100"
                            leaveTo="opacity-0"
                          >
                            <Listbox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm z-10">
                              {people.map((person) => (
                                <Listbox.Option
                                  key={person.id}
                                  className={({ active }) =>
                                    classNames(
                                      active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                      'relative cursor-default select-none py-2 pl-3 pr-2'
                                    )
                                  }
                                  value={person}
                                >
                                  {({ selected, active }) => (
                                    <>
                                      <span className={classNames(selected ? 'font-semibold' : 'font-normal', 'block truncate')}>
                                        {person.name}
                                      </span>
                                    </>
                                  )}
                                </Listbox.Option>
                              ))}
                            </Listbox.Options>
                          </Transition>
                        </div>
                      </>
                    )}
                  </Listbox>
                  <div className="relative mt-2 flex items-center">
                    <label
                      htmlFor="codeRepositorySearch"
                      className="absolute -top-2 left-2 inline-block bg-white px-1 text-xs font-medium text-gray-900"
                    >
                      Code Repository
                    </label>
                    <input
                      type="text"
                      id="codeRepositorySearch"
                      placeholder="search code repository"
                      value={searchCodeRepo}
                      onChange={e => { setSearchCodeRepo(e.target.value); }}
                      onKeyDown={e => {
                        // if (e.key == "Enter") {
                        //   fetchNamespace()
                        // }
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
              <div className="flex flex-col">
                <button data-tooltip-target="tooltip-hover" data-tooltip-trigger="hover" type="button" className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                // onClick={() => { setCreateNamespaceModal(true) }}
                >Sync</button>
              </div>
              <div className="pr-2 flex flex-col">
                <button className="my-auto block px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                // onClick={() => { setCreateNamespaceModal(true) }}
                >Pull credential</button>
              </div>
            </div>
          </main>
          {/* part 1 end */}
          {/* part 2 begin */}
          <div>
            <div className="align-middle inline-block min-w-full border-b border-gray-200">
              <table className="min-w-full flex-1">
                <thead>
                  <tr>
                    <th className="sticky top-0 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <span className="lg:pl-2">Name</span>
                    </th>
                    <th className="sticky top-0 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      <OrderHeader text={"Image Repo Count"} orderStatus={sortOrder} setOrder={e => {
                        setSortOrder(e);
                        setSortName("");
                      }} />
                    </th>
                    <th className="sticky top-0 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                      Action
                    </th>
                  </tr>
                </thead>
                {/* <tbody className="bg-white divide-y divide-gray-100 max-h-max">
                  {
                    namespaceList.items?.map((namespace, index) => {
                      return (
                        <TableItem key={namespace.id} index={index} namespace={namespace} localServer={localServer} setRefresh={setRefresh} />
                      );
                    })
                  }
                </tbody> */}
              </table>
            </div>
          </div>
          {/* part 2 end */}
          {/* part 3 begin */}
          <div style={{ marginTop: "auto" }}>
            <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
          </div>
          {/* part 3 end */}
        </div>
      </div>
    </Fragment>
  )
}

function TableItem({ localServer, index }: { localServer: string, index: number }) {
  return (
    <div>

    </div>
  )
}