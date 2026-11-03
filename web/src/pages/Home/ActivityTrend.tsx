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

import { useId, useMemo, useState } from "react";

import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
} from "@/components/ui/card";
import { useTranslation } from "@/i18n/useTranslation";
import { IDailyActivity } from "@/interfaces";
import { cn } from "@/lib/utils";

import CardTitleWithHint from "./CardTitleWithHint";
import { useElementWidth } from "./chart";

const chartHeight = 200;
// The right padding holds the push axis labels: both series keep their own scale
// because pulls outnumber pushes by an order of magnitude.
const chartPadding = { top: 16, right: 40, bottom: 26, left: 44 };
const gridSteps = 4;
// The tooltip is sized in its own box, so these keep it clear of the hovered point and the edges.
const tooltipGap = 14;
const tooltipHalfWidth = 80;

const ranges = [7, 30] as const;

type Range = (typeof ranges)[number];

type ActivityTrendProps = {
  items?: IDailyActivity[];
};

// ActivityTrend plots daily pulls and pushes for the last week or month.
export default function ActivityTrend({ items }: ActivityTrendProps) {
  const { locale, t } = useTranslation();
  const [containerRef, width] = useElementWidth();
  const areaId = useId();
  const [range, setRange] = useState<Range>(30);
  const [hovered, setHovered] = useState<number | null>(null);

  const points = useMemo(
    () => (items ?? []).slice(-range).map(item => ({
      date: item.date,
      push: item.push_count,
      pull: item.pull_count,
    })),
    [items, range],
  );

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
    // Counts are integers, so the axes are rounded to a multiple of the grid steps; niceCeil
    // would produce fractions like 0.25 pull when the window is empty.
    const pullMax = Math.max(gridSteps, Math.ceil(Math.max(...points.map(point => point.pull)) / gridSteps) * gridSteps);
    const pushMax = Math.max(gridSteps, Math.ceil(Math.max(...points.map(point => point.push)) / gridSteps) * gridSteps);
    const stepX = points.length > 1 ? innerWidth / (points.length - 1) : 0;
    const baseline = chartPadding.top + innerHeight;

    const positions = points.map((point, index) => ({
      ...point,
      x: chartPadding.left + (points.length > 1 ? stepX * index : innerWidth / 2),
      pullY: chartPadding.top + innerHeight - (point.pull / pullMax) * innerHeight,
      pushY: chartPadding.top + innerHeight - (point.push / pushMax) * innerHeight,
    }));

    const line = (key: "pullY" | "pushY") =>
      positions.map((point, index) => `${index === 0 ? "M" : "L"}${point.x.toFixed(1)},${point[key].toFixed(1)}`).join(" ");

    const pullLine = line("pullY");
    const area = `${pullLine} L${positions[positions.length - 1].x.toFixed(1)},${baseline} L${positions[0].x.toFixed(1)},${baseline} Z`;

    const gridLines = Array.from({ length: gridSteps + 1 }, (_, index) => ({
      pull: (pullMax / gridSteps) * index,
      push: (pushMax / gridSteps) * index,
      y: baseline - (innerHeight / gridSteps) * index,
    }));

    // Keep the axis readable: at most seven labels, always ending on the latest day.
    const tickEvery = Math.max(1, Math.ceil(positions.length / 6));
    const ticks = positions.filter(
      (_, index) => index % tickEvery === 0 || index === positions.length - 1,
    );

    return { positions, pullLine, pushLine: line("pushY"), area, gridLines, ticks, innerHeight };
  }, [points, width]);

  const hoveredPoint = geometry !== undefined && hovered !== null ? geometry.positions[hovered] : undefined;
  const tooltipPlacement = hoveredPoint === undefined ? undefined : {
    left: Math.min(Math.max(hoveredPoint.x, tooltipHalfWidth), Math.max(width - tooltipHalfWidth, tooltipHalfWidth)),
    top: hoveredPoint.pullY - tooltipGap,
    above: hoveredPoint.pullY - tooltipGap - 72 >= 0,
  };

  const totals = useMemo(() => points.reduce(
    (sum, point) => ({ pull: sum.pull + point.pull, push: sum.push + point.push }),
    { pull: 0, push: 0 },
  ), [points]);

  return (
    <Card className="h-full gap-[1.2rem]">
      <CardHeader>
        <CardTitleWithHint title={t("home.activityTrend")} hint={t("home.activityTrendDescription")} />
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
      <CardContent className="flex flex-1 flex-col gap-2">
        <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
          <span className="inline-flex items-center gap-1.5">
            <span className="h-0.5 w-4 rounded-full bg-primary" />
            {t("home.activityTrendPull")}
            <span className="font-medium tabular-nums text-foreground">{numberFormatter.format(totals.pull)}</span>
          </span>
          <span className="inline-flex items-center gap-1.5">
            <span className="w-4 border-t border-dashed border-muted-foreground" />
            {t("home.activityTrendPush")}
            <span className="font-medium tabular-nums text-foreground">{numberFormatter.format(totals.push)}</span>
          </span>
        </div>

        <div ref={containerRef} className="relative flex-1">
          {geometry === undefined ? (
            <div className="flex items-center justify-center text-xs text-muted-foreground" style={{ height: chartHeight }}>
              {items === undefined ? null : t("home.emptyActivity")}
            </div>
          ) : (
            <>
              <svg
                width={width}
                height={chartHeight}
                role="img"
                aria-label={t("home.activityTrend")}
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
                    <stop offset="0%" stopColor="currentColor" stopOpacity={0.18} />
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
                      {numberFormatter.format(line.pull)}
                    </text>
                    <text
                      x={width - chartPadding.right + 8}
                      y={line.y + 3}
                      textAnchor="start"
                      className="fill-muted-foreground text-[10px] tabular-nums"
                    >
                      {numberFormatter.format(line.push)}
                    </text>
                  </g>
                ))}

                <path d={geometry.area} fill={`url(#${areaId})`} stroke="none" />
                <path
                  d={geometry.pullLine}
                  className="stroke-primary"
                  fill="none"
                  strokeWidth={1.5}
                  strokeLinejoin="round"
                  strokeLinecap="round"
                />
                <path
                  d={geometry.pushLine}
                  className="stroke-muted-foreground/70"
                  fill="none"
                  strokeWidth={1.25}
                  strokeDasharray="4 3"
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
                    <circle cx={hoveredPoint.x} cy={hoveredPoint.pullY} r={3} className="fill-primary" />
                    <circle cx={hoveredPoint.x} cy={hoveredPoint.pushY} r={3} className="fill-muted-foreground" />
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
                    "pointer-events-none absolute z-10 -translate-x-1/2 rounded-md bg-popover/80 px-2 py-1 text-xs shadow-md ring-1 ring-foreground/10 backdrop-blur-sm",
                    tooltipPlacement.above && "-translate-y-full",
                  )}
                  style={{
                    left: tooltipPlacement.left,
                    top: tooltipPlacement.above ? tooltipPlacement.top : hoveredPoint.pullY + tooltipGap,
                  }}
                >
                  <div className="font-medium text-popover-foreground">
                    {dateFormatter.format(new Date(hoveredPoint.date))}
                  </div>
                  <div className="mt-0.5 flex items-center gap-1.5 text-muted-foreground">
                    <span className="size-1.5 rounded-full bg-primary" />
                    {t("home.activityTrendPull")}
                    <span className="ml-auto font-medium tabular-nums text-foreground">
                      {numberFormatter.format(hoveredPoint.pull)}
                    </span>
                  </div>
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <span className="size-1.5 rounded-full bg-muted-foreground" />
                    {t("home.activityTrendPush")}
                    <span className="ml-auto font-medium tabular-nums text-foreground">
                      {numberFormatter.format(hoveredPoint.push)}
                    </span>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
