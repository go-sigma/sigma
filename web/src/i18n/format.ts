import dayjs from "dayjs";

import { Locale } from "../stores";

export function formatDateTime(value: string | number | Date, _locale: Locale) {
  return dayjs(value).format("YYYY-MM-DD HH:mm:ss");
}

export function formatRelativeTime(value: string | number | Date, _locale: Locale) {
  return dayjs(value).fromNow();
}
