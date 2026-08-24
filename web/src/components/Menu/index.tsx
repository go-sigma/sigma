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
import { createContext, useContext, useEffect, useState } from "react";
import { Link, useNavigate } from 'react-router-dom';

import Regex from "../../utils/regex";
import Toast from "../../components/Notification";
import { LocaleSwitcher, ThemeSwitcher } from "../../components/Preferences";
import { useTranslation } from "../../i18n/useTranslation";
import { IEndpoint, IHTTPError, INamespaceItem, INamespaceList, ISystemConfig, IUserSelf, IVersion } from "../../interfaces";
import { setupAutoRefreshToken, teardownAutoRefreshToken } from "../../utils/refreshToken";

import { Button } from "@/components/ui/button";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  useSidebar,
} from "@/components/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  ChevronUp,
  List,
  Folder,
  Tag,
  Server,
  Settings,
  Users,
  FileText,
  Webhook,
  ExternalLink,
  User,
  Lock,
  Info,
  LogOut,
  Circle,
} from "lucide-react";

export const AppLayoutMenuContext = createContext(false);

type MenuProps = {
  localServer: string;
  item: string;
  namespace?: string;
  namespace_id?: string;
  repository?: string;
  repository_id?: string;
  tag?: string;
  selfClick?: boolean;
  persistent?: boolean;
};

export default function Menu(props: MenuProps) {
  const hasAppLayoutMenu = useContext(AppLayoutMenuContext);
  if (hasAppLayoutMenu && !props.persistent) {
    return null;
  }

  return (
    <SidebarProvider defaultOpen>
      <MenuContent {...props} />
    </SidebarProvider>
  );
}

