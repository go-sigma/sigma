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
import { useDebounce } from "react-use";
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';

import Header from "../../components/Header";
import IMenu from "../../components/Menu";
import Notification from "../../components/Notification";
import OrderHeader from "../../components/OrderHeader";
import Pagination from "../../components/Pagination";
import Quota from "../../components/Quota";
import QuotaSimple from "../../components/QuotaSimple";
import Settings from "../../Settings";
import calcUnit from "../../utils/calcUnit";
import { useTranslation } from "../../i18n/useTranslation";
import { IHTTPError, INamespaceItem, IOrder, IRepositoryItem, IRepositoryList, IUserSelf } from "../../interfaces";
import { NamespaceRole, UserRole } from "../../interfaces/enums";
import TableItemDropdown from "../../components/Menu/TableItemDropdown";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export default function ({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
  const [repositoryList, setRepositoryList] = useState<IRepositoryList>({} as IRepositoryList);
  const [refresh, setRefresh] = useState({});
  const [page, setPage] = useState(1);
  const [searchRepository, setSearchRepository] = useState("");
  const [total, setTotal] = useState(0);

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId = searchParams.get('namespace_id');
  const [namespaceObj, setNamespaceObj] = useState<INamespaceItem>({} as INamespaceItem);

  useEffect(() => {
    if (!namespaceId) return;
    axios.get(`${localServer}/api/v1/namespaces/${namespaceId}`).then(response => {
      if (response.status == 200) setNamespaceObj(response.data as INamespaceItem);
    }).catch(() => {});
  }, []);

  const [repositoryText, setRepositoryText] = useState("");
  const [repositoryTextValid, setRepositoryTextValid] = useState(true);
  useEffect(() => { repositoryText != "" && setRepositoryTextValid(/^[a-z][0-9a-z-]{0,20}$/.test(repositoryText)) }, [repositoryText])
  const [descriptionText, setDescriptionText] = useState("");
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => { descriptionText != "" && setDescriptionTextValid(/^.{0,30}$/.test(descriptionText)) }, [descriptionText]);
  const [tagCountLimit, setTagCountLimit] = useState<string | number>(0);
  const [tagCountLimitValid, setTagCountLimitValid] = useState(true);
  useEffect(() => { setTagCountLimitValid(Number.isInteger(tagCountLimit) && parseInt(tagCountLimit.toString()) >= 0) }, [tagCountLimit])
  const [realSizeLimit, setRealSizeLimit] = useState(0);
  const [sizeLimit, setSizeLimit] = useState<string | number>(0);
  const [sizeLimitValid, setSizeLimitValid] = useState(true);
  const [sizeLimitUnit, setSizeLimitUnit] = useState("MiB");
  useEffect(() => { setSizeLimitValid(Number.isInteger(sizeLimit) && parseInt(sizeLimit.toString()) >= 0) }, [sizeLimit])
  useEffect(() => {
    let sl = 0;
    if (Number.isInteger(sizeLimit)) sl = parseInt(sizeLimit.toString());
    switch (sizeLimitUnit) {
      case "MiB": setRealSizeLimit(sl * 1 << 20); break;
      case "GiB": setRealSizeLimit(sl * 1 << 30); break;
      case "TiB": setRealSizeLimit(sl * 1 << 40); break;
    }
  }, [sizeLimit, sizeLimitUnit]);
  const [createRepositoryModal, setCreateRepositoryModal] = useState(false);

  const createRepository = () => {
    if (!(repositoryTextValid && descriptionTextValid && sizeLimitValid && tagCountLimitValid)) {
      Notification({ level: "warning", title: t("common.formValidateFailed"), message: t("common.checkForm") });
      return;
    }
    setCreateRepositoryModal(false);
    axios.post(localServer + `/api/v1/namespaces/${namespaceId}/repositories/`, {
      name: namespace + "/" + repositoryText, description: descriptionText, size_limit: realSizeLimit, tag_limit: tagCountLimit,
    } as IRepositoryItem).then(response => {
      if (response.status === 201) {
        setRepositoryText(""); setDescriptionText(""); setTagCountLimit(0); setSizeLimit(0);
        setRefresh({});
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    })
  }

  const validateRepository = () => {
    if (repositoryText === "") return;
    axios.get(localServer + `/api/v1/validators/reference?reference=${namespace}/${repositoryText}`).then(response => {
      setRepositoryTextValid(response?.status === 204);
    }).catch(() => setRepositoryTextValid(false));
  }
  useDebounce(validateRepository, 300, [repositoryText]);

  const [sizeOrder, setSizeOrder] = useState(IOrder.None);
  const [tagCountOrder, setTagCountOrder] = useState(IOrder.None);
  const [createdAtOrder, setCreatedAtOrder] = useState(IOrder.None);
  const [updatedAtOrder, setUpdatedAtOrder] = useState(IOrder.None);
  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");

  const resetOrder = () => {
    setSizeOrder(IOrder.None); setTagCountOrder(IOrder.None); setCreatedAtOrder(IOrder.None); setUpdatedAtOrder(IOrder.None);
  }

  const fetchRepository = () => {
    let url = localServer + `/api/v1/namespaces/${namespaceId}/repositories/?limit=${Settings.PageSize}&page=${page}`;
    if (searchRepository !== "") url += `&name=${searchRepository}`;
    if (sortName !== "") url += `&sort=${sortName}&method=${sortOrder.toString()}`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        setRepositoryList(response.data as IRepositoryList);
        setTotal((response.data as IRepositoryList).total);
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  }

  useEffect(fetchRepository, [refresh, page]);

  const [userObj, setUserObj] = useState<IUserSelf>({} as IUserSelf);

  useEffect(() => {
    axios.get(localServer + "/api/v1/users/self").then(response => {
      if (response.status === 200) setUserObj(response.data as IUserSelf);
    }).catch(() => {});
  }, []);

  const canManage = userObj.role == UserRole.Admin || userObj.role == UserRole.Root ||
    (namespaceObj.role != undefined && (namespaceObj.role == NamespaceRole.Admin || namespaceObj.role == NamespaceRole.Manager));

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet><title>{t("common.repositories")}</title></Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background">
        <IMenu localServer={localServer} item="repositories" namespace={namespace} namespace_id={namespaceId || ""} />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title={t("header.repository")}
              props={
                <Tabs value="repositoryList" className="h-full">
                  <TabsList className="h-full">
                    <TabsTrigger value="summary" render={<Link to={`/namespaces/${namespace}/namespace-summary?namespace_id=${namespaceId}`} />}>{t("common.summary")}</TabsTrigger>
                    <TabsTrigger value="repositoryList">{t("common.repositoryList")}</TabsTrigger>
                    <TabsTrigger value="members" render={<Link to={`/namespaces/${namespace}/members?namespace_id=${namespaceId}`} />}>{t("common.members")}</TabsTrigger>
                    <TabsTrigger value="daemon" render={<Link to={`/namespaces/${namespace}/daemon-tasks?namespace_id=${namespaceId}`} />}>{t("common.daemonTask")}</TabsTrigger>
                    <TabsTrigger value="webhook" render={<Link to={`/namespaces/${namespace}/webhooks?namespace_id=${namespaceId}`} />}>{t("common.webhook")}</TabsTrigger>
                  </TabsList>
                </Tabs>
              }
            />
            <div className="pt-1 pb-1 flex justify-between items-center min-h-[60px]">
              <div className="pr-2 pl-2">
                <div className="relative flex items-center">
                  <Label htmlFor="repositorySearch" className="absolute -top-2 left-2 inline-block bg-background px-1 text-xs font-medium text-foreground z-10">
                    {t("common.repository")}
                  </Label>
                  <Input id="repositorySearch" placeholder={t("repository.searchPlaceholder")} value={searchRepository}
                    onChange={e => setSearchRepository(e.target.value)} onKeyDown={e => { if (e.key == "Enter") fetchRepository() }}
                    className="h-10 pr-14" />
                  <kbd className="absolute inset-y-0 right-0 flex items-center py-1.5 pr-3 text-xs text-muted-foreground">enter</kbd>
                </div>
              </div>
              <div className="pr-2 pl-2">
                <Button onClick={() => canManage && setCreateRepositoryModal(true)} disabled={!canManage}>{t("common.create")}</Button>
              </div>
            </div>
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className="w-full">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead><span className="lg:pl-2">{t("repository.table.name")}</span></TableHead>
                    <TableHead className="text-right"><OrderHeader text={t("repository.table.size")} orderStatus={sizeOrder} setOrder={e => { resetOrder(); setSizeOrder(e); setSortOrder(e); setSortName("size"); }} /></TableHead>
                    <TableHead className="text-right"><OrderHeader text={t("repository.table.tagCount")} orderStatus={tagCountOrder} setOrder={e => { resetOrder(); setTagCountOrder(e); setSortOrder(e); setSortName("tag_count"); }} /></TableHead>
                    <TableHead className="text-right"><OrderHeader text={t("repository.table.createdAt")} orderStatus={createdAtOrder} setOrder={e => { resetOrder(); setCreatedAtOrder(e); setSortOrder(e); setSortName("created_at"); }} /></TableHead>
                    <TableHead className="text-right"><OrderHeader text={t("repository.table.updatedAt")} orderStatus={updatedAtOrder} setOrder={e => { resetOrder(); setUpdatedAtOrder(e); setSortOrder(e); setSortName("updated_at"); }} /></TableHead>
                    <TableHead className="text-right">{t("common.action")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {repositoryList.items?.map((repo, index) => (
                    <TableItem key={index} localServer={localServer} index={index} user={userObj} namespace={namespaceObj} repository={repo} setRefresh={setRefresh} />
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
          <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
        </div>
      </div>

      {/* Create Repository Modal */}
      <Dialog open={createRepositoryModal} onOpenChange={setCreateRepositoryModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader><DialogTitle>{t("common.create")} Repository</DialogTitle></DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label><span className="text-destructive">*</span>Name</Label>
              <div className="flex">
                <span className="inline-flex items-center rounded-l-md border border-r-0 border-border bg-muted px-3 text-sm text-muted-foreground">{namespace}/</span>
                <Input placeholder="2-20 lowercase characters" value={repositoryText} onChange={e => setRepositoryText(e.target.value)} className="rounded-l-none" data-invalid={!repositoryTextValid ? true : undefined} />
              </div>
              {!repositoryTextValid && <p className="text-xs text-destructive">Not a valid repository name.</p>}
            </div>
            <div className="grid gap-2">
              <Label>Description</Label>
              <Input placeholder="30 characters" value={descriptionText} onChange={e => setDescriptionText(e.target.value)} data-invalid={!descriptionTextValid ? true : undefined} />
              {!descriptionTextValid && <p className="text-xs text-destructive">Not a valid description.</p>}
            </div>
            <div className="grid gap-2">
              <Label>Size limit</Label>
              <div className="flex gap-2">
                <Input type="number" placeholder="0 means no limit" value={sizeLimit} onChange={e => setSizeLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} data-invalid={!sizeLimitValid ? true : undefined} className="flex-1" />
                <Select value={sizeLimitUnit} onValueChange={(v) => { if (v) setSizeLimitUnit(v); }}>
                  <SelectTrigger className="w-20"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="MiB">MiB</SelectItem>
                    <SelectItem value="GiB">GiB</SelectItem>
                    <SelectItem value="TiB">TiB</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              {!sizeLimitValid && <p className="text-xs text-destructive">Not a valid size limit.</p>}
            </div>
            <div className="grid gap-2">
              <Label>Tag count limit</Label>
              <Input type="number" placeholder="0 means no limit" value={tagCountLimit} onChange={e => setTagCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} data-invalid={!tagCountLimitValid ? true : undefined} />
              {!tagCountLimitValid && <p className="text-xs text-destructive">Not a valid tag count limit.</p>}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateRepositoryModal(false)}>Cancel</Button>
            <Button onClick={createRepository}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Fragment>
  )
}

function TableItem({ localServer, index, user, namespace: ns, repository, setRefresh }: {
  localServer: string; index: number; user: IUserSelf; namespace: INamespaceItem;
  repository: IRepositoryItem; setRefresh: (param: any) => void;
}) {
  const navigate = useNavigate();

  const [deleteRepositoryModal, setDeleteRepositoryModal] = useState(false);
  const [descriptionText, setDescriptionText] = useState(repository.description);
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => { descriptionText != "" && setDescriptionTextValid(/^.{0,30}$/.test(descriptionText)) }, [descriptionText]);
  const [tagCountLimit, setTagCountLimit] = useState<string | number>(repository.tag_limit);
  const [tagCountLimitValid, setTagCountLimitValid] = useState(true);
  useEffect(() => { setTagCountLimitValid(Number.isInteger(tagCountLimit) && parseInt(tagCountLimit.toString()) >= 0) }, [tagCountLimit])
  const [realSizeLimit, setRealSizeLimit] = useState(repository.size_limit);
  let calcUnitObj = calcUnit(repository.size_limit);
  const [sizeLimit, setSizeLimit] = useState<string | number>(calcUnitObj.size);
  const [sizeLimitValid, setSizeLimitValid] = useState(true);
  const [sizeLimitUnit, setSizeLimitUnit] = useState(calcUnitObj.unit);
  useEffect(() => { setSizeLimitValid(Number.isInteger(sizeLimit) && parseInt(sizeLimit.toString()) >= 0) }, [sizeLimit]);
  useEffect(() => {
    let sl = 0;
    if (Number.isInteger(sizeLimit)) sl = parseInt(sizeLimit.toString());
    switch (sizeLimitUnit) {
      case "MiB": setRealSizeLimit(sl * 1 << 20); break;
      case "GiB": setRealSizeLimit(sl * 1 << 30); break;
      case "TiB": setRealSizeLimit(sl * 1 << 40); break;
    }
  }, [sizeLimit, sizeLimitUnit]);
  const [updateRepositoryModal, setUpdateRepositoryModal] = useState(false);

  const canManage = user.role == UserRole.Admin || user.role == UserRole.Root ||
    (ns.role != undefined && (ns.role == NamespaceRole.Admin || ns.role == NamespaceRole.Manager));

  const updateRepository = () => {
    if (!(descriptionTextValid && sizeLimitValid && tagCountLimitValid)) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    setUpdateRepositoryModal(false);
    axios.put(localServer + `/api/v1/namespaces/${ns.id}/repositories/${repository.id}`, {
      description: descriptionText, size_limit: realSizeLimit, tag_limit: tagCountLimit,
    } as IRepositoryItem).then(response => {
      if (response.status === 204) setRefresh({});
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    })
  }

  const deleteRepository = () => {
    axios.delete(localServer + `/api/v1/namespaces/${ns.id}/repositories/${repository.id}`).then(response => {
      if (response.status === 204) setRefresh({});
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    })
  }

  return (
    <>
      <TableRow>
        <TableCell className="cursor-pointer" onClick={() => navigate(`/namespaces/${ns.name}/repository/tags?namespace_id=${repository.namespace_id}&repository=${repository.name}&repository_id=${repository.id}`)}>
          <div className="truncate hover:text-muted-foreground">
            <span className="font-medium">{repository.name}</span>
            <span className="text-muted-foreground font-normal ml-4">{repository.description}</span>
          </div>
        </TableCell>
        <TableCell className="text-right"><Quota current={repository.size} limit={repository.size_limit} /></TableCell>
        <TableCell className="text-right"><QuotaSimple current={repository.tag_count} limit={repository.tag_limit} /></TableCell>
        <TableCell className="text-right text-muted-foreground">{dayjs.utc(repository.created_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}</TableCell>
        <TableCell className="text-right text-muted-foreground">{dayjs.utc(repository.updated_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}</TableCell>
        <TableCell className="text-center">
          <TableItemDropdown
            items={[
              { name: "Update", disable: !canManage, onClick: () => setUpdateRepositoryModal(true) },
              { name: "Delete", disable: !canManage, warn: true, onClick: () => setDeleteRepositoryModal(true) },
            ]}
          />
        </TableCell>
      </TableRow>

      {/* Update Repository Modal */}
      <Dialog open={updateRepositoryModal} onOpenChange={setUpdateRepositoryModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader><DialogTitle>Update Repository</DialogTitle></DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label>Name</Label>
              <Input value={repository.name} disabled />
            </div>
            <div className="grid gap-2">
              <Label>Description</Label>
              <Input placeholder="30 characters" value={descriptionText} onChange={e => setDescriptionText(e.target.value)} data-invalid={!descriptionTextValid ? true : undefined} />
            </div>
            <div className="grid gap-2">
              <Label>Size limit</Label>
              <div className="flex gap-2">
                <Input type="number" placeholder="0 means no limit" value={sizeLimit} onChange={e => setSizeLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} data-invalid={!sizeLimitValid ? true : undefined} className="flex-1" />
                <Select value={sizeLimitUnit} onValueChange={(v) => { if (v) setSizeLimitUnit(v); }}>
                  <SelectTrigger className="w-20"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="MiB">MiB</SelectItem>
                    <SelectItem value="GiB">GiB</SelectItem>
                    <SelectItem value="TiB">TiB</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid gap-2">
              <Label>Tag count limit</Label>
              <Input type="number" placeholder="0 means no limit" value={tagCountLimit} onChange={e => setTagCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} data-invalid={!tagCountLimitValid ? true : undefined} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setUpdateRepositoryModal(false)}>Cancel</Button>
            <Button onClick={updateRepository}>Update</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Repository Dialog */}
      <AlertDialog open={deleteRepositoryModal} onOpenChange={setDeleteRepositoryModal}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete repository</AlertDialogTitle>
            <AlertDialogDescription>Are you sure you want to delete the repository <strong>{repository.name}</strong>?</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={deleteRepository} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}