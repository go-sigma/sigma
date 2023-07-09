/**
 * Copyright 2023 XImager
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

import Settings from "../../Settings";

export default function ({ current, limit }: { current: number, limit: number }) {
  const threshold = limit !== 0 ? (current / limit > 1 ? 1 : current / limit) : 0;

  return (
    <div className={limit === 0 ? "text-right text-sm" : "text-right"}>
      {
        limit === 0 ? (
          <>{humanFormat(current)}</>
        ) : (
          <>
            <div className="mb-1 text-sm font-medium">
              {humanFormat(current)} / {humanFormat(limit)} (<span className={threshold > Settings.QuotaThreshold ? "text-red-700 dark:text-red-500" : "text-blue-700 dark:text-blue-500"}>{(current / limit * 100 > 100 ? 100 : current / limit * 100).toFixed(1)}%</span>)
            </div>
          </>
        )
      }
    </div>
  );
}