function MenuContent({ localServer, item, namespace, namespace_id, repository, selfClick }: MenuProps) {
  const { t } = useTranslation();
  const [menuActive, setMenuActive] = useState(item === "" ? "home" : item);
  const navigate = useNavigate();

  useEffect(() => {
    setMenuActive(item === "" ? "home" : item);
  }, [item]);

  const [isAnonymous, setIsAnonymous] = useState(false);
  const [userID, setUserID] = useState(0);
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [refresh, setRefresh] = useState({});

  useEffect(() => {
    axios.get(localServer + "/api/v1/users/self").then(response => {
      if (response.status === 200) {
        const user = response.data as IUserSelf;
        setUsername(user.username);
        setEmail(user.email);
        setUserID(user.id);
        if (user.role === "Anonymous") {
          setIsAnonymous(true);
        }
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
    setupAutoRefreshToken(localServer, logout);
    const visibilitychangeHandler = () => {
      if (document.hidden) {
        teardownAutoRefreshToken();
      } else {
        setupAutoRefreshToken(localServer, logout);
      }
    }
    document.addEventListener("visibilitychange", visibilitychangeHandler);
    return () => {
      document.removeEventListener("visibilitychange", visibilitychangeHandler);
    }
  }, [refresh]);

  const [config, setConfig] = useState<ISystemConfig>({
    daemon: { builder: true }
  } as ISystemConfig);

  useEffect(() => {
    axios.get(localServer + "/api/v1/systems/config").then(response => {
      if (response.status === 200) {
        setConfig(response.data as ISystemConfig);
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  }, []);

  const [hotNamespaceList, setHotNamespaceList] = useState<INamespaceItem[]>([]);

  useEffect(() => {
    if (!isAnonymous) {
      axios.get(localServer + "/api/v1/namespaces/hot").then(response => {
        if (response.status === 200) {
          setHotNamespaceList((response.data as INamespaceList).items);
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
        }
      }).catch(error => {
        const errorcode = error.response?.data as IHTTPError;
        Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
      });
    }
  }, []);

  const logout = () => {
    let tokens: string[] = [localStorage.getItem("token") || "", localStorage.getItem("refresh_token") || ""];
    axios.post(localServer + "/api/v1/users/logout", { tokens }).then(() => {
      localStorage.removeItem("token");
      localStorage.removeItem("refresh_token");
      navigate("/login");
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  }

  const [endpoint, setEndpoint] = useState("");

  useEffect(() => {
    axios.get(localServer + `/api/v1/systems/endpoint`).then(response => {
      if (response.status === 200) {
        setEndpoint((response.data as IEndpoint).endpoint);
      }
    }).catch(() => {});
  }, []);

  const [updateProfileModal, setUpdateProfileModal] = useState(false);
  const [updatePasswordModal, setUpdatePasswordModal] = useState(false);
  const [aboutModal, setAboutModal] = useState(false);
  const [version, setVersion] = useState<IVersion>();

  useEffect(() => {
    axios.get(`${localServer}/api/v1/systems/version`).then(response => {
      if (response?.status === 200) {
        setVersion(response.data as IVersion);
      }
    }).catch(() => {});
  }, [])

  return (
    <>
      <Sidebar collapsible="icon">
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" render={<Link to={isAnonymous ? "/namespaces" : "/home"} />}>
                <img className="size-8" src="/title.svg" alt="sigma" />
                <span className="font-semibold">sigma</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>

        <SidebarContent>
          <SidebarGroup>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton render={<Link to="/namespaces" onClick={() => setMenuActive("namespaces")} />} isActive={menuActive === "namespaces"}>
                  <List />
                  <span>{t("menu.namespaces")}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>

              {(menuActive === "repositories" || menuActive === "tags" || menuActive === "artifacts") && (
                <SidebarMenuItem>
                  <SidebarMenuButton render={<Link to={`/namespaces/${namespace}/repositories?namespace_id=${namespace_id}`} onClick={(e) => { setMenuActive("repositories"); if (item === "repositories" && selfClick !== true) e.preventDefault(); }} />} isActive={menuActive === "repositories"} className="ml-4">
                    <Folder />
                    <span>{t("common.repositories")}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )}

              {(menuActive === "tags" || menuActive === "artifacts") && (
                <SidebarMenuItem>
                  <SidebarMenuButton render={<Link to={`/namespaces/${namespace}/repository/tags?repository=${repository}`} onClick={(e) => { setMenuActive("tags"); if (item === "tags") e.preventDefault(); }} />} isActive={menuActive === "tags"} className="ml-8">
                    <Tag />
                    <span>{t("common.tags")}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )}

              {!isAnonymous && config.daemon?.builder && (
                <SidebarMenuItem>
                  <SidebarMenuButton render={<Link to="/coderepos" onClick={() => setMenuActive("coderepos")} />} isActive={menuActive === "coderepos"}>
                    <Server />
                    <span>{t("menu.codeRepository")}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )}

              {!isAnonymous && (
                <SidebarMenuItem>
                  <SidebarMenuButton render={<Link to="/settings/users" onClick={() => setMenuActive("users")} />} isActive={menuActive === "settings" || menuActive === "users" || menuActive === "daemon-tasks" || menuActive === "webhooks"}>
                    <Settings />
                    <span>{t("menu.setting")}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )}

              {(menuActive === "settings" || menuActive === "users" || menuActive === "daemon-tasks" || menuActive === "webhooks") && (
                <>
                  <SidebarMenuItem>
                    <SidebarMenuButton render={<Link to="/settings/users" onClick={() => setMenuActive("users")} />} isActive={menuActive === "users"} className="ml-4">
                      <Users />
                      <span>{t("common.users")}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                  <SidebarMenuItem>
                    <SidebarMenuButton render={<Link to="/settings/daemon-tasks?namespace_id=0" onClick={() => setMenuActive("daemon-tasks")} />} isActive={menuActive === "daemon-tasks"} className="ml-4">
                      <FileText />
                      <span>{t("menu.daemonTask")}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                  <SidebarMenuItem>
                    <SidebarMenuButton render={<Link to="/settings/webhooks?namespace_id=0" onClick={() => setMenuActive("webhooks")} />} isActive={menuActive === "webhooks"} className="ml-4">
                      <Webhook />
                      <span>{t("menu.webhook")}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </>
              )}
            </SidebarMenu>
          </SidebarGroup>

          {!isAnonymous && hotNamespaceList.length > 0 && (
            <SidebarGroup>
              <SidebarGroupLabel>{t("menu.hotNamespace")}</SidebarGroupLabel>
              <SidebarMenu>
                {hotNamespaceList.map((ns: INamespaceItem, index: number) => (
                  <SidebarMenuItem key={ns.id}>
                    <SidebarMenuButton render={<Link to={`/namespaces/${ns.name}/repositories?namespace_id=${ns.id}`} onClick={(e) => { if (item === "repositories" && ns.name === namespace) e.preventDefault(); }} />}>
                      <Circle
                        className={`size-2.5 fill-current ${
                          index === 0 ? "text-red-500" : index === 1 ? "text-amber-500" : "text-indigo-500"
                        }`}
                      />
                      <span className="truncate">{ns.name}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroup>
          )}
        </SidebarContent>

        <SidebarFooter>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton render={<a href={`${endpoint}/swagger/index.html`} target="_blank" rel="noreferrer" />}>
                <ExternalLink />
                <span>{t("menu.apiDocs")}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
            <SidebarMenuItem>
              <div className="px-2 py-1">
                <ThemeSwitcher />
              </div>
            </SidebarMenuItem>
            <SidebarMenuItem>
              <div className="px-2 py-1">
                <LocaleSwitcher />
              </div>
            </SidebarMenuItem>
            <SidebarMenuItem>
              <DropdownMenu>
                <DropdownMenuTrigger render={<SidebarMenuButton className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground" />}>
                  <Avatar className="size-6">
                    <AvatarFallback className="text-xs">{username.charAt(0).toUpperCase()}</AvatarFallback>
                  </Avatar>
                  <span>{username}</span>
                  <ChevronUp className="ml-auto" />
                </DropdownMenuTrigger>
                <DropdownMenuContent side="top" className="w-(--sidebar-width) min-w-56">
                  {!isAnonymous && (
                    <>
                      <DropdownMenuItem onClick={() => setUpdateProfileModal(true)}>
                        <User />
                        <span>{t("menu.updateProfile")}</span>
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => setUpdatePasswordModal(true)}>
                        <Lock />
                        <span>{t("menu.resetPassword")}</span>
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                    </>
                  )}
                  <DropdownMenuItem onClick={logout}>
                    <LogOut />
                    <span>{t("menu.logout")}</span>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => setAboutModal(true)}>
                    <Info />
                    <span>{t("common.about")}</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarFooter>
      </Sidebar>

      {/* Update Profile Modal */}
      <UpdateProfileModal
        open={updateProfileModal}
        onOpenChange={setUpdateProfileModal}
        localServer={localServer}
        username={username}
        email={email}
        onUpdated={() => setRefresh({})}
      />

      {/* Update Password Modal */}
      <UpdatePasswordModal
        open={updatePasswordModal}
        onOpenChange={setUpdatePasswordModal}
        localServer={localServer}
      />

      {/* About Modal */}
      <Dialog open={aboutModal} onOpenChange={setAboutModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("common.about")}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            {t("menu.aboutDescription", { version: version?.version || t("common.notAvailable") })}
          </p>
          <DialogFooter>
            <Button onClick={() => setAboutModal(false)}>{t("common.confirm")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

function UpdateProfileModal({ open, onOpenChange, localServer, username: initialUsername, email: initialEmail, onUpdated }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  localServer: string;
  username: string;
  email: string;
  onUpdated: () => void;
}) {
  const { t } = useTranslation();
  const [usernameInput, setUsernameInput] = useState(initialUsername);
  const [usernameInputValid, setUsernameInputValid] = useState(true);
  const [emailInput, setEmailInput] = useState(initialEmail);
  const [emailInputValid, setEmailInputValid] = useState(true);

  useEffect(() => {
    setUsernameInput(initialUsername);
    setEmailInput(initialEmail);
  }, [initialUsername, initialEmail]);

  useEffect(() => {
    if (usernameInput.length > 0) {
      setUsernameInputValid(Regex.Username.test(usernameInput));
    }
  }, [usernameInput]);

  useEffect(() => {
    if (emailInput.length > 0) {
      setEmailInputValid(Regex.Email.test(emailInput));
    }
  }, [emailInput]);

  const updateUser = () => {
    if (!(usernameInputValid && emailInputValid)) {
      Toast({ level: "warning", title: t("common.formValidateFailed"), message: t("common.checkForm") });
      return;
    }
    const data: { [key: string]: any } = {};
    if (usernameInput.length > 0) data.username = usernameInput;
    if (emailInput.length > 0) data.email = emailInput;
    if (_.size(data) > 0) {
      axios.put(`${localServer}/api/v1/users/self`, data).then(response => {
        if (response?.status === 204) {
          Toast({ level: "success", title: t("common.success"), message: t("menu.profileUpdated") });
          onOpenChange(false);
          onUpdated();
        }
      }).catch(error => {
        const errorcode = error.response?.data as IHTTPError;
        Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("menu.updateProfile")}</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="usernameInput">{t("common.username")}</Label>
            <Input
              id="usernameInput"
              value={usernameInput}
              onChange={e => setUsernameInput(e.target.value)}
              data-invalid={!usernameInputValid ? true : undefined}
            />
            {!usernameInputValid && <p className="text-xs text-destructive">{t("menu.invalidUsername")}</p>}
          </div>
          <div className="grid gap-2">
            <Label htmlFor="emailInput">{t("common.email")}</Label>
            <Input
              id="emailInput"
              value={emailInput}
              onChange={e => setEmailInput(e.target.value)}
              data-invalid={!emailInputValid ? true : undefined}
            />
            {!emailInputValid && <p className="text-xs text-destructive">{t("menu.invalidEmail")}</p>}
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>{t("common.cancel")}</Button>
          <Button onClick={updateUser}>{t("common.update")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function UpdatePasswordModal({ open, onOpenChange, localServer }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  localServer: string;
}) {
  const { t } = useTranslation();
  const [passwordInput, setPasswordInput] = useState("");
  const [passwordInputValid, setPasswordInputValid] = useState(true);
  const [repeatPasswordInput, setRepeatPasswordInput] = useState("");
  const [repeatPasswordInputValid, setRepeatPasswordInputValid] = useState(true);

  useEffect(() => {
    if (passwordInput.length > 0) {
      axios.get(`${localServer}/api/v1/validators/password?password=${passwordInput}`).then(response => {
        setPasswordInputValid(response?.status === 204);
      }).catch(() => setPasswordInputValid(false));
    }
  }, [passwordInput]);

  useEffect(() => {
    setRepeatPasswordInputValid(passwordInput === repeatPasswordInput);
  }, [passwordInput, repeatPasswordInput]);

  const updatePassword = () => {
    if (!(passwordInputValid && repeatPasswordInputValid)) {
      Toast({ level: "warning", title: t("common.formValidateFailed"), message: t("common.checkForm") });
      return;
    }
    axios.put(`${localServer}/api/v1/users/self/reset-password`, { password: passwordInput }).then(response => {
      if (response?.status === 204) {
        Toast({ level: "success", title: t("common.success"), message: t("menu.passwordUpdated") });
        onOpenChange(false);
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Toast({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("menu.resetPassword")}</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="passwordInput">{t("common.password")}</Label>
            <Input
              id="passwordInput"
              type="password"
              value={passwordInput}
              onChange={e => setPasswordInput(e.target.value)}
              data-invalid={!passwordInputValid ? true : undefined}
            />
            {!passwordInputValid && <p className="text-xs text-destructive">{t("menu.invalidPassword")}</p>}
          </div>
          <div className="grid gap-2">
            <Label htmlFor="repeatPasswordInput">{t("menu.repeatPassword")}</Label>
            <Input
              id="repeatPasswordInput"
              type="password"
              value={repeatPasswordInput}
              onChange={e => setRepeatPasswordInput(e.target.value)}
              data-invalid={!repeatPasswordInputValid ? true : undefined}
            />
            {!repeatPasswordInputValid && <p className="text-xs text-destructive">{t("menu.invalidRepeatPassword")}</p>}
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>{t("common.cancel")}</Button>
          <Button onClick={updatePassword}>{t("common.update")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}