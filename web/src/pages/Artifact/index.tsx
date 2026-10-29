/**
 * Copyright 2026 sigma
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
import humanFormat from "human-format";
import { CheckIcon, Info } from "lucide-react";
import { Fragment, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { useParams, useSearchParams } from "react-router-dom";
import { useCopyToClipboard } from "react-use";

import Header from "@/components/Header";
import IMenu from "@/components/Menu";
import Notification from "@/components/Notification";
import Pagination from "@/components/Pagination";
import RepositoryTabs from "@/components/RepositoryTabs";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Tooltip as UITooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useTranslation } from "@/i18n/useTranslation";
import {
  IArtifactLayers,
  IArtifactVulnerabilities,
  IEndpoint,
  IHTTPError,
  ISystemConfig,
} from "@/interfaces";
import Settings from "@/Settings";
import { trimHTTP } from "@/utils";
import Vulnerabilities, { sortVulnerabilities } from "./Vulnerabilities";

// pullReference builds the digest reference of the artifact. The page is opened from a tag row, but
// the reference shown here always points at the manifest digest, so a moved tag cannot change it.
function pullReference(endpoint: string, repository: string, digest: string) {
  const image = endpoint === "" ? repository : `${trimHTTP(endpoint)}/${repository}`;
  return `${image}@${digest}`;
}

// layerIndexByDigest maps a layer digest to its 1-based position in the manifest, so a vulnerability
// can name the layer that introduced the vulnerable package.
function layerIndexByDigest(layers?: IArtifactLayers) {
  const indexes = new Map<string, number>();
  for (const layer of layers?.layers ?? []) {
    indexes.set(layer.digest, layer.index);
  }
  return indexes;
}

export default function Artifact({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();

  const repository = searchParams.get("repository") ?? "";
  const repositoryId = searchParams.get("repository_id") ?? "";
  const namespaceId = searchParams.get("namespace_id") ?? "";
  const digest = searchParams.get("digest") ?? "";

  const [endpoint, setEndpoint] = useState("");
  const [gotConfig, setGotConfig] = useState(false);
  const [config, setConfig] = useState<ISystemConfig>({ daemon: { builder: false } } as ISystemConfig);
  const [layers, setLayers] = useState<IArtifactLayers>();
  const [vulnerabilities, setVulnerabilities] = useState<IArtifactVulnerabilities>();

  // The vulnerabilities are rendered on one tab and paged by the page itself, so the pagination can
  // stay pinned to the bottom of the page instead of scrolling away with the table.
  const [tab, setTab] = useState("layers");
  const [vulnerabilityPage, setVulnerabilityPage] = useState(1);

  const [, copyToClipboard] = useCopyToClipboard();
  const [copiedKey, setCopiedKey] = useState("");
  const copiedTimer = useRef<number | undefined>(undefined);

  // copiedKey marks the element whose copy just happened, so the confirmation is shown next to what
  // was copied instead of in a floating tooltip.
  const flashCopied = useCallback((key: string) => {
    setCopiedKey(key);
    window.clearTimeout(copiedTimer.current);
    copiedTimer.current = window.setTimeout(() => setCopiedKey(""), 1600);
  }, []);

  useEffect(() => () => window.clearTimeout(copiedTimer.current), []);

  const notify = useCallback((error: unknown) => {
    const errorcode = ((error as { response?: { data?: IHTTPError } }).response?.data ?? {}) as IHTTPError;
    Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
  }, []);

  useEffect(() => {
    axios.get(localServer + "/api/v1/systems/endpoint")
      .then(response => {
        if (response.status === 200) {
          setEndpoint((response.data as IEndpoint).endpoint);
        }
      })
      .catch(notify);
    axios.get(localServer + "/api/v1/systems/config")
      .then(response => {
        if (response.status === 200) {
          setConfig(response.data as ISystemConfig);
          setGotConfig(true);
        }
      })
      .catch(notify);
  }, [localServer, notify]);

  useEffect(() => {
    if (!namespaceId || !repositoryId || digest === "") {
      return;
    }
    axios.get(`${localServer}/api/v1/namespaces/${namespaceId}/repositories/${repositoryId}/artifacts/${encodeURIComponent(digest)}/layers`)
      .then(response => {
        if (response.status === 200) {
          setLayers(response.data as IArtifactLayers);
        }
      })
      .catch(notify);
  }, [localServer, namespaceId, repositoryId, digest, notify]);

  useEffect(() => {
    if (!namespaceId || !repositoryId || digest === "") {
      return;
    }
    axios.get(`${localServer}/api/v1/namespaces/${namespaceId}/repositories/${repositoryId}/artifacts/${encodeURIComponent(digest)}/vulnerabilities`)
      .then(response => {
        if (response.status === 200) {
          setVulnerabilities(response.data as IArtifactVulnerabilities);
        }
      })
      .catch(notify);
  }, [localServer, namespaceId, repositoryId, digest, notify]);

  const pull = pullReference(endpoint, repository, digest);

  const vulnerabilityRows = useMemo(
    () => sortVulnerabilities(vulnerabilities?.items ?? []),
    [vulnerabilities],
  );
  const vulnerabilityPageCount = Math.max(1, Math.ceil(vulnerabilityRows.length / Settings.PageSize));
  // The rows can shrink when the report is reloaded, so the page is clamped instead of trusted.
  const vulnerabilityCurrentPage = Math.min(vulnerabilityPage, vulnerabilityPageCount);
  const vulnerabilityPageRows = vulnerabilityRows.slice(
    (vulnerabilityCurrentPage - 1) * Settings.PageSize,
    vulnerabilityCurrentPage * Settings.PageSize,
  );
  const showPagination = tab === "vulnerabilities" && vulnerabilities?.status === "Success" && vulnerabilityRows.length > 0;

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>{t("artifact.title")}</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white dark:bg-gray-950">
        <IMenu
          localServer={localServer}
          item="tags"
          namespace={namespace}
          namespace_id={namespaceId}
          repository={repository}
          repository_id={repositoryId}
        />
        <div className="flex flex-col w-0 flex-1 overflow-hidden max-h-screen">
          <main>
            <Header
              title={t("artifact.title")}
              props={
                gotConfig && (
                  <RepositoryTabs
                    namespace={namespace || ""}
                    namespaceId={namespaceId}
                    repository={repository}
                    repositoryId={repositoryId}
                    active="tags"
                    showBuilder={config.daemon.builder}
                  />
                )
              }
            />
          </main>
          <div className="flex-1 overflow-y-auto px-4 py-4">
            <div className="rounded-lg border border-border bg-card p-4">
              <div className="flex items-center gap-1.5">
                <div className="text-xs font-medium text-muted-foreground">{t("artifact.pullReference")}</div>
                <UITooltip>
                  <TooltipTrigger>
                    <Info className="size-3.5 text-muted-foreground" aria-hidden="true" />
                  </TooltipTrigger>
                  <TooltipContent>{t("artifact.pullHint")}</TooltipContent>
                </UITooltip>
              </div>
              <code
                className="relative mt-1 block cursor-pointer overflow-hidden break-all rounded-md bg-gray-700 px-3 py-2 pr-16 font-mono text-xs text-gray-50"
                onClick={() => {
                  copyToClipboard(pull);
                  flashCopied("pull");
                }}
              >
                {pull}
                {copiedKey === "pull" ? (
                  <span className="absolute inset-y-0 right-0 flex items-center gap-1 bg-gray-700 px-2 text-green-400">
                    <CheckIcon className="size-3.5" />
                    {t("common.copied")}
                  </span>
                ) : null}
              </code>
              <dl className="mt-4 grid grid-cols-2 gap-4 text-sm sm:grid-cols-3 lg:grid-cols-5">
                <div className="col-span-2 min-w-0">
                  <dt className="text-xs text-muted-foreground">{t("artifact.digest")}</dt>
                  <dd className="truncate font-mono text-xs" title={digest}>{digest}</dd>
                </div>
                <div className="min-w-0">
                  <dt className="text-xs text-muted-foreground">{t("artifact.mediaType")}</dt>
                  <dd className="truncate text-xs" title={layers?.media_type}>{layers?.media_type ?? "-"}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground">{t("artifact.size")}</dt>
                  <dd>{layers === undefined ? "-" : humanFormat(layers.blob_size || 0)}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground">{t("artifact.osArch")}</dt>
                  <dd>{layers === undefined || !layers.os ? "-" : `${layers.os}/${layers.architecture}`}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground">{t("artifact.distro")}</dt>
                  <dd>{layers?.distro === "" || layers === undefined ? "-" : layers.distro}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground">{t("artifact.pushedAt")}</dt>
                  <dd>{layers === undefined || layers.pushed_at === "" ? "-" : layers.pushed_at}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground">{t("artifact.pullTimes")}</dt>
                  <dd>{layers?.pull_times ?? 0}</dd>
                </div>
              </dl>
            </div>

            <Tabs value={tab} onValueChange={value => setTab(String(value))} className="mt-4">
              <TabsList>
                <TabsTrigger value="layers">
                  {t("artifact.layers")}
                  <TabInfoHint hint={t("artifact.layersHint")} />
                  {layers === undefined ? null : ` (${layers.layers?.length ?? 0})`}
                </TabsTrigger>
                <TabsTrigger value="vulnerabilities">
                  {t("artifact.vulnerabilities")}
                  <TabInfoHint hint={t("artifact.sortHint")} />
                  {vulnerabilities === undefined ? null : ` (${vulnerabilities.total})`}
                </TabsTrigger>
              </TabsList>
              <TabsContent value="layers">
                <LayersTable layers={layers} />
              </TabsContent>
              <TabsContent value="vulnerabilities">
                <Vulnerabilities
                  report={vulnerabilities}
                  rows={vulnerabilityPageRows}
                  layerIndex={layerIndexByDigest(layers)}
                />
              </TabsContent>
            </Tabs>
          </div>
          {showPagination && (
            <Pagination
              limit={Settings.PageSize}
              page={vulnerabilityCurrentPage}
              setPage={setVulnerabilityPage}
              total={vulnerabilityRows.length}
            />
          )}
        </div>
      </div>
    </Fragment>
  );
}

// TabInfoHint is the info icon of a tab: it explains what the tab shows. The tab itself renders a
// button, so the tooltip trigger is a span to keep the markup valid.
function TabInfoHint({ hint }: { hint: string }) {
  return (
    <UITooltip>
      <TooltipTrigger render={<span className="inline-flex items-center" />}>
        <Info className="size-3.5 text-muted-foreground" aria-hidden="true" />
      </TooltipTrigger>
      <TooltipContent>{hint}</TooltipContent>
    </UITooltip>
  );
}

// LayersTable lists every layer of the manifest on one page: an image has a handful of layers, and
// the build command of a layer is the reason to look at it.
function LayersTable({ layers }: { layers?: IArtifactLayers }) {
  const { t } = useTranslation();

  if (layers === undefined) {
    return null;
  }
  if ((layers.layers?.length ?? 0) === 0) {
    return <p className="py-6 text-sm text-muted-foreground">{t("artifact.layersEmpty")}</p>;
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-border">
      <table className="min-w-full">
        <thead>
          <tr className="bg-muted/40">
            <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
              {t("artifact.layer")}
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
              {t("artifact.digest")}
            </th>
            <th className="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">
              {t("artifact.size")}
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
              {t("artifact.mediaType")}
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
              {t("artifact.layerCommand")}
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {layers.layers.map(layer => (
            <tr key={layer.digest} className="align-top">
              <td className="px-4 py-3 text-sm text-muted-foreground">{layer.index}</td>
              <td className="px-4 py-3">
                <code className="block max-w-88 truncate font-mono text-xs" title={layer.digest}>{layer.digest}</code>
              </td>
              <td className="whitespace-nowrap px-4 py-3 text-right text-sm">{humanFormat(layer.size || 0)}</td>
              <td className="px-4 py-3 text-xs text-muted-foreground">{layer.media_type}</td>
              <td className="px-4 py-3">
                <code className="block max-w-144 whitespace-pre-wrap wrap-break-word font-mono text-xs">
                  {layer.command === "" ? "-" : layer.command}
                </code>
                {layer.created_at === "" ? null : (
                  <div className="mt-1 text-[11px] text-muted-foreground">{layer.created_at}</div>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
