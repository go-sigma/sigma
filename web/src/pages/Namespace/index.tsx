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
import dayjs from 'dayjs';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { useNavigate } from 'react-router-dom';

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
import { IHTTPError, INamespaceItem, INamespaceList, IOrder, IUserSelf } from "../../interfaces";
import { NamespaceRole, UserRole } from "../../interfaces/enums";
import TableItemDropdown from "../../components/Menu/TableItemDropdown";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
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
import { Badge } from "@/components/ui/badge";

export default function Namespace({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
  const [namespaceList, setNamespaceList] = useState<INamespaceList>({} as INamespaceList);

  const [namespaceText, setNamespaceText] = useState("");
  const [namespaceTextValid, setNamespaceTextValid] = useState(true);
  useEffect(() => { namespaceText != "" && setNamespaceTextValid(/^[a-z][0-9a-z-]{0,20}$/.test(namespaceText)) }, [namespaceText])
  const [descriptionText, setDescriptionText] = useState("");
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => { descriptionText != "" && setDescriptionTextValid(/^.{0,30}$/.test(descriptionText)) }, [descriptionText]);
  const [repositoryCountLimit, setRepositoryCountLimit] = useState<string | number>(0);
  const [repositoryCountLimitValid, setRepositoryCountLimitValid] = useState(true);
  useEffect(() => { setRepositoryCountLimitValid(Number.isInteger(repositoryCountLimit) && parseInt(repositoryCountLimit.toString()) >= 0) }, [repositoryCountLimit]);
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
    if (Number.isInteger(sizeLimit)) {
      sl = parseInt(sizeLimit.toString());
    }
    switch (sizeLimitUnit) {
      case "MiB": setRealSizeLimit(sl * 1 << 20); break;
      case "GiB": setRealSizeLimit(sl * 1 << 30); break;
      case "TiB": setRealSizeLimit(sl * 1 << 40); break;
    }
  }, [sizeLimit, sizeLimitUnit])
  const [namespaceVisibility, setNamespaceVisibility] = useState("private");

  const [refresh, setRefresh] = useState({});
  const [page, setPage] = useState(1);
  const [searchNamespace, setSearchNamespace] = useState("");
  const [total, setTotal] = useState(0);

  const [sizeOrder, setSizeOrder] = useState(IOrder.None);
  const [repositoryCountOrder, setRepositoryOrder] = useState(IOrder.None);
  const [tagCountOrder, setTagCountOrder] = useState(IOrder.None);
  const [createdAtOrder, setCreatedAtOrder] = useState(IOrder.None);
  const [updatedAtOrder, setUpdatedAtOrder] = useState(IOrder.None);
  const [sortOrder, setSortOrder] = useState(IOrder.None);
  const [sortName, setSortName] = useState("");

  const resetOrder = () => {
    setSizeOrder(IOrder.None);
    setRepositoryOrder(IOrder.None);
    setTagCountOrder(IOrder.None);
    setCreatedAtOrder(IOrder.None);
    setUpdatedAtOrder(IOrder.None);
  }

  const [createNamespaceModal, setCreateNamespaceModal] = useState(false);

  const fetchNamespace = () => {
    let url = localServer + `/api/v1/namespaces/?limit=${Settings.PageSize}&page=${page}`;
    if (searchNamespace !== "") url += `&name=${searchNamespace}`;
    if (sortName !== "") url += `&sort=${sortName}&method=${sortOrder.toString()}`
    axios.get(url).then(response => {
      if (response?.status === 200) {
        setNamespaceList(response.data as INamespaceList);
        setTotal((response.data as INamespaceList).total);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  }

  useEffect(() => { fetchNamespace() }, [refresh, page, sortOrder, sortName]);

  const [userObj, setUserObj] = useState<IUserSelf>({} as IUserSelf);

  useEffect(() => {
    axios.get(localServer + "/api/v1/users/self").then(response => {
      if (response.status === 200) {
        setUserObj(response.data as IUserSelf);
      }
    }).catch(() => {});
  }, []);

  const createNamespace = () => {
    if (!(namespaceTextValid && descriptionTextValid && sizeLimitValid && repositoryCountLimitValid && tagCountLimitValid)) {
      Notification({ level: "warning", title: t("common.formValidateFailed"), message: t("common.checkForm") });
      return;
    }
    setCreateNamespaceModal(false);
    axios.post(localServer + '/api/v1/namespaces/', {
      name: namespaceText, description: descriptionText, size_limit: realSizeLimit,
      repository_limit: repositoryCountLimit, tag_limit: tagCountLimit, visibility: namespaceVisibility,
    } as INamespaceItem).then(response => {
      if (response.status === 201) {
        setNamespaceText(""); setDescriptionText(""); setNamespaceVisibility("private");
        setRepositoryCountLimit(0); setTagCountLimit(0); setSizeLimit(0);
        setRefresh({});
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    })
  }

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet><title>{t("common.namespaces")}</title></Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background">
        <IMenu localServer={localServer} item="namespaces" />
        <div className="flex flex-col flex-1 max-h-screen">
          <main className="relative z-0 focus:outline-none">
            <Header title={t("header.namespace")} />
            <div className="pt-2 pb-2 flex justify-between items-center">
              <div className="pr-2 pl-2">
                <div className="relative flex items-center">
                  <Label htmlFor="namespaceSearch" className="absolute -top-2 left-2 inline-block bg-background px-1 text-xs font-medium text-foreground z-10">
                    {t("common.namespace")}
                  </Label>
                  <Input
                    id="namespaceSearch"
                    placeholder={t("namespace.searchPlaceholder")}
                    value={searchNamespace}
                    onChange={e => setSearchNamespace(e.target.value)}
                    onKeyDown={e => { if (e.key == "Enter") fetchNamespace() }}
                    className="h-10 pr-14"
                  />
                  <kbd className="absolute inset-y-0 right-0 flex items-center py-1.5 pr-3 text-xs text-muted-foreground">enter</kbd>
                </div>
              </div>
              <div className="pr-2 pl-2">
                <Button onClick={() => setCreateNamespaceModal(true)}>{t("common.create")}</Button>
              </div>
            </div>
          </main>
          <div className="flex-1 flex overflow-y-auto">
            <div className="w-full">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead><span className="lg:pl-2">{t("namespace.table.name")}</span></TableHead>
                    <TableHead className="text-right"><OrderHeader text={t("namespace.table.size")} orderStatus={sizeOrder} setOrder={e => { resetOrder(); setSizeOrder(e); setSortOrder(e); setSortName("size"); }} /></TableHead>
                    <TableHead className="text-right"><OrderHeader text={t("namespace.table.repositoryCount")} orderStatus={repositoryCountOrder} setOrder={e => { resetOrder(); setRepositoryOrder(e); setSortOrder(e); setSortName("repository_count"); }} /></TableHead>
                    <TableHead className="text-right"><OrderHeader text={t("namespace.table.tagCount")} orderStatus={tagCountOrder} setOrder={e => { resetOrder(); setTagCountOrder(e); setSortOrder(e); setSortName("tag_count"); }} /></TableHead>
                    <TableHead className="text-right">{t("namespace.table.visibility")}</TableHead>
                    <TableHead className="text-right"><OrderHeader text={t("namespace.table.createdAt")} orderStatus={createdAtOrder} setOrder={e => { resetOrder(); setCreatedAtOrder(e); setSortOrder(e); setSortName("created_at"); }} /></TableHead>
                    <TableHead className="text-right"><OrderHeader text={t("namespace.table.updatedAt")} orderStatus={updatedAtOrder} setOrder={e => { resetOrder(); setUpdatedAtOrder(e); setSortOrder(e); setSortName("updated_at"); }} /></TableHead>
                    <TableHead className="text-right">{t("common.action")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {namespaceList.items?.map((ns, index) => (
                    <TableItem key={ns.id} index={index} user={userObj} namespace={ns} localServer={localServer} setRefresh={setRefresh} />
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
          <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
        </div>
      </div>

      {/* Create Namespace Modal */}
      <Dialog open={createNamespaceModal} onOpenChange={setCreateNamespaceModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{t("common.create")} {t("common.namespace")}</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label><span className="text-destructive">*</span>Name</Label>
              <Input placeholder="2-20 lowercase characters" value={namespaceText} onChange={e => setNamespaceText(e.target.value)} data-invalid={!namespaceTextValid ? true : undefined} />
              {!namespaceTextValid && <p className="text-xs text-destructive">Not a valid namespace name, 2-20 lowercase characters.</p>}
            </div>
            <div className="grid gap-2">
              <Label>Description</Label>
              <Input placeholder="30 characters" value={descriptionText} onChange={e => setDescriptionText(e.target.value)} data-invalid={!descriptionTextValid ? true : undefined} />
              {!descriptionTextValid && <p className="text-xs text-destructive">Not a valid description, max 30 characters.</p>}
            </div>
            <div className="grid gap-2">
              <Label>Visibility</Label>
              <Select value={namespaceVisibility} onValueChange={(v) => { if (v) setNamespaceVisibility(v); }}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="private">Private</SelectItem>
                  <SelectItem value="public">Public</SelectItem>
                </SelectContent>
              </Select>
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
              {!sizeLimitValid && <p className="text-xs text-destructive">Not a valid size limit, should be non-negative integer.</p>}
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Repository count limit</Label>
                <Input type="number" placeholder="0 means no limit" value={repositoryCountLimit} onChange={e => setRepositoryCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} data-invalid={!repositoryCountLimitValid ? true : undefined} />
                {!repositoryCountLimitValid && <p className="text-xs text-destructive">Not a valid repository count limit.</p>}
              </div>
              <div className="grid gap-2">
                <Label>Tag count limit</Label>
                <Input type="number" placeholder="0 means no limit" value={tagCountLimit} onChange={e => setTagCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} data-invalid={!tagCountLimitValid ? true : undefined} />
                {!tagCountLimitValid && <p className="text-xs text-destructive">Not a valid tag count limit.</p>}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateNamespaceModal(false)}>Cancel</Button>
            <Button onClick={createNamespace}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Fragment>
  )
}

function TableItem({ localServer, index, user, namespace: ns, setRefresh }: { localServer: string, index: number, user: IUserSelf, namespace: INamespaceItem, setRefresh: (param: any) => void }) {
  const navigate = useNavigate();

  const [updateNamespaceModal, setUpdateNamespaceModal] = useState(false);
  const [deleteNamespaceModal, setDeleteNamespaceModal] = useState(false);

  const [descriptionText, setDescriptionText] = useState(ns.description);
  const [descriptionTextValid, setDescriptionTextValid] = useState(true);
  useEffect(() => { descriptionText != "" && setDescriptionTextValid(/^.{0,30}$/.test(descriptionText)) }, [descriptionText]);
  const [repositoryCountLimit, setRepositoryCountLimit] = useState<string | number>(ns.repository_limit);
  const [repositoryCountLimitValid, setRepositoryCountLimitValid] = useState(true);
  useEffect(() => { setRepositoryCountLimitValid(Number.isInteger(repositoryCountLimit) && parseInt(repositoryCountLimit.toString()) >= 0) }, [repositoryCountLimit])
  const [tagCountLimit, setTagCountLimit] = useState<string | number>(ns.tag_limit);
  const [tagCountLimitValid, setTagCountLimitValid] = useState(true);
  useEffect(() => { setTagCountLimitValid(Number.isInteger(tagCountLimit) && parseInt(tagCountLimit.toString()) >= 0) }, [tagCountLimit])
  let calcUnitObj = calcUnit(ns.size_limit);
  const [realSizeLimit, setRealSizeLimit] = useState(0);
  const [sizeLimit, setSizeLimit] = useState<string | number>(calcUnitObj.size);
  const [sizeLimitValid, setSizeLimitValid] = useState(true);
  const [sizeLimitUnit, setSizeLimitUnit] = useState(calcUnitObj.unit);
  useEffect(() => { setSizeLimitValid(Number.isInteger(sizeLimit) && parseInt(sizeLimit.toString()) >= 0) }, [sizeLimit])
  useEffect(() => {
    let sl = 0;
    if (Number.isInteger(sizeLimit)) { sl = parseInt(sizeLimit.toString()); }
    switch (sizeLimitUnit) {
      case "MiB": setRealSizeLimit(sl * 1 << 20); break;
      case "GiB": setRealSizeLimit(sl * 1 << 30); break;
      case "TiB": setRealSizeLimit(sl * 1 << 40); break;
    }
  }, [sizeLimit, sizeLimitUnit])
  const [namespaceVisibility, setNamespaceVisibility] = useState("private");

  const canManage = user.role == UserRole.Admin || user.role == UserRole.Root || (ns.role != undefined && (ns.role == NamespaceRole.Admin || ns.role == NamespaceRole.Manager));

  const updateNamespace = () => {
    setUpdateNamespaceModal(false);
    axios.put(localServer + `/api/v1/namespaces/${ns.id}`, {
      description: descriptionText, size_limit: realSizeLimit,
      repository_limit: repositoryCountLimit, tag_limit: tagCountLimit, visibility: namespaceVisibility,
    } as INamespaceItem).then(response => {
      if (response.status === 204) { setRefresh({}); }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    })
  }

  const deleteNamespace = () => {
    axios.delete(localServer + `/api/v1/namespaces/${ns.id}`).then(response => {
      if (response.status === 204) { setRefresh({}); }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    })
  }

  return (
    <>
      <TableRow className="align-middle">
        <TableCell className="cursor-pointer" onClick={() => navigate(`/namespaces/${ns.name}/repositories?namespace_id=${ns.id}`)}>
          <div className="truncate hover:text-muted-foreground">
            <span className="font-medium">{ns.name}</span>
            <span className="text-muted-foreground font-normal ml-4">{ns.description}</span>
          </div>
        </TableCell>
        <TableCell className="text-right"><Quota current={ns.size} limit={ns.size_limit} /></TableCell>
        <TableCell className="text-right"><QuotaSimple current={ns.repository_count} limit={ns.repository_limit} /></TableCell>
        <TableCell className="text-right"><QuotaSimple current={ns.tag_count} limit={ns.tag_limit} /></TableCell>
        <TableCell className="text-right"><Badge variant="outline" className="capitalize">{ns.visibility}</Badge></TableCell>
        <TableCell className="text-right text-muted-foreground">{dayjs.utc(ns.created_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}</TableCell>
        <TableCell className="text-right text-muted-foreground">{dayjs.utc(ns.updated_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}</TableCell>
        <TableCell className="text-center">
          <TableItemDropdown
            index={index}
            items={[
              { name: "Update", disable: !canManage, onClick: () => { setUpdateNamespaceModal(true); setNamespaceVisibility(ns.visibility); } },
              { name: "Delete", disable: !canManage, warn: true, onClick: () => setDeleteNamespaceModal(true) },
            ]}
          />
        </TableCell>
      </TableRow>

      {/* Update Namespace Modal */}
      <Dialog open={updateNamespaceModal} onOpenChange={setUpdateNamespaceModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader><DialogTitle>Update Namespace</DialogTitle></DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label>Name</Label>
              <Input value={ns.name} disabled />
            </div>
            <div className="grid gap-2">
              <Label>Description</Label>
              <Input placeholder="30 characters" value={descriptionText} onChange={e => setDescriptionText(e.target.value)} data-invalid={!descriptionTextValid ? true : undefined} />
              {!descriptionTextValid && <p className="text-xs text-destructive">Not a valid description, max 30 characters.</p>}
            </div>
            <div className="grid gap-2">
              <Label>Visibility</Label>
              <Select value={namespaceVisibility} onValueChange={(v) => { if (v) setNamespaceVisibility(v); }}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="private">Private</SelectItem>
                  <SelectItem value="public">Public</SelectItem>
                </SelectContent>
              </Select>
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
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Repository count limit</Label>
                <Input type="number" placeholder="0 means no limit" value={repositoryCountLimit} onChange={e => setRepositoryCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} data-invalid={!repositoryCountLimitValid ? true : undefined} />
              </div>
              <div className="grid gap-2">
                <Label>Tag count limit</Label>
                <Input type="number" placeholder="0 means no limit" value={tagCountLimit} onChange={e => setTagCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} data-invalid={!tagCountLimitValid ? true : undefined} />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setUpdateNamespaceModal(false)}>Cancel</Button>
            <Button onClick={updateNamespace}>Update</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Namespace Dialog */}
      <AlertDialog open={deleteNamespaceModal} onOpenChange={setDeleteNamespaceModal}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete namespace</AlertDialogTitle>
            <AlertDialogDescription>Are you sure you want to delete the namespace <strong>{ns.name}</strong>?</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={deleteNamespace} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}