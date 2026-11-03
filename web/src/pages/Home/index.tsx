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
import { Activity, ArrowUpRight, Boxes, ChevronRight, HardDrive, Tags } from "lucide-react";
import { Fragment, useEffect, useMemo, useState, type ComponentType, type ReactNode } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { Link, useNavigate } from "react-router-dom";

import Header from "@/components/Header";
import Menu from "@/components/Menu";
import Notification from "@/components/Notification";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useTranslation } from "@/i18n/useTranslation";
import { IDailyActivity, IDailyActivityList, IHTTPError, IOverview } from "@/interfaces";
import { cn } from "@/lib/utils";

import ActivityHeatmap from "./ActivityHeatmap";
import ActivityTrend from "./ActivityTrend";
import BuildTrend from "./BuildTrend";
import CardTitleWithHint from "./CardTitleWithHint";
import QuickStart from "./QuickStart";
import { formatBytes } from "./format";

// The shares of a namespace quota that deserve a warning color.
const quotaWarning = 0.8;
// The heatmap and the trend chart share one activity window.
const activityDays = 371;

export default function Home({ localServer }: { localServer: string }) {
  const { locale, t } = useTranslation();
  const navigate = useNavigate();
  const [overview, setOverview] = useState<IOverview>();
  const [activity, setActivity] = useState<IDailyActivity[]>();

  useEffect(() => {
    let cancelled = false;
    axios.get(`${localServer}/api/v1/overview`).then(response => {
      if (!cancelled && response.status === 200) {
        setOverview(response.data as IOverview);
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: t("home.loadFailed"), message: errorcode?.description });
    });
    axios.get(`${localServer}/api/v1/activity/daily?days=${activityDays}`).then(response => {
      if (!cancelled && response.status === 200) {
        setActivity((response.data as IDailyActivityList).items ?? []);
      }
    }).catch(() => {
      // The activity charts simply stay empty; the overview already reported the failure.
    });
    return () => {
      cancelled = true;
    };
  }, [localServer, t]);

  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);

  const loading = overview === undefined;
  const formatCount = (value: number | undefined) => (value === undefined ? "—" : numberFormatter.format(value));

  const storagePercent = overview !== undefined && overview.storage.capacity > 0
    ? Math.min((overview.storage.used / overview.storage.capacity) * 100, 100)
    : undefined;

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>{t("header.home")}</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background">
        <Menu item="home" localServer={localServer} />
        <div className="flex flex-col flex-1 h-screen">
          <main className="relative z-0 flex min-h-0 flex-1 flex-col focus:outline-none" tabIndex={0}>
            <Header title={t("header.home")} />
            <div className="min-h-0 flex-1 overflow-y-auto">
              <div className="mx-auto w-full max-w-[1440px] space-y-4 px-[0.8rem] pb-10 pt-4 sm:px-[1.2rem] lg:px-[1.6rem]">
                <div className="grid gap-x-4 gap-y-3 sm:grid-cols-2 xl:grid-cols-4">
                  <MetricCard
                    icon={HardDrive}
                    label={t("home.metricStorage")}
                    value={formatBytes(overview?.storage.used ?? 0)}
                    footer={
                      <div className="mt-3 space-y-1.5">
                        {storagePercent !== undefined && (
                          <div className="h-1.5 w-full overflow-hidden rounded-full bg-muted">
                            <div
                              className={cn(
                                "h-full rounded-full transition-all",
                                storagePercent >= quotaWarning * 100 ? "bg-destructive" : "bg-primary",
                              )}
                              style={{ width: `${storagePercent}%` }}
                            />
                          </div>
                        )}
                        <div className="text-xs text-muted-foreground">
                          {storagePercent !== undefined
                            ? t("home.metricStorageHint", {
                              limit: formatBytes(overview?.storage.capacity ?? 0),
                              percent: storagePercent.toFixed(1),
                            })
                            : t(overview?.scope === "global" ? "home.metricStorageNoCapacity" : "home.metricStorageSelfHint", {
                              used: formatBytes(overview?.storage.used ?? 0),
                            })}
                        </div>
                      </div>
                    }
                  />
                  <MetricCard
                    icon={Boxes}
                    label={t("home.metricNamespaces")}
                    value={formatCount(overview?.counts.namespaces)}
                    hint={t("home.metricNamespacesHint", {
                      repositories: numberFormatter.format(overview?.counts.repositories ?? 0),
                    })}
                  />
                  <MetricCard
                    icon={Tags}
                    label={t("home.metricTags")}
                    value={formatCount(overview?.counts.tags)}
                    hint={t("home.metricTagsHint", {
                      repositories: numberFormatter.format(overview?.counts.repositories ?? 0),
                    })}
                  />
                  <MetricCard
                    icon={Activity}
                    label={t("home.metricPushes")}
                    value={formatCount(overview?.activity.push)}
                    hint={
                      <span className="inline-flex items-center gap-1">
                        <ArrowUpRight className="size-3.5 text-green-600 dark:text-green-500" />
                        {t("home.metricPushesDelta", { delta: `${overview?.activity.push_delta ?? 0}%` })}
                      </span>
                    }
                  />
                </div>

                <QuickStart localServer={localServer} />

                <div className="grid gap-4 lg:grid-cols-2">
                  <ActivityHeatmap items={activity} days={activityDays} />
                  <ActivityTrend items={activity} />
                </div>

                <div className="grid gap-4 lg:auto-rows-[26rem] lg:grid-cols-2">
                  <Card className="h-full gap-[1.2rem]">
                    <CardHeader>
                      <CardTitleWithHint
                        title={t("home.topNamespaces")}
                        hint={t("home.topNamespacesDescription")}
                      />
                      <CardAction>
                        <Button variant="ghost" size="xs" render={<Link to="/namespaces" />}>
                          {t("home.viewAll")}
                          <ChevronRight />
                        </Button>
                      </CardAction>
                    </CardHeader>
                    <CardContent className="flex min-h-0 flex-1 flex-col px-2 pb-2 [&_[data-slot=table-container]]:min-h-0 [&_[data-slot=table-container]]:flex-1 [&_[data-slot=table-container]]:overflow-y-auto">
                      <Table>
                        <TableHeader className="sticky top-0 z-10 [&_th]:bg-muted">
                          <TableRow>
                            <TableHead>{t("home.topNamespacesName")}</TableHead>
                            <TableHead className="text-right">{t("home.topNamespacesRepositories")}</TableHead>
                            <TableHead className="text-right">{t("home.topNamespacesTags")}</TableHead>
                            <TableHead className="w-[120px] text-right">{t("home.topNamespacesStorage")}</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {(overview?.namespaces ?? []).map(namespace => {
                            const percent = namespace.size_limit > 0
                              ? Math.min((namespace.size / namespace.size_limit) * 100, 100)
                              : 0;
                            const nearLimit = namespace.size_limit > 0 && percent >= quotaWarning * 100;
                            return (
                              <TableRow
                                key={namespace.id}
                                className="cursor-pointer"
                                onClick={() => navigate(`/namespaces/${namespace.name}/repositories`)}
                              >
                                <TableCell>
                                  <div className="flex items-center gap-2">
                                    <span className="font-medium text-foreground">{namespace.name}</span>
                                    <Badge variant="outline" className="capitalize text-muted-foreground">
                                      {namespace.visibility}
                                    </Badge>
                                  </div>
                                </TableCell>
                                <TableCell className="text-right tabular-nums text-muted-foreground">
                                  {numberFormatter.format(namespace.repositories)}
                                </TableCell>
                                <TableCell className="text-right tabular-nums text-muted-foreground">
                                  {numberFormatter.format(namespace.tags)}
                                </TableCell>
                                <TableCell className="text-right">
                                  {namespace.size_limit > 0 ? (
                                    <div className="flex items-center justify-end gap-2">
                                      <div className="h-1.5 min-w-0 flex-1 overflow-hidden rounded-full bg-muted">
                                        <div
                                          className={cn("h-full rounded-full", nearLimit ? "bg-destructive" : "bg-primary")}
                                          style={{ width: `${percent}%` }}
                                        />
                                      </div>
                                      <span
                                        className={cn(
                                          "shrink-0 text-xs tabular-nums",
                                          nearLimit ? "text-destructive" : "text-muted-foreground",
                                        )}
                                      >
                                        {percent.toFixed(0)}%
                                      </span>
                                    </div>
                                  ) : (
                                    <span className="text-xs tabular-nums text-muted-foreground">
                                      {formatBytes(namespace.size)}
                                    </span>
                                  )}
                                </TableCell>
                              </TableRow>
                            );
                          })}
                          {!loading && (overview?.namespaces ?? []).length === 0 && (
                            <TableRow>
                              <TableCell colSpan={4} className="py-8 text-center text-muted-foreground">
                                {t("home.emptyNamespaces")}
                              </TableCell>
                            </TableRow>
                          )}
                        </TableBody>
                      </Table>
                    </CardContent>
                  </Card>

                  <BuildTrend localServer={localServer} />
                </div>
              </div>
            </div>
          </main>
        </div>
      </div>
    </Fragment>
  );
}

type MetricCardProps = {
  icon: ComponentType<{ className?: string }>;
  label: string;
  value: ReactNode;
  hint?: ReactNode;
  footer?: ReactNode;
};

function MetricCard({ icon: Icon, label, value, hint, footer }: MetricCardProps) {
  return (
    <Card className="gap-0 py-0">
      <div className="p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="truncate text-xs font-medium text-muted-foreground">{label}</div>
            <div className="mt-1.5 text-2xl font-semibold tracking-tight tabular-nums text-foreground">
              {value}
            </div>
          </div>
          <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
            <Icon className="size-4" />
          </span>
        </div>
        {footer}
        {hint !== undefined && <div className="mt-3 text-xs text-muted-foreground">{hint}</div>}
      </div>
    </Card>
  );
}
