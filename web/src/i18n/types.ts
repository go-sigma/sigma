import { Locale } from "../stores";
import { MessageKey } from "./messages";

export type { Locale, MessageKey };
export type MessageParams = Record<string, string | number>;
export type Messages = Record<MessageKey, string>;
