import { useCallback } from "react";

import { useUiStore } from "../stores";
import { MessageKey, MessageParams } from "./types";
import { messagesByLocale } from "./messages";

function interpolate(message: string, params?: MessageParams) {
  if (params === undefined) {
    return message;
  }

  return message.replace(/\{(\w+)}/g, (token, key) => {
    const value = params[key];
    return value === undefined ? token : String(value);
  });
}

export function useTranslation() {
  const locale = useUiStore((state) => state.locale);

  const t = useCallback((key: MessageKey, params?: MessageParams) => {
    const messages = messagesByLocale[locale];
    const fallback = messagesByLocale["en-US"];
    const message = messages[key] ?? fallback[key];

    if (message === undefined) {
      if (import.meta.env.DEV) {
        console.warn(`missing i18n message: ${key}`);
      }
      return key;
    }

    return interpolate(message, params);
  }, [locale]);

  return { locale, t };
}
