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

import dayjs from "dayjs";
import { useMemo, useRef, useState, type MouseEvent } from "react";

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

const cellLevels = ["bg-muted", "bg-primary/20", "bg-primary/40", "bg-primary/60", "bg-primary"];

// minMonthGap keeps short month names from colliding when two labels land next to each other.
const minMonthGap = 3;
// The tooltip is sized in its own box, so these keep it clear of the hovered cell and the edges.
const tooltipGap = 6;
const tooltipHalfWidth = 76;
// Below this offset the tooltip would be clipped by the card, so it flips under the cell.
const tooltipFlipThreshold = 56;

function levelOf(count: number) {
  if (count <= 0) {
    return 0;
  }
  if (count <= 3) {
    return 1;
  }
  if (count <= 6) {
    return 2;
  }
  if (count <= 9) {
    return 3;
  }
  return 4;
}

type HeatmapDay = {
  date: string;
  count?: number;
};

type HeatmapWeek = {
  start: string;
  days: HeatmapDay[];
};

type MonthLabel = {
  date: string;
  column: number;
  label: string;
};

type HoveredDay = {
  date: string;
  count: number;
  left: number;
  top: number;
  above: boolean;
};

type ActivityHeatmapProps = {
  items?: IDailyActivity[];
  days: number;
};

