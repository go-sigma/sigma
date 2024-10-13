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

// import * as monaco from 'monaco-editor';
import _ from 'lodash';
import axios from "axios";
// import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
// import Editor, { loader } from '@monaco-editor/react';
import CodeMirror from '@uiw/react-codemirror';
import { StreamLanguage } from '@codemirror/language';
import { dockerFile as dockerfileModel } from '@codemirror/legacy-modes/mode/dockerfile';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid';
import { Combobox, Listbox, Transition } from '@headlessui/react';
import { Fragment, useEffect, useState } from 'react';
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';

import Header from '../../components/Header';
import HeaderMenu from '../../components/Menu';
import Settings from '../../Settings';
import Toast from "../../components/Notification";
import { IBuilderItem, ICodeRepositoryBranchItem, ICodeRepositoryBranchList, ICodeRepositoryItem, ICodeRepositoryList, ICodeRepositoryOwnerItem, ICodeRepositoryOwnerList, ICodeRepositoryProviderItem, ICodeRepositoryProviderList, IHTTPError, INamespaceItem, INamespaceList, IRepositoryItem, IRepositoryList } from '../../interfaces';

// self.MonacoEnvironment = {
//   getWorker() {
//     return new editorWorker();
//   },
// };

// loader.config({ monaco });

const supportPlatforms = [
  { id: 1, name: 'linux/amd64' },
  { id: 4, name: 'linux/arm64' },
  { id: 5, name: 'linux/riscv64' },
  { id: 6, name: 'linux/ppc64le' },
  { id: 7, name: 'linux/s390x' },
  { id: 8, name: 'linux/386' },
  { id: 9, name: 'linux/mips64le' },
  { id: 10, name: 'linux/mips64' },
  { id: 11, name: 'linux/arm/v7' },
  { id: 12, name: 'linux/arm/v6' },
];

const supportCredential = [
  { id: 1, name: 'none' },
  { id: 2, name: 'username' },
  { id: 3, name: 'token' },
  { id: 4, name: 'ssh' },
];

const supportBuilderSource = [
  { id: 1, name: 'CodeRepository' },
  { id: 2, name: 'SelfCodeRepository' },
  { id: 3, name: 'Dockerfile' },
];

