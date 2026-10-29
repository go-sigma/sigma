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

import { ChevronUpDownIcon } from '@heroicons/react/20/solid';
import { Info } from 'lucide-react';
import axios from "axios";
import { orderBy } from "lodash-es";
import { Fragment, useCallback, useEffect, useState } from 'react';
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';

import ShikiDockerfileEditor from '@/components/CodeEditor/ShikiDockerfileEditor';
import CredentialDialog from '@/components/CredentialDialog';
import { useNamespaceId } from '@/hooks/useNamespaceId';
import Header from '@/components/Header';
import HeaderMenu from '@/components/Menu';
import Toast from "@/components/Notification";
import RepositoryTabs from "@/components/RepositoryTabs";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Input } from "@/components/ui/input";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useTranslation } from "@/i18n/useTranslation";
import { IBuilderItem, ICodeRepositoryBranchItem, ICodeRepositoryBranchList, ICodeRepositoryItem, ICodeRepositoryList, ICodeRepositoryOwnerItem, ICodeRepositoryOwnerList, ICodeRepositoryProviderItem, ICodeRepositoryProviderList, IEndpoint, IHTTPError, INamespaceItem, INamespaceList, IOrder, IRepositoryItem, IRepositoryList } from '@/interfaces';

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

// pickerLimit is the page size of every picker on this page. The API caps a page at 100
// entries, so ask for as much as it allows.
const pickerLimit = 100;

const supportBuilderSource = [
  { id: 1, name: 'CodeRepository' },
  { id: 2, name: 'SelfCodeRepository' },
  { id: 3, name: 'Dockerfile' },
];

