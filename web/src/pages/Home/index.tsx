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

import { Activity, ArrowUpRight, Boxes, ChevronRight, HardDrive, Tags } from "lucide-react";
import { Fragment, useMemo, type ComponentType, type ReactNode } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { Link, useNavigate } from "react-router-dom";

import Header from "@/components/Header";
import Menu from "@/components/Menu";
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
import { cn } from "@/lib/utils";

import ActivityHeatmap from "./ActivityHeatmap";
import ActivityTrend from "./ActivityTrend";
import QuickStart from "./QuickStart";
import BuildTrend from "./BuildTrend";
import CardTitleWithHint from "./CardTitleWithHint";
import { formatBytes, mockDailyActivity, mockOverview, mockTopNamespaces } from "./mock";

// The shares of a namespace quota that deserve a warning color.
const quotaWarning = 0.8;

export default function Home({ localServer }: { localServer: string }) {
  const { locale, t } = useTranslation();
  const navigate = useNavigate();

  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);
  const storagePercent = (mockOverview.storage.used / mockOverview.storage.limit) * 100;
  const pushes30d = useMemo(
    () => mockDailyActivity(30).reduce((total, point) => total + point.push, 0),
    [],
  );

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
                    value={formatBytes(mockOverview.storage.used)}
                    footer={
                      <div className="mt-3 space-y-1.5">
                        <div className="h-1.5 w-full overflow-hidden rounded-full bg-muted">
                          <div
                            className={cn(
                              "h-full rounded-full transition-all",
                              storagePercent >= quotaWarning * 100 ? "bg-destructive" : "bg-primary",
                            )}
                            style={{ width: `${Math.min(storagePercent, 100)}%` }}
                          />
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {t("home.metricStorageHint", {
                            limit: formatBytes(mockOverview.storage.limit),
                            percent: storagePercent.toFixed(1),
                          })}
                        </div>
                      </div>
                    }
                  />
                  <MetricCard
                    icon={Boxes}
                    label={t("home.metricNamespaces")}
                    value={numberFormatter.format(mockOverview.namespaces)}
                    hint={t("home.metricNamespacesHint", {
                      repositories: numberFormatter.format(mockOverview.repositories),
                    })}
                  />
                  <MetricCard
                    icon={Tags}
                    label={t("home.metricTags")}
                    value={numberFormatter.format(mockOverview.tags)}
                    hint={t("home.metricTagsHint", {
                      repositories: numberFormatter.format(mockOverview.repositories),
                    })}
                  />
                  <MetricCard
                    icon={Activity}
                    label={t("home.metricPushes")}
                    value={numberFormatter.format(pushes30d)}
                    hint={
                      <span className="inline-flex items-center gap-1">
                        <ArrowUpRight className="size-3.5 text-green-600 dark:text-green-500" />
                        {t("home.metricPushesDelta", { delta: `+${mockOverview.pushesDelta}%` })}
                      </span>
                    }
                  />
                </div>

                <QuickStart localServer={localServer} />

                <div className="grid gap-4 lg:grid-cols-2">
                  <ActivityHeatmap />
                  <ActivityTrend />
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
                            <TableHead className="w-[240px]">{t("home.topNamespacesStorage")}</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {mockTopNamespaces.map(namespace => {
                            const percent = Math.min((namespace.used / namespace.limit) * 100, 100);
                            const nearLimit = percent >= quotaWarning * 100;
                            return (
                              <TableRow
                                key={namespace.name}
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
                                <TableCell>
                                  <div className="flex items-center gap-2">
                                    <div className="h-1.5 w-full overflow-hidden rounded-full bg-muted">
                                      <div
                                        className={cn(
                                          "h-full rounded-full",
                                          nearLimit ? "bg-destructive" : "bg-primary",
                                        )}
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
                                </TableCell>
                              </TableRow>
                            );
                          })}
                        </TableBody>
                      </Table>
                    </CardContent>
                  </Card>

                  <BuildTrend />
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
