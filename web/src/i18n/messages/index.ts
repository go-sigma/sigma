import enUS from "./en-US";
import zhCN from "./zh-CN";

export const messagesByLocale = {
  "en-US": enUS,
  "zh-CN": zhCN,
} as const;

export type MessageKey = keyof typeof enUS;
