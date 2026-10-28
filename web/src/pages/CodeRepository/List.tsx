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
import dayjs from "dayjs";
import { orderBy } from "lodash-es";
import { CornerDownLeft, RefreshCw } from "lucide-react";
import { Fragment, useEffect, useRef, useState } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { useParams } from "react-router-dom";

import Header from "@/components/Header";
import HeaderMenu from "@/components/Menu";
import Toast from "@/components/Notification";
import Pagination from "@/components/Pagination";
import RelativeTime from "@/components/RelativeTime";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useTranslation } from "@/i18n/useTranslation";
import {
  ICodeRepositoryItem,
  ICodeRepositoryList,
  ICodeRepositoryOwnerItem,
  ICodeRepositoryOwnerList,
  ICodeRepositoryProviderItem,
  ICodeRepositoryProviderList,
  IOrder,
} from "@/interfaces";

// codeRepositoryPageSize is the number of code repositories shown per page.
const codeRepositoryPageSize = 20;

export default function ({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
  const { provider } = useParams<{ provider: string }>();

  const coderepoRef = useRef<HTMLDivElement>(null);

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const [searchCodeRepo, setSearchCodeRepo] = useState("");
  const [searchCodeRepoEvent, setSearchCodeRepoEvent] = useState(0);

  const [codeRepositoryOwners, setCodeRepositoryOwners] = useState<
    ICodeRepositoryOwnerItem[]
  >([]);
  const [organization, setOrganization] = useState("");

  const [repositories, setRepositories] = useState<ICodeRepositoryItem[]>([]);
  const [refresh, setRefresh] = useState({});

  useEffect(() => {
    if (!provider) return;
    axios
      .get(`${localServer}/api/v1/coderepos/${provider}/owners`)
      .then((response) => {
        if (response.status == 200) {
          const data = response.data as ICodeRepositoryOwnerList;
          setCodeRepositoryOwners(orderBy(data.items, ["is_org"]));
          for (let i = 0; i < data.items.length; i++) {
            if (!data.items[i].is_org) {
              setOrganization(data.items[i].owner);
              break;
            }
          }
        }
      })
      .catch(() => {});
  }, [provider, refresh, localServer]);

  const [providerItem, setProviderItem] =
    useState<ICodeRepositoryProviderItem>();
  const [refreshUser3rdparty, setRefreshUser3rdparty] = useState({});
  useEffect(() => {
    if (!provider) return;
    axios
      .get(`${localServer}/api/v1/coderepos/providers`)
      .then((response) => {
        if (response.status == 200) {
          const items = (response.data as ICodeRepositoryProviderList).items;
          setProviderItem(items.find((item) => item.provider === provider));
        }
      })
      .catch(() => {});
  }, [provider, refresh, refreshUser3rdparty, localServer]);

  useEffect(() => {
    const timer = setInterval(() => setRefreshUser3rdparty({}), 5000);
    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!provider || organization == "") return;
    const params = new URLSearchParams({
      owner: organization,
      limit: String(codeRepositoryPageSize),
      page: String(page),
      sort: "pushed_at",
      method: String(IOrder.Desc),
    });
    if (searchCodeRepo != "") {
      params.set("name", searchCodeRepo);
    }
    const url = `${localServer}/api/v1/coderepos/${provider}?${params.toString()}`;
    axios
      .get(url)
      .then((response) => {
        if (response.status == 200) {
          setRepositories((response.data as ICodeRepositoryList).items);
          setTotal((response.data as ICodeRepositoryList).total);
        }
      })
      .catch(() => {});
  }, [
    provider,
    organization,
    page,
    searchCodeRepoEvent,
    localServer,
    searchCodeRepo,
  ]);

  const setPageAndScrollTop = (nextPage: number) => {
    if (coderepoRef?.current) coderepoRef.current.scrollTop = 0;
    setPage(nextPage);
  };

  const crResync = () => {
    if (!provider) return;
    if (providerItem?.cr_last_update_status == "Doing") {
      Toast({
        level: "warning",
        title: "Code repository is already synchronizing",
        message: "",
      });
      return;
    }
    axios
      .get(`${localServer}/api/v1/coderepos/${provider}/resync`)
      .then((response) => {
        if (response.status == 202) {
          setTimeout(() => setRefresh({}), 200);
          Toast({
            level: "success",
            title: "Code Repository is synchronizing",
            message: "",
          });
        }
      })
      .catch(() => {});
  };

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>{t("common.codeRepository")}</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background">
        <HeaderMenu localServer={localServer} item="coderepos" />
        <div className="flex flex-col flex-1 max-h-screen">
          <main className="relative focus:outline-none" tabIndex={0}>
            <Header title={t("header.codeRepository")} />
            <div className="pt-4 pb-4 flex justify-between items-center">
              <div className="px-4 flex-1">
                <div className="flex items-end gap-2">
                  <div className="relative flex items-center w-40">
                    <Label
                      htmlFor="codeRepositorySearch"
                      className="absolute -top-2 left-2 inline-block bg-background px-1 text-xs font-medium text-foreground z-10"
                    >
                      Organization
                    </Label>
                    <Select
                      value={organization}
                      onValueChange={(v) => {
                        if (v) setOrganization(v);
                      }}
                    >
                      <SelectTrigger id="codeRepositorySearch" className="h-10 w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent align="start">
                        {codeRepositoryOwners.map((cro) => (
                          <SelectItem key={cro.id} value={cro.owner}>
                            {cro.owner}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="relative flex items-center flex-1 max-w-xs">
                    <Label
                      htmlFor="codeRepoSearchInput"
                      className="absolute -top-2 left-2 inline-block bg-background px-1 text-xs font-medium text-foreground z-10"
                    >
                      {t("common.codeRepository")}
                    </Label>
                    <Input
                      id="codeRepoSearchInput"
                      placeholder={t("codeRepository.searchPlaceholder")}
                      value={searchCodeRepo}
                      onChange={(e) => setSearchCodeRepo(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key == "Enter")
                          setSearchCodeRepoEvent(searchCodeRepoEvent + 1);
                      }}
                      className="h-10 pr-14"
                    />
                    <kbd
                      className="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground"
                      aria-hidden="true"
                    >
                      <CornerDownLeft className="size-3.5" />
                    </kbd>
                  </div>
                </div>
              </div>
              <div className="px-4">
                <Tooltip>
                  <TooltipTrigger>
                    <Button onClick={crResync}>
                      <RefreshCw data-icon="inline-start" />
                      Sync
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    {providerItem?.cr_last_update_status || ""}
                    {providerItem?.cr_last_update_status == "Failed" &&
                    providerItem?.cr_last_update_message != ""
                      ? ", " + providerItem?.cr_last_update_message
                      : ""}
                    {providerItem?.cr_last_update_timestamp
                      ? `. Last updated ${dayjs.utc(providerItem.cr_last_update_timestamp).fromNow()}`
                      : ""}
                  </TooltipContent>
                </Tooltip>
              </div>
            </div>
          </main>
          <div ref={coderepoRef} className="flex-1 flex overflow-y-auto">
            <div className="w-full">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead className="text-right">Last Update</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {repositories?.map((repository) => (
                    <TableItem key={repository.id} repository={repository} />
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
          <Pagination
            limit={codeRepositoryPageSize}
            page={page}
            setPage={setPageAndScrollTop}
            total={total}
          />
        </div>
      </div>
    </Fragment>
  );
}

function TableItem({ repository }: { repository: ICodeRepositoryItem }) {
  return (
    <TableRow className="align-middle">
      <TableCell
        className="cursor-pointer"
        onClick={() => window.open(repository.clone_url, "_blank")}
      >
        <div className="truncate">
          <span className="font-medium">{repository.name}</span>
          <span className="text-muted-foreground font-normal ml-2">
            {repository.clone_url}
          </span>
        </div>
      </TableCell>
      <TableCell className="text-right text-muted-foreground cursor-pointer">
        <RelativeTime time={repository.pushed_at} />
      </TableCell>
    </TableRow>
  );
}
