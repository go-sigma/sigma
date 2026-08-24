/**
 * Copyright 2023 sigma
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

import humanFormat from 'human-format';
import { Badge } from "@/components/ui/badge";

import Settings from "../../Settings";

export default function ({ current, limit }: { current: number, limit: number }) {
  const threshold = limit !== 0 ? (current / limit > 1 ? 1 : current / limit) : 0;
  const percentage = limit !== 0 ? Math.min((current / limit * 100), 100) : 0;

  return (
    <div className="text-right text-sm">
      {
        limit === 0 ? (
          <>{humanFormat(current)}</>
        ) : (
          <div className="flex items-center justify-end gap-2">
            <span className="text-sm">{humanFormat(current)} / {humanFormat(limit)}</span>
            <Badge variant={threshold > Settings.QuotaThreshold ? "destructive" : "secondary"}>
              {percentage.toFixed(1)}%
            </Badge>
          </div>
        )
      }
    </div>
  );
}