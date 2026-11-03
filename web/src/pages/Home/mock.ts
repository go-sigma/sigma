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

// The home dashboard renders sample data for now. Every value here mirrors the
// shape of a future `GET /api/v1/overview` response, so wiring the page to the
// backend means replacing this module and nothing else.

const gib = 1024 ** 3;

// mulberry32 is a tiny deterministic PRNG: the sample data stays stable across
// renders and reloads, which keeps the charts comparable while developing.
function mulberry32(seed: number) {
  let state = seed >>> 0;
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let value = state;
    value = Math.imul(value ^ (value >>> 15), value | 1);
    value ^= value + Math.imul(value ^ (value >>> 7), value | 61);
    return ((value ^ (value >>> 14)) >>> 0) / 4294967296;
  };
}

export type DailyActivity = {
  date: string;
  push: number;
  pull: number;
};

export type BuildDay = {
  date: string;
  success: number;
  failed: number;
  // duration is the average build duration of the day, in seconds.
  duration: number;
};

export type TopNamespace = {
  name: string;
  visibility: "public" | "private";
  repositories: number;
  tags: number;
  used: number;
  limit: number;
};

// mockDailyActivity returns one point per day, oldest first. Weekends are
// quieter and the whole series drifts up slowly so the trend chart has a shape.
export function mockDailyActivity(days: number): DailyActivity[] {
  const random = mulberry32(20261103);
  const today = dayjs().startOf("day");
  const points: DailyActivity[] = [];
  let weekFactor = 1;

  for (let offset = days - 1; offset >= 0; offset--) {
    const day = today.subtract(offset, "day");
    const weekend = day.day() === 0 || day.day() === 6 ? 0.45 : 1;
    const growth = 0.75 + ((days - offset) / days) * 0.5;
    if (offset % 7 === 0) {
      // Vary the volume week by week so the weekly summaries are not identical.
      weekFactor = 0.7 + random() * 0.7;
    }
    points.push({
      date: day.format("YYYY-MM-DD"),
      push: Math.max(0, Math.round((2 + random() * 8) * weekend * growth * weekFactor)),
      pull: Math.round((90 + random() * 420) * weekend * growth * weekFactor),
    });
  }

  return points;
}

export const mockTopNamespaces: TopNamespace[] = [
  { name: "library", visibility: "public", repositories: 8, tags: 96, used: 5.06 * gib, limit: 20 * gib },
  { name: "demo", visibility: "private", repositories: 5, tags: 31, used: 3.2 * gib, limit: 10 * gib },
  { name: "infra", visibility: "private", repositories: 6, tags: 14, used: 2.4 * gib, limit: 10 * gib },
  { name: "sandbox", visibility: "public", repositories: 3, tags: 5, used: 1.64 * gib, limit: 2 * gib },
  { name: "ml", visibility: "private", repositories: 1, tags: 4, used: 0.86 * gib, limit: 5 * gib },
  { name: "devteam", visibility: "private", repositories: 2, tags: 9, used: 0.42 * gib, limit: 5 * gib },
  { name: "edge", visibility: "public", repositories: 2, tags: 7, used: 0.31 * gib, limit: 2 * gib },
  { name: "qa", visibility: "private", repositories: 1, tags: 3, used: 0.12 * gib, limit: 2 * gib },
  { name: "tosone", visibility: "private", repositories: 1, tags: 2, used: 0.1 * gib, limit: 1 * gib },
  { name: "archive", visibility: "private", repositories: 3, tags: 6, used: 0.09 * gib, limit: 1 * gib },
];

// mockBuildActivity returns one point per day, oldest first: how many builds
// succeeded, how many failed, and the average build duration in seconds.
export function mockBuildActivity(days: number): BuildDay[] {
  const random = mulberry32(20261104);
  const today = dayjs().startOf("day");
  const points: BuildDay[] = [];
  let failureBurst = 0;

  for (let offset = days - 1; offset >= 0; offset--) {
    const day = today.subtract(offset, "day");
    const weekend = day.day() === 0 || day.day() === 6 ? 0.4 : 1;
    const success = Math.round((1 + random() * 7) * weekend);
    // Failures come in small bursts so the failure line has a shape instead of noise.
    if (failureBurst === 0 && random() < 0.16) {
      failureBurst = 1 + Math.floor(random() * 3);
    }
    const failed = failureBurst > 0 ? Math.min(failureBurst, 2) : 0;
    failureBurst = Math.max(0, failureBurst - 1);
    points.push({
      date: day.format("YYYY-MM-DD"),
      success,
      failed,
      duration: Math.round(110 + random() * 240 + success * 9),
    });
  }

  return points;
}

export const mockOverview = {
  version: "v0.9.0",
  commit: "78e4659",
  storage: {
    used: mockTopNamespaces.reduce((total, item) => total + item.used, 0),
    limit: 100 * gib,
  },
  namespaces: mockTopNamespaces.length,
  repositories: mockTopNamespaces.reduce((total, item) => total + item.repositories, 0),
  tags: mockTopNamespaces.reduce((total, item) => total + item.tags, 0),
  pushesDelta: 12.4,
};

// formatBytes renders a byte count with binary units, e.g. "5.1 GiB".
export function formatBytes(bytes: number) {
  const units = ["B", "KiB", "MiB", "GiB", "TiB", "PiB"];
  let value = bytes;
  let unit = 0;

  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }

  const digits = unit === 0 || value >= 100 ? 0 : 1;
  return `${value.toFixed(digits)} ${units[unit]}`;
}
