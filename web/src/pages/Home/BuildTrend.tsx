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
import { useEffect, useId, useMemo, useState } from "react";

import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
} from "@/components/ui/card";
import { useTranslation } from "@/i18n/useTranslation";
import { IDailyBuildList } from "@/interfaces";
import { cn } from "@/lib/utils";

import CardTitleWithHint from "./CardTitleWithHint";
import { useElementWidth } from "./chart";

const chartHeight = 220;
// The right padding only keeps the last point clear of the card edge.
const chartPadding = { top: 16, right: 12, bottom: 26, left: 36 };
const gridSteps = 4;
// The tooltip is sized in its own box, so these keep it clear of the hovered point and the edges.
const tooltipGap = 14;
const tooltipHalfWidth = 84;

const ranges = [7, 30] as const;

type Range = (typeof ranges)[number];

// formatDuration renders a number of seconds the way a build log reads it.
function formatDuration(seconds: number) {
  const total = Math.max(0, Math.round(seconds));
  const minutes = Math.floor(total / 60);
  const rest = total % 60;
  if (minutes === 0) {
    return `${rest}s`;
  }
  if (rest === 0) {
    return `${minutes}m`;
  }
  return `${minutes}m ${rest}s`;
}

type BuildTrendProps = {
  localServer: string;
};

