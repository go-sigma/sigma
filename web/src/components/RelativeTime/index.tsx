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
import "dayjs/locale/zh-cn";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { type Locale, useUiStore } from "../../stores";

const dayjsLocales: Record<Locale, string> = {
  "en-US": "en",
  "zh-CN": "zh-cn",
};

// RelativeTime renders a human readable time distance (e.g. "3 hours ago")
// and reveals the exact timestamp in a tooltip on hover.
export default function RelativeTime({ time, fallback = "-" }: { time?: string | Date | null; fallback?: string }) {
  const locale = useUiStore((state) => state.locale);

  if (!time) {
    return <span>{fallback}</span>;
  }

  const local = dayjs.utc(time).tz(dayjs.tz.guess());
  const exact = local.format("YYYY-MM-DD HH:mm:ss");
  const relative = local.locale(dayjsLocales[locale] ?? "en").fromNow();

  return (
    <Tooltip>
      <TooltipTrigger render={<span />} className="cursor-default">
        {relative}
      </TooltipTrigger>
      <TooltipContent>{exact}</TooltipContent>
    </Tooltip>
  );
}
