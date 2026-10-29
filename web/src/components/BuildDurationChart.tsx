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
import dayjs from "dayjs";
import { Info } from "lucide-react";
import { useEffect, useId, useMemo, useRef, useState } from "react";

import Notification from "@/components/Notification";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useTranslation } from "@/i18n/useTranslation";
import { formatDuration, runnerDurationMs } from "@/lib/duration";
import { IBuilderRunnerList, IHTTPError } from "@/interfaces";

// sampleLimit is the number of builds the chart plots: the most recent ones.
const sampleLimit = 100;

const chartHeight = 220;
// The vertical padding keeps the topmost gridline label and the zero line clear of the card
// edges; the left padding holds the duration labels.
const chartPadding = { top: 20, right: 16, bottom: 28, left: 56 };
const gridSteps = 4;
// The tooltip is sized in its own box, so these keep it clear of the hovered point and the edges.
const tooltipGap = 16;
const tooltipHeight = 54;
const tooltipHalfWidth = 72;

type BuildSample = {
  id: string;
  label: string;
  duration: number;
  createdAt: string;
};

// niceCeil rounds a maximum up to a value that divides evenly into the grid steps.
function niceCeil(value: number) {
  if (value <= 0) {
    return 1;
  }
  const magnitude = 10 ** Math.floor(Math.log10(value));
  const normalized = value / magnitude;
  const step = normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 5 ? 5 : 10;
  const rounded = step * magnitude;
  return rounded < value ? value : rounded;
}

