import { MessageKey } from "./messages";
import { Locale } from "@/stores";

export type { Locale, MessageKey };
export type MessageParams = Record<string, string | number>;
export type Messages = Record<MessageKey, string>;
