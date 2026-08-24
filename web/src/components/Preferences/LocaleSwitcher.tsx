import { Locale, useUiStore } from "../../stores";
import { useTranslation } from "../../i18n/useTranslation";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";

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
      <div className="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        {t("preferences.language")}
      </div>
      <ToggleGroup
        value={[locale]}
        onValueChange={(values) => {
          if (values.length > 0) setLocale(values[0] as Locale);
        }}
        className="w-full"
      >
        {locales.map((item) => (
          <ToggleGroupItem key={item} value={item} className="flex-1" size="sm">
            {t(localeLabels[item])}
          </ToggleGroupItem>
        ))}
      </ToggleGroup>
    </div>
  );
}