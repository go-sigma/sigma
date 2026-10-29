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
import { EllipsisVertical, FileText, Info, RotateCw, Square } from "lucide-react";
import { useEffect, useState } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { useNavigate } from 'react-router-dom';

import Header from "@/components/Header";
import Menu from "@/components/Menu";
import Notification from "@/components/Notification";
import OrderHeader from "@/components/OrderHeader";
import Pagination from "@/components/Pagination";
import BuildDurationChart from "@/components/BuildDurationChart";
import RepositoryTabs from "@/components/RepositoryTabs";
import RunnerLogView from "@/components/RunnerLogView";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Tooltip as UITooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useTranslation } from "@/i18n/useTranslation";
import {
  IBuilderItem,
  IBuilderRunnerItem,
  IBuilderRunnerList,
  IHTTPError,
  IOrder,
  IRepositoryItem,
} from "@/interfaces";
import Settings from "@/Settings";
import { Tooltip } from '@/utils';

export default function ({ localServer }: { localServer: string }) {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repository_id = searchParams.get("repository_id") || "";
  const namespaceId = searchParams.get("namespace_id") || "";

  const [builderObj, setBuilderObj] = useState<IBuilderItem>();
  const [repositoryObj, setRepositoryObj] = useState<IRepositoryItem>();

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const [branchText, setBranchText] = useState("");
  const [branchTextValid, setBranchTextValid] = useState(true);
  useEffect(() => {
    if (branchText != "") {
      setBranchTextValid(/^[a-zA-Z0-9_-]{1,64}$/.test(branchText));
    }
  }, [branchText]);

  useEffect(() => {
    axios.get(localServer + `/api/v1/namespaces/${namespaceId}/repositories/${repository_id}`).then(response => {
      if (response?.status === 200) {
        const r = response.data as IRepositoryItem;
        setRepositoryObj(r);
        if (r.builder !== undefined && r.builder !== null) {
          setBuilderObj(r.builder);
        }
        setBranchText(r.builder?.scm_branch || "");
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespace, repository_id, localServer, namespaceId]);

  const [runnerObjs, setRunnerObjs] = useState<IBuilderRunnerItem[]>()

  const [createRunnerModal, setCreateRunnerModal] = useState(false);
  const [tagTemplateText, setTagTemplateText] = useState("");
  const [tagTemplateTextValid, setTagTemplateTextValid] = useState(true);
  useEffect(() => {
    if (tagTemplateText == "") {
      return;
    }
    axios.get(localServer + `/api/v1/validators/tag?tag=${tagTemplateText}`).then(response => {
      if (response?.status === 204) {
        setTagTemplateTextValid(true);
      } else {
        setTagTemplateTextValid(false);
      }
    }).catch(error => {
      console.log(error);
      setTagTemplateTextValid(false);
    });
  }, [tagTemplateText, localServer]);

  const [descriptionText, setDescriptionText] = useState("");
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => {
    if (descriptionText != "") {
      setDescriptionTextValid(/^.{0,50}$/.test(descriptionText));
    }
  }, [descriptionText]);

  useEffect(() => {
    if (repositoryObj != undefined) {
      let tag = localStorage.getItem(`${repositoryObj?.name}-tag`);
      if (tag != null) {
        setTagTemplateText(tag);
      }
    }
  }, [repositoryObj]);

  const createRunner = () => {
    if (builderObj == undefined) {
      return;
    }
    if (!(tagTemplateTextValid && branchTextValid && descriptionTextValid)) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    if (builderObj.source !== "Dockerfile" && branchText === "") {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    const data: { [key: string]: any } = {
      raw_tag: tagTemplateText,
    };
    if (builderObj.source !== "Dockerfile") {
      data["scm_branch"] = branchText;
    }
    if (descriptionText !== "") {
      data["description"] = descriptionText;
    }
    if (repositoryObj != undefined && repositoryObj?.name != "") {
      localStorage.setItem(`${repositoryObj?.name}-tag`, `${tagTemplateText}`);
    }
    axios.post(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj.id}/runners/run`, data).then(response => {
      if (response?.status === 201) {
        // The new run shows up in the list below, so close the dialog instead of following its log.
        setCreateRunnerModal(false);
        setDescriptionText("");
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const [refreshState, setRefreshState] = useState({});

  // logRunner is the run shown in the log dialog; undefined keeps the dialog closed.
  const [logRunner, setLogRunner] = useState<IBuilderRunnerItem>();

  const rerunRunner = (runner: IBuilderRunnerItem) => {
    axios.get(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${runner.builder_id}/runners/${runner.id}/rerun`).then(response => {
      if (response?.status === 200) {
        // The new run shows up in the list below.
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const stopRunner = (runner: IBuilderRunnerItem) => {
    axios.get(localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${runner.builder_id}/runners/${runner.id}/stop`).then(response => {
      if (response?.status === 204) {
        setLogRunner(previous => (previous !== undefined && previous.id === runner.id ? { ...previous, status: "Stopping" } : previous));
        setRefreshState({});
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }

  const [createdAtOrder, setCreatedAtOrder] = useState(IOrder.None);
  const [costOrder, setCostOrder] = useState(IOrder.None);
  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");

  const resetOrder = () => {
    setCostOrder(IOrder.None);
    setCreatedAtOrder(IOrder.None);
  }

  useEffect(() => {
    if (builderObj === undefined) {
      return;
    }
    let url = localServer + `/api/v1/namespaces/${repositoryObj?.namespace_id}/repositories/${repository_id}/builders/${builderObj.id}/runners/?limit=${Settings.PageSize}&page=${page}`
    if (sortName !== "" && sortOrder !== IOrder.None) {
      url += `&sort=${sortName}&method=${sortOrder.toString()}`
    }
    axios.get(url).then(response => {
      if (response?.status === 200) {
        const r = response.data as IBuilderRunnerList;
        setRunnerObjs(r.items);
        setTotal(r.total);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }, [namespace, repository_id, builderObj, refreshState, sortOrder, sortName, localServer, page, repositoryObj?.namespace_id]);

  return (
    <>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Runners</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen max-h-screen flex overflow-hidden bg-background">
        <Menu localServer={localServer} item="repositories" namespace={namespace} repository={repositoryObj?.name} selfClick={true} />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title="Repository"
              props={
                <RepositoryTabs
                  namespace={namespace || ""}
                  namespaceId={namespaceId}
                  repository={repositoryObj?.name || ""}
                  repositoryId={repository_id}
                  active="builder"
                />
              } />
            <div className="pt-2 pb-2 flex flex-row-reverse justify-between">
              <div className="px-4">
                {
                  builderObj === undefined ? (
                    <button className="my-auto px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/80 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring sm:order-1 sm:ml-3"
                      onClick={() => { navigate(`/namespaces/${namespace}/repository/builder?namespace=${namespace}&namespace_id=${repositoryObj?.namespace_id}&repository=${repositoryObj?.name}&repository_id=${repositoryObj?.id}&namespace_stick=true&repository_stick=true`); }}
                    >Configure</button>
                  ) : null
                }
                {
                  builderObj === undefined ? null : (
                    <button className="my-auto px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/80 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring sm:order-1 sm:ml-3"
                      onClick={() => { navigate(`/builders/setup/${repositoryObj?.builder?.id}?namespace=${namespace}&namespace_id=${repositoryObj?.namespace_id}&repository=${repositoryObj?.name}&repository_id=${repositoryObj?.id}&namespace_stick=true&repository_stick=true&back_to=/namespaces/${namespace}/repository/builder?repository=${repositoryObj?.name}%26repository_id=${repositoryObj?.id}%26namespace_id=${namespaceId}`); }}
                    >Update</button>
                  )
                }
                {
                  builderObj === undefined ? null : (
                    <button className="my-auto px-4 py-2 h-10 border border-transparent shadow-sm text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/80 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring sm:order-1 sm:ml-3"
                      onClick={() => { setCreateRunnerModal(true) }}
                    >Build</button>
                  )
                }
              </div>
            </div>
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className="border-t w-full">
              {
                builderObj === undefined ? (
                  <div className="my-2 mx-2 text-gray-500 text-md">
                    Please <Link to={`/namespaces/${namespace}/repository/builder?namespace=${namespace}&namespace_id=${repositoryObj?.namespace_id}&repository=${repositoryObj?.name}&repository_id=${repositoryObj?.id}&namespace_stick=true&repository_stick=true`} className="hover:underline-offset-1">configure</Link> the builder first.
                  </div>
                ) : (
                  <>
                    <div className="px-4 pt-4 pb-4">
                      <BuildDurationChart
                        localServer={localServer}
                        namespaceId={repositoryObj?.namespace_id || namespaceId}
                        repositoryId={repository_id}
                        builderId={builderObj.id}
                      />
                    </div>
                    <table className="min-w-full flex-1">
                    <thead>
                      <tr>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <span className="lg:pl-2">Tag</span>
                        </th>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-left text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <span>Status</span>
                        </th>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <OrderHeader text={"Elapsed"}
                            orderStatus={costOrder} setOrder={e => {
                              resetOrder();
                              setCostOrder(e);
                              setSortOrder(e);
                              setSortName("duration");
                            }} />
                        </th>
                        <th className="sticky top-0 z-10 px-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          <OrderHeader text={"Created at"}
                            orderStatus={createdAtOrder} setOrder={e => {
                              resetOrder();
                              setCreatedAtOrder(e);
                              setSortOrder(e);
                              setSortName("created_at");
                            }} />
                        </th>
                        <th className="sticky top-0 z-10 pr-6 py-3 border-gray-200 bg-gray-100 text-right text-xs font-medium text-gray-500 tracking-wider whitespace-nowrap">
                          Action
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-100 max-h-max">
                      {
                        runnerObjs?.map(runnerObj => {
                          return (
                            <TableItem key={runnerObj.id} runnerObj={runnerObj} onViewLog={setLogRunner} onRerun={rerunRunner} onStop={stopRunner} />
                          );
                        })
                      }
                    </tbody>
                  </table>
                  </>
                )
              }
            </div>
          </div>
          {
            builderObj === undefined ? null : (
              <div style={{ marginTop: "auto" }}>
                <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
              </div>
            )
          }
        </div>
      </div >

      <Dialog open={createRunnerModal} onOpenChange={setCreateRunnerModal}>
        <DialogContent className="sm:max-w-lg" aria-label="Build">
          <DialogHeader>
            <DialogTitle>Build</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4">
            <div className="grid gap-2">
              <div className="flex items-center gap-1.5">
                <Label htmlFor="tagTemplate">
                  <span className="text-destructive">*</span>Tag
                </Label>
                <UITooltip>
                  <TooltipTrigger>
                    <Info className="size-4 text-muted-foreground" aria-hidden="true" />
                  </TooltipTrigger>
                  <TooltipContent>{t("builder.tagTemplateHint")}</TooltipContent>
                </UITooltip>
              </div>
              <Input
                type="text"
                id="tagTemplate"
                name="tagTemplate"
                placeholder="tag template"
                autoComplete="off"
                value={tagTemplateText}
                aria-invalid={!tagTemplateTextValid}
                onChange={e => setTagTemplateText(e.target.value)}
              />
              {tagTemplateTextValid ? null : (
                <p className="text-xs text-destructive">
                  Not a valid tag template, you can try 'main', '&#123;&#123;.ScmTag&#125;&#125;', '&#123;&#123;.ScmBranch&#125;&#125;'.
                </p>
              )}
            </div>
            {
              builderObj?.source !== "Dockerfile" ? (
                <div className="grid gap-2">
                  <Label htmlFor="branch">
                    <span className="text-destructive">*</span>Branch
                  </Label>
                  <Input
                    type="text"
                    id="branch"
                    name="branch"
                    placeholder="1-64 characters"
                    value={branchText}
                    aria-invalid={!branchTextValid}
                    onChange={e => setBranchText(e.target.value)}
                  />
                  {branchTextValid ? null : (
                    <p className="text-xs text-destructive">
                      Not a valid branch, you can try 'main', 'master', 'dev'.
                    </p>
                  )}
                </div>
              ) : null
            }
            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                name="description"
                placeholder="30 characters"
                value={descriptionText}
                aria-invalid={!descriptionTextValid}
                onChange={e => setDescriptionText(e.target.value)}
              />
              {descriptionTextValid ? null : (
                <p className="text-xs text-destructive">
                  Not a valid description, max 50 characters.
                </p>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateRunnerModal(false)}>
              Cancel
            </Button>
            <Button onClick={() => createRunner()}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={logRunner !== undefined} onOpenChange={(open) => { if (!open) setLogRunner(undefined); }}>
        <DialogContent className="sm:max-w-5xl" aria-label="Build log">
          <DialogHeader>
            <DialogTitle>Build log</DialogTitle>
          </DialogHeader>
          {logRunner === undefined ? null : (
            <RunnerLogView
              localServer={localServer}
              namespaceId={repositoryObj?.namespace_id || namespaceId}
              repositoryId={repository_id}
              builderId={logRunner.builder_id}
              runnerId={logRunner.id}
              status={logRunner.status}
              className="h-[60vh] w-full rounded-none bg-black p-1"
            />
          )}
          <DialogFooter>
            <Button
              variant="destructive"
              disabled={logRunner === undefined || !["Pending", "Building"].includes(logRunner.status)}
              onClick={() => { if (logRunner !== undefined) stopRunner(logRunner); }}
            >
              Stop
            </Button>
            <Button variant="outline" onClick={() => setLogRunner(undefined)}>
              Cancel
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

    </>
  )
}

function TableItem({ runnerObj, onViewLog, onRerun, onStop }: {
  runnerObj: IBuilderRunnerItem;
  onViewLog: (runner: IBuilderRunnerItem) => void;
  onRerun: (runner: IBuilderRunnerItem) => void;
  onStop: (runner: IBuilderRunnerItem) => void;
}) {
  return (
    <tr className="align-middle">
      <td className="px-6 py-4 w-5/6 whitespace-nowrap text-sm font-medium text-gray-900 cursor-pointer"
        onClick={() => onViewLog(runnerObj)}
      >
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {runnerObj.tag !== null ? runnerObj.tag : runnerObj.raw_tag}
              <span className="text-gray-500 font-normal ml-2">{runnerObj.description}</span>
            </span>
          </div>
        </div>
      </td>
      <td className={`px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-left${runnerObj.status_message ? " cursor-pointer" : ""}`}>
        {runnerObj.status_message ? (
          <div id={`tooltip-message-${runnerObj.id}`} role="tooltip" className="fixed z-10 invisible inline-block max-w-md px-3 py-2 text-left text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip whitespace-pre-wrap break-words dark:bg-gray-700">
            {runnerObj.status_message}
          </div>
        ) : null}
        <div className={runnerObj.status === "Failed" ? "text-red-600" : ""}
          id={`target-tooltip-${runnerObj.id}`}
          // The failure reason is only rendered when the run stored one, so an empty tooltip
          // cannot pop up for runs without a message.
          onClick={runnerObj.status_message ? e => {
            let tooltip = new Tooltip(document.getElementById(`tooltip-message-${runnerObj.id}`),
              document.getElementById(`target-tooltip-${runnerObj.id}`), { triggerType: "click" });
            tooltip.show();
            e.stopPropagation();
          } : undefined}
        >
          {runnerObj.status}
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {runnerObj.duration || "-"}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right cursor-pointer">
        {dayjs.utc(runnerObj.created_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-right">
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button variant="ghost" size="icon-sm" aria-label="Open actions">
                <EllipsisVertical className="size-4" />
              </Button>
            }
          />
          <DropdownMenuContent align="end" className="w-40!">
            <DropdownMenuItem onClick={() => onViewLog(runnerObj)}>
              <FileText />
              <span>View log</span>
            </DropdownMenuItem>
            {
              runnerObj?.status === "Success" || runnerObj?.status === "Failed" || runnerObj?.status === "Stopped" ? (
                <DropdownMenuItem onClick={() => onRerun(runnerObj)}>
                  <RotateCw />
                  <span>Rerun</span>
                </DropdownMenuItem>
              ) : (
                <DropdownMenuItem disabled={runnerObj?.status === "Stopping"} onClick={() => onStop(runnerObj)}>
                  <Square />
                  <span>Stop</span>
                </DropdownMenuItem>
              )
            }
          </DropdownMenuContent>
        </DropdownMenu>
      </td>
    </tr>
  )
}
