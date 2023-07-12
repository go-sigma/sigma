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

import { ISizeWithUnit } from "../interfaces";

export default function (size: number): ISizeWithUnit {
  let unit = "MiB";
  let result = 0;
  let m = (size / (1 << 20));
  if (m < 1024) {
    unit = "MiB";
    result = m;
  } else {
    m = (size / (1 << 30));
    if (m < 1024) {
      unit = "GiB"
      result = m;
    } else {
      m = (size / (1 << 40));
      unit = "TiB"
      result = m;
    }
  }
  return {
    unit: unit,
    size: result,
  }
}