export default function ({ localServer }: { localServer: string }) {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const { id } = useParams<{ id?: string }>();

  const [selectedPlatforms, setSelectedPlatforms] = useState([supportPlatforms[0]])

  useEffect(() => {
    navigate(`?${searchParams.toString()}`);
  }, [searchParams]);

  const [namespaceSearch, setNamespaceSearch] = useState('');
  const [namespaceList, setNamespaceList] = useState<INamespaceItem[]>();
  const [namespaceSelected, setNamespaceSelected] = useState<INamespaceItem>({
    name: searchParams.get('namespace') || "",
    id: parseInt(searchParams.get('namespace_id') || "") || 0,
  } as INamespaceItem);

  let backTo = searchParams.get('back_to') || "";

  useEffect(() => {
    const back = () => {
      navigate(backTo);
    }
    window.addEventListener('popstate', back);
    return () => {
      window.removeEventListener('popstate', back);
    }
  }, []);

  useEffect(() => {
    let url = `${localServer}/api/v1/namespaces/?limit=${Settings.AutoCompleteSize}`;
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
  const [repositoryList, setRepositoryList] = useState<IRepositoryItem[]>();
  const [repositorySelected, setRepositorySelected] = useState<IRepositoryItem>({
    name: searchParams.get('repository') || "",
    id: parseInt(searchParams.get('repository_id') || "") || 0,
  } as IRepositoryItem);

  useEffect(() => {
    if (namespaceSelected.name == undefined || namespaceSelected.name.length == 0) {
      return;
    }
    let url = localServer + `/api/v1/namespaces/${namespaceSelected.id}/repositories/?limit=${Settings.AutoCompleteSize}`;
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
    let url = `${localServer}/api/v1/coderepos/${codeRepositoryProviderSelected.provider}/repos/${codeRepositorySelected.id}/branches`;
    axios.get(url).then(response => {
      if (response.status == 200) {
        const data = response.data as ICodeRepositoryBranchList;
        setCodeRepositoryBranchList(data.items);
        setCodeRepositoryBranchFilteredList(data.items);
        if (searchParams.get('code_repository_branch_name') !== null) {
          for (let i = 0; i < data.total; i++) {
            if (data.items[i].name === searchParams.get('code_repository_branch_name')) {
              setCodeRepositoryBranchSelected(data.items[i]);
              setSearchParams({
                ...Object.fromEntries(searchParams.entries()),
                code_repository_branch_name: data.items[i].name,
                code_repository_branch_id: data.items[i].id.toString(),
              });
              break;
            }
          }
        }
        if (searchParams.get('cron_branch_name') !== null) {
          for (let i = 0; i < data.total; i++) {
            if (data.items[i].name === searchParams.get('cron_branch_name')) {
              setCodeRepositoryBranchSelected(data.items[i]);
              setSearchParams({
                ...Object.fromEntries(searchParams.entries()),
                cron_branch_name: data.items[i].name,
                cron_branch_id: data.items[i].id.toString(),
              });
              break;
            }
          }
        }
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

  const [dockerfileContext, setDockerfileContext] = useState('.');
  const [dockerfilePath, setDockerfilePath] = useState('Dockerfile');

  const [tagEvent, SetTagEvent] = useState(false);
  const [tagEventTagTemplate, setTagEventTagTemplate] = useState('');

  const [mergeEvent, setMergeEvent] = useState(false);
  const [mergeEventBranch, setMergeEventBranch] = useState('');
  const [mergeEventTagTemplate, setMergeEventTagTemplate] = useState('');
  const [mergeEventBranchSearch, setMergeEventBranchSearch] = useState('');
  const [mergeEventBranchFilteredList, setMergeEventBranchFilteredList] = useState<ICodeRepositoryBranchItem[]>();
  const [mergeEventBranchSelected, setMergeEventBranchSelected] = useState<ICodeRepositoryBranchItem>({
    name: searchParams.get('merge_event_branch_name') || '',
    id: parseInt(searchParams.get('merge_event_branch_id') || "") || 0,
  } as ICodeRepositoryBranchItem);
  useEffect(() => {
    if (codeRepositoryBranchList?.length == undefined || codeRepositoryBranchList?.length == 0) {
      return;
    }
    if (mergeEventBranchSearch == '') {
      setMergeEventBranchFilteredList(codeRepositoryBranchList);
    }
    let result = [];
    for (let i = 0; i < (codeRepositoryBranchList?.length || 0); i++) {
      if (codeRepositoryBranchList[i].name.toLowerCase().includes(codeRepositoryBranchSearch.toLocaleLowerCase())) {
        result.push(codeRepositoryBranchList[i])
      }
    }
    setMergeEventBranchFilteredList(result);
  }, [mergeEventBranchSearch, codeRepositoryBranchList]);

  const [cronBuild, setCronBuild] = useState(false);
  const [cronExpr, setCronExpr] = useState('');
  const [cronBranch, setCronBranch] = useState('');
  const [cronTagTemplate, setCronTagTemplate] = useState('');
  const [cronBranchSearch, setCronBranchSearch] = useState('');
  const [cronBranchFilteredList, setCronBranchFilteredList] = useState<ICodeRepositoryBranchItem[]>();
  const [cronBranchSelected, setCronBranchSelected] = useState<ICodeRepositoryBranchItem>({
    name: searchParams.get('cron_branch_name') || '',
    id: parseInt(searchParams.get('cron_branch_id') || "") || 0,
  } as ICodeRepositoryBranchItem);
  useEffect(() => {
    if (codeRepositoryBranchList?.length == undefined || codeRepositoryBranchList?.length == 0) {
      return;
    }
    if (mergeEventBranchSearch == '') {
      setCronBranchFilteredList(codeRepositoryBranchList);
    }
    let result = [];
    for (let i = 0; i < (codeRepositoryBranchList?.length || 0); i++) {
      if (codeRepositoryBranchList[i].name.toLowerCase().includes(codeRepositoryBranchSearch.toLocaleLowerCase())) {
        result.push(codeRepositoryBranchList[i])
      }
    }
    setCronBranchFilteredList(result);
  }, [cronBranchSearch, codeRepositoryBranchList]);

  const [builderSource, setBuilderSource] = useState(searchParams.get('builder_source') || 'CodeRepository');
  const [dockerfile, setDockerfile] = useState('FROM alpine:latest');
  const [customRepositoryCloneUrl, setCustomRepositoryCloneUrl] = useState('');
  const [customRepositoryBranch, setCustomRepositoryBranch] = useState('');
  const [customRepositoryCredential, setCustomRepositoryCredential] = useState('username');
  const [customRepositoryUsername, setCustomRepositoryUsername] = useState('');
  const [customRepositoryPassword, setCustomRepositoryPassword] = useState('');
  const [customRepositoryToken, setCustomRepositoryToken] = useState('');
  const [customRepositorySshKey, setCustomRepositorySshKey] = useState('');

  const setupBuilder = () => {
    const data: { [key: string]: any } = {
      namespace_id: namespaceSelected.id,
      repository_id: repositorySelected.id,
      source: builderSource,
    };
    if (builderSource === 'Dockerfile') {
      data['dockerfile'] = btoa(dockerfile);
    }
    if (builderSource === 'CodeRepository') {
      data['scm_provider'] = codeRepositoryProviderSelected.provider;
      data['code_repository_id'] = codeRepositorySelected.id;
    }
    if (builderSource === 'SelfCodeRepository') {
      data['scm_repository'] = customRepositoryCloneUrl;
      data['scm_credential_type'] = customRepositoryCredential;
      if (customRepositoryCredential === 'token') {
        data['scm_token'] = customRepositoryToken;
      }
      if (customRepositoryCredential === 'ssh') {
        data['scm_ssh_key'] = customRepositorySshKey;
      }
      if (customRepositoryCredential === 'username') {
        data['scm_username'] = customRepositoryUsername;
        data['scm_password'] = customRepositoryPassword;
      }
    }
    if (builderSource === 'CodeRepository') {
      data['scm_branch'] = codeRepositoryBranchSelected.name;
    }
    if (builderSource === 'SelfCodeRepository') {
      data['scm_branch'] = customRepositoryBranch;
    }
    if (builderSource === 'CodeRepository' || builderSource === 'SelfCodeRepository') {
      data['scm_depth'] = depth;
      data['scm_submodule'] = submodule;
    }
    if (cronBuild) {
      data['cron_rule'] = cronExpr;
      if (builderSource === 'CodeRepository' || builderSource === 'SelfCodeRepository') {
        data['cron_branch'] = cronBranch;
      }
      data['cron_tag_template'] = cronTagTemplate;
    }
    if (mergeEvent) {
      data['webhook_branch_name'] = mergeEventBranch;
      data['webhook_branch_tag_template'] = mergeEventBranch;
    }
    if (tagEvent) {
      data['webhook_tag_tag_template'] = tagEventTagTemplate;
    }
    if (builderSource === 'CodeRepository' || builderSource === 'SelfCodeRepository') {
      data['buildkit_context'] = dockerfileContext;
      data['buildkit_dockerfile'] = dockerfilePath;
    }
    let ps: string[] = [];
    for (let i = 0; i < selectedPlatforms.length; i++) {
      ps.push(selectedPlatforms[i].name);
    }
    data['buildkit_platforms'] = ps;
    console.log("create builder body:", data);
    if (id === undefined) {
      let url = `${localServer}/api/v1/namespaces/${namespaceSelected.id}/repositories/${repositorySelected.id}/builders/`;
      axios.post(url, data as IBuilderItem).then(response => {
        if (response?.status === 201) {
          navigate(`/namespaces/${namespaceSelected.name}/repository/runners?repository_id=${repositorySelected.id}&repository=${repositorySelected.name}&namespace_id=${namespaceSelected.id}`, { replace: true });
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
        }
      }).catch(error => {
        const errorcode = error.response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      });
    } else {
      let url = `${localServer}/api/v1/namespaces/${namespaceSelected.id}/repositories/${searchParams.get('repository_id')}/builders/${id}`;
      axios.put(url, data as IBuilderItem).then(response => {
        if (response?.status === 204) {
          navigate(`/namespaces/${namespaceSelected.name}/repository/runners?repository_id=${repositorySelected.id}&repository=${repositorySelected.name}&namespace_id=${namespaceSelected.id}`, { replace: true });
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
        }
      }).catch(error => {
        const errorcode = error.response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      });
    }
  }

  const [codeRepositoryInit, setCodeRepositoryInit] = useState(0);
  useEffect(() => {
    if (codeRepositoryInit != 0) {
      axios.get(`${localServer}/api/v1/coderepos/${codeRepositoryProviderSelected.provider}/repos/${codeRepositoryInit}`).then(response => {
        let codeRepositoryItem = response.data as ICodeRepositoryItem;
        setCodeRepositoryOwnerSelected({
          owner: codeRepositoryItem.owner,
          id: codeRepositoryItem.owner_id,
        } as ICodeRepositoryOwnerItem);
        setCodeRepositorySelected({
          name: codeRepositoryItem.name,
          id: codeRepositoryItem.id,
        } as ICodeRepositoryItem)
      }).catch(error => {
        const errorcode = error.response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      });
    }
  }, [codeRepositoryInit]);

  const [codeRepositoryBranchInit, setCodeRepositoryBranchInit] = useState("");
  useEffect(() => {
    if (codeRepositoryInit != 0 && codeRepositoryBranchInit !== "") {
      axios.get(`${localServer}/api/v1/coderepos/${codeRepositoryProviderSelected.provider}/repos/${codeRepositoryInit}/branches/${codeRepositoryBranchInit}`).then(response => {
        let codeRepositoryBranch = response.data as ICodeRepositoryBranchItem;
        setCodeRepositoryBranchSelected({
          name: codeRepositoryBranch.name,
          id: codeRepositoryBranch.id,
        } as ICodeRepositoryBranchItem);
      }).catch(error => {
        const errorcode = error.response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      });
    }
  }, [codeRepositoryInit, codeRepositoryBranchInit]);

  useEffect(() => {
    if (id === undefined) {
      return;
    }
    axios.get(`${localServer}/api/v1/namespaces/${namespaceSelected.id}/repositories/${searchParams.get('repository_id')}`).then(response => {
      if (response?.status === 200) {
        let repositoryItem = response.data as IRepositoryItem;
        let builderItem = repositoryItem.builder || {} as IBuilderItem;
        setBuilderSource(builderItem.source);
        setSearchParams({
          ...Object.fromEntries(searchParams.entries()),
          builder_source: builderItem.source,
          cron_enabled: builderItem.cron_rule !== null ? "true" : "false",
        });
        if (builderItem.scm_branch !== undefined) {
          setSearchParams({
            ...Object.fromEntries(searchParams.entries()),
            code_repository_branch_name: builderItem.scm_branch || '',
          });
        }
        if (builderItem.code_repository_id !== undefined) {
          setCodeRepositoryInit(builderItem.code_repository_id || 0);
          setCodeRepositoryBranchInit(builderItem.scm_branch || "");
          setCodeRepositoryProviderSelected({
            provider: builderItem.scm_provider || "",
          } as ICodeRepositoryProviderItem)
        }
        let ps = "";
        let platforms: {
          id: number;
          name: string;
        }[] = [];
        for (let i = 0; i < builderItem.buildkit_platforms.length; i++) {
          if (i == 0) {
            ps += builderItem.buildkit_platforms[i];
          } else {
            ps += `,${builderItem.buildkit_platforms[i]}`
          }
          for (let j = 0; j < supportPlatforms.length; j++) {
            if (supportPlatforms[j].name === builderItem.buildkit_platforms[i]) {
              platforms.push(supportPlatforms[j]);
              break;
            }
          }
        }
        setSearchParams({
          ...Object.fromEntries(searchParams.entries()),
          platforms: ps,
        });
        setSelectedPlatforms(platforms);
        setDockerfile(atob(builderItem.dockerfile || ''));
        setCronBuild(builderItem.cron_rule !== null);
        if (builderItem.cron_rule !== null) {
          setCronExpr(builderItem.cron_rule || '');
          setCronTagTemplate(builderItem.cron_tag_template || '');
          if (builderItem.source === 'SelfCodeRepository') {
            setCronBranch(builderItem.cron_branch || '');
            setSearchParams({
              ...Object.fromEntries(searchParams.entries()),
              cron_branch_name: builderItem.cron_branch || '',
            });
          } else if (builderItem.source === 'CodeRepository') {
            setSearchParams({
              ...Object.fromEntries(searchParams.entries()),
              cron_branch_name: builderItem.cron_branch || '',
            });
          }
        }
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [id]);

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Setup Builder</title>
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
          <div className="flex-1 flex flex-col overflow-y-auto w-full">
            <div className="py-6 px-8 border-gray-200 border-b w-full">
              <h2 className="text-base font-semibold leading-7 text-gray-900">OCI Repository</h2>
              <p className="mt-1 text-sm leading-6 text-gray-600">Builder will push the artifact to this repository.</p>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                <div className="sm:col-span-1">
                  <label htmlFor="namespace" className="block text-sm font-medium leading-6 text-gray-900">
                    Namespace
                  </label>
                  <div className="mt-2">
                    <Combobox value={namespaceSelected}
                      disabled={(searchParams.get('namespace_stick') || '') === 'true'}
                      onChange={(namespace: INamespaceItem) => {
                        setSearchParams({
                          ...Object.fromEntries(searchParams.entries()),
                          namespace: namespace.name,
                          namespace_id: namespace.id.toString(),
                          repository: '',
                          repository_id: '',
                        });
                        setRepositorySelected({} as IRepositoryItem); // clear the repo selected
                        setNamespaceSelected(namespace);
                      }}>
                      <div className="relative mt-1">
                        <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                          <Combobox.Input
                            id="namespace"
                            className={() => {
                              let cursor = ''
                              if ((searchParams.get('namespace_stick') || '') === 'true') {
                                cursor = 'cursor-not-allowed ';
                              }
                              return cursor + "w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                            }}
                            displayValue={(namespace: INamespaceItem) => namespace.name}
                            onChange={event => {
                              setNamespaceSearch(event.target.value);
                            }}
                          />
                          <Combobox.Button
                            className={() => {
                              let cursor = ''
                              if ((searchParams.get('namespace_stick') || '') === 'true') {
                                cursor = 'cursor-not-allowed ';
                              }
                              return cursor + "absolute inset-y-0 right-0 flex items-center pr-2"
                            }}
                          >
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
                          <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
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
                    <Combobox value={repositorySelected}
                      disabled={(searchParams.get('repository_stick') || '') === 'true'}
                      onChange={(repo: IRepositoryItem) => {
                        setSearchParams({
                          ...Object.fromEntries(searchParams.entries()),
                          repository: repo.name,
                          repository_id: repo.id.toString(),
                        });
                        setRepositorySelected(repo);
                      }}>
                      <div className="relative mt-1">
                        <div className="w-full relative overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm cursor-pointer" >
                          <Combobox.Input
                            id="repository"
                            className={() => {
                              let cursor = ''
                              if ((searchParams.get('repository_stick') || '') === 'true') {
                                cursor = 'cursor-not-allowed ';
                              }
                              return cursor + "w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                            }}
                            displayValue={(repository: IRepositoryItem) => {
                              if (namespaceSelected.name != undefined && repository.name != undefined) {
                                return repository.name.substring(namespaceSelected.name.length + 1)
                              }
                              return "";
                            }}
                            onChange={(event) => {
                              setRepositorySearch(event.target.value);
                            }}
                          />
                          <Combobox.Button
                            className={() => {
                              let cursor = ''
                              if ((searchParams.get('repository_stick') || '') === 'true') {
                                cursor = 'cursor-not-allowed ';
                              }
                              return cursor + "absolute inset-y-0 right-0 flex items-center pr-2"
                            }}
                          >
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
                          <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
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
                  <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                    Builder Source
                  </label>
                  <div className="mt-2 flex flex-row items-center h-[36px]">
                    <Listbox
                      disabled={(searchParams.get('code_repository_stick') || '') === 'true'}
                      value={builderSource}
                      onChange={(source: string) => {
                        setSearchParams({
                          ...Object.fromEntries(searchParams.entries()),
                          builder_source: source,
                        });
                        setBuilderSource(source);
                      }}>
                      <div className="relative mt-1 w-full">
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
                          <span className="block truncate">{builderSource}</span>
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
                              supportBuilderSource.map(source => (
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
              </div>
              {
                builderSource !== 'Dockerfile' ? (
                  <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                    {
                      builderSource === 'CodeRepository' ? (
                        <div className="sm:col-span-1">
                          <label htmlFor="codeProviders" className="block text-sm font-medium leading-6 text-gray-900">
                            Provider
                          </label>
                          <div className="mt-2">
                            <Combobox
                              disabled={(searchParams.get('code_repository_stick') || '') === 'true'}
                              value={codeRepositoryProviderSelected}
                              onChange={(provider: ICodeRepositoryProviderItem) => {
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  provider: provider.provider,
                                });
                                setCodeRepositoryOwnerSelected({} as ICodeRepositoryOwnerItem); // clear the selected
                                setCodeRepositorySelected({} as ICodeRepositoryItem);
                                setCodeRepositoryBranchSelected({} as ICodeRepositoryBranchItem);
                                setMergeEventBranchSelected({} as ICodeRepositoryBranchItem);
                                setCronBranchSelected({} as ICodeRepositoryBranchItem);
                                setCodeRepositoryProviderSelected(provider);
                              }}>
                              <div className="relative mt-1">
                                <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                                  <Combobox.Input
                                    id="codeProviders"
                                    className={() => {
                                      let cursor = ''
                                      if ((searchParams.get('code_repository_stick') || '') === 'true') {
                                        cursor = 'cursor-not-allowed ';
                                      }
                                      return cursor + "w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                                    }}
                                    displayValue={(provider: ICodeRepositoryProviderItem) => provider.provider}
                                    onChange={(event) => { }}
                                  />
                                  <Combobox.Button
                                    className={() => {
                                      let cursor = ''
                                      if ((searchParams.get('code_repository_stick') || '') === 'true') {
                                        cursor = 'cursor-not-allowed ';
                                      }
                                      return cursor + "absolute inset-y-0 right-0 flex items-center pr-2"
                                    }}
                                  >
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
                                  <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
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
                                            <span className={`block truncate font-normal`}>
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
                      ) : null
                    }
                    {
                      builderSource === 'CodeRepository' ? (
                        <div className="sm:col-span-1">
                          <label htmlFor="codeOwners" className="block text-sm font-medium leading-6 text-gray-900">
                            Organization
                          </label>
                          <div className="mt-2">
                            <Combobox
                              disabled={(searchParams.get('code_repository_stick') || '') === 'true'}
                              value={codeRepositoryOwnerSelected}
                              onChange={(owner: ICodeRepositoryOwnerItem) => {
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  code_repository_owner: owner.owner,
                                  code_repository_owner_id: owner.id.toString(),
                                });
                                setCodeRepositorySelected({} as ICodeRepositoryItem);
                                setCodeRepositoryBranchSelected({} as ICodeRepositoryBranchItem);
                                setMergeEventBranchSelected({} as ICodeRepositoryBranchItem);
                                setCronBranchSelected({} as ICodeRepositoryBranchItem);
                                setCodeRepositoryOwnerSelected(owner);
                              }}>
                              <div className="relative mt-1">
                                <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                                  <Combobox.Input
                                    id="codeOwners"
                                    className={() => {
                                      let cursor = ''
                                      if ((searchParams.get('code_repository_stick') || '') === 'true') {
                                        cursor = 'cursor-not-allowed ';
                                      }
                                      return cursor + "w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                                    }}
                                    displayValue={(owner: ICodeRepositoryOwnerItem) => owner.owner}
                                    onChange={(event) => {
                                      setCodeRepositoryOwnerSearch(event.target.value);
                                    }}
                                  />
                                  <Combobox.Button
                                    className={() => {
                                      let cursor = ''
                                      if ((searchParams.get('code_repository_stick') || '') === 'true') {
                                        cursor = 'cursor-not-allowed ';
                                      }
                                      return cursor + "absolute inset-y-0 right-0 flex items-center pr-2"
                                    }}
                                  >
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
                                  <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
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
                      ) : null
                    }
                    {
                      builderSource === 'CodeRepository' ? (
                        <div className="sm:col-span-2">
                          <label htmlFor="coderepos" className="block text-sm font-medium leading-6 text-gray-900">
                            Repository
                          </label>
                          <div className="mt-2">
                            <Combobox
                              disabled={(searchParams.get('code_repository_stick') || '') === 'true'}
                              value={codeRepositorySelected}
                              onChange={(cr: ICodeRepositoryItem) => {
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  code_repository_name: cr.name,
                                  code_repository_id: cr.id.toString(),
                                });
                                setCodeRepositoryBranchSelected({} as ICodeRepositoryBranchItem);
                                setMergeEventBranchSelected({} as ICodeRepositoryBranchItem);
                                setCronBranchSelected({} as ICodeRepositoryBranchItem);
                                setCodeRepositorySelected(cr);
                              }}>
                              <div className="relative mt-1">
                                <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                                  <Combobox.Input
                                    id="coderepos"
                                    className={() => {
                                      let cursor = ''
                                      if ((searchParams.get('code_repository_stick') || '') === 'true') {
                                        cursor = 'cursor-not-allowed ';
                                      }
                                      return cursor + "w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                                    }}
                                    // className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                                    displayValue={(cr: ICodeRepositoryItem) => cr.name}
                                    onChange={(event) => {
                                      setCodeRepositorySearch(event.target.value);
                                    }}
                                  />
                                  <Combobox.Button
                                    className={() => {
                                      let cursor = ''
                                      if ((searchParams.get('code_repository_stick') || '') === 'true') {
                                        cursor = 'cursor-not-allowed ';
                                      }
                                      return cursor + "absolute inset-y-0 right-0 flex items-center pr-2"
                                    }}
                                  >
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
                                  <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
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
                      ) : null
                    }
                    {
                      builderSource === 'CodeRepository' ? (
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
                                <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
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
                                  <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
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
                      ) : null
                    }
                    {
                      builderSource === 'SelfCodeRepository' ? (
                        <div className="sm:col-span-2">
                          <label htmlFor="customRepositoryCloneUrl" className="block text-sm font-medium leading-6 text-gray-900">
                            Repository
                          </label>
                          <div className="mt-2 flex flex-row items-center h-[36px]">
                            <input
                              type="text"
                              name="customRepositoryCloneUrl"
                              id="customRepositoryCloneUrl"
                              className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                              value={customRepositoryCloneUrl}
                              onChange={e => {
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  custom_repository_clone_url: e.target.value,
                                });
                                setCustomRepositoryCloneUrl(e.target.value)
                              }}
                            />
                          </div>
                        </div>
                      ) : null
                    }
                    {
                      builderSource === 'SelfCodeRepository' ? (
                        <div className="sm:col-span-1">
                          <label htmlFor="customRepositoryBranch" className="block text-sm font-medium leading-6 text-gray-900">
                            Branch
                          </label>
                          <div className="mt-2 flex flex-row items-center h-[36px]">
                            <input
                              type="text"
                              name="customRepositoryBranch"
                              id="customRepositoryBranch"
                              className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                              value={customRepositoryBranch}
                              onChange={e => {
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  custom_repository_branch: e.target.value,
                                });
                                setCustomRepositoryBranch(e.target.value)
                              }}
                            />
                          </div>
                        </div>
                      ) : null
                    }
                  </div>
                ) : null
              }
              {
                builderSource === 'SelfCodeRepository' ? (
                  <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                    <div className="sm:col-span-1">
                      <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                        Credential
                      </label>
                      <div className="mt-2 flex flex-row items-center h-[36px]">
                        <Listbox value={customRepositoryCredential} onChange={(cred: string) => {
                          setSearchParams({
                            ...Object.fromEntries(searchParams.entries()),
                            custom_repository_credential: cred,
                          });
                          setCustomRepositoryCredential(cred);
                        }}>
                          <div className="relative mt-1 w-full">
                            <Listbox.Button className="relative w-full rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm cursor-pointer">
                              <span className="block truncate">{customRepositoryCredential}</span>
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
                                  supportCredential.map((credential) => (
                                    <Listbox.Option key={credential.id} value={credential.name} className={({ active }) =>
                                      `relative cursor-pointer select-none py-2 pl-10 pr-4 ${active ? 'bg-gray-100 text-gray-900' : 'text-gray-900'
                                      }`
                                    }>
                                      {({ selected }) => (
                                        <>
                                          <span
                                            className={`block truncate ${selected ? 'font-medium' : 'font-normal'
                                              }`}
                                          >
                                            {credential.name}
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

                    {
                      customRepositoryCredential == 'username' ? (
                        <div className='sm:col-span-1'>
                          <label htmlFor='customRepositoryUsername' className='block text-sm font-medium leading-6 text-gray-900'>
                            Username
                          </label>
                          <div className='mt-2 h-[36px]'>
                            <input
                              type='text'
                              name='customRepositoryUsername'
                              id='customRepositoryUsername'
                              className='block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
                              value={customRepositoryUsername}
                              onChange={e => {
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  custom_repository_username: e.target.value,
                                });
                                setCustomRepositoryUsername(e.target.value)
                              }}
                            />
                          </div>
                        </div>
                      ) : null
                    }
                    {
                      customRepositoryCredential == 'username' ? (
                        <div className='sm:col-span-1'>
                          <label htmlFor="customRepositoryPassword" className="block text-sm font-medium leading-6 text-gray-900">
                            Password
                          </label>
                          <div className='mt-2 h-[36px]'>
                            <input
                              type='password'
                              autoComplete='new-password'
                              name='customRepositoryPassword'
                              id='customRepositoryPassword'
                              className='block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
                              value={customRepositoryPassword}
                              onChange={e => setCustomRepositoryPassword(e.target.value)}
                            />
                          </div>
                        </div>
                      ) : null
                    }

                    {
                      customRepositoryCredential == 'token' ? (
                        <div className='sm:col-span-1'>
                          <label htmlFor='customRepositoryToken' className="block text-sm font-medium leading-6 text-gray-900">
                            Token
                          </label>
                          <div className='mt-2 h-[36px]'>
                            <input
                              type="text"
                              name="customRepositoryToken"
                              id="customRepositoryToken"
                              className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                              value={customRepositoryToken}
                              onChange={e => setCustomRepositoryToken(e.target.value)}
                            />
                          </div>
                        </div>
                      ) : null
                    }

                    {
                      customRepositoryCredential == 'ssh' ? (
                        <div className="sm:col-span-2">
                          <label htmlFor="customRepositorySshKey" className="block text-sm font-medium leading-6 text-gray-900">
                            SSH Key
                          </label>
                          <div className='mt-2 h-[36px]'>
                            <input
                              type="text"
                              name="customRepositorySshKey"
                              id="customRepositorySshKey"
                              className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                              value={customRepositorySshKey}
                              onChange={e => setCustomRepositorySshKey(e.target.value)}
                            />
                          </div>
                        </div>
                      ) : null
                    }
                  </div>
                ) : null
              }
              {
                builderSource !== 'Dockerfile' ? (
                  <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                    <div className="sm:col-span-1">
                      <label className="block text-sm font-medium leading-6 text-gray-900">
                        Clone Submodule
                      </label>
                      <div className="mt-2 flex flex-row items-center h-[36px]">
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input type="checkbox" checked={submodule} className="sr-only peer" onChange={e => {
                            setSearchParams({
                              ...Object.fromEntries(searchParams.entries()),
                              code_repository_clone_submodule: e.target.checked ? "true" : "false",
                            });
                            setSubmodule(e.target.checked);
                          }} />
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
                          onChange={e => {
                            let depth = Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value);
                            setSearchParams({
                              ...Object.fromEntries(searchParams.entries()),
                              depth: depth.toString(),
                            });
                            setDepth(depth);
                          }}
                        />
                      </div>
                    </div>
                  </div>
                ) : null
              }
              {
                builderSource === 'Dockerfile' ? (
                  <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                    <div className="sm:col-span-4">
                      <label className="block text-sm font-medium leading-6 text-gray-900">
                        Dockerfile
                      </label>
                      <div className="mt-2 flex flex-row items-center">
                      <CodeMirror value={"FROM alpine:latest"} height="200px" extensions={[StreamLanguage.define(dockerfileModel)]} />
                        {/* <Editor
                          theme="vs-dark"
                          height="40vh"
                          defaultLanguage="dockerfile"
                          defaultValue="FROM alpine:latest"
                          value={dockerfile}
                          className='text-xs'
                          onChange={v => {
                            setDockerfile(v || '');
                          }}
                          options={{
                            fontSize: 14,
                            padding: {
                              top: 5,
                            },
                            minimap: {
                              enabled: false
                            },
                            contextmenu: false,
                          }}
                        /> */}
                      </div>
                    </div>
                  </div>
                ) : null
              }
            </div>
            <div className="py-6 px-8 border-gray-200 border-b w-full">
              <h2 className="text-base font-semibold leading-7 text-gray-900">Build Options</h2>
              <p className="mt-1 text-sm leading-6 text-gray-600">The image build options.</p>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                {
                  builderSource !== 'Dockerfile' ? (
                    <div className="sm:col-span-1">
                      <label htmlFor="dockerfileContext" className="block text-sm font-medium leading-6 text-gray-900">
                        Context
                      </label>
                      <div className="mt-2 flex flex-row items-center h-[36px]">
                        <input
                          type="text"
                          name="dockerfileContext"
                          id="dockerfileContext"
                          className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                          value={dockerfileContext}
                          onChange={e => {
                            setSearchParams({
                              ...Object.fromEntries(searchParams.entries()),
                              dockerfile_context: e.target.value,
                            });
                            setDockerfileContext(e.target.value);
                          }}
                        />
                      </div>
                    </div>
                  ) : null
                }
                {
                  builderSource !== 'Dockerfile' ? (
                    <div className="sm:col-span-1">
                      <label htmlFor="dockerfilePath" className="block text-sm font-medium leading-6 text-gray-900">
                        Dockerfile Path
                      </label>
                      <div className="mt-2 flex flex-row items-center h-[36px]">
                        <input
                          type="text"
                          name="dockerfilePath"
                          id="dockerfilePath"
                          className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                          value={dockerfilePath}
                          onChange={e => {
                            setSearchParams({
                              ...Object.fromEntries(searchParams.entries()),
                              dockerfile_path: e.target.value,
                            });
                            setDockerfilePath(e.target.value);
                          }}
                        />
                      </div>
                    </div>
                  ) : null
                }
                <div className="sm:col-span-2">
                  <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                    Platforms
                  </label>
                  <div className="mt-2 flex flex-row items-center h-[36px]">
                    <Listbox value={selectedPlatforms} onChange={platforms => {
                      let ps = "";
                      for (let i = 0; i < platforms.length; i++) {
                        if (i == 0) {
                          ps += platforms[i].name;
                        } else {
                          ps += `,${platforms[i].name}`
                        }
                      }
                      setSearchParams({
                        ...Object.fromEntries(searchParams.entries()),
                        platforms: ps,
                      });
                      setSelectedPlatforms(platforms);
                    }} multiple>
                      <div className="relative mt-1 w-full">
                        <Listbox.Button className="relative w-full rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm cursor-pointer">
                          <span className="block truncate">{selectedPlatforms.map((platform) => platform.name).join(', ')}</span>
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
                              supportPlatforms.map((platform) => (
                                <Listbox.Option key={platform.name} value={platform} className={({ active }) =>
                                  `relative cursor-pointer select-none py-2 pl-10 pr-4 ${active ? 'bg-gray-100 text-gray-900' : 'text-gray-900'
                                  }`
                                }>
                                  {({ selected }) => (
                                    <>
                                      <span
                                        className={`block truncate ${selected ? 'font-medium' : 'font-normal'
                                          }`}
                                      >
                                        {platform.name}
                                      </span>
                                      {selected ? (
                                        <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-gray-600">
                                          <CheckIcon className="h-5 w-5" aria-hidden="true" />
                                        </span>
                                      ) : null}
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
            </div>
            <div className="py-6 px-8 border-gray-200 border-b w-full">
              <h2 className="text-base font-semibold leading-7 text-gray-900">Cron build</h2>
              <p className="mt-1 text-sm leading-6 text-gray-600">Cron build image.</p>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                <div className="sm:col-span-1">
                  <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                    Enabled
                  </label>
                  <div className="mt-2 flex flex-row items-center h-[36px]">
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" checked={cronBuild} className="sr-only peer" onChange={e => {
                        setSearchParams({
                          ...Object.fromEntries(searchParams.entries()),
                          cron_enabled: e.target.checked ? "true" : "false",
                        });
                        setCronBuild(e.target.checked);
                      }} />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                    </label>
                  </div>
                </div>
                {
                  cronBuild ? (
                    <div className="sm:col-span-1">
                      <label htmlFor="mergeEventBranch" className="block text-sm font-medium leading-6 text-gray-900">
                        Expression
                      </label>
                      <div className="mt-2 flex flex-row items-center h-[36px]">
                        <input
                          type="text"
                          name="mergeEventBranch"
                          id="mergeEventBranch"
                          className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                          value={cronExpr}
                          onChange={e => {
                            setSearchParams({
                              ...Object.fromEntries(searchParams.entries()),
                              cron_expr: e.target.value,
                            });
                            setCronExpr(e.target.value);
                          }}
                        />
                      </div>
                    </div>
                  ) : null
                }
                {
                  cronBuild ? (builderSource === 'CodeRepository' ? (
                    <div className="sm:col-span-1">
                      <label htmlFor="cronBranch" className="block text-sm font-medium leading-6 text-gray-900">
                        Branch
                      </label>
                      <div className="mt-2">
                        <Combobox value={cronBranchSelected} onChange={(branch: ICodeRepositoryBranchItem) => {
                          setSearchParams({
                            ...Object.fromEntries(searchParams.entries()),
                            cron_branch_name: branch.name,
                            cron_branch_id: branch.id.toString(),
                          });
                          setCronBranchSelected(branch);
                        }}>
                          <div className="relative mt-1">
                            <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                              <Combobox.Input
                                id="cronBranch"
                                className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                                displayValue={(branch: ICodeRepositoryBranchItem) => branch.name}
                                onChange={e => {
                                  setCronBranchSearch(e.target.value);
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
                              afterLeave={() => setCronBranchSearch('')}
                            >
                              <Combobox.Options className="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                                {
                                  cronBranchFilteredList?.length === 0 ? (
                                    <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                      Nothing found.
                                    </div>
                                  ) : (
                                    cronBranchFilteredList?.map(branch => (
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
                  ) : (
                    builderSource === 'SelfCodeRepository' ? (
                      <div className="sm:col-span-1">
                        <label htmlFor="mergeEventTagTemplate" className="block text-sm font-medium leading-6 text-gray-900">
                          Branch
                        </label>
                        <div className="mt-2 flex flex-row items-center h-[36px]">
                          <input
                            type="text"
                            name="mergeEventTagTemplate"
                            id="mergeEventTagTemplate"
                            className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            value={cronBranch}
                            onChange={e => setCronBranch(e.target.value)}
                          />
                        </div>
                      </div>
                    ) : null)
                  ) : null
                }
                {
                  cronBuild ? (
                    <div className="sm:col-span-2">
                      <label htmlFor="mergeEventTagTemplate" className="block text-sm font-medium leading-6 text-gray-900">
                        OCI tag template
                      </label>
                      <div className="mt-2 flex flex-row items-center h-[36px]">
                        <input
                          type="text"
                          name="mergeEventTagTemplate"
                          id="mergeEventTagTemplate"
                          className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                          value={cronTagTemplate}
                          onChange={e => setCronTagTemplate(e.target.value)}
                        />
                      </div>
                    </div>
                  ) : null
                }
              </div>
            </div>
            {
              builderSource === 'CodeRepository11' ? ( // TODO: show this section in code repository
                <div className="py-6 px-8 border-gray-200 border-b w-full">
                  <h2 className="text-base font-semibold leading-7 text-gray-900">Webhook</h2>
                  <p className="mt-1 text-sm leading-6 text-gray-600">Webhook event trigger image build.</p>
                  <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                    <div className="sm:col-span-1">
                      <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                        Merge Event
                      </label>
                      <div className="mt-2 flex flex-row items-center h-[36px]">
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input type="checkbox" checked={mergeEvent} className="sr-only peer" onChange={e => { setMergeEvent(e.target.checked); }} />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                        </label>
                      </div>
                    </div>
                    {
                      mergeEvent ? (builderSource === 'CodeRepository11' ? (// TODO: show this section in code repository
                        <div className="sm:col-span-1">
                          <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                            Branch
                          </label>
                          <div className="mt-2">
                            <Combobox value={mergeEventBranchSelected} onChange={(branch: ICodeRepositoryBranchItem) => {
                              setSearchParams({
                                ...Object.fromEntries(searchParams.entries()),
                                merge_event_branch_name: branch.name,
                                merge_event_branch_id: branch.id.toString(),
                              });
                              setMergeEventBranchSelected(branch);
                            }}>
                              <div className="relative mt-1">
                                <div className="w-full relative cursor-default overflow-hidden rounded-lg bg-white text-left shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-teal-300 sm:text-sm">
                                  <Combobox.Input
                                    id="branch"
                                    className="w-full border-none py-2 pl-3 pr-10 text-sm leading-5 text-gray-900 focus:ring-0"
                                    displayValue={(branch: ICodeRepositoryBranchItem) => branch.name}
                                    onChange={(event) => {
                                      setMergeEventBranchSearch(event.target.value);
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
                                  afterLeave={() => setMergeEventBranchSearch('')}
                                >
                                  <Combobox.Options className="absolute mt-1 max-h-60 bottom-12 w-full overflow-auto rounded-md bg-white py-1 text-base ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                                    {
                                      mergeEventBranchFilteredList?.length === 0 ? (
                                        <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                                          Nothing found.
                                        </div>
                                      ) : (
                                        mergeEventBranchFilteredList?.map(branch => (
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
                      ) : (
                        builderSource === 'SelfCodeRepository' ? (
                          <div className="sm:col-span-1">
                            <label htmlFor="mergeEventBranch" className="block text-sm font-medium leading-6 text-gray-900">
                              Branch
                            </label>
                            <div className="mt-2 flex flex-row items-center h-[36px]">
                              <input
                                type="text"
                                name="mergeEventBranch"
                                id="mergeEventBranch"
                                className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                value={mergeEventBranch}
                                onChange={e => setMergeEventBranch(e.target.value)}
                              />
                            </div>
                          </div>
                        ) : null)
                      ) : null
                    }
                    {
                      mergeEvent ? (
                        <div className="sm:col-span-2">
                          <label htmlFor="mergeEventTagTemplate" className="block text-sm font-medium leading-6 text-gray-900">
                            OCI tag template
                          </label>
                          <div className="mt-2 flex flex-row items-center h-[36px]">
                            <input
                              type="text"
                              name="mergeEventTagTemplate"
                              id="mergeEventTagTemplate"
                              className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                              value={mergeEventTagTemplate}
                              onChange={e => setMergeEventTagTemplate(e.target.value)}
                            />
                          </div>
                        </div>
                      ) : null
                    }
                  </div>
                  <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                    <div className="sm:col-span-1">
                      <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                        Tag Event
                      </label>
                      <div className="mt-2 flex flex-row items-center h-[36px]">
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input type="checkbox" checked={tagEvent} className="sr-only peer" onChange={e => { SetTagEvent(e.target.checked); }} />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                        </label>
                      </div>
                    </div>
                    {
                      tagEvent ? (
                        <div className="sm:col-span-2">
                          <label htmlFor="tagEventTagTemplate" className="block text-sm font-medium leading-6 text-gray-900">
                            OCI tag template
                          </label>
                          <div className="mt-2 flex flex-row items-center h-[36px]">
                            <input
                              type="text"
                              name="tagEventTagTemplate"
                              id="tagEventTagTemplate"
                              className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                              value={tagEventTagTemplate}
                              onChange={e => setTagEventTagTemplate(e.target.value)}
                            />
                          </div>
                        </div>
                      ) : null
                    }
                  </div>
                </div>
              ) : null
            }
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
                  onClick={e => {
                    if (backTo !== "") {
                      navigate(backTo);
                    }
                  }}
                >
                  Cancel
                </button>
                <button
                  className="relative ml-3 inline-flex items-center rounded-md border px-4 py-2 text-sm font-medium text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500"
                  onClick={setupBuilder}
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