// ActivityHeatmap plots the daily pushes of the requested window as a contribution grid.
export default function ActivityHeatmap({ items, days }: ActivityHeatmapProps) {
  const { locale, t } = useTranslation();

  const { weeks, monthLabels, total, thisWeek, lastWeek, busiest } = useMemo(() => {
    const points = (items ?? []).map(item => ({ date: item.date, push: item.push_count }));
    if (points.length === 0) {
      return {
        weeks: [] as HeatmapWeek[],
        monthLabels: [] as MonthLabel[],
        total: 0,
        thisWeek: 0,
        lastWeek: 0,
        busiest: undefined as { date: string; push: number } | undefined,
      };
    }

    const byDate = new Map(points.map(point => [point.date, point]));
    const start = dayjs(points[0].date).startOf("week");
    const end = dayjs(points[points.length - 1].date);
    const monthFormatter = new Intl.DateTimeFormat(locale, { month: "short" });
    const sum = (window: typeof points) => window.reduce((accumulator, point) => accumulator + point.push, 0);

    const columns: HeatmapWeek[] = [];
    const labels: MonthLabel[] = [];
    let lastLabeledMonth = -1;

    for (let cursor = start; !cursor.isAfter(end, "day"); cursor = cursor.add(7, "day")) {
      const week: HeatmapDay[] = [];
      for (let dayOffset = 0; dayOffset < 7; dayOffset++) {
        const date = cursor.add(dayOffset, "day").format("YYYY-MM-DD");
        week.push({ date, count: byDate.get(date)?.push });
      }

      const firstInRange = week.find(day => day.count !== undefined);
      if (firstInRange !== undefined) {
        const month = dayjs(firstInRange.date).month();
        const lastLabel = labels[labels.length - 1];
        if (month !== lastLabeledMonth && (lastLabel === undefined || columns.length - lastLabel.column >= minMonthGap)) {
          labels.push({
            date: firstInRange.date,
            column: columns.length,
            label: monthFormatter.format(new Date(firstInRange.date)),
          });
          lastLabeledMonth = month;
        }
      }
      columns.push({ start: cursor.format("YYYY-MM-DD"), days: week });
    }

    return {
      weeks: columns,
      monthLabels: labels,
      total: sum(points),
      thisWeek: sum(points.slice(-7)),
      lastWeek: sum(points.slice(-14, -7)),
      busiest: points.reduce((best, point) => (point.push > best.push ? point : best), points[0]),
    };
  }, [items, locale]);

  const dateFormatter = useMemo(
    () => new Intl.DateTimeFormat(locale, { dateStyle: "medium" }),
    [locale],
  );

  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);
  const shortDateFormatter = useMemo(
    () => new Intl.DateTimeFormat(locale, { month: "2-digit", day: "2-digit" }),
    [locale],
  );

  const labelByColumn = useMemo(
    () => new Map(monthLabels.map(item => [item.column, item])),
    [monthLabels],
  );

  const gridRef = useRef<HTMLDivElement | null>(null);
  const [hoveredDay, setHoveredDay] = useState<HoveredDay>();

  // One delegated handler keeps the cells from each owning a listener: the cell under the
  // pointer is read from the dataset and the tooltip is placed relative to the grid.
  const trackHover = (event: MouseEvent<HTMLDivElement>) => {
    const cell = (event.target as HTMLElement).closest<HTMLElement>("[data-heatmap-date]");
    const grid = gridRef.current;
    if (cell === null || cell.dataset.heatmapDate === undefined || grid === null) {
      setHoveredDay(undefined);
      return;
    }

    const date = cell.dataset.heatmapDate;
    const cellRect = cell.getBoundingClientRect();
    const gridRect = grid.getBoundingClientRect();
    const left = Math.min(
      Math.max(cellRect.left + cellRect.width / 2 - gridRect.left, tooltipHalfWidth),
      Math.max(gridRect.width - tooltipHalfWidth, tooltipHalfWidth),
    );
    const top = cellRect.top - gridRect.top;

    setHoveredDay(current => {
      if (current?.date === date && current.count === Number(cell.dataset.heatmapCount)) {
        return current;
      }
      return {
        date,
        count: Number(cell.dataset.heatmapCount ?? "0"),
        left,
        top,
        above: top >= tooltipFlipThreshold,
      };
    });
  };

  return (
    <Card className="h-full gap-[1.2rem]">
      <CardHeader>
        <CardTitleWithHint
          title={t("home.activityHeatmap")}
          hint={t("home.activityHeatmapDescription", { days })}
        />
        <CardAction>
          <span className="text-xs text-muted-foreground">
            {t("home.activityHeatmapTotal", { count: total })}
          </span>
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col justify-between gap-4">
        {weeks.length === 0 ? (
          <div className="flex items-center justify-center py-10 text-xs text-muted-foreground">
            {t("home.emptyActivity")}
          </div>
        ) : (
          <div className="flex flex-col gap-1">
            <div className="grid grid-flow-col gap-[3px] [grid-auto-columns:minmax(0,1fr)]">
              {weeks.map((week, column) => {
                const label = labelByColumn.get(column);
                return (
                  <div key={week.start} className="relative h-3.5">
                    {label !== undefined && (
                      <span className="absolute left-0 top-0 whitespace-nowrap text-[10px] leading-none text-muted-foreground">
                        {label.label}
                      </span>
                    )}
                  </div>
                );
              })}
            </div>

            <div
              ref={gridRef}
              className="relative grid grid-flow-col grid-rows-7 gap-[3px] [grid-auto-columns:minmax(0,1fr)]"
              role="img"
              aria-label={t("home.activityHeatmap")}
              onMouseMove={trackHover}
              onMouseLeave={() => setHoveredDay(undefined)}
            >
              {weeks.flatMap(week => week.days.map(day => {
                if (day.count === undefined) {
                  return <div key={day.date} className="aspect-square w-full" />;
                }
                return (
                  <div
                    key={day.date}
                    data-heatmap-date={day.date}
                    data-heatmap-count={day.count}
                    className={cn(
                      "aspect-square w-full rounded-[2px] transition-shadow",
                      cellLevels[levelOf(day.count)],
                      hoveredDay?.date === day.date && "ring-1 ring-foreground/40 ring-offset-1 ring-offset-background",
                    )}
                  />
                );
              }))}

              {hoveredDay !== undefined && (
                <div
                  className={cn(
                    "pointer-events-none absolute z-10 -translate-x-1/2 whitespace-nowrap rounded-md bg-popover/90 px-2 py-1 text-xs shadow-md ring-1 ring-foreground/10 backdrop-blur-sm",
                    hoveredDay.above && "-translate-y-full",
                  )}
                  style={{
                    left: hoveredDay.left,
                    top: hoveredDay.above ? hoveredDay.top - tooltipGap : hoveredDay.top + tooltipGap + 12,
                  }}
                >
                  <div className="font-medium text-popover-foreground">
                    {dateFormatter.format(new Date(hoveredDay.date))}
                  </div>
                  <div className="text-muted-foreground">
                    {t("home.activityHeatmapCount", { count: hoveredDay.count })}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        <div className="mt-1 flex items-center justify-end gap-1.5 text-[10px] text-muted-foreground">
          <span>{t("home.activityHeatmapLess")}</span>
          {cellLevels.map(level => (
            <span key={level} className={cn("size-[10px] rounded-[2px]", level)} />
          ))}
          <span>{t("home.activityHeatmapMore")}</span>
        </div>

        <dl className="grid grid-cols-3 gap-3 border-t border-border pt-3">
          <div>
            <dt className="text-xs text-muted-foreground">{t("home.activityHeatmapThisWeek")}</dt>
            <dd className="mt-0.5 text-sm font-medium tabular-nums text-foreground">
              {numberFormatter.format(thisWeek)}
            </dd>
          </div>
          <div>
            <dt className="text-xs text-muted-foreground">{t("home.activityHeatmapLastWeek")}</dt>
            <dd className="mt-0.5 text-sm font-medium tabular-nums text-foreground">
              {numberFormatter.format(lastWeek)}
            </dd>
          </div>
          <div>
            <dt className="text-xs text-muted-foreground">{t("home.activityHeatmapBusiest")}</dt>
            <dd className="mt-0.5 flex items-baseline gap-1.5">
              <span className="text-sm font-medium tabular-nums text-foreground">
                {numberFormatter.format(busiest?.push ?? 0)}
              </span>
              {busiest !== undefined && (
                <span className="text-xs text-muted-foreground">
                  {shortDateFormatter.format(new Date(busiest.date))}
                </span>
              )}
            </dd>
          </div>
        </dl>
      </CardContent>
    </Card>
  );
}
