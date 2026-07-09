import { Locale, useUiStore } from "../../stores";
import { useTranslation } from "../../i18n/useTranslation";

const locales: Locale[] = ["en-US", "zh-CN"];

const localeLabels: Record<Locale, "preferences.locale.english" | "preferences.locale.chinese"> = {
  "en-US": "preferences.locale.english",
  "zh-CN": "preferences.locale.chinese",
};

export default function LocaleSwitcher() {
  const locale = useUiStore((state) => state.locale);
  const setLocale = useUiStore((state) => state.setLocale);
  const { t } = useTranslation();

  return (
    <div className="space-y-2">
      <div className="px-1 text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
        {t("preferences.language")}
      </div>
      <div className="grid grid-cols-2 rounded-lg bg-gray-100 p-1 text-xs font-medium text-gray-600 dark:bg-gray-800 dark:text-gray-300">
        {locales.map((item) => (
          <button
            key={item}
            type="button"
            className={`rounded-md px-2 py-1.5 transition ${locale === item ? "bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-white" : "hover:text-gray-900 dark:hover:text-white"}`}
            onClick={() => setLocale(item)}
          >
            {t(localeLabels[item])}
          </button>
        ))}
      </div>
    </div>
  );
}
