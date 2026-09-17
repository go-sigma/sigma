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

import { z } from "zod";

import { ISizeWithUnit } from "../interfaces";

/** Remove the http(s) scheme from an endpoint. */
export function trimHTTP(str: string) {
  if (str.startsWith("http://")) {
    return str.substring(7);
  } else if (str.startsWith("https://")) {
    return str.substring(8);
  }
  return str;
}

/** Convert a byte size into a human friendly value with its unit. */
export function calcUnit(size: number): ISizeWithUnit {
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

const emailSchema = z.email();

/** Validate an email address (zod) — replaces the ported HTML5 pattern. */
export const isEmail = (value: string) => emailSchema.safeParse(value).success;

/** Shared validation patterns. */
export const Regex = {
  Username: /^[a-zA-Z0-9_\-@#$%]+$/,
  Password: /^(?=.*?[A-Z])(?=.*?[a-z])(?=.*?[0-9])(?=.*?[#?!@$ %^&*-]).{8,}$/,
};

/** Minimal DOM tooltip used by the legacy pages. */
export class Tooltip {
  private target: HTMLElement | null;
  private trigger: HTMLElement | null;
  private timer?: number;

  constructor(target: HTMLElement | null, trigger: HTMLElement | null, _options?: unknown) {
    this.target = target;
    this.trigger = trigger;
  }

  show() {
    if (this.target === null || this.trigger === null) {
      return;
    }

    const rect = this.trigger.getBoundingClientRect();
    this.target.style.position = "fixed";
    this.target.style.left = `${rect.left + rect.width / 2}px`;
    this.target.style.top = `${rect.top - 8}px`;
    this.target.style.transform = "translate(-50%, -100%)";
    this.target.classList.remove("invisible", "opacity-0");
    this.target.classList.add("visible", "opacity-100");

    window.clearTimeout(this.timer);
    this.timer = window.setTimeout(() => {
      this.hide();
    }, 1600);
  }

  hide() {
    if (this.target === null) {
      return;
    }
    this.target.classList.remove("visible", "opacity-100");
    this.target.classList.add("invisible", "opacity-0");
  }
}