// useElementWidth tracks the width of an element so the svg can be drawn in real pixels: scaling
// a fixed viewBox would stretch the stroke widths and the axis labels.
function useElementWidth() {
  const ref = useRef<HTMLDivElement | null>(null);
  const [width, setWidth] = useState(0);

  useEffect(() => {
    const element = ref.current;
    if (element === null) {
      return;
    }
    const observer = new ResizeObserver(entries => {
      for (const entry of entries) {
        setWidth(entry.contentRect.width);
      }
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  return [ref, width] as const;
}

type BuildDurationChartProps = {
  localServer: string;
  namespaceId: string;
  repositoryId: string;
  builderId: string;
  // refreshKey changes whenever the parent reloads the run list, which re-reads the samples so the
  // chart keeps up with a running build.
  refreshKey?: unknown;
};

// BuildDurationChart plots the duration of the most recent builds of a repository builder.
export default function BuildDurationChart({ localServer, namespaceId, repositoryId, builderId, refreshKey }: BuildDurationChartProps) {
  const { t } = useTranslation();
  const [containerRef, width] = useElementWidth();
  const fillId = useId();
  const [hovered, setHovered] = useState<number | null>(null);
  const [samples, setSamples] = useState<BuildSample[]>();

  useEffect(() => {
    if (builderId === "") {
      return;
    }
    const url = `${localServer}/api/v1/namespaces/${namespaceId}/repositories/${repositoryId}/builders/${builderId}/runners/?limit=${sampleLimit}&sort=created_at&method=desc`;
    axios.get(url).then(response => {
      if (response?.status !== 200) {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
        return;
      }
      // The API answers newest first; the chart reads left to right, oldest to newest.
      const items = (response.data as IBuilderRunnerList).items ?? [];
      const points: BuildSample[] = [];
      for (const item of items.toReversed()) {
        const duration = runnerDurationMs(item);
        if (duration === undefined) {
          continue;
        }
        points.push({
          id: item.id,
          label: item.tag ?? item.raw_tag,
          duration,
          createdAt: item.created_at,
        });
      }
      setSamples(points);
    }).catch(error => {
      const errorcode = (error.response?.data ?? {}) as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
    // refreshKey is not read by the effect body: it is bumped by the parent so the samples are
    // re-read whenever the run list polls.
    // oxlint-disable-next-line react/exhaustive-effect-dependencies
  }, [localServer, namespaceId, repositoryId, builderId, refreshKey]);

  const geometry = useMemo(() => {
    if (samples === undefined || samples.length === 0 || width === 0) {
      return undefined;
    }
    const innerWidth = Math.max(width - chartPadding.left - chartPadding.right, 1);
    const innerHeight = chartHeight - chartPadding.top - chartPadding.bottom;
    const maxDuration = niceCeil(Math.max(...samples.map(sample => sample.duration)));
    const stepX = samples.length > 1 ? innerWidth / (samples.length - 1) : 0;

    const points = samples.map((sample, index) => ({
      ...sample,
      x: chartPadding.left + (samples.length > 1 ? stepX * index : innerWidth / 2),
      y: chartPadding.top + innerHeight - (sample.duration / maxDuration) * innerHeight,
    }));

    const line = points.map((point, index) => `${index === 0 ? "M" : "L"}${point.x.toFixed(1)},${point.y.toFixed(1)}`).join(" ");
    const area = `${line} L${points[points.length - 1].x.toFixed(1)},${chartPadding.top + innerHeight} L${points[0].x.toFixed(1)},${chartPadding.top + innerHeight} Z`;
    const gridLines = Array.from({ length: gridSteps + 1 }, (_, index) => ({
      value: (maxDuration / gridSteps) * index,
      y: chartPadding.top + innerHeight - (innerHeight / gridSteps) * index,
    }));

    return { points, line, area, gridLines, innerHeight };
  }, [samples, width]);

  const hoveredPoint = geometry !== undefined && hovered !== null ? geometry.points[hovered] : undefined;

  // The tooltip floats above the hovered point and flips below it near the top of the chart, so it
  // does not sit on top of the line it describes.
  const tooltipPlacement = hoveredPoint === undefined ? undefined : {
    left: Math.min(Math.max(hoveredPoint.x, tooltipHalfWidth), Math.max(width - tooltipHalfWidth, tooltipHalfWidth)),
    top: hoveredPoint.y - tooltipGap,
    above: hoveredPoint.y - tooltipGap - tooltipHeight >= 0,
  };

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-1.5">
          <h2 className="text-sm font-medium text-foreground">{t("builder.buildDuration")}</h2>
          <Tooltip>
            <TooltipTrigger>
              <Info className="size-4 text-muted-foreground" aria-hidden="true" />
            </TooltipTrigger>
            <TooltipContent>{t("builder.buildDurationDescription", { count: sampleLimit })}</TooltipContent>
          </Tooltip>
        </div>
        {samples !== undefined && samples.length > 0 && (
          <p className="text-xs text-muted-foreground">
            {t("builder.buildDurationLatest", { duration: formatDuration(samples[samples.length - 1].duration) })}
          </p>
        )}
      </div>
      <div ref={containerRef} className="relative mt-3">
        {geometry === undefined ? (
          <div className="flex items-center justify-center text-xs text-muted-foreground" style={{ height: chartHeight }}>
            {samples === undefined ? null : t("builder.buildDurationEmpty")}
          </div>
        ) : (
          <>
            <svg
              width={width}
              height={chartHeight}
              role="img"
              aria-label={t("builder.buildDuration")}
              className="block text-primary"
              onMouseLeave={() => setHovered(null)}
              onMouseMove={event => {
                const bounds = event.currentTarget.getBoundingClientRect();
                const offsetX = event.clientX - bounds.left;
                let nearest = 0;
                for (let index = 1; index < geometry.points.length; index++) {
                  if (Math.abs(geometry.points[index].x - offsetX) < Math.abs(geometry.points[nearest].x - offsetX)) {
                    nearest = index;
                  }
                }
                setHovered(nearest);
              }}
            >
              <defs>
                <linearGradient id={fillId} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="currentColor" stopOpacity={0.25} />
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
                    className="fill-muted-foreground text-[10px]"
                  >
                    {formatDuration(line.value)}
                  </text>
                </g>
              ))}
              <path d={geometry.area} fill={`url(#${fillId})`} stroke="none" />
              <path d={geometry.line} className="stroke-primary" fill="none" strokeWidth={1.5} strokeLinejoin="round" strokeLinecap="round" />
              {geometry.points.map((point, index) => (
                <circle
                  key={point.id}
                  cx={point.x}
                  cy={point.y}
                  r={hovered === index ? 3.5 : 2}
                  className="fill-black dark:fill-white"
                />
              ))}
              {hoveredPoint !== undefined && (
                <line
                  x1={hoveredPoint.x}
                  x2={hoveredPoint.x}
                  y1={chartPadding.top}
                  y2={chartPadding.top + geometry.innerHeight}
                  className="stroke-muted-foreground/40"
                  strokeWidth={1}
                  strokeDasharray="3 3"
                />
              )}
            </svg>
            {hoveredPoint !== undefined && tooltipPlacement !== undefined && (
              <div
                className={`pointer-events-none absolute z-10 -translate-x-1/2 rounded-md bg-popover/80 px-2 py-1 text-xs shadow-md ring-1 ring-foreground/10 backdrop-blur-sm ${tooltipPlacement.above ? "-translate-y-full" : ""}`}
                style={{
                  left: tooltipPlacement.left,
                  top: tooltipPlacement.above ? tooltipPlacement.top : hoveredPoint.y + tooltipGap,
                }}
              >
                <div className="font-medium text-popover-foreground">{hoveredPoint.label}</div>
                <div className="text-muted-foreground">{formatDuration(hoveredPoint.duration)}</div>
                <div className="text-muted-foreground">
                  {dayjs.utc(hoveredPoint.createdAt).tz(dayjs.tz.guess()).format("MM-DD HH:mm:ss")}
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
