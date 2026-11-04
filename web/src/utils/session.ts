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

import { toast } from "sonner";

import { messagesByLocale } from "@/i18n/messages";
import { useUiStore } from "@/stores";

// A request burst that hits an expired access token fails once per request, and every failing
// caller reports its own toast. The window keeps the burst to a single toast while still letting
// a later failure (for example a wrong password on the login page) surface normally.
const suppressionWindowMs = 2000;

let expiredAt = 0;
let notifying = false;

/** sessionExpiredBurst reports whether the app is still handling the token expiry burst. */
export function sessionExpiredBurst() {
  return expiredAt > 0 && Date.now()-expiredAt < suppressionWindowMs;
}

/** clearSession drops the stored credentials of the signed out user. */
export function clearSession() {
  localStorage.removeItem("token");
  localStorage.removeItem("refresh_token");
  localStorage.removeItem("user");
  localStorage.removeItem("username");
  localStorage.removeItem("email");
}

/**
 * notifySessionExpired reports an expired session once: it clears the credentials, shows a single
 * toast and runs onExpire (navigation, timer teardown). Concurrent 401 responses are ignored.
 */
export function notifySessionExpired(onExpire: () => void) {
  if (notifying) {
    return;
  }
  notifying = true;
  expiredAt = Date.now();

  const locale = useUiStore.getState().locale;
  const messages = messagesByLocale[locale] ?? messagesByLocale["en-US"];
  clearSession();
  toast.warning(messages["auth.sessionExpired"], {
    description: messages["auth.sessionExpiredDescription"],
    duration: 4000,
  });
  onExpire();

  // A new session can expire again later, so stop ignoring after the burst settles.
  window.setTimeout(() => {
    notifying = false;
  }, suppressionWindowMs);
}

/** resetSessionExpired clears the burst state after a successful sign in. */
export function resetSessionExpired() {
  expiredAt = 0;
  notifying = false;
}
