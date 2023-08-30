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

import _ from 'lodash';
import axios from "axios";
import { Listbox, Transition, Combobox } from '@headlessui/react';
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Fragment, useEffect, useRef, useState } from 'react';
import { ChevronUpDownIcon, CheckIcon } from '@heroicons/react/20/solid';
import { Link, useSearchParams, useParams, useNavigate } from 'react-router-dom';

import Settings from '../../Settings';
import Header from '../../components/Header';
import HeaderMenu from '../../components/Menu';
import Toast from "../../components/Notification";

import { IHTTPError, INamespaceList, INamespace, IRepository, IRepositoryList, ICodeRepositoryOwnerItem, ICodeRepositoryItem, ICodeRepositoryProviderItem, ICodeRepositoryProviderList, ICodeRepositoryList, ICodeRepositoryOwnerList, ICodeRepositoryBranchItem, ICodeRepositoryBranchList } from '../../interfaces';

export default function ({ localServer }: { localServer: string }) {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  useEffect(() => {
    navigate(`?${searchParams.toString()}`);
  }, [searchParams]);

  const [namespaceSearch, setNamespaceSearch] = useState('');
  const [namespaceList, setNamespaceList] = useState<INamespace[]>();
  const [namespaceSelected, setNamespaceSelected] = useState<INamespace>({
    name: searchParams.get('namespace') || "",
    id: parseInt(searchParams.get('namespace_id') || "") || 0,
  } as INamespace);

  useEffect(() => {
    let url = localServer + `/api/v1/namespaces/?limit=5`;
    if (namespaceSearch != null && namespaceSearch !== "") {
      url += `&name=${namespaceSearch}`;
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const namespaceList = response.data as INamespaceList;
        setNamespaceList(namespaceList.items);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespaceSearch]);

  const [repositorySearch, setRepositorySearch] = useState('');
  const [repositoryList, setRepositoryList] = useState<IRepository[]>();
  const [repositorySelected, setRepositorySelected] = useState<IRepository>({
    name: searchParams.get('repository') || "",
    id: parseInt(searchParams.get('repository_id') || "") || 0,
  } as IRepository);

  useEffect(() => {
    if (namespaceSelected.name == undefined || namespaceSelected.name.length == 0) {
      return;
    }
    let url = localServer + `/api/v1/namespaces/${namespaceSelected.name}/repositories/?limit=5`;
    if (repositorySearch != null && repositorySearch !== "") {
      url += `&name=${repositorySearch}`;
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const repositoryList = response.data as IRepositoryList;
        setRepositoryList(repositoryList.items);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespaceSelected, repositorySearch]);

  const [codeRepositoryProviderList, setCodeRepositoryProviderList] = useState<ICodeRepositoryProviderItem[]>();
  const [codeRepositoryProviderSelected, setCodeRepositoryProviderSelected] = useState<ICodeRepositoryProviderItem>({
    provider: searchParams.get('provider') || "",
  } as ICodeRepositoryProviderItem);

  useEffect(() => {
    let url = localServer + `/api/v1/coderepos/providers`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const providerList = response.data as ICodeRepositoryProviderList;
        setCodeRepositoryProviderList(providerList.items);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, []);

  const [codeRepositoryOwnerSearch, setCodeRepositoryOwnerSearch] = useState('');
  const [codeRepositoryOwnerList, setCodeRepositoryOwnerList] = useState<ICodeRepositoryOwnerItem[]>();
  const [codeRepositoryOwnerFilteredList, setCodeRepositoryOwnerFilteredList] = useState<ICodeRepositoryOwnerItem[]>();
  const [codeRepositoryOwnerSelected, setCodeRepositoryOwnerSelected] = useState<ICodeRepositoryOwnerItem>({
    owner: searchParams.get('code_repository_owner') || '',
    id: parseInt(searchParams.get('code_repository_owner_id') || "") || 0,
  } as ICodeRepositoryOwnerItem);

  useEffect(() => {
    if (codeRepositoryProviderSelected.provider == undefined || codeRepositoryProviderSelected.provider == "") {
      return;
    }
    axios.get(`${localServer}/api/v1/coderepos/${codeRepositoryProviderSelected.provider}/owners`).then(response => {
      if (response.status == 200) {
        const data = response.data as ICodeRepositoryOwnerList;
        setCodeRepositoryOwnerList(_.orderBy(data.items, ['is_org']));
        setCodeRepositoryOwnerFilteredList(_.orderBy(data.items, ['is_org']));
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [codeRepositoryProviderSelected]);

  useEffect(() => {
    if (codeRepositoryOwnerList?.length == undefined || codeRepositoryOwnerList?.length == 0) {
      return;
    }
    if (codeRepositoryOwnerSearch == '') {
      setCodeRepositoryOwnerFilteredList(codeRepositoryOwnerList);
      return;
    }
    let result = [];
    for (let i = 0; i < (codeRepositoryOwnerList?.length || 0); i++) {
      if (codeRepositoryOwnerList[i].owner.toLowerCase().includes(codeRepositoryOwnerSearch.toLocaleLowerCase())) {
        result.push(codeRepositoryOwnerList[i])
      }
    }
    setCodeRepositoryOwnerFilteredList(result);
  }, [codeRepositoryOwnerSearch])

  const [codeRepositorySearch, setCodeRepositorySearch] = useState('');
  const [codeRepositoryList, setCodeRepositoryList] = useState<ICodeRepositoryItem[]>();
  const [codeRepositorySelected, setCodeRepositorySelected] = useState<ICodeRepositoryItem>({
    name: searchParams.get('code_repository_name') || '',
    id: parseInt(searchParams.get('code_repository_id') || "") || 0,
  } as ICodeRepositoryItem);

  useEffect(() => {
    if (codeRepositoryProviderSelected.provider == undefined || codeRepositoryProviderSelected.provider == "") {
      return;
    }
    if (codeRepositoryOwnerSelected.owner == undefined || codeRepositoryOwnerSelected.owner == "") {
      return;
    }
    let url = `${localServer}/api/v1/coderepos/${codeRepositoryProviderSelected.provider}?owner=${codeRepositoryOwnerSelected.owner}&limit=5`;
    if (codeRepositorySearch != null && codeRepositorySearch !== "") {
      url += `&name=${codeRepositorySearch}`;
    }
    axios.get(url).then(response => {
      if (response.status == 200) {
        const data = response.data as ICodeRepositoryList;
        setCodeRepositoryList(data.items);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [codeRepositoryOwnerSelected, codeRepositoryProviderSelected, codeRepositorySearch]);

  const [codeRepositoryBranchSearch, setCodeRepositoryBranchSearch] = useState('');
  const [codeRepositoryBranchList, setCodeRepositoryBranchList] = useState<ICodeRepositoryBranchItem[]>();
  const [codeRepositoryBranchFilteredList, setCodeRepositoryBranchFilteredList] = useState<ICodeRepositoryBranchItem[]>();
  const [codeRepositoryBranchSelected, setCodeRepositoryBranchSelected] = useState<ICodeRepositoryBranchItem>({
    name: searchParams.get('code_repository_branch_name') || '',
    id: parseInt(searchParams.get('code_repository_branch_id') || "") || 0,
  } as ICodeRepositoryBranchItem);

  useEffect(() => {
    if (codeRepositorySelected.id == undefined || codeRepositorySelected.id == 0) {
      return;
    }
    let url = `${localServer}/api/v1/coderepos/${codeRepositorySelected.id}/branches`;
    axios.get(url).then(response => {
      if (response.status == 200) {
        const data = response.data as ICodeRepositoryBranchList;
        setCodeRepositoryBranchList(data.items);
        setCodeRepositoryBranchFilteredList(data.items);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [codeRepositorySelected]);

  useEffect(() => {
    if (codeRepositoryBranchList?.length == undefined || codeRepositoryBranchList?.length == 0) {
      return;
    }
    if (codeRepositoryBranchSearch == '') {
      setCodeRepositoryBranchFilteredList(codeRepositoryBranchList);
    }
    let result = [];
    for (let i = 0; i < (codeRepositoryBranchList?.length || 0); i++) {
      if (codeRepositoryBranchList[i].name.toLowerCase().includes(codeRepositoryBranchSearch.toLocaleLowerCase())) {
        result.push(codeRepositoryBranchList[i])
      }
    }
    setCodeRepositoryBranchFilteredList(result);
  }, [codeRepositoryBranchSearch]);

  const [submodule, setSubmodule] = useState(false);
  const [depth, setDepth] = useState<string | number>(0);

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Code Repositories</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white min-w-[1600px]">
        <HeaderMenu localServer={localServer} item="coderepos" />
        <div className="flex flex-col flex-1 max-h-screen">
          {/* part 1 begin */}
          <main className="relative focus:outline-none" tabIndex={0}>
            <Header title="Setup builder" />
          </main>
          {/* part 1 end */}
          {/* part 2 begin */}
          <div className="flex-1 flex flex-col w-full">
            <div className="py-6 px-8 border-gray-200 border-b w-full">
              <h2 className="text-base font-semibold leading-7 text-gray-900">OCI Repository</h2>
              <p className="mt-1 text-sm leading-6 text-gray-600">Builder will push the artifact to this repository.</p>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                <div className="sm:col-span-1">
                  <label htmlFor="namespace" className="block text-sm font-medium leading-6 text-gray-900">
                    Namespace
                  </label>
                  <div className="mt-2">
                    <Combobox value={namespaceSelected} onChange={(namespace: INamespace) => {
                      setSearchParams({
                        ...Object.fromEntries(searchParams.entries()),
                        namespace: namespace.name,
                        namespace_id: namespace.id.toString(),
                        repository: '',
                        repository_id: '',
                      });
                      setRepositorySelected({} as IRepository); // clear the repo selected
                      setNamespaceSelected(namespace);
                    }}>
                      <div className="relative mt-1">
                        <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow-md focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                          <Combobox.Input
                            id="namespace"
                            className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                            displayValue={(namespace: INamespace) => namespace.name}
                            onChange={(event) => {
                              setNamespaceSearch(event.target.value);
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
                          afterLeave={() => setNamespaceSearch('')}
                        >
                          <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                            {
                              namespaceList?.length === 0 ? (
                                <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                  Nothing found.
                                </div>
                              ) : (
                                namespaceList?.map(namespace => (
                                  <Combobox.Option
                                    key={namespace.id}
                                    className={({ active }) =>
                                      `relative cursor-pointer select-none py-2 pl-4 pr-4 ${active ? 'bg-gray-200 text-gray-800' : 'text-gray-900'
                                      }`
                                    }
                                    value={namespace}
                                  >
                                    <span className={`block truncate  font-normal`}>
                                      {namespace.name}
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
                <div className="sm:col-span-1">
                  <label htmlFor="repository" className="block text-sm font-medium leading-6 text-gray-900">
                    Repository
                  </label>
                  <div className="mt-2">
                    <Combobox value={repositorySelected} onChange={(repo: IRepository) => {
                      setSearchParams({
                        ...Object.fromEntries(searchParams.entries()),
                        repository: repo.name,
                        repository_id: repo.id.toString(),
                      });
                      setRepositorySelected(repo);
                    }}>
                      <div className="relative mt-1">
                        <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow-md focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                          <Combobox.Input
                            id="repository"
                            className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                            displayValue={(repository: IRepository) => {
                              if (namespaceSelected.name != undefined && repository.name != undefined) {
                                return repository.name.substring(namespaceSelected.name.length + 1)
                              }
                              return "";
                            }}
                            onChange={(event) => {
                              setRepositorySearch(event.target.value);
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
                          afterLeave={() => setRepositorySearch('')}
                        >
                          <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                            {
                              repositoryList?.length === 0 ? (
                                <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                  Nothing found.
                                </div>
                              ) : (
                                repositoryList?.map(repository => (
                                  <Combobox.Option
                                    key={repository.id}
                                    className={({ active }) =>
                                      `relative cursor-pointer select-none py-2 pl-4 pr-4 ${active ? 'bg-gray-200 text-gray-800' : 'text-gray-900'
                                      }`
                                    }
                                    value={repository}
                                  >
                                    <span className={`block truncate font-normal`}>
                                      {repository.name.substring(namespaceSelected.name.length + 1)}
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
              </div>
            </div>
            <div className="py-6 px-8 border-gray-200 border-b w-full">
              <h2 className="text-base font-semibold leading-7 text-gray-900">Code Repository</h2>
              <p className="mt-1 text-sm leading-6 text-gray-600">Builder will clone source code from here.</p>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                <div className="sm:col-span-1">
                  <label htmlFor="codeProviders" className="block text-sm font-medium leading-6 text-gray-900">
                    Provider
                  </label>
                  <div className="mt-2">
                    <Combobox value={codeRepositoryProviderSelected} onChange={(provider: ICodeRepositoryProviderItem) => {
                      setSearchParams({
                        ...Object.fromEntries(searchParams.entries()),
                        provider: provider.provider,
                      });
                      setCodeRepositoryOwnerSelected({} as ICodeRepositoryOwnerItem); // clear the selected
                      setCodeRepositorySelected({} as ICodeRepositoryItem);
                      setCodeRepositoryBranchSelected({} as ICodeRepositoryBranchItem);
                      setCodeRepositoryProviderSelected(provider);
                    }}>
                      <div className="relative mt-1">
                        <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow-md focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                          <Combobox.Input
                            id="codeProviders"
                            className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                            displayValue={(provider: ICodeRepositoryProviderItem) => provider.provider}
                            onChange={(event) => { }}
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
                        >
                          <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                            {
                              codeRepositoryProviderList?.length === 0 ? (
                                <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                  Nothing found.
                                </div>
                              ) : (
                                codeRepositoryProviderList?.map(provider => (
                                  <Combobox.Option
                                    key={provider.provider}
                                    className={({ active }) =>
                                      `relative cursor-pointer select-none py-2 pl-4 pr-4 ${active ? 'bg-gray-200 text-gray-800' : 'text-gray-900'
                                      }`
                                    }
                                    value={provider}
                                  >
                                    <span className={`block truncate  font-normal`}>
                                      {provider.provider}
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
                <div className="sm:col-span-1">
                  <label htmlFor="codeOwners" className="block text-sm font-medium leading-6 text-gray-900">
                    Organization
                  </label>
                  <div className="mt-2">
                    <Combobox value={codeRepositoryOwnerSelected} onChange={(owner: ICodeRepositoryOwnerItem) => {
                      setSearchParams({
                        ...Object.fromEntries(searchParams.entries()),
                        code_repository_owner: owner.owner,
                        code_repository_owner_id: owner.id.toString(),
                      });
                      setCodeRepositorySelected({} as ICodeRepositoryItem);
                      setCodeRepositoryBranchSelected({} as ICodeRepositoryBranchItem);
                      setCodeRepositoryOwnerSelected(owner);
                    }}>
                      <div className="relative mt-1">
                        <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow-md focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                          <Combobox.Input
                            id="codeOwners"
                            className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                            displayValue={(owner: ICodeRepositoryOwnerItem) => owner.owner}
                            onChange={(event) => {
                              setCodeRepositoryOwnerSearch(event.target.value);
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
                          afterLeave={() => setCodeRepositoryOwnerSearch('')}
                        >
                          <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                            {
                              codeRepositoryOwnerFilteredList?.length === 0 ? (
                                <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                  Nothing found.
                                </div>
                              ) : (
                                codeRepositoryOwnerFilteredList?.map(owner => (
                                  <Combobox.Option
                                    key={owner.id}
                                    className={({ active }) =>
                                      `relative cursor-pointer select-none py-2 pl-4 pr-4 ${active ? 'bg-gray-200 text-gray-800' : 'text-gray-900'
                                      }`
                                    }
                                    value={owner}
                                  >
                                    <span className={`block truncate  font-normal`}>
                                      {owner.owner}
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
                <div className="sm:col-span-2">
                  <label htmlFor="coderepos" className="block text-sm font-medium leading-6 text-gray-900">
                    Repository
                  </label>
                  <div className="mt-2">
                    <Combobox value={codeRepositorySelected} onChange={(cr: ICodeRepositoryItem) => {
                      setSearchParams({
                        ...Object.fromEntries(searchParams.entries()),
                        code_repository_name: cr.name,
                        code_repository_id: cr.id.toString(),
                      });
                      setCodeRepositoryBranchSelected({} as ICodeRepositoryBranchItem);
                      setCodeRepositorySelected(cr);
                    }}>
                      <div className="relative mt-1">
                        <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow-md focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                          <Combobox.Input
                            id="coderepos"
                            className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                            displayValue={(cr: ICodeRepositoryItem) => cr.name}
                            onChange={(event) => {
                              setCodeRepositorySearch(event.target.value);
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
                          afterLeave={() => setCodeRepositorySearch('')}
                        >
                          <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                            {
                              codeRepositoryList?.length === 0 ? (
                                <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                  Nothing found.
                                </div>
                              ) : (
                                codeRepositoryList?.map(cr => (
                                  <Combobox.Option
                                    key={cr.id}
                                    className={({ active }) =>
                                      `relative cursor-pointer select-none py-2 pl-4 pr-4 ${active ? 'bg-gray-200 text-gray-800' : 'text-gray-900'
                                      }`
                                    }
                                    value={cr}
                                  >
                                    <span className={`block truncate  font-normal`}>
                                      {cr.name} <span className='text-gray-400'>{cr.clone_url}</span>
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
                <div className="sm:col-span-1">
                  <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                    Branch
                  </label>
                  <div className="mt-2">
                    <Combobox value={codeRepositoryBranchSelected} onChange={(branch: ICodeRepositoryBranchItem) => {
                      setSearchParams({
                        ...Object.fromEntries(searchParams.entries()),
                        code_repository_branch_name: branch.name,
                        code_repository_branch_id: branch.id.toString(),
                      });
                      setCodeRepositoryBranchSelected(branch);
                    }}>
                      <div className="relative mt-1">
                        <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow-md focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                          <Combobox.Input
                            id="branch"
                            className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                            displayValue={(branch: ICodeRepositoryBranchItem) => branch.name}
                            onChange={(event) => {
                              setCodeRepositoryBranchSearch(event.target.value);
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
                          afterLeave={() => setCodeRepositoryBranchSearch('')}
                        >
                          <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                            {
                              codeRepositoryBranchFilteredList?.length === 0 ? (
                                <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                  Nothing found.
                                </div>
                              ) : (
                                codeRepositoryBranchFilteredList?.map(branch => (
                                  <Combobox.Option
                                    key={branch.id}
                                    className={({ active }) =>
                                      `relative cursor-pointer select-none py-2 pl-4 pr-4 ${active ? 'bg-gray-200 text-gray-800' : 'text-gray-900'
                                      }`
                                    }
                                    value={branch}
                                  >
                                    <span className={`block truncate  font-normal`}>
                                      {branch.name}
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
              </div>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                <div className="sm:col-span-1">
                  <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                    Clone Submodule
                  </label>
                  <div className="mt-2 flex flex-row items-center h-[36px]">
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" checked={submodule} className="sr-only peer" onChange={e => { setSubmodule(e.target.checked); }} />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                    </label>
                  </div>
                </div>
                <div className="sm:col-span-1">
                  <label htmlFor="depth" className="block text-sm font-medium leading-6 text-gray-900">
                    Clone Depth
                  </label>
                  <div className="mt-2 h-[36px]">
                    <input
                      type="number"
                      name="depth"
                      id="depth"
                      className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                      value={depth}
                      onChange={e => setDepth(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))}
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
          {/* part 2 end */}
          {/* part 3 begin */}
          <div style={{ marginTop: "auto" }}>
            <div
              className="flex flex-2 items-center justify-between border-gray-200 px-4 py-3 sm:px-6 border-t-0 bg-slate-100"
              aria-label="Pagination"
            >
              <div>
              </div>
              <div className="flex flex-1 justify-between sm:justify-end">
                <button
                  className="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  className="relative ml-3 inline-flex items-center rounded-md border px-4 py-2 text-sm font-medium text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500"
                >
                  Setup
                </button>
              </div>
            </div>
          </div>
          {/* part 3 end */}
        </div>
      </div>
    </Fragment >
  );
}
