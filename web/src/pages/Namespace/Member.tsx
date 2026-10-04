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

import {
  ChevronUpDownIcon,
  EllipsisVerticalIcon,
} from "@heroicons/react/20/solid";
import axios from "axios";
import { CornerDownLeft } from "lucide-react";
import { Fragment, useCallback, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import Toast from "react-hot-toast";
import { useParams, useSearchParams } from "react-router-dom";

import Header from "@/components/Header";
import IMenu from "@/components/Menu";
import NamespaceTabs from "@/components/NamespaceTabs";
import Notification from "@/components/Notification";
import Pagination from "@/components/Pagination";
import RelativeTime from "@/components/RelativeTime";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  IHTTPError,
  INamespaceItem,
  INamespaceRoleItem as INamespaceMemberItem,
  INamespaceRoleList,
  IUserItem,
  IUserList,
} from "@/interfaces";
import Settings from "@/Settings";

const namespaceRoles = [
  { id: 1, name: "NamespaceAdmin" },
  { id: 2, name: "NamespaceManager" },
  { id: 3, name: "NamespaceReader" },
];

export default function Member({ localServer }: { localServer: string }) {
  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const namespaceId =
    searchParams.get("namespace_id") == null
      ? 0
      : parseInt(searchParams.get("namespace_id") || "");

  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [memberSearch, setUsernameSearch] = useState("");
  const [createUserNamespaceModal, setCreateUserNamespaceModal] =
    useState(false);
  const [namespaceObj, setNamespaceObj] = useState<INamespaceItem>(
    {} as INamespaceItem,
  );
  const [memberList, setMemberList] = useState<INamespaceRoleList>(
    {} as INamespaceRoleList,
  );

  useEffect(() => {
    if (namespaceId == 0) {
      return;
    }
    let url = `${localServer}/api/v1/namespaces/${namespaceId}`;
    axios
      .get(url)
      .then((response) => {
        if (response?.status === 200) {
          const namespaceData = response.data as INamespaceItem;
          setNamespaceObj(namespaceData);
        } else if (response.status !== 404) {
          const errorcode = response.data as IHTTPError;
          Notification({
            level: "warning",
            title: errorcode?.title,
            message: errorcode?.description,
          });
        }
      })
      .catch((error) => {
        // A missing namespace means there are no members, so stay silent.
        if (error.response?.status === 404) {
          return;
        }
        const errorcode = error.response?.data as IHTTPError;
        Notification({
          level: "warning",
          title: errorcode?.title,
          message: errorcode?.description,
        });
      });
  }, [namespaceId, localServer]);

  const fetchMembers = useCallback(() => {
    let url = `${localServer}/api/v1/namespaces/${namespaceId}/members/?limit=${Settings.AutoCompleteSize}`;
    if (memberSearch !== "") {
      url += `&name=${memberSearch}`;
    }
    axios
      .get(url)
      .then((response) => {
        if (response?.status === 200) {
          const namespaceRoleList = response.data as INamespaceRoleList;
          setMemberList(namespaceRoleList);
          setTotal(namespaceRoleList.total);
        } else if (response.status !== 404) {
          const errorcode = response.data as IHTTPError;
          Notification({
            level: "warning",
            title: errorcode?.title,
            message: errorcode?.description,
          });
        }
      })
      .catch((error) => {
        // An empty member list is not an error, so stay silent.
        if (error.response?.status === 404) {
          return;
        }
        const errorcode = error.response?.data as IHTTPError;
        Notification({
          level: "warning",
          title: errorcode?.title,
          message: errorcode?.description,
        });
      });
  }, [localServer, memberSearch, namespaceId]);

  useEffect(() => {
    fetchMembers();
  }, [fetchMembers]);

  const [userSearch, setUserSearch] = useState("");
  const [userList, setUserList] = useState<IUserItem[]>();
  const [userSelected, setUserSelected] = useState<IUserItem>({} as IUserItem);
  const userSelectedValid = userSelected.username !== undefined;
  const [addNamespaceRoleRole, setAddNamespaceRoleRole] =
    useState("NamespaceReader");

  useEffect(() => {
    let url = `${localServer}/api/v1/users/?limit=${Settings.AutoCompleteSize}&without_admin=true`;
    if (userSearch !== "") {
      url += `&name=${userSearch}`;
    }
    axios
      .get(url)
      .then((response) => {
        if (response?.status === 200) {
          const namespaceList = response.data as IUserList;
          setUserList(namespaceList.items);
        } else {
          const errorcode = response.data as IHTTPError;
          Notification({
            level: "warning",
            title: errorcode.title,
            message: errorcode.description,
          });
        }
      })
      .catch((error) => {
        const errorcode = error.response?.data as IHTTPError;
        Notification({
          level: "warning",
          title: errorcode?.title,
          message: errorcode?.description,
        });
      });
  }, [userSearch, localServer]);

  const addMember = () => {
    if (userSelected.username === undefined) {
      return;
    }
    let url = `${localServer}/api/v1/namespaces/${namespaceId}/members/`;
    axios
      .post(url, {
        user_id: userSelected.id,
        role: addNamespaceRoleRole,
      })
      .then((response) => {
        if (response?.status === 201) {
          Toast.success("Add member to namespace success");
          setCreateUserNamespaceModal(false);
          fetchMembers();
        } else {
          const errorcode = response.data as IHTTPError;
          Notification({
            level: "warning",
            title: errorcode.title,
            message: errorcode.description,
          });
        }
      })
      .catch((error) => {
        const errorcode = error.response.data as IHTTPError;
        Notification({
          level: "warning",
          title: errorcode.title,
          message: errorcode.description,
        });
      });
  };

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Namespace Member</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item="Repository" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header
              title="Namespace - Member"
              props={
                <NamespaceTabs
                  namespace={namespace}
                  namespaceId={namespaceId ? namespaceId.toString() : ""}
                  active="members"
                />
              }
            />
            <div className="pt-4 pb-4 flex justify-between items-center">
              <div className="px-4">
                <div className="relative flex items-center">
                  <Label
                    htmlFor="usernameSearch"
                    className="absolute -top-2 left-2 inline-block bg-background px-1 text-xs font-medium text-foreground z-10"
                  >
                    Username
                  </Label>
                  <Input
                    id="usernameSearch"
                    placeholder="search username"
                    value={memberSearch}
                    onChange={(e) => setUsernameSearch(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key == "Enter") fetchMembers();
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
              <div className="px-4">
                <Button onClick={() => setCreateUserNamespaceModal(true)}>
                  Add
                </Button>
              </div>
            </div>
          </main>
          <div className="flex-1 flex overflow-y-auto">
            <div className="w-full">
              <Table>
                <TableHeader className="[&_th]:font-normal">
                  <TableRow>
                    <TableHead>Username</TableHead>
                    <TableHead className="text-right">Role</TableHead>
                    <TableHead className="text-right">Added</TableHead>
                    <TableHead className="text-right">Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {memberList.items?.map((member) => (
                    <TableItem
                      key={member.id}
                      localServer={localServer}
                      namespace={namespaceObj}
                      userSelectedArg={
                        {
                          username: member.username,
                          id: member.user_id,
                        } as IUserItem
                      }
                      member={member}
                      onChanged={fetchMembers}
                    />
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
          <Pagination
            limit={Settings.PageSize}
            page={page}
            setPage={setPage}
            total={total}
          />
        </div>
      </div>
      <Dialog
        open={createUserNamespaceModal}
        onOpenChange={setCreateUserNamespaceModal}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogTitle className="border-b pb-4">Add member</DialogTitle>
          <div className="flex flex-col gap-0 mt-4">
            <div className="grid grid-cols-6 gap-4">
              <div className="col-span-2 flex flex-row">
                <label
                  htmlFor="usernameText"
                  className="block text-sm font-medium leading-6 text-gray-900 my-auto"
                >
                  <div className="flex">
                    <span className="text-red-600">*</span>
                    <span className="leading-6 ">User</span>
                    <span>:</span>
                  </div>
                </label>
              </div>
              <div className="col-span-4">
                <Popover>
                  <PopoverTrigger
                    render={
                      <Button
                        variant="outline"
                        className="w-full justify-between font-normal"
                      >
                        {userSelected?.username || "Select user"}
                        <ChevronUpDownIcon
                          className="size-4 text-muted-foreground"
                          aria-hidden="true"
                        />
                      </Button>
                    }
                  />
                  <PopoverContent
                    align="start"
                    className="w-(--anchor-width) p-0"
                  >
                    <Command shouldFilter={false}>
                      <CommandInput
                        placeholder="Search user"
                        value={userSearch}
                        onValueChange={setUserSearch}
                      />
                      <CommandList>
                        <CommandEmpty>Nothing found.</CommandEmpty>
                        <CommandGroup>
                          {userList?.map((user) => (
                            <CommandItem
                              key={user.id}
                              value={user.username}
                              onSelect={() => setUserSelected(user)}
                            >
                              {user.username}
                            </CommandItem>
                          ))}
                        </CommandGroup>
                      </CommandList>
                    </Command>
                  </PopoverContent>
                </Popover>
              </div>
            </div>
            <div className="grid grid-cols-6 gap-4">
              <div className="col-span-2"></div>
              <div className="col-span-4">
                {userSelectedValid ? null : (
                  <p className="mt-1 text-xs text-red-600">
                    <span>Please select a user.</span>
                  </p>
                )}
              </div>
            </div>
            <div className="grid grid-cols-6 gap-4 mt-4">
              <div className="col-span-2 flex flex-row">
                <label
                  htmlFor="usernameText"
                  className="block text-sm font-medium leading-6 text-gray-900 my-auto"
                >
                  <div className="flex">
                    <span className="text-red-600">*</span>
                    <span className="leading-6 ">Role</span>
                    <span>:</span>
                  </div>
                </label>
              </div>
              <div className="col-span-4">
                <Select
                  value={addNamespaceRoleRole}
                  onValueChange={(source) => {
                    if (source) setAddNamespaceRoleRole(source);
                  }}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {namespaceRoles.map((source) => (
                      <SelectItem key={source.name} value={source.name}>
                        {source.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setCreateUserNamespaceModal(false)}
            >
              Cancel
            </Button>
            <Button onClick={() => addMember()}>Add</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Fragment>
  );
}

function TableItem({
  localServer,
  namespace,
  userSelectedArg,
  member,
  onChanged,
}: {
  localServer: string;
  namespace: INamespaceItem;
  userSelectedArg: IUserItem;
  member: INamespaceMemberItem;
  onChanged: () => void;
}) {
  const [updateUserNamespaceModal, setUpdateUserNamespaceModal] =
    useState(false);
  const [userSelected] = useState<IUserItem>(userSelectedArg);
  const [addNamespaceRoleRole, setAddNamespaceRoleRole] = useState(member.role);

  const deleteMember = () => {
    let url = `${localServer}/api/v1/namespaces/${namespace.id}/members/?user_id=${userSelected.id}`;
    axios
      .delete(url)
      .then((response) => {
        if (response?.status === 204) {
          Toast.success("Delete member success");
          onChanged();
        } else {
          const errorcode = response.data as IHTTPError;
          Notification({
            level: "warning",
            title: errorcode.title,
            message: errorcode.description,
          });
        }
      })
      .catch((error) => {
        const errorcode = error.response.data as IHTTPError;
        Notification({
          level: "warning",
          title: errorcode.title,
          message: errorcode.description,
        });
      });
  };

  const updateMember = () => {
    let url = `${localServer}/api/v1/namespaces/${namespace.id}/members/`;
    axios
      .put(url, {
        user_id: userSelected.id,
        role: addNamespaceRoleRole,
      })
      .then((response) => {
        if (response?.status === 204) {
          Toast.success("Update member success");
          setUpdateUserNamespaceModal(false);
          onChanged();
        } else {
          const errorcode = response.data as IHTTPError;
          Notification({
            level: "warning",
            title: errorcode.title,
            message: errorcode.description,
          });
        }
      })
      .catch((error) => {
        const errorcode = error.response.data as IHTTPError;
        Notification({
          level: "warning",
          title: errorcode.title,
          message: errorcode.description,
        });
      });
  };

  return (
    <>
      <TableRow className="align-middle">
        <TableCell>{member.username}</TableCell>
        <TableCell className="text-right text-muted-foreground">
          {member.role.substring(9)}
        </TableCell>
        <TableCell className="text-right text-muted-foreground">
          <RelativeTime time={member.created_at} />
        </TableCell>
        <TableCell className="text-right">
          <div className="flex items-center justify-end">
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    className="text-muted-foreground"
                  >
                    <span className="sr-only">Open options</span>
                    <EllipsisVerticalIcon
                      className="size-5"
                      aria-hidden="true"
                    />
                  </Button>
                }
              />
              <DropdownMenuContent align="end" className="w-28">
                <DropdownMenuItem
                  onClick={() => setUpdateUserNamespaceModal(true)}
                >
                  Update
                </DropdownMenuItem>
                <DropdownMenuItem
                  variant="destructive"
                  onClick={() => deleteMember()}
                >
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </TableCell>
      </TableRow>

      <Dialog
        open={updateUserNamespaceModal}
        onOpenChange={setUpdateUserNamespaceModal}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogTitle className="border-b pb-4">Update member</DialogTitle>
          <div className="flex flex-col gap-0 mt-4">
            {/* <div className="grid grid-cols-6 gap-4">
                <div className="col-span-2 flex flex-row">
                  <label htmlFor="usernameText" className="block text-sm font-medium leading-6 text-gray-900 my-auto">
                    <div className="flex">
                      <span className="text-red-600">*</span>
                      <span className="leading-6 ">User</span>
                      <span>:</span>
                    </div>
                  </label>
                </div>
                <div className="col-span-4">
                  <Popover>
                    <PopoverTrigger
                      render={
                        <Button variant="outline" className="w-full justify-between font-normal">
                          {userSelected?.username || "Select user"}
                          <ChevronUpDownIcon className="size-4 text-muted-foreground" aria-hidden="true" />
                        </Button>
                      }
                    />
                    <PopoverContent align="start" className="w-(--anchor-width) p-0">
                      <Command shouldFilter={false}>
                        <CommandInput
                          placeholder="Search user"
                          value={userSearch}
                          onValueChange={setUserSearch}
                        />
                        <CommandList>
                          <CommandEmpty>Nothing found.</CommandEmpty>
                          <CommandGroup>
                            {userList?.map(user => (
                              <CommandItem
                                key={user.id}
                                value={user.username}
                                onSelect={() => setUserSelected(user)}
                              >
                                {user.username}
                              </CommandItem>
                            ))}
                          </CommandGroup>
                        </CommandList>
                      </Command>
                    </PopoverContent>
                  </Popover>
                </div>
              </div> */}
            {/* <div className="grid grid-cols-6 gap-4">
                <div className="col-span-2"></div>
                <div className="col-span-4">
                  {
                    userSelectedValid ? null : (
                      <p className="mt-1 text-xs text-red-600">
                        <span>
                          Please select a user.
                        </span>
                      </p>
                    )
                  }
                </div>
              </div> */}
            <div className="grid grid-cols-6 gap-4">
              <div className="col-span-2 flex flex-row">
                <label
                  htmlFor="usernameText"
                  className="block text-sm font-medium leading-6 text-gray-900 my-auto"
                >
                  <div className="flex">
                    <span className="text-red-600">*</span>
                    <span className="leading-6 ">Role</span>
                    <span>:</span>
                  </div>
                </label>
              </div>
              <div className="col-span-4">
                <Select
                  value={addNamespaceRoleRole}
                  onValueChange={(source) => {
                    if (source) setAddNamespaceRoleRole(source);
                  }}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {namespaceRoles.map((source) => (
                      <SelectItem key={source.name} value={source.name}>
                        {source.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setUpdateUserNamespaceModal(false)}
            >
              Cancel
            </Button>
            <Button onClick={() => updateMember()}>Update</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