export default function ({ localServer, embedded = false, onConfigured }: { localServer: string; embedded?: boolean; onConfigured?: () => void }) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const { id, namespace: routeNamespace } = useParams<{ id?: string; namespace?: string }>();

  const [selectedPlatforms, setSelectedPlatforms] = useState([supportPlatforms[0]])

  // openPicker is the branch picker that is currently open; the branch pickers refresh their
  // list when they open, and select controls close themselves after a value is picked.
  const [openPicker, setOpenPicker] = useState<string | null>(null);

  // The namespace may arrive as a name only (menu link without an id); resolve it so the
  // repository picker below has something to query.
  const resolvedNamespaceId = useNamespaceId(localServer, searchParams.get('namespace') || '');

  const [namespaceList, setNamespaceList] = useState<INamespaceItem[]>();
  const [namespaceSelected, setNamespaceSelected] = useState<INamespaceItem>({
    name: searchParams.get('namespace') || "",
    id: searchParams.get('namespace_id') || "",
  } as INamespaceItem);

  useEffect(() => {
    if (namespaceSelected.id || !resolvedNamespaceId) {
      return;
    }
    setNamespaceSelected(previous => ({ ...previous, id: resolvedNamespaceId }));
  }, [namespaceSelected.id, resolvedNamespaceId]);

  // `back_to` is only used by the Cancel button below; the browser back button is left alone so
  // it undoes the navigation that opened this page instead of rewriting the current entry.
  const backTo = searchParams.get('back_to') || "";

  // The repository of the embedded page is fixed by the route; the standalone page picks it in the
  // form below.
  const embeddedRepository = searchParams.get('repository') || "";
  const embeddedRepositoryId = searchParams.get('repository_id') || "";
  const embeddedNamespaceId = searchParams.get('namespace_id') || "";

  useEffect(() => {
    const url = `${localServer}/api/v1/namespaces/?limit=${pickerLimit}`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const namespaceListData = response.data as INamespaceList;
        setNamespaceList(orderBy(namespaceListData.items, ['name']));
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [localServer]);

  // The repository picker only lists repositories of the selected namespace, so its prefix is
  // redundant: show "test" instead of "library/test".
  const shortRepositoryName = (name: string) => {
    const prefix = `${namespaceSelected.name}/`;
    return name.startsWith(prefix) ? name.slice(prefix.length) : name;
  };

  const [repositoryList, setRepositoryList] = useState<IRepositoryItem[]>();
  const [repositorySelected, setRepositorySelected] = useState<IRepositoryItem>({
    name: searchParams.get('repository') || "",
    id: searchParams.get('repository_id') || "",
  } as IRepositoryItem);

  useEffect(() => {
    if (!namespaceSelected.id) {
      return;
    }
    const url = localServer + `/api/v1/namespaces/${namespaceSelected.id}/repositories/?limit=${pickerLimit}`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const repositoryListData = response.data as IRepositoryList;
        setRepositoryList(orderBy(repositoryListData.items, ['name']));
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespaceSelected, localServer]);

  const [codeRepositoryProviderList, setCodeRepositoryProviderList] = useState<ICodeRepositoryProviderItem[]>();
  const [codeRepositoryProviderSelected, setCodeRepositoryProviderSelected] = useState<ICodeRepositoryProviderItem>({
    provider: searchParams.get('provider') || "",
  } as ICodeRepositoryProviderItem);

  // Opened from the "add access token" hint next to the provider picker.
  const [credentialDialogOpen, setCredentialDialogOpen] = useState(false);
  const [endpoint, setEndpoint] = useState("");

  const loadCodeRepositoryProviders = useCallback(() => {
    axios.get(`${localServer}/api/v1/coderepos/providers`).then(response => {
      if (response?.status === 200) {
        setCodeRepositoryProviderList((response.data as ICodeRepositoryProviderList).items);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [localServer]);

  useEffect(() => {
    loadCodeRepositoryProviders();
  }, [loadCodeRepositoryProviders]);

  useEffect(() => {
    axios.get(`${localServer}/api/v1/systems/endpoint`).then(response => {
      if (response.status === 200) setEndpoint((response.data as IEndpoint).endpoint);
    }).catch(() => {});
  }, [localServer]);

  // A provider without a stored credential cannot list organizations or repositories, so the
  // dependent pickers stay disabled until it is configured in settings.
  const providerCredentialMissing =
    codeRepositoryProviderSelected.provider !== "" &&
    codeRepositoryProviderList !== undefined &&
    !(codeRepositoryProviderList.find(item => item.provider === codeRepositoryProviderSelected.provider)?.has_credential ?? false);

  const [codeRepositoryOwnerList, setCodeRepositoryOwnerList] = useState<ICodeRepositoryOwnerItem[]>();
  const [codeRepositoryOwnerSelected, setCodeRepositoryOwnerSelected] = useState<ICodeRepositoryOwnerItem>({
    owner: searchParams.get('code_repository_owner') || '',
    id: searchParams.get('code_repository_owner_id') || "",
  } as ICodeRepositoryOwnerItem);

  useEffect(() => {
    if (codeRepositoryProviderSelected.provider == undefined || codeRepositoryProviderSelected.provider == "") {
      return;
    }
    if (providerCredentialMissing) {
      setCodeRepositoryOwnerList(undefined);
      return;
    }
    axios.get(`${localServer}/api/v1/coderepos/${codeRepositoryProviderSelected.provider}/owners`).then(response => {
      if (response.status == 200) {
        const data = response.data as ICodeRepositoryOwnerList;
        setCodeRepositoryOwnerList(orderBy(data.items, ['is_org']));
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [codeRepositoryProviderSelected, providerCredentialMissing, localServer]);

  const [codeRepositoryList, setCodeRepositoryList] = useState<ICodeRepositoryItem[]>();
  const [codeRepositorySelected, setCodeRepositorySelected] = useState<ICodeRepositoryItem>({
    name: searchParams.get('code_repository_name') || '',
    id: searchParams.get('code_repository_id') || "",
  } as ICodeRepositoryItem);
  // The repository picker searches on the server, so the name is part of the list query below.
  const [codeRepositorySearch, setCodeRepositorySearch] = useState('');
  const [codeRepositoryPickerOpen, setCodeRepositoryPickerOpen] = useState(false);
  // Closing the picker also drops the search term so it reopens with the full list.
  const closeCodeRepositoryPicker = () => {
    setCodeRepositoryPickerOpen(false);
    setCodeRepositorySearch('');
  };

  useEffect(() => {
    if (codeRepositoryProviderSelected.provider == undefined || codeRepositoryProviderSelected.provider == "") {
      return;
    }
    if (codeRepositoryOwnerSelected.owner == undefined || codeRepositoryOwnerSelected.owner == "") {
      return;
    }
    if (providerCredentialMissing) {
      setCodeRepositoryList(undefined);
      return;
    }
    // The picker searches on the server so repositories past the first page stay reachable, and
    // lists the most recently updated repository first. pushed_at is the provider side update
    // time, updated_at only tracks the sync of the local row.
    const params = new URLSearchParams({
      owner: codeRepositoryOwnerSelected.owner,
      limit: String(pickerLimit),
      sort: 'pushed_at',
      method: IOrder.Desc,
    });
    if (codeRepositorySearch !== '') {
      params.set('name', codeRepositorySearch);
    }
    const url = `${localServer}/api/v1/coderepos/${codeRepositoryProviderSelected.provider}?${params.toString()}`;
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
  }, [codeRepositoryOwnerSelected, codeRepositoryProviderSelected, providerCredentialMissing, codeRepositorySearch, localServer]);

  const [codeRepositoryBranchList, setCodeRepositoryBranchList] = useState<ICodeRepositoryBranchItem[]>();
  const [codeRepositoryBranchSelected, setCodeRepositoryBranchSelected] = useState<ICodeRepositoryBranchItem>({
    name: searchParams.get('code_repository_branch_name') || '',
  } as ICodeRepositoryBranchItem);

  useEffect(() => {
    if (!codeRepositorySelected.id) {
      return;
    }
    let url = `${localServer}/api/v1/coderepos/${codeRepositoryProviderSelected.provider}/repos/${codeRepositorySelected.id}/branches`;
    axios.get(url).then(response => {
      if (response.status == 200) {
        const data = response.data as ICodeRepositoryBranchList;
        setCodeRepositoryBranchList(data.items);
        if (searchParams.get('code_repository_branch_name') !== null) {
          for (let i = 0; i < data.total; i++) {
            if (data.items[i].name === searchParams.get('code_repository_branch_name')) {
              setCodeRepositoryBranchSelected(data.items[i]);
              setSearchParams({
                ...Object.fromEntries(searchParams.entries()),
                code_repository_branch_name: data.items[i].name,
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
  }, [codeRepositorySelected, codeRepositoryProviderSelected.provider, localServer, searchParams, setSearchParams]);

  // Branches change far more often than repositories, so refresh them when a branch picker
  // is opened instead of only using the snapshot taken when the repository was chosen.
  const refreshCodeRepositoryBranches = () => {
    const provider = codeRepositoryProviderSelected.provider;
    const repositoryID = codeRepositorySelected.id;
    if (!provider || !repositoryID) {
      return;
    }
    axios
      .get(`${localServer}/api/v1/coderepos/${provider}/repos/${repositoryID}/branches`)
      .then(response => {
        if (response.status == 200) {
          const data = response.data as ICodeRepositoryBranchList;
          setCodeRepositoryBranchList(data.items);
        }
      })
      .catch(error => {
        const errorcode = error.response?.data as IHTTPError;
        Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
      });
  };

  const [submodule, setSubmodule] = useState(false);
  const [depth, setDepth] = useState<number>(() => {
    const depthParam = searchParams.get('code_repository_clone_depth');
    const parsed = depthParam === null ? NaN : parseInt(depthParam, 10);
    return Number.isNaN(parsed) || parsed < 0 ? 0 : parsed;
  });

  const [dockerfileContext, setDockerfileContext] = useState('.');
  const [dockerfilePath, setDockerfilePath] = useState('Dockerfile');

  const [tagEvent, SetTagEvent] = useState(false);
  const [tagEventTagTemplate, setTagEventTagTemplate] = useState('');

  const [mergeEvent, setMergeEvent] = useState(false);
  const [mergeEventBranch, setMergeEventBranch] = useState('');
  const [mergeEventTagTemplate, setMergeEventTagTemplate] = useState('');
  const [mergeEventBranchSelected, setMergeEventBranchSelected] = useState<ICodeRepositoryBranchItem>({
    name: searchParams.get('merge_event_branch_name') || '',
  } as ICodeRepositoryBranchItem);
  const [cronBuild, setCronBuild] = useState(false);
  const [cronExpr, setCronExpr] = useState('');
  const [cronBranch, setCronBranch] = useState('');
  const [cronTagTemplate, setCronTagTemplate] = useState('');
  const [cronBranchSelected, setCronBranchSelected] = useState<ICodeRepositoryBranchItem>({
    name: searchParams.get('cron_branch_name') || '',
  } as ICodeRepositoryBranchItem);
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
          if (embedded) {
            // The repository tab swaps the form for the run list once the builder exists.
            onConfigured?.();
            return;
          }
          navigate(`/namespaces/${namespaceSelected.name}/repository/builder?repository_id=${repositorySelected.id}&repository=${repositorySelected.name}&namespace_id=${namespaceSelected.id}`, { replace: true });
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
          navigate(`/namespaces/${namespaceSelected.name}/repository/builder?repository_id=${repositorySelected.id}&repository=${repositorySelected.name}&namespace_id=${namespaceSelected.id}`, { replace: true });
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

  const [codeRepositoryInit, setCodeRepositoryInit] = useState("");
  useEffect(() => {
    if (codeRepositoryInit) {
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
  }, [codeRepositoryInit, codeRepositoryProviderSelected.provider, localServer]);

  const [codeRepositoryBranchInit, setCodeRepositoryBranchInit] = useState("");
  useEffect(() => {
    if (codeRepositoryInit && codeRepositoryBranchInit !== "") {
      axios.get(`${localServer}/api/v1/coderepos/${codeRepositoryProviderSelected.provider}/repos/${codeRepositoryInit}/branches/${codeRepositoryBranchInit}`).then(response => {
        let codeRepositoryBranch = response.data as ICodeRepositoryBranchItem;
        setCodeRepositoryBranchSelected({ name: codeRepositoryBranch.name } as ICodeRepositoryBranchItem);
      }).catch(error => {
        const errorcode = error.response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      });
    }
  }, [codeRepositoryInit, codeRepositoryBranchInit, codeRepositoryProviderSelected.provider, localServer]);

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
          setCodeRepositoryInit(builderItem.code_repository_id || "");
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
  }, [id, localServer, namespaceSelected.id, searchParams, setSearchParams]);

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>{t("header.setupBuilder")}</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background">
        {embedded ? (
          <HeaderMenu localServer={localServer} item="repositories" namespace={routeNamespace} repository={embeddedRepository} />
        ) : (
          <HeaderMenu localServer={localServer} item="coderepos" />
        )}
        <div className="flex flex-col flex-1 max-h-screen">
          {/* part 1 begin */}
          <main className="relative focus:outline-none" tabIndex={0}>
            {embedded ? (
              <Header
                title={t("common.builder")}
                props={
                  <RepositoryTabs
                    namespace={routeNamespace || ""}
                    namespaceId={embeddedNamespaceId}
                    repository={embeddedRepository}
                    repositoryId={embeddedRepositoryId}
                    active="builder"
                  />
                }
              />
            ) : (
              <Header title={t("header.setupBuilder")} />
            )}
          </main>
          {/* part 1 end */}
          {/* part 2 begin */}
          <div className="flex-1 flex flex-col overflow-y-auto w-full">
            <div className="py-6 px-8 border-gray-200 border-b w-full">
              <div className="flex items-center gap-1.5">
                <h2 className="text-base font-semibold leading-7 text-gray-900 dark:text-gray-100">{t("builder.ociRepository")}</h2>
                <Tooltip>
                  <TooltipTrigger>
                    <Info className="size-4 text-muted-foreground" aria-hidden="true" />
                  </TooltipTrigger>
                  <TooltipContent>{t("builder.ociRepositoryDescription")}</TooltipContent>
                </Tooltip>
              </div>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                <div className="sm:col-span-1">
                  <label htmlFor="namespace" className="block text-sm font-medium leading-6 text-gray-900">
                    Namespace
                  </label>
                  <div className="mt-2">
                    <Select
                      value={namespaceSelected?.name || ""}
                      onValueChange={(value) => {
                        if (!value) {
                          return;
                        }
                        const namespace = namespaceList?.find(item => item.name === value);
                        if (!namespace) {
                          return;
                        }
                        setSearchParams({
                          ...Object.fromEntries(searchParams.entries()),
                          namespace: namespace.name,
                          namespace_id: namespace.id.toString(),
                          repository: '',
                          repository_id: '',
                        });
                        setRepositorySelected({} as IRepositoryItem); // clear the repo selected
                        setNamespaceSelected(namespace);
                      }}
                    >
                      <SelectTrigger className="w-full" disabled={(searchParams.get('namespace_stick') || '') === 'true'}>
                        <SelectValue placeholder="Select" />
                      </SelectTrigger>
                      {namespaceList?.length ? (
                        <SelectContent align="start">
                          {namespaceList.map(item => (
                            <SelectItem key={item.id} value={item.name}>
                              {item.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      ) : null}
                    </Select>
                  </div>
                </div>
                <div className="sm:col-span-1">
                  <label htmlFor="repository" className="block text-sm font-medium leading-6 text-gray-900">
                    Repository
                  </label>
                  <div className="mt-2">
                    <Select
                      value={repositorySelected?.name || ""}
                      onValueChange={(value) => {
                        if (!value) {
                          return;
                        }
                        const repo = repositoryList?.find(item => item.name === value);
                        if (!repo) {
                          return;
                        }
                        setSearchParams({
                          ...Object.fromEntries(searchParams.entries()),
                          repository: repo.name,
                          repository_id: repo.id.toString(),
                        });
                        setRepositorySelected(repo);
                      }}
                    >
                      <SelectTrigger className="w-full" disabled={(searchParams.get('repository_stick') || '') === 'true'}>
                        <SelectValue placeholder="Select">
                          {(value: string | null) => (value ? shortRepositoryName(value) : "Select")}
                        </SelectValue>
                      </SelectTrigger>
                      {repositoryList?.length ? (
                        <SelectContent align="start">
                          {repositoryList.map(item => (
                            <SelectItem key={item.id} value={item.name}>
                              {shortRepositoryName(item.name)}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      ) : null}
                    </Select>
                  </div>
                </div>
              </div>
            </div>
            <div className="py-6 px-8 border-gray-200 border-b w-full">
              <div className="flex items-center gap-1.5">
                <h2 className="text-base font-semibold leading-7 text-gray-900 dark:text-gray-100">{t("builder.codeRepository")}</h2>
                <Tooltip>
                  <TooltipTrigger>
                    <Info className="size-4 text-muted-foreground" aria-hidden="true" />
                  </TooltipTrigger>
                  <TooltipContent>{t("builder.codeRepositoryDescription")}</TooltipContent>
                </Tooltip>
              </div>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                <div className="sm:col-span-1">
                  <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                    Builder Source
                  </label>
                  <div className="mt-2 flex flex-row items-center h-9">
                    <Select
                      value={builderSource}
                      onValueChange={(value) => {
                        if (value == null) {
                          return;
                        }
                        const source = value;
                        setSearchParams({
                        ...Object.fromEntries(searchParams.entries()),
                        builder_source: source,
                        });
                        setBuilderSource(source);
                      }}
                    >
                      <SelectTrigger className="w-full" disabled={(searchParams.get('code_repository_stick') || '') === 'true'}>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {supportBuilderSource.map(source => (
                          <SelectItem key={source.name} value={source.name}>
                            {source.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </div>
              {
                builderSource !== 'Dockerfile' ? (
                  <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                    {
                      builderSource === 'CodeRepository' ? (
                        <div className="sm:col-span-3">
                          <label htmlFor="codeProviders" className="block text-sm font-medium leading-6 text-gray-900">
                            Provider
                          </label>
                          <div className="mt-2">
                            <Select
                              value={codeRepositoryProviderSelected?.provider || ""}
                              onValueChange={(value) => {
                                if (!value) {
                                  return;
                                }
                                const provider = codeRepositoryProviderList?.find(item => item.provider === value);
                                if (!provider) {
                                  return;
                                }
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  provider: provider.provider,
                                });
                                setCodeRepositoryOwnerSelected({} as ICodeRepositoryOwnerItem); // clear the selected
                                setCodeRepositorySelected({} as ICodeRepositoryItem);
                                setCodeRepositoryBranchSelected({} as ICodeRepositoryBranchItem);
                                setMergeEventBranchSelected({} as ICodeRepositoryBranchItem);
                                setCronBranchSelected({} as ICodeRepositoryBranchItem);
                                setCodeRepositorySearch('');
                                setCodeRepositoryProviderSelected(provider);
                              }}
                            >
                              <SelectTrigger id="codeProviders" className="w-full" disabled={(searchParams.get('code_repository_stick') || '') === 'true'}>
                                <SelectValue placeholder="Select">
                                  {(value: string | null) => codeRepositoryProviderList?.find(item => item.provider === value)?.display_name || value || "Select"}
                                </SelectValue>
                              </SelectTrigger>
                              {codeRepositoryProviderList?.length ? (
                                <SelectContent align="start">
                                  {codeRepositoryProviderList.map(item => (
                                    <SelectItem key={item.provider} value={item.provider}>
                                      {item.display_name || item.provider}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              ) : null}
                            </Select>
                            {codeRepositoryProviderList?.length === 0 && (
                              <p className="mt-1 text-xs text-muted-foreground">
                                {t("codeRepository.emptyProviders")}{" "}
                                <a href="/coderepos" className="underline">
                                  {t("codeRepository.addCredential")}
                                </a>
                              </p>
                            )}
                            {providerCredentialMissing && (
                              <p className="mt-1 text-xs text-muted-foreground">
                                {t("codeRepository.credentialMissing")}
                                {": "}
                                {codeRepositoryProviderSelected.display_name || codeRepositoryProviderSelected.provider}
                                {" · "}
                                <button
                                  type="button"
                                  className="cursor-pointer underline"
                                  onClick={() => setCredentialDialogOpen(true)}
                                >
                                  {t("codeRepository.addCredential")}
                                </button>
                              </p>
                            )}
                          </div>
                        </div>
                      ) : null
                    }
                    {
                      builderSource === 'CodeRepository' ? (
                        <div className="sm:col-span-3">
                          <label htmlFor="codeOwners" className="block text-sm font-medium leading-6 text-gray-900">
                            Organization
                          </label>
                          <div className="mt-2">
                            <Select
                              value={codeRepositoryOwnerSelected?.owner || ""}
                              onValueChange={(value) => {
                                if (!value) {
                                  return;
                                }
                                const owner = codeRepositoryOwnerList?.find(item => item.owner === value);
                                if (!owner) {
                                  return;
                                }
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  code_repository_owner: owner.owner,
                                  code_repository_owner_id: owner.id.toString(),
                                });
                                setCodeRepositorySelected({} as ICodeRepositoryItem);
                                setCodeRepositoryBranchSelected({} as ICodeRepositoryBranchItem);
                                setMergeEventBranchSelected({} as ICodeRepositoryBranchItem);
                                setCronBranchSelected({} as ICodeRepositoryBranchItem);
                                setCodeRepositorySearch('');
                                setCodeRepositoryOwnerSelected(owner);
                              }}
                            >
                              <SelectTrigger className="w-full" disabled={providerCredentialMissing || (searchParams.get('code_repository_stick') || '') === 'true'}>
                                <SelectValue placeholder="Select" />
                              </SelectTrigger>
                              {codeRepositoryOwnerList?.length ? (
                                <SelectContent align="start">
                                  {codeRepositoryOwnerList.map(item => (
                                    <SelectItem key={item.id} value={item.owner}>
                                      {item.owner}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              ) : null}
                            </Select>
                          </div>
                        </div>
                      ) : null
                    }
                    {
                      builderSource === 'CodeRepository' ? (
                        <div className="sm:col-span-3">
                          <label htmlFor="coderepos" className="block text-sm font-medium leading-6 text-gray-900">
                            Repository
                          </label>
                          <div className="mt-2">
                            <Popover
                              open={codeRepositoryPickerOpen}
                              onOpenChange={(open) => {
                                if (open) {
                                  setCodeRepositoryPickerOpen(true);
                                } else {
                                  closeCodeRepositoryPicker();
                                }
                              }}
                            >
                              <PopoverTrigger
                                render={
                                  <Button
                                    id="coderepos"
                                    variant="outline"
                                    className="w-full justify-between font-normal"
                                    disabled={providerCredentialMissing || (searchParams.get('code_repository_stick') || '') === 'true'}
                                  >
                                    <span className="truncate">{codeRepositorySelected?.name || "Select"}</span>
                                    <ChevronUpDownIcon className="size-4 text-muted-foreground" aria-hidden="true" />
                                  </Button>
                                }
                              />
                              <PopoverContent align="start" className="w-(--anchor-width) p-0">
                                <Command shouldFilter={false}>
                                  <CommandInput
                                    placeholder={t("codeRepository.searchPlaceholder")}
                                    value={codeRepositorySearch}
                                    onValueChange={setCodeRepositorySearch}
                                  />
                                  <CommandList>
                                    <CommandEmpty>{t("common.noResult")}</CommandEmpty>
                                    <CommandGroup>
                                      {codeRepositoryList?.map(item => (
                                        <CommandItem
                                          key={item.id}
                                          value={item.name}
                                          onSelect={() => {
                                            setSearchParams({
                                              ...Object.fromEntries(searchParams.entries()),
                                              code_repository_name: item.name,
                                              code_repository_id: item.id.toString(),
                                            });
                                            setCodeRepositoryBranchSelected({} as ICodeRepositoryBranchItem);
                                            setMergeEventBranchSelected({} as ICodeRepositoryBranchItem);
                                            setCronBranchSelected({} as ICodeRepositoryBranchItem);
                                            setCodeRepositorySelected(item);
                                            closeCodeRepositoryPicker();
                                          }}
                                        >
                                          {item.name}
                                        </CommandItem>
                                      ))}
                                    </CommandGroup>
                                  </CommandList>
                                </Command>
                              </PopoverContent>
                            </Popover>
                          </div>
                        </div>
                      ) : null
                    }
                    {
                      builderSource === 'CodeRepository' ? (
                        <div className="sm:col-span-3">
                          <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                            Branch
                          </label>
                          <div className="mt-2">
                            <Select
                              value={codeRepositoryBranchSelected?.name || ""}
                              onValueChange={(value) => {
                                if (!value) {
                                  return;
                                }
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  code_repository_branch_name: value,
                                });
                                setCodeRepositoryBranchSelected({ name: value } as ICodeRepositoryBranchItem);
                              }}
                              open={openPicker === "branch"}
                              onOpenChange={(open) => {
                                setOpenPicker(open ? "branch" : null);
                                // Branches change far more often than repositories, so refresh them on open.
                                if (open) refreshCodeRepositoryBranches();
                              }}
                            >
                              <SelectTrigger className="w-full">
                                <SelectValue placeholder="Select" />
                              </SelectTrigger>
                              {codeRepositoryBranchList?.length ? (
                                <SelectContent align="start">
                                  {codeRepositoryBranchList.map(item => (
                                    <SelectItem key={item.name} value={item.name}>
                                      {item.name}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              ) : null}
                            </Select>
                          </div>
                        </div>
                      ) : null
                    }
                    {
                      builderSource === 'SelfCodeRepository' ? (
                        <div className="sm:col-span-4">
                          <label htmlFor="customRepositoryCloneUrl" className="block text-sm font-medium leading-6 text-gray-900">
                            Repository
                          </label>
                          <div className="mt-2">
                            <Input
                              type="text"
                              name="customRepositoryCloneUrl"
                              id="customRepositoryCloneUrl"
                              placeholder="https://github.com/go-sigma/sigma.git"
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
                        <div className="sm:col-span-2">
                          <label htmlFor="customRepositoryBranch" className="block text-sm font-medium leading-6 text-gray-900">
                            Branch
                          </label>
                          <div className="mt-2">
                            <Input
                              type="text"
                              name="customRepositoryBranch"
                              id="customRepositoryBranch"
                              placeholder="main"
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
                    <div className="sm:col-span-2">
                      <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                        Credential
                      </label>
                      <div className="mt-2">
                        <Select
                          value={customRepositoryCredential}
                          onValueChange={(value) => {
                            if (value == null) {
                              return;
                            }
                            const cred = value;
                            setSearchParams({
                            ...Object.fromEntries(searchParams.entries()),
                            custom_repository_credential: cred,
                            });
                            setCustomRepositoryCredential(cred);
                          }}
                        >
                          <SelectTrigger className="w-full">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {supportCredential.map(credential => (
                              <SelectItem key={credential.name} value={credential.name}>
                                {credential.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                    </div>

                    {
                      customRepositoryCredential == 'username' ? (
                        <div className='sm:col-span-2'>
                          <label htmlFor='customRepositoryUsername' className='block text-sm font-medium leading-6 text-gray-900'>
                            Username
                          </label>
                          <div className='mt-2'>
                            <Input
                              type='text'
                              name='customRepositoryUsername'
                              id='customRepositoryUsername'
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
                        <div className='sm:col-span-2'>
                          <label htmlFor="customRepositoryPassword" className="block text-sm font-medium leading-6 text-gray-900">
                            Password
                          </label>
                          <div className='mt-2'>
                            <Input
                              type='password'
                              autoComplete='new-password'
                              name='customRepositoryPassword'
                              id='customRepositoryPassword'
                              value={customRepositoryPassword}
                              onChange={e => setCustomRepositoryPassword(e.target.value)}
                            />
                          </div>
                        </div>
                      ) : null
                    }

                    {
                      customRepositoryCredential == 'token' ? (
                        <div className='sm:col-span-2'>
                          <label htmlFor='customRepositoryToken' className="block text-sm font-medium leading-6 text-gray-900">
                            Token
                          </label>
                          <div className='mt-2'>
                            <Input
                              type="password"
                              autoComplete="new-password"
                              name="customRepositoryToken"
                              id="customRepositoryToken"
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
                          <div className='mt-2'>
                            <Input
                              type="text"
                              name="customRepositorySshKey"
                              id="customRepositorySshKey"
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
                      <div className="mt-2 flex flex-row items-center h-9">
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input type="checkbox" checked={submodule} className="sr-only peer" onChange={e => {
                            setSearchParams({
                              ...Object.fromEntries(searchParams.entries()),
                              code_repository_clone_submodule: e.target.checked ? "true" : "false",
                            });
                            setSubmodule(e.target.checked);
                          }} />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-ring rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-primary peer-checked:after:bg-primary-foreground"></div>
                        </label>
                      </div>
                    </div>
                    <div className="sm:col-span-1">
                      <label htmlFor="depth" className="block text-sm font-medium leading-6 text-gray-900">
                        Clone Depth
                      </label>
                      <div className="mt-2">
                        <Input
                          type="number"
                          name="depth"
                          id="depth"
                          min={0}
                          step={1}
                          placeholder="0"
                          value={depth}
                          onChange={e => {
                            const parsed = parseInt(e.target.value, 10);
                            const depthValue = Number.isNaN(parsed) || parsed < 0 ? 0 : parsed;
                            setSearchParams({
                              ...Object.fromEntries(searchParams.entries()),
                              code_repository_clone_depth: depthValue.toString(),
                            });
                            setDepth(depthValue);
                          }}
                        />
                        <p className="mt-1 text-xs text-muted-foreground">
                          {t("builder.cloneDepthHint")}
                        </p>
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
                        <ShikiDockerfileEditor
                          height="40vh"
                          value={dockerfile}
                          onChange={setDockerfile}
                        />
                      </div>
                    </div>
                  </div>
                ) : null
              }
            </div>
            <div className="py-6 px-8 border-gray-200 border-b w-full">
              <div className="flex items-center gap-1.5">
                <h2 className="text-base font-semibold leading-7 text-gray-900 dark:text-gray-100">{t("builder.buildOptions")}</h2>
                <Tooltip>
                  <TooltipTrigger>
                    <Info className="size-4 text-muted-foreground" aria-hidden="true" />
                  </TooltipTrigger>
                  <TooltipContent>{t("builder.buildOptionsDescription")}</TooltipContent>
                </Tooltip>
              </div>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                {
                  builderSource !== 'Dockerfile' ? (
                    <div className="sm:col-span-2">
                      <label htmlFor="dockerfileContext" className="block text-sm font-medium leading-6 text-gray-900">
                        Context
                      </label>
                      <div className="mt-2">
                        <Input
                          type="text"
                          name="dockerfileContext"
                          id="dockerfileContext"
                          placeholder="."
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
                    <div className="sm:col-span-2">
                      <label htmlFor="dockerfilePath" className="block text-sm font-medium leading-6 text-gray-900">
                        Dockerfile Path
                      </label>
                      <div className="mt-2">
                        <Input
                          type="text"
                          name="dockerfilePath"
                          id="dockerfilePath"
                          placeholder="Dockerfile"
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
                  <label htmlFor="platforms" className="block text-sm font-medium leading-6 text-gray-900">
                    Platforms
                  </label>
                  <div className="mt-2">
                    <DropdownMenu>
                      <DropdownMenuTrigger
                        render={
                          <Button id="platforms" variant="outline" className="w-full justify-between font-normal">
                            {selectedPlatforms.map((platform) => platform.name).join(', ') || "Select"}
                            <ChevronUpDownIcon className="size-4 text-muted-foreground" aria-hidden="true" />
                          </Button>
                        }
                      />
                      <DropdownMenuContent align="start" className="w-(--anchor-width)">
                        {supportPlatforms.map(item => (
                          <DropdownMenuCheckboxItem
                            key={item.name}
                            checked={selectedPlatforms.some(selected => selected.name === item.name)}
                            onCheckedChange={(checked) => {
                              const platforms = checked
                                ? [...selectedPlatforms, item]
                                : selectedPlatforms.filter(selected => selected.name !== item.name);
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
                            }}
                          >
                            {item.name}
                          </DropdownMenuCheckboxItem>
                        ))}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </div>
              </div>
            </div>
            <div className="py-6 px-8 border-gray-200 border-b w-full">
              <div className="flex items-center gap-1.5">
                <h2 className="text-base font-semibold leading-7 text-gray-900 dark:text-gray-100">{t("builder.cronBuild")}</h2>
                <Tooltip>
                  <TooltipTrigger>
                    <Info className="size-4 text-muted-foreground" aria-hidden="true" />
                  </TooltipTrigger>
                  <TooltipContent>{t("builder.cronBuildDescription")}</TooltipContent>
                </Tooltip>
              </div>
              <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                <div className="sm:col-span-1">
                  <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                    Enabled
                  </label>
                  <div className="mt-2 flex flex-row items-center h-9">
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" checked={cronBuild} className="sr-only peer" onChange={e => {
                        setSearchParams({
                          ...Object.fromEntries(searchParams.entries()),
                          cron_enabled: e.target.checked ? "true" : "false",
                        });
                        setCronBuild(e.target.checked);
                      }} />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-ring rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-primary peer-checked:after:bg-primary-foreground"></div>
                    </label>
                  </div>
                </div>
                {
                  cronBuild ? (
                    <div className="sm:col-span-1">
                      <div className="flex items-center gap-1.5">
                        <label htmlFor="mergeEventBranch" className="block text-sm font-medium leading-6 text-gray-900">
                          Expression
                        </label>
                        <Tooltip>
                          <TooltipTrigger>
                            <Info className="size-4 text-muted-foreground" aria-hidden="true" />
                          </TooltipTrigger>
                          <TooltipContent>{t("builder.cronExpressionHint")}</TooltipContent>
                        </Tooltip>
                      </div>
                      <div className="mt-2">
                        <Input
                          type="text"
                          name="mergeEventBranch"
                          id="mergeEventBranch"
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
                        <Select
                          value={cronBranchSelected?.name || ""}
                          onValueChange={(value) => {
                            if (!value) {
                              return;
                            }
                            setSearchParams({
                              ...Object.fromEntries(searchParams.entries()),
                              cron_branch_name: value,
                            });
                            setCronBranchSelected({ name: value } as ICodeRepositoryBranchItem);
                          }}
                          open={openPicker === "cronBranch"}
                          onOpenChange={(open) => {
                            setOpenPicker(open ? "cronBranch" : null);
                            if (open) refreshCodeRepositoryBranches();
                          }}
                        >
                          <SelectTrigger className="w-full">
                            <SelectValue placeholder="Select" />
                          </SelectTrigger>
                          {codeRepositoryBranchList?.length ? (
                            <SelectContent align="start">
                              {codeRepositoryBranchList.map(item => (
                                <SelectItem key={item.name} value={item.name}>
                                  {item.name}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          ) : null}
                        </Select>
                      </div>
                    </div>
                  ) : (
                    builderSource === 'SelfCodeRepository' ? (
                      <div className="sm:col-span-1">
                        <label htmlFor="cronBranch" className="block text-sm font-medium leading-6 text-gray-900">
                          Branch
                        </label>
                        <div className="mt-2">
                          <Input
                            type="text"
                            name="cronBranch"
                            id="cronBranch"
                            placeholder="main"
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
                      <div className="flex items-center gap-1.5">
                        <label htmlFor="mergeEventTagTemplate" className="block text-sm font-medium leading-6 text-gray-900">
                          OCI tag template
                        </label>
                        <Tooltip>
                          <TooltipTrigger>
                            <Info className="size-4 text-muted-foreground" aria-hidden="true" />
                          </TooltipTrigger>
                          <TooltipContent>{t("builder.tagTemplateHint")}</TooltipContent>
                        </Tooltip>
                      </div>
                      <div className="mt-2">
                        <Input
                          type="text"
                          name="mergeEventTagTemplate"
                          id="mergeEventTagTemplate"
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
                  <div className="flex items-center gap-1.5">
                    <h2 className="text-base font-semibold leading-7 text-gray-900 dark:text-gray-100">{t("common.webhook")}</h2>
                    <Tooltip>
                      <TooltipTrigger>
                        <Info className="size-4 text-muted-foreground" aria-hidden="true" />
                      </TooltipTrigger>
                      <TooltipContent>{t("builder.webhookDescription")}</TooltipContent>
                    </Tooltip>
                  </div>
                  <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                    <div className="sm:col-span-1">
                      <label htmlFor="branch" className="block text-sm font-medium leading-6 text-gray-900">
                        Merge Event
                      </label>
                      <div className="mt-2 flex flex-row items-center h-9">
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input type="checkbox" checked={mergeEvent} className="sr-only peer" onChange={e => { setMergeEvent(e.target.checked); }} />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-ring rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-primary peer-checked:after:bg-primary-foreground"></div>
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
                            <Select
                              value={mergeEventBranchSelected?.name || ""}
                              onValueChange={(value) => {
                                if (!value) {
                                  return;
                                }
                                setSearchParams({
                                  ...Object.fromEntries(searchParams.entries()),
                                  merge_event_branch_name: value,
                                });
                                setMergeEventBranchSelected({ name: value } as ICodeRepositoryBranchItem);
                              }}
                              open={openPicker === "mergeEventBranch"}
                              onOpenChange={(open) => {
                                setOpenPicker(open ? "mergeEventBranch" : null);
                                if (open) refreshCodeRepositoryBranches();
                              }}
                            >
                              <SelectTrigger className="w-full">
                                <SelectValue placeholder="Select" />
                              </SelectTrigger>
                              {codeRepositoryBranchList?.length ? (
                                <SelectContent align="start">
                                  {codeRepositoryBranchList.map(item => (
                                    <SelectItem key={item.name} value={item.name}>
                                      {item.name}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              ) : null}
                            </Select>
                          </div>
                        </div>
                      ) : (
                        builderSource === 'SelfCodeRepository' ? (
                          <div className="sm:col-span-1">
                            <label htmlFor="mergeEventBranch" className="block text-sm font-medium leading-6 text-gray-900">
                              Branch
                            </label>
                            <div className="mt-2">
                              <Input
                                type="text"
                                name="mergeEventBranch"
                                id="mergeEventBranch"
                                placeholder="main"
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
                          <div className="mt-2 flex flex-row items-center h-9">
                            <input
                              type="text"
                              name="mergeEventTagTemplate"
                              id="mergeEventTagTemplate"
                              className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-ring sm:text-sm sm:leading-6"
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
                      <div className="mt-2 flex flex-row items-center h-9">
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input type="checkbox" checked={tagEvent} className="sr-only peer" onChange={e => { SetTagEvent(e.target.checked); }} />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-ring rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-primary peer-checked:after:bg-primary-foreground"></div>
                        </label>
                      </div>
                    </div>
                    {
                      tagEvent ? (
                        <div className="sm:col-span-2">
                          <label htmlFor="tagEventTagTemplate" className="block text-sm font-medium leading-6 text-gray-900">
                            OCI tag template
                          </label>
                          <div className="mt-2 flex flex-row items-center h-9">
                            <input
                              type="text"
                              name="tagEventTagTemplate"
                              id="tagEventTagTemplate"
                              className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-ring sm:text-sm sm:leading-6"
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
                  onClick={() => {
                    if (backTo !== "") {
                      navigate(backTo);
                      return;
                    }
                    // Opened as a repository tab: go back the way the user came.
                    navigate(-1);
                  }}
                >
                  Cancel
                </button>
                <button
                  className="relative ml-3 inline-flex items-center rounded-md border px-4 py-2 text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/80 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring"
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
      {credentialDialogOpen && (
        <CredentialDialog
          localServer={localServer}
          provider={
            codeRepositoryProviderList?.find(item => item.provider === codeRepositoryProviderSelected.provider) ??
            codeRepositoryProviderSelected
          }
          endpoint={endpoint}
          onClose={() => setCredentialDialogOpen(false)}
          onSaved={() => {
            setCredentialDialogOpen(false);
            // Reload so has_credential flips and the organization picker unlocks.
            loadCodeRepositoryProviders();
          }}
        />
      )}
    </Fragment >
  );
}
