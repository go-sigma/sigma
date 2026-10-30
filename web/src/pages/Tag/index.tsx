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
  EllipsisVerticalIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/20/solid";
import axios from "axios";
import dayjs from "dayjs";
import humanFormat from "human-format";
import { Fragment, useCallback, useEffect, useRef, useState } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { CheckIcon, CornerDownLeft } from "lucide-react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { useCopyToClipboard } from "react-use";

import Header from "@/components/Header";
import IMenu from "@/components/Menu";
import Notification from "@/components/Notification";
import Toast from "@/components/Notification";
import Pagination from "@/components/Pagination";
import RepositoryTabs from "@/components/RepositoryTabs";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
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
import { useTranslation } from "@/i18n/useTranslation";
import {
  IArtifact,
  IEndpoint,
  IHTTPError,
  IImageConfig,
  INamespaceItem,
  ISbom,
  ISystemConfig,
  ITagItem,
  ITagList,
  IUserSelf,
  IVuln,
} from "@/interfaces";
import Settings from "@/Settings";
import { trimHTTP } from "@/utils";
import distros, { distroName } from "@/utils/distros";

export default function Tag({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [tagList, setTagList] = useState<ITagList>({} as ITagList);
  const [refresh, setRefresh] = useState({});
  const [page, setPage] = useState(1);
  const [searchTag, setSearchTag] = useState("");
  const [total, setTotal] = useState(0);

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repository = searchParams.get("repository");
  const repositoryId = searchParams.get("repository_id");
  const namespaceId = searchParams.get("namespace_id");

  const [, copyToClipboard] = useCopyToClipboard();

  // copiedKey marks the element whose copy just happened, so the copy confirmation is shown next
  // to what was copied instead of in a floating tooltip.
  const [copiedKey, setCopiedKey] = useState("");
  const copiedTimer = useRef<number | undefined>(undefined);

  const flashCopied = useCallback((key: string) => {
    setCopiedKey(key);
    window.clearTimeout(copiedTimer.current);
    copiedTimer.current = window.setTimeout(() => setCopiedKey(""), 1600);
  }, []);

  useEffect(() => () => window.clearTimeout(copiedTimer.current), []);

  // openArtifact opens the detail page of the artifact of one row of the tag card.
  const openArtifact = (artifact: IArtifact) => {
    const params = new URLSearchParams({
      namespace_id: namespaceId || "",
      repository_id: repositoryId || "",
      repository: repository || "",
      digest: artifact.digest,
    });
    navigate(`/namespaces/${namespace}/repository/artifact?${params.toString()}`);
  };

  // pullText is the image reference shown in the tag card. The pull verb is left out so the whole
  // reference fits the card, and copying it yields exactly what is displayed.
  const pullText = (tag: ITagItem) => {
    const image = `${trimHTTP(endpoint)}/${repository}`;
    if (
      tag.artifact.config_media_type ===
      "application/vnd.cncf.helm.config.v1+json"
    ) {
      return `${image} --version ${tag.name}`;
    }
    return `${image}:${tag.name}`;
  };

  const [endpoint, setEndpoint] = useState("");

  const [, setUserObj] = useState<IUserSelf>({} as IUserSelf);

  const [deleteTagModal, setDeleteTagModal] = useState(false);
  const [deleteTagModalTag, setDeleteTagModalTag] = useState<ITagItem>(
    {} as ITagItem,
  );

  useEffect(() => {
    axios
      .get(localServer + "/api/v1/users/self")
      .then((response) => {
        if (response.status === 200) {
          const user = response.data as IUserSelf;
          setUserObj(user);
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({
            level: "warning",
            title: errorcode.title,
            message: errorcode.description,
          });
        }
      })
      .catch((error) => {
        const errorcode = error.response.data as IHTTPError;
        Toast({
          level: "warning",
          title: errorcode.title,
          message: errorcode.description,
        });
      });
  }, [localServer]);

  const [, setNamespaceObj] = useState<INamespaceItem>({} as INamespaceItem);

  useEffect(() => {
    if (namespaceId == null || namespaceId == "") {
      return;
    }
    axios
      .get(`${localServer}/api/v1/namespaces/${namespaceId}`)
      .then((response) => {
        if (response.status == 200) {
          const namespaceObj = response.data as INamespaceItem;
          setNamespaceObj(namespaceObj);
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
  }, [localServer, namespaceId]);

  useEffect(() => {
    let url = localServer + `/api/v1/systems/endpoint`;
    axios
      .get(url)
      .then((response) => {
        if (response.status === 200) {
          let e = response.data as IEndpoint;
          setEndpoint(e.endpoint);
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({
            level: "warning",
            title: errorcode.title,
            message: errorcode.description,
          });
        }
      })
      .catch((error) => {
        const errorcode = error.response.data as IHTTPError;
        Toast({
          level: "warning",
          title: errorcode.title,
          message: errorcode.description,
        });
      });
  }, [localServer]);

  const fetchTags = useCallback(() => {
    // Without a resolved namespace and repository there is nothing to ask for; building the URL
    // anyway sent "/namespaces/null/..." and surfaced a 404 toast.
    if (!namespaceId || !repositoryId || !repository) {
      return;
    }
    let url =
      localServer +
      `/api/v1/namespaces/${namespaceId}/repositories/${repositoryId}/tags/?repository=${repository}&limit=${Settings.PageSize}&page=${page}&type=Image&type=ImageIndex&type=Chart&type=Sif`;
    if (searchTag !== "") {
      url += `&name=${searchTag}`;
    }
    axios
      .get(url)
      .then((response) => {
        if (response.status === 200) {
          const tagListData = response.data as ITagList;
          setTagList(tagListData);
          setTotal(tagListData.total);
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({
            level: "warning",
            title: errorcode.title,
            message: errorcode.description,
          });
        }
      })
      .catch((error) => {
        const errorcode = error.response.data as IHTTPError;
        Toast({
          level: "warning",
          title: errorcode.title,
          message: errorcode.description,
        });
      });
  }, [localServer, namespaceId, repositoryId, repository, page, searchTag]);

  useEffect(fetchTags, [fetchTags, refresh]);

  const [gotConfig, setGotConfig] = useState(false);
  const [config, setConfig] = useState<ISystemConfig>({
    daemon: {
      builder: true,
    },
  } as ISystemConfig);

  useEffect(() => {
    axios
      .get(localServer + "/api/v1/systems/config")
      .then((response) => {
        if (response.status === 200) {
          const configData = response.data as ISystemConfig;
          setConfig(configData);
          setGotConfig(true);
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({
            level: "warning",
            title: errorcode.title,
            message: errorcode.description,
          });
        }
      })
      .catch((error) => {
        const errorcode = error.response.data as IHTTPError;
        Toast({
          level: "warning",
          title: errorcode.title,
          message: errorcode.description,
        });
      });
  }, [localServer]);

  const deleteTag = (tagId: number) => {
    axios
      .delete(
        `${localServer}/api/v1/namespaces/${namespaceId}/repositories/${repositoryId}/tags/${tagId}`,
      )
      .then((response) => {
        if (response.status === 204) {
          setRefresh({});
        } else {
          const errorcode = response.data as IHTTPError;
          Toast({
            level: "warning",
            title: errorcode.title,
            message: errorcode.description,
          });
        }
      })
      .catch((error) => {
        const errorcode = error.response.data as IHTTPError;
        Toast({
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
          <title>{t("common.tags")}</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white dark:bg-gray-950">
        <IMenu
          localServer={localServer}
          item="tags"
          namespace={namespace}
          namespace_id={namespaceId || ""}
          repository={repository || ""}
          repository_id={repositoryId || ""}
        />
        <div className="flex flex-col w-0 flex-1 overflow-hidden h-screen">
          <main className="">
            <Header
              title={t("common.tags")}
              props={
                gotConfig && (
                  <RepositoryTabs
                    namespace={namespace || ""}
                    namespaceId={namespaceId || ""}
                    repository={repository || ""}
                    repositoryId={repositoryId || ""}
                    active="tags"
                    showBuilder={config.daemon.builder}
                  />
                )
              }
            />
            <div className="pt-4 pb-4 flex">
              <div className="px-4">
                <div className="flex gap-4">
                  <div className="relative flex items-center">
                    <Label
                      htmlFor="tagSearch"
                      className="absolute -top-2 left-2 inline-block bg-background px-1 text-xs font-medium text-foreground z-10"
                    >
                      {t("common.tag")}
                    </Label>
                    <Input
                      id="tagSearch"
                      placeholder={t("tag.searchPlaceholder")}
                      value={searchTag}
                      onChange={(e) => {
                        setSearchTag(e.target.value);
                      }}
                      onKeyDown={(e) => {
                        if (e.key == "Enter") {
                          fetchTags();
                        }
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
            </div>
          </main>
          <div className="flex-1 overflow-y-auto">
            <Dialog open={deleteTagModal} onOpenChange={setDeleteTagModal}>
              <DialogContent className="sm:max-w-lg">
                <div className="flex items-start gap-4">
                  <div className="mx-auto flex size-10 shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0">
                    <ExclamationTriangleIcon
                      className="size-6 text-red-600"
                      aria-hidden="true"
                    />
                  </div>
                  <div className="text-center sm:text-left">
                    <DialogTitle>Delete tag</DialogTitle>
                    <DialogDescription className="mt-2">
                      Are you sure you want to delete the tag{" "}
                      <span className="text-foreground font-medium">
                        {deleteTagModalTag.name}
                      </span>
                    </DialogDescription>
                  </div>
                </div>
                <DialogFooter>
                  <Button
                    variant="outline"
                    onClick={() => setDeleteTagModal(false)}
                  >
                    Cancel
                  </Button>
                  <Button
                    variant="destructive"
                    onClick={() => {
                      setDeleteTagModal(false);
                      deleteTag(deleteTagModalTag.id);
                    }}
                  >
                    Delete
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
            <div className="flex flex-col border-b border-gray-200">
              {tagList.items?.map((tag) => {
                return tag.artifact.type === "cosign" ? null : (
                  <div
                    key={tag.id}
                    className="p-4 border-t border-gray-200 hover:shadow-md last:hover:shadow-none"
                  >
                    {/* first row begin */}
                    <div className="flex">
                      <div className="flex-1 flex items-center gap-2">
                        <span
                          className="font-semibold text-gray-600 cursor-pointer"
                          onClick={() => {
                            copyToClipboard(`${tag.name}`);
                            flashCopied(`tag-${tag.id}`);
                          }}
                        >
                          {tag.name}
                        </span>
                        {copiedKey === `tag-${tag.id}` ? (
                          <span className="flex items-center gap-1 text-xs text-green-600 dark:text-green-500">
                            <CheckIcon className="size-3.5" />
                            {t("common.copied")}
                          </span>
                        ) : null}
                      </div>
                      <div>
                        <code
                          className="relative block text-xs bg-gray-700 p-2 text-gray-50 cursor-pointer rounded-md w-70 text-ellipsis whitespace-nowrap overflow-hidden"
                          onClick={() => {
                            copyToClipboard(pullText(tag));
                            flashCopied(`pull-${tag.id}`);
                          }}
                        >
                          {pullText(tag)}
                          {copiedKey === `pull-${tag.id}` ? (
                            <span className="absolute inset-y-0 right-0 flex items-center gap-1 bg-gray-700 px-2 text-green-400">
                              <CheckIcon className="size-3.5" />
                              {t("common.copied")}
                            </span>
                          ) : null}
                        </code>
                      </div>
                      <div className="flex flex-col justify-center">
                        <DropdownMenu>
                          <DropdownMenuTrigger
                            render={
                              <Button
                                variant="ghost"
                                size="icon-sm"
                                className="text-gray-500"
                              >
                                <span className="sr-only">Open options</span>
                                <EllipsisVerticalIcon
                                  className="size-5"
                                  aria-hidden="true"
                                />
                              </Button>
                            }
                          />
                          <DropdownMenuContent align="end" className="w-20">
                            <DropdownMenuItem
                              variant="destructive"
                              onClick={() => {
                                setDeleteTagModal(true);
                                setDeleteTagModalTag(tag);
                              }}
                            >
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </div>
                    {/* first row end */}

                    {/* second row begin */}
                    <div className="text-xs text-gray-600">
                      Last pushed{" "}
                      <span className="font-semibold">
                        {dayjs().to(
                          dayjs.utc(tag.pushed_at).tz(dayjs.tz.guess()),
                        )}
                      </span>
                    </div>
                    <div className="mt-2 text-xs text-gray-600">
                      Pull times{" "}
                      <span className="font-semibold">
                        {tag.pull_times === undefined ? 0 : tag.pull_times}
                      </span>
                    </div>
                    {/* second row end */}

                    {/* third row begin */}
                    <table className="mt-2 min-w-full">
                      <thead>
                        <tr className="">
                          <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                            Digest
                          </th>
                          <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                            Type
                          </th>
                          <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                            Distro
                          </th>
                          <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                            Os/Arch
                          </th>
                          <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                            Signing
                          </th>
                          <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                            Last pull
                          </th>
                          {/* <th className="pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                              Pull Times
                            </th> */}
                          <th className="pb-1 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                            Vulnerabilities
                          </th>
                          <th className="pb-1 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                            Size
                          </th>
                        </tr>
                      </thead>
                      <TableItem
                        artifact={tag.artifact}
                        artifacts={tag.artifacts}
                        onOpen={openArtifact}
                      />
                    </table>
                    {/* third row end */}
                  </div>
                );
              })}
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
    </Fragment>
  );
}

// ManifestJSON is the subset of an image manifest or index the tag list reads.
type ManifestJSON = {
  mediaType?: string;
  config?: { mediaType?: string };
  layers?: { mediaType?: string; annotations?: Record<string, string> }[];
};

// parseArtifactJSON parses one of the JSON encoded artifact fields and returns the fallback when
// the field is empty or not valid JSON. Artifacts do not always carry a config, sbom or
// vulnerability report (raw blobs, for example) and the API answers with an empty string for
// those, which JSON.parse rejects.
function parseArtifactJSON<T>(text: string | undefined | null, fallback: T): T {
  if (!text) {
    return fallback;
  }
  try {
    return JSON.parse(text) as T;
  } catch {
    return fallback;
  }
}

function skipManifest(raw: string) {
  const artifactObj = parseArtifactJSON<ManifestJSON>(raw, {});
  if (artifactObj.config?.mediaType !== "application/vnd.oci.image.config.v1+json") {
    return false;
  }
  const layers = artifactObj.layers;
  if (layers?.length !== 1 || layers[0].mediaType !== "application/vnd.in-toto+json") {
    return false;
  }
  return (layers[0].annotations?.["in-toto.io/predicate-type"] ?? "") !== "";
}

function TableItem({
  artifact,
  artifacts,
  onOpen,
}: {
  artifact: IArtifact;
  artifacts: IArtifact[];
  onOpen: (artifact: IArtifact) => void;
}) {
  const artifactObj = parseArtifactJSON<ManifestJSON>(artifact.raw, {});

  return (
    <tbody>
      {artifactObj.mediaType === "application/vnd.oci.image.manifest.v1+json" ||
        artifactObj.mediaType ===
        "application/vnd.docker.distribution.manifest.v2+json" ||
        artifact.config_media_type ==
        "application/vnd.cncf.helm.config.v1+json" ? (
        <DetailItem artifact={artifact} onOpen={() => onOpen(artifact)} />
      ) : artifactObj.mediaType ===
        "application/vnd.docker.distribution.manifest.list.v2+json" ||
        artifactObj.mediaType === "application/vnd.oci.image.index.v1+json" ? (
        (artifacts ?? []).map((item: IArtifact) => {
          return (
            !skipManifest(item.raw) && (
              <DetailItem key={item.digest} artifact={item} onOpen={() => onOpen(item)} />
            )
          );
        })
      ) : (
        <tr></tr>
      )}
    </tbody>
  );
}

function cutDigest(digest: string) {
  if (digest === undefined) {
    return "";
  }
  if (digest.indexOf(":") < 0) {
    return "";
  }
  return digest.substring(digest.indexOf(":") + 1, digest.indexOf(":") + 13);
}

function DetailItem({ artifact, onOpen }: { artifact: IArtifact; onOpen: () => void }) {
  const sbomObj = parseArtifactJSON<ISbom>(artifact.sbom, {} as ISbom);
  const vulnerabilityObj = parseArtifactJSON<IVuln>(
    artifact.vulnerability,
    {} as IVuln,
  );
  const imageConfigObj = parseArtifactJSON<IImageConfig>(
    artifact.config_raw,
    {} as IImageConfig,
  );
  return (
    <tr className="hover:bg-gray-50 cursor-pointer" onClick={onOpen}>
      <td className="text-left w-45">
        <code className="text-xs underline underline-offset-1 text-primary hover:text-primary/80">
          {cutDigest(artifact.digest)}
        </code>
      </td>
      <td className="text-left text-xs w-45 capitalize">{artifact.type}</td>
      <td className="text-left text-xs w-45">
        <div className="flex gap-1">
          {distros(sbomObj.distro?.name) === "" ? (
            ""
          ) : (
            <img
              src={"/distros/" + distros(sbomObj.distro.name)}
              alt={sbomObj.distro.name}
              className="w-4 h-4 inline relative"
            />
          )}
          <div className="">
            {distroName(sbomObj.distro?.name) === ""
              ? "-"
              : distroName(sbomObj.distro.name) + " " + sbomObj.distro.version}
          </div>
        </div>
      </td>
      <td className="text-left text-xs w-45">
        {imageConfigObj.os === undefined ||
          imageConfigObj.architecture === undefined ||
          imageConfigObj.os === "" ||
          imageConfigObj.architecture === "" ? (
          "-"
        ) : (
          <span>
            {imageConfigObj.os}/{imageConfigObj.architecture}
          </span>
        )}
      </td>
      <td className="text-left text-xs w-45">Verified</td>
      <td className="text-left text-xs w-45">
        {(artifact.pull_times || 0) > 0
          ? dayjs().to(dayjs(artifact.last_pull))
          : "Never pulled"}
      </td>
      {/* <td className="text-right text-xs w-45">
        {artifact.pull_times}
      </td> */}
      <td className="text-left text-xs w-55">
        <span className="bg-red-800 text-white text-xs font-medium mr-1 px-2 py-0.5 dark:bg-red-900 dark:text-red-300">
          <span>{vulnerabilityObj.critical || 0}</span> C
        </span>
        <span className="bg-red-300 text-gray-800 text-xs font-medium mr-1 px-2 py-0.5 dark:bg-red-900 dark:text-red-300">
          {vulnerabilityObj.high || 0} H
        </span>
        <span className="bg-amber-400 text-gray-800 text-xs font-medium mr-1 px-2 py-0.5 dark:bg-red-900 dark:text-red-300">
          {vulnerabilityObj.medium || 0} M
        </span>
        <span className="bg-amber-200 text-gray-800 text-xs font-medium px-2 py-0.5 dark:bg-red-900 dark:text-red-300">
          {vulnerabilityObj.low || 0} L
        </span>
      </td>
      <td className="text-right text-xs w-45">
        {humanFormat(artifact.blob_size || 0)}
      </td>
    </tr>
  );
}
