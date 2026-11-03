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

import { useEffect, useRef, useState } from "react";

// niceCeil rounds a maximum up to a value that divides evenly into the grid steps.
export function niceCeil(value: number) {
  if (value <= 0) {
    return 1;
  }
  const magnitude = 10 ** Math.floor(Math.log10(value));
  const normalized = value / magnitude;
  const step = normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 5 ? 5 : 10;
  const rounded = step * magnitude;
  return rounded < value ? value : rounded;
}

// useElementWidth tracks the width of an element so a chart can be drawn in real pixels: scaling
// a fixed viewBox would stretch the stroke widths and the axis labels.
export function useElementWidth() {
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