// BuildTrend plots daily build successes and failures of the current user or, for
// administrators, of the whole instance.
export default function BuildTrend({ localServer }: BuildTrendProps) {
  const { locale, t } = useTranslation();
  const [containerRef, width] = useElementWidth();
  const areaId = useId();
  const [range, setRange] = useState<Range>(30);
  const [hovered, setHovered] = useState<number | null>(null);
  const [data, setData] = useState<IDailyBuildList>();

  useEffect(() => {
    let cancelled = false;
    axios.get(`${localServer}/api/v1/builds/daily?days=${range}`).then(response => {
      if (!cancelled && response.status === 200) {
        setData(response.data as IDailyBuildList);
      }
    }).catch(() => {
      // The card stays empty; the overview request already reported the failure.
    });
    return () => {
      cancelled = true;
    };
  }, [localServer, range]);

  const points = useMemo(() => data?.items ?? [], [data]);
  const summary = data?.summary;

  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);
  const dateFormatter = useMemo(
    () => new Intl.DateTimeFormat(locale, { dateStyle: "medium" }),
    [locale],
  );
  const tickFormatter = useMemo(
    () => new Intl.DateTimeFormat(locale, { month: "2-digit", day: "2-digit" }),
    [locale],
  );

  const geometry = useMemo(() => {
    if (width === 0 || points.length === 0) {
      return undefined;
    }
    const innerWidth = Math.max(width - chartPadding.left - chartPadding.right, 1);
    const innerHeight = chartHeight - chartPadding.top - chartPadding.bottom;
    // Build counts are integers, so the count axis is rounded to a multiple of the
    // grid steps instead of using niceCeil, which would produce half builds like 7.5.
    const rawCountMax = Math.max(...points.map(point => Math.max(point.success, point.failed)));
    const countMax = Math.max(gridSteps, Math.ceil(rawCountMax / gridSteps) * gridSteps);
    const stepX = points.length > 1 ? innerWidth / (points.length - 1) : 0;
    const baseline = chartPadding.top + innerHeight;
    const y = (value: number) => chartPadding.top + innerHeight - (value / countMax) * innerHeight;

    const positions = points.map((point, index) => ({
      ...point,
      x: chartPadding.left + (points.length > 1 ? stepX * index : innerWidth / 2),
      successY: y(point.success),
      failedY: y(point.failed),
    }));

    const line = (key: "successY" | "failedY") =>
      positions.map((point, index) => `${index === 0 ? "M" : "L"}${point.x.toFixed(1)},${point[key].toFixed(1)}`).join(" ");

    const successLine = line("successY");
    const area = `${successLine} L${positions[positions.length - 1].x.toFixed(1)},${baseline} L${positions[0].x.toFixed(1)},${baseline} Z`;

    const gridLines = Array.from({ length: gridSteps + 1 }, (_, index) => ({
      count: (countMax / gridSteps) * index,
      y: baseline - (innerHeight / gridSteps) * index,
    }));

    // Keep the axis readable: at most seven labels, always ending on the latest day.
    const tickEvery = Math.max(1, Math.ceil(positions.length / 6));
    const ticks = positions.filter(
      (_, index) => index % tickEvery === 0 || index === positions.length - 1,
    );

    return { positions, successLine, failedLine: line("failedY"), area, gridLines, ticks, innerHeight };
  }, [points, width]);

  const hoveredPoint = geometry !== undefined && hovered !== null ? geometry.positions[hovered] : undefined;
  const tooltipPlacement = hoveredPoint === undefined ? undefined : {
    left: Math.min(Math.max(hoveredPoint.x, tooltipHalfWidth), Math.max(width - tooltipHalfWidth, tooltipHalfWidth)),
    top: hoveredPoint.successY - tooltipGap,
    above: hoveredPoint.successY - tooltipGap - 86 >= 0,
  };

  const total = (summary?.success ?? 0) + (summary?.failed ?? 0);

  return (
    <Card className="h-full gap-[1.2rem]">
      <CardHeader>
        <CardTitleWithHint
          title={t("home.buildTrend")}
          hint={t("home.buildTrendDescription", { days: range })}
        />
        <CardAction>
          <div className="flex items-center gap-0.5 rounded-lg bg-muted p-0.5">
            {ranges.map(item => (
              <button
                key={item}
                type="button"
                onClick={() => {
                  setHovered(null);
                  setRange(item);
                }}
                className={cn(
                  "rounded-md px-2 py-0.5 text-xs font-medium transition-colors",
                  range === item
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground",
                )}
              >
                {t(item === 7 ? "home.range7" : "home.range30")}
              </button>
            ))}
          </div>
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
            <span className="inline-flex items-center gap-1.5">
              <span className="h-0.5 w-4 rounded-full bg-primary" />
              {t("home.buildTrendSuccess")}
              <span className="font-medium tabular-nums text-foreground">{numberFormatter.format(summary?.success ?? 0)}</span>
            </span>
            <span className="inline-flex items-center gap-1.5">
              <span className="h-0.5 w-4 rounded-full bg-destructive" />
              {t("home.buildTrendFailed")}
              <span className="font-medium tabular-nums text-foreground">{numberFormatter.format(summary?.failed ?? 0)}</span>
            </span>
          </div>

          <div ref={containerRef} className="relative mt-2">
            {geometry === undefined ? (
              <div className="flex items-center justify-center text-xs text-muted-foreground" style={{ height: chartHeight }}>
                {data === undefined ? null : t("home.emptyBuilds")}
              </div>
            ) : (
              <>
                <svg
                  width={width}
                  height={chartHeight}
                  role="img"
                  aria-label={t("home.buildTrend")}
                  className="block text-primary"
                  onMouseLeave={() => setHovered(null)}
                  onMouseMove={event => {
                    const bounds = event.currentTarget.getBoundingClientRect();
                    const offsetX = event.clientX - bounds.left;
                    let nearest = 0;
                    for (let index = 1; index < geometry.positions.length; index++) {
                      if (Math.abs(geometry.positions[index].x - offsetX) < Math.abs(geometry.positions[nearest].x - offsetX)) {
                        nearest = index;
                      }
                    }
                    setHovered(nearest);
                  }}
                >
                  <defs>
                    <linearGradient id={areaId} x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="currentColor" stopOpacity={0.16} />
                      <stop offset="100%" stopColor="currentColor" stopOpacity={0} />
                    </linearGradient>
                  </defs>

                  {geometry.gridLines.map(line => (
                    <g key={line.y}>
                      <line
                        x1={chartPadding.left}
                        x2={width - chartPadding.right}
                        y1={line.y}
                        y2={line.y}
                        className="stroke-border"
                        strokeWidth={1}
                      />
                      <text
                        x={chartPadding.left - 8}
                        y={line.y + 3}
                        textAnchor="end"
                        className="fill-primary text-[10px] tabular-nums"
                      >
                        {numberFormatter.format(line.count)}
                      </text>
                    </g>
                  ))}

                  <path d={geometry.area} fill={`url(#${areaId})`} stroke="none" />
                  <path
                    d={geometry.failedLine}
                    className="stroke-destructive"
                    fill="none"
                    strokeWidth={1.5}
                    strokeLinejoin="round"
                    strokeLinecap="round"
                  />
                  <path
                    d={geometry.successLine}
                    className="stroke-primary"
                    fill="none"
                    strokeWidth={1.5}
                    strokeLinejoin="round"
                    strokeLinecap="round"
                  />

                  {hoveredPoint !== undefined && (
                    <>
                      <line
                        x1={hoveredPoint.x}
                        x2={hoveredPoint.x}
                        y1={chartPadding.top}
                        y2={chartPadding.top + geometry.innerHeight}
                        className="stroke-muted-foreground/40"
                        strokeWidth={1}
                        strokeDasharray="3 3"
                      />
                      <circle cx={hoveredPoint.x} cy={hoveredPoint.successY} r={3} className="fill-primary" />
                      <circle cx={hoveredPoint.x} cy={hoveredPoint.failedY} r={3} className="fill-destructive" />
                    </>
                  )}

                  {geometry.ticks.map(point => (
                    <text
                      key={point.date}
                      x={point.x}
                      y={chartHeight - 6}
                      textAnchor="middle"
                      className="fill-muted-foreground text-[10px] tabular-nums"
                    >
                      {tickFormatter.format(new Date(point.date))}
                    </text>
                  ))}
                </svg>

                {hoveredPoint !== undefined && tooltipPlacement !== undefined && (
                  <div
                    className={cn(
                      "pointer-events-none absolute z-10 -translate-x-1/2 whitespace-nowrap rounded-md bg-popover/90 px-2 py-1 text-xs shadow-md ring-1 ring-foreground/10 backdrop-blur-sm",
                      tooltipPlacement.above && "-translate-y-full",
                    )}
                    style={{
                      left: tooltipPlacement.left,
                      top: tooltipPlacement.above ? tooltipPlacement.top : hoveredPoint.successY + tooltipGap,
                    }}
                  >
                    <div className="font-medium text-popover-foreground">
                      {dateFormatter.format(new Date(hoveredPoint.date))}
                    </div>
                    <div className="mt-0.5 flex items-center gap-1.5 text-muted-foreground">
                      <span className="size-1.5 rounded-full bg-primary" />
                      {t("home.buildTrendSuccess")}
                      <span className="ml-auto font-medium tabular-nums text-foreground">
                        {numberFormatter.format(hoveredPoint.success)}
                      </span>
                    </div>
                    <div className="flex items-center gap-1.5 text-muted-foreground">
                      <span className="size-1.5 rounded-full bg-destructive" />
                      {t("home.buildTrendFailed")}
                      <span className="ml-auto font-medium tabular-nums text-foreground">
                        {numberFormatter.format(hoveredPoint.failed)}
                      </span>
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
        </div>

        <dl className="grid grid-cols-3 gap-3 border-t border-border pt-3">
          <div>
            <dt className="text-xs text-muted-foreground">{t("home.buildTrendRate")}</dt>
            <dd className="mt-0.5 text-sm font-medium tabular-nums text-foreground">
              {total === 0 ? "—" : `${(summary?.success_rate ?? 0).toFixed(1)}%`}
            </dd>
          </div>
          <div>
            <dt className="text-xs text-muted-foreground">{t("home.buildTrendFastest")}</dt>
            <dd className="mt-0.5 text-sm font-medium tabular-nums text-foreground">
              {total === 0 ? "—" : formatDuration(summary?.fastest ?? 0)}
            </dd>
          </div>
          <div>
            <dt className="text-xs text-muted-foreground">{t("home.buildTrendSlowest")}</dt>
            <dd className="mt-0.5 text-sm font-medium tabular-nums text-foreground">
              {total === 0 ? "—" : formatDuration(summary?.slowest ?? 0)}
            </dd>
          </div>
        </dl>
      </CardContent>
    </Card>
  );
}
