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
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { Link, useParams } from 'react-router-dom';

import Header from "../../components/Header";
import Menu from "../../components/Menu";
import OrderHeader from "../../components/OrderHeader";
import Pagination from "../../components/Pagination";
import QuotaSimple from "../../components/QuotaSimple";
import Regex from "../../utils/regex";
import Settings from "../../Settings";
import Toast from "../../components/Notification";
import { useTranslation } from "../../i18n/useTranslation";
import { IHTTPError, IOrder, IUserItem, IUserList } from "../../interfaces";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import {
  Dialog as ShadDialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Info } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

const supportRoles = [
  { id: 1, name: 'Admin' },
  { id: 2, name: 'User' },
];

const supportStatus = [
  { id: 1, name: 'Active' },
  { id: 2, name: 'Inactive' },
];

export default function ({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
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
      axios.post(localServer + `/api/v1/validators/password`, { password: passwordText }).then(response => {
        setPasswordTextValid(response?.status === 204);
      }).catch(() => setPasswordTextValid(false));
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
    if (searchUsername !== "") url += `&name=${searchUsername}`;
    if (sortName !== "") url += `&sort=${sortName}&method=${sortOrder.toString()}`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        setUserList(response.data as IUserList);
        setTotal((response.data as IUserList).total);
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  }, [refresh]);

  const [role, setRole] = useState("User");

  const createUser = () => {
    if (!(usernameTextValid && passwordTextValid)) {
      Toast({ level: "warning", title: t("common.formValidateFailed"), message: t("common.checkForm") });
      return;
    }
    axios.post(localServer + `/api/v1/users/`, {
      username: usernameText, password: passwordText, email: emailInput,
      namespace_limit: namespaceCountLimit, role: role,
    }).then(response => {
      if (response?.status === 201) {
        Toast({ level: "success", title: t("common.success"), message: t("setting.createUserSuccess") });
        setCreateUserModal(false);
        setRefresh({});
        setUsernameText(""); setPasswordText(""); setEmailInput(""); setNamespaceCountLimit(0); setRole("User");
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  }

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet><title>{t("header.settingUsers")}</title></Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background">
        <Menu localServer={localServer} item="users" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title={t("header.settingUsers")} />
            <div className="pt-2 pb-2 flex justify-between">
              <div className="pr-2 pl-2 flex gap-2">
                <div className="relative flex items-center">
                  <Label htmlFor="usernameSearch" className="absolute -top-2 left-2 inline-block bg-background px-1 text-xs font-medium text-foreground z-10">Username</Label>
                  <Input id="usernameSearch" placeholder="search username" value={searchUsername} onChange={e => setSearchUsername(e.target.value)} onKeyDown={e => { if (e.key == "Enter") setRefresh({}); }} className="h-10 pr-14" />
                  <kbd className="absolute inset-y-0 right-0 flex items-center py-1.5 pr-3 text-xs text-muted-foreground">enter</kbd>
                </div>
              </div>
              <div className="pr-2 pl-2">
                <Button onClick={() => setCreateUserModal(true)}>Create</Button>
              </div>
            </div>
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className="w-full">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead><span className="lg:pl-2">Username</span></TableHead>
                    <TableHead>Namespace</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right"><OrderHeader text="Last Login" orderStatus={lastLoginOrder} setOrder={e => { resetOrder(); setLastLoginOrder(e); setSortOrder(e); setSortName("last_login"); setRefresh({}); }} /></TableHead>
                    <TableHead className="text-right"><OrderHeader text="Created at" orderStatus={createdAtOrder} setOrder={e => { resetOrder(); setCreatedAtOrder(e); setSortOrder(e); setSortName("created_at"); setRefresh({}); }} /></TableHead>
                    <TableHead className="text-right">Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {userList.items?.map(userObj => (
                    <TableItemRow key={userObj.id} localServer={localServer} user={userObj} setRefresh={setRefresh} />
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
          <Pagination limit={Settings.PageSize} page={page} setPage={setPage} total={total} />
        </div>
      </div>

      {/* Create User Modal */}
      <ShadDialog open={createUserModal} onOpenChange={setCreateUserModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader><DialogTitle>Create User</DialogTitle></DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label><span className="text-destructive">*</span>Username</Label>
              <Input placeholder="username" value={usernameText} onChange={e => setUsernameText(e.target.value)} data-invalid={!usernameTextValid ? true : undefined} />
              {!usernameTextValid && <p className="text-xs text-destructive">Not a valid username.</p>}
            </div>
            <div className="grid gap-2">
              <Label><span className="text-destructive">*</span>Password</Label>
              <Input type="password" placeholder="password" value={passwordText} onChange={e => setPasswordText(e.target.value)} data-invalid={!passwordTextValid ? true : undefined} />
              {!passwordTextValid && <p className="text-xs text-destructive">Password is invalid.</p>}
            </div>
            <div className="grid gap-2">
              <Label><span className="text-destructive">*</span>Email</Label>
              <Input placeholder="email" value={emailInput} onChange={e => setEmailInput(e.target.value)} data-invalid={!emailInputValid ? true : undefined} />
              {!emailInputValid && <p className="text-xs text-destructive">Email is invalid.</p>}
            </div>
            {role === "User" && (
              <div className="grid gap-2">
                <Label className="flex items-center gap-1">
                  Namespace count limit
                  <Tooltip>
                    <TooltipTrigger><Info className="h-3.5 w-3.5 text-muted-foreground" /></TooltipTrigger>
                    <TooltipContent>0 means no limit</TooltipContent>
                  </Tooltip>
                </Label>
                <Input type="number" placeholder="0 means no limit" value={namespaceCountLimit} onChange={e => setNamespaceCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} data-invalid={!namespaceCountLimitValid ? true : undefined} />
                {!namespaceCountLimitValid && <p className="text-xs text-destructive">Not a valid namespace count limit.</p>}
              </div>
            )}
            <div className="grid gap-2">
              <Label>Role</Label>
              <Select value={role} onValueChange={(v) => { if (v) setRole(v); }}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {supportRoles.map(r => <SelectItem key={r.name} value={r.name}>{r.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateUserModal(false)}>Cancel</Button>
            <Button onClick={createUser}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </ShadDialog>
    </Fragment>
  );
}

function TableItemRow({ localServer, user, setRefresh }: { localServer: string, user: IUserItem, setRefresh: (param: any) => void }) {
  const [status, setStatus] = useState(user.status === "" ? "Active" : user.status);
  const [role, setRole] = useState(user.role === "" ? "Normal" : user.role);
  const [usernameText, setUsernameText] = useState(user.username);
  const [usernameTextValid, setUsernameTextValid] = useState(true);
  useEffect(() => {
    if (usernameText.length > 0) setUsernameTextValid(Regex.Username.test(usernameText));
  }, [usernameText]);
  const [passwordText, setPasswordText] = useState("");
  const [passwordTextValid, setPasswordTextValid] = useState(true);
  useEffect(() => {
    if (passwordText.length > 0) {
      axios.post(localServer + `/api/v1/validators/password`, { password: passwordText }).then(response => {
        setPasswordTextValid(response?.status === 204);
      }).catch(() => setPasswordTextValid(false));
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
    const data: { [key: string]: any } = { email: emailInput, username: usernameText, status: status, namespace_limit: namespaceCountLimit };
    if (passwordText.length != 0) data["password"] = passwordText;
    axios.put(localServer + `/api/v1/users/${user.id}`, data).then(response => {
      if (response?.status === 204) {
        Toast({ level: "success", title: "Success", message: "Update user info success" });
        setUpdateUserModal(false);
        setRefresh({});
        setPasswordText("");
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  }

  return (
    <>
      <TableRow className="align-middle">
        <TableCell>
          <div className="truncate">
            <span className="font-medium">{user.username}</span>
            <span className="text-muted-foreground font-normal ml-4">{user.email}</span>
          </div>
        </TableCell>
        <TableCell><QuotaSimple current={user.namespace_count} limit={user.namespace_limit} /></TableCell>
        <TableCell><Badge variant="outline">{user.status}</Badge></TableCell>
        <TableCell className="text-right text-muted-foreground">{dayjs.utc(user.last_login).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}</TableCell>
        <TableCell className="text-right text-muted-foreground">{dayjs.utc(user.created_at).tz(dayjs.tz.guess()).format("YYYY-MM-DD HH:mm:ss")}</TableCell>
        <TableCell className="text-center">
          <Button variant="ghost" size="sm" onClick={() => setUpdateUserModal(true)}>Update</Button>
        </TableCell>
      </TableRow>

      <ShadDialog open={updateUserModal} onOpenChange={setUpdateUserModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader><DialogTitle>Update User</DialogTitle></DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label><span className="text-destructive">*</span>Username</Label>
              <Input value={usernameText} onChange={e => setUsernameText(e.target.value)} data-invalid={!usernameTextValid ? true : undefined} />
            </div>
            <div className="grid gap-2">
              <Label>Password</Label>
              <Input type="password" placeholder="Leave empty to keep current" value={passwordText} onChange={e => setPasswordText(e.target.value)} data-invalid={passwordText.length > 0 && !passwordTextValid ? true : undefined} />
            </div>
            <div className="grid gap-2">
              <Label><span className="text-destructive">*</span>Email</Label>
              <Input value={emailInput} onChange={e => setEmailInput(e.target.value)} data-invalid={!emailInputValid ? true : undefined} />
            </div>
            <div className="grid gap-2">
              <Label>Namespace count limit</Label>
              <Input type="number" placeholder="0 means no limit" value={namespaceCountLimit} onChange={e => setNamespaceCountLimit(Number.isNaN(parseInt(e.target.value)) ? "" : parseInt(e.target.value))} />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Role</Label>
                <Select value={role} onValueChange={(v) => { if (v) setRole(v); }}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {supportRoles.map(r => <SelectItem key={r.name} value={r.name}>{r.name}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label>Status</Label>
                <Select value={status} onValueChange={(v) => { if (v) setStatus(v); }}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {supportStatus.map(s => <SelectItem key={s.name} value={s.name}>{s.name}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setUpdateUserModal(false)}>Cancel</Button>
            <Button onClick={updateUser}>Update</Button>
          </DialogFooter>
        </DialogContent>
      </ShadDialog>
    </>
  );
}