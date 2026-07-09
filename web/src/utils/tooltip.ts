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
