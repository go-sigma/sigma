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

// formatDuration renders a duration in milliseconds the way the build tables and charts show it.
export function formatDuration(ms: number) {
  if (ms <= 0) {
    return "0s";
  }
  if (ms < 1000) {
    return `${Math.round(ms)}ms`;
  }
  const seconds = ms / 1000;
  if (seconds < 60) {
    return `${seconds.toFixed(seconds < 10 ? 1 : 0)}s`;
  }
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${Math.round(seconds % 60)}s`;
}

// runnerDurationMs returns the duration of a build run in milliseconds. The API records a
// duration for some runs only, so fall back to the span between the start and end timestamps.
export function runnerDurationMs(runner: {
  raw_duration?: number;
  started_at?: number;
  ended_at?: number;
}) {
  if (runner.raw_duration != null && runner.raw_duration > 0) {
    return runner.raw_duration;
  }
  if (runner.started_at == null || runner.ended_at == null) {
    return undefined;
  }
  const elapsed = runner.ended_at - runner.started_at;
  return elapsed > 0 ? elapsed : undefined;
}
