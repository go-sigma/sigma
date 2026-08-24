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

import { Fragment, useEffect, useRef, useState } from 'react';
import { Helmet, HelmetProvider } from 'react-helmet-async';
import {
  ICodeRepositoryItem,
  ICodeRepositoryList,
  ICodeRepositoryOwnerItem,
  ICodeRepositoryOwnerList,
  ICodeRepositoryUser3rdParty,
  IHTTPError,
  IOrder
} from '../../interfaces';
import { useNavigate, useParams } from 'react-router-dom';

import Header from '../../components/Header';
import HeaderMenu from '../../components/Menu';
import Pagination from '../../components/Pagination';
import Settings from '../../Settings';
import Toast from "../../components/Notification";
import { useTranslation } from "../../i18n/useTranslation";
import _ from 'lodash';
import axios from 'axios';
import dayjs from 'dayjs';

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { RefreshCw, UserPlus } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

export default function ({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
  const { provider } = useParams<{ provider: string }>();

  const coderepoRef = useRef<HTMLDivElement>(null);

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const [searchCodeRepo, setSearchCodeRepo] = useState("");
  const [searchCodeRepoEvent, setSearchCodeRepoEvent] = useState(0);

  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");

  const [codeRepositoryOwners, setCodeRepositoryOwners] = useState<ICodeRepositoryOwnerItem[]>([]);
  const [organization, setOrganization] = useState("");

  const [repositories, setRepositories] = useState<ICodeRepositoryItem[]>([]);
  const [refresh, setRefresh] = useState({});

  useEffect(() => {
    if (!provider) return;
    axios.get(`${localServer}/api/v1/coderepos/${provider}/owners`).then(response => {
      if (response.status == 200) {
        const data = response.data as ICodeRepositoryOwnerList;
        setCodeRepositoryOwners(_.orderBy(data.items, ['is_org']));
        for (let i = 0; i < data.items.length; i++) {
          if (!data.items[i].is_org) {
            setOrganization(data.items[i].owner);
            break;
          }
        }
      }
    }).catch(() => {});
  }, [provider, refresh]);

  const [user3rdparty, setUser3rdparty] = useState<ICodeRepositoryUser3rdParty>();
  const [refreshUser3rdparty, setRefreshUser3rdparty] = useState({});
  useEffect(() => {
    if (!provider) return;
    axios.get(`${localServer}/api/v1/coderepos/${provider}/user3rdparty`).then(response => {
      if (response.status == 200) setUser3rdparty(response.data as ICodeRepositoryUser3rdParty);
    }).catch(() => {});
  }, [provider, refresh, refreshUser3rdparty]);

  useEffect(() => {
    const timer = setInterval(() => setRefreshUser3rdparty({}), 5000);
    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!provider || organization == "") return;
    let url = `${localServer}/api/v1/coderepos/${provider}?owner=${organization}&limit=${Settings.PageSize}&page=${page}`;
    if (searchCodeRepo != "") {
      url = `${localServer}/api/v1/coderepos/${provider}?owner=${organization}&name=${searchCodeRepo}&limit=${Settings.PageSize}&page=${page}`;
    }
    axios.get(url).then(response => {
      if (response.status == 200) {
        setRepositories((response.data as ICodeRepositoryList).items);
        setTotal((response.data as ICodeRepositoryList).total);
      }
    }).catch(() => {});
  }, [provider, organization, page, searchCodeRepoEvent]);

  const setPageAndScrollTop = (page: number) => {
    if (coderepoRef?.current) coderepoRef.current.scrollTop = 0;
    setPage(page);
  }

  const crResync = () => {
    if (!provider) return;
    if (user3rdparty?.cr_last_update_status == "Doing") {
      Toast({ level: "warning", title: "Code repository is already synchronizing", message: "" });
      return;
    }
    axios.get(`${localServer}/api/v1/coderepos/${provider}/resync`).then(response => {
      if (response.status == 202) {
        setTimeout(() => setRefresh({}), 200);
        Toast({ level: "success", title: "Code Repository is synchronizing", message: "" });
      }
    }).catch(() => {});
  }

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet><title>{t("common.codeRepository")}</title></Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background min-w-[1600px]">
        <HeaderMenu localServer={localServer} item="coderepos" />
        <div className="flex flex-col flex-1 max-h-screen">
          <main className="relative focus:outline-none" tabIndex={0}>
            <Header title={t("header.codeRepository")} />
            <div className="pt-2 pb-2 flex justify-between items-center">
              <div className="pr-2 pl-2 flex-1">
                <div className="flex items-end gap-2">
                  <div className="grid gap-1 relative w-40">
                    <Label htmlFor="codeRepositorySearch" className="text-xs text-foreground">Organization</Label>
                    <Select value={organization} onValueChange={(v) => { if (v) setOrganization(v); }}>
                      <SelectTrigger id="codeRepositorySearch" className="h-10"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {codeRepositoryOwners.map(cro => (
                          <SelectItem key={cro.id} value={cro.owner}>{cro.owner}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-1 relative flex-1 max-w-xs">
                    <Label htmlFor="codeRepoSearchInput" className="text-xs text-foreground">Code Repository</Label>
                    <Input
                      id="codeRepoSearchInput"
                      placeholder="search code repository"
                      value={searchCodeRepo}
                      onChange={e => setSearchCodeRepo(e.target.value)}
                      onKeyDown={e => { if (e.key == "Enter") setSearchCodeRepoEvent(searchCodeRepoEvent + 1); }}
                      className="h-10 pr-14"
                    />
                  </div>
                </div>
              </div>
              <div>
                <Tooltip>
                  <TooltipTrigger>
                    <Button onClick={crResync}>
                      <RefreshCw data-icon="inline-start" />
                      Sync
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    {user3rdparty?.cr_last_update_status || ""}{user3rdparty?.cr_last_update_status == "Failed" && user3rdparty?.cr_last_update_message != "" ? ", " + user3rdparty?.cr_last_update_message : ""}. Last updated {dayjs().to(dayjs(user3rdparty?.cr_last_update_timestamp))}
                  </TooltipContent>
                </Tooltip>
              </div>
              <div className="pr-2 flex flex-col">
                <Button variant="outline">
                  <UserPlus data-icon="inline-start" />
                  Clone credential
                </Button>
              </div>
            </div>
          </main>
          <div ref={coderepoRef} className="flex-1 flex overflow-y-auto">
            <div className="w-full">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead><span className="lg:pl-2">Name</span></TableHead>
                    <TableHead className="text-right">Oci Repo Count</TableHead>
                    <TableHead className="text-right">Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {repositories?.map((repository) => (
                    <TableItem key={repository.id} provider={provider || ""} repository={repository} />
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
          <div style={{ marginTop: "auto" }}>
            <Pagination limit={Settings.PageSize} page={page} setPage={setPageAndScrollTop} total={total} />
          </div>
        </div>
      </div>
    </Fragment>
  )
}

function TableItem({ provider, repository }: { provider: string, repository: ICodeRepositoryItem }) {
  const navigate = useNavigate();
  return (
    <TableRow className="align-middle">
      <TableCell className="cursor-pointer" onClick={() => window.open(repository.clone_url, "_blank")}>
        <div className="truncate">
          <span className="font-medium">{repository.name}</span>
          <span className="text-muted-foreground font-normal ml-2">{repository.clone_url}</span>
        </div>
      </TableCell>
      <TableCell className="text-right text-muted-foreground cursor-pointer">{repository.oci_repo_count}</TableCell>
      <TableCell className="text-right text-muted-foreground cursor-pointer hover:text-foreground"
        onClick={() => {
          navigate(`/builders/setup?builder_source=CodeRepository&provider=${provider}&code_repository_stick=true&code_repository_owner=${repository.owner}&code_repository_owner_id=${repository.owner_id}&code_repository_name=${repository.name}&code_repository_id=${repository.id}&back_to=/coderepos/${provider}`)
        }}
      >
        Setup
      </TableCell>
    </TableRow>
  )
}