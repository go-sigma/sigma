import { useUiStore, ThemeMode } from "../../stores";
import { useTranslation } from "../../i18n/useTranslation";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { Sun, Moon, Monitor } from "lucide-react";

const themeModes: ThemeMode[] = ["light", "dark", "system"];

const themeIcons: Record<ThemeMode, typeof Sun> = {
  light: Sun,
  dark: Moon,
  system: Monitor,
};

const themeMessageKeys: Record<ThemeMode, "preferences.theme.light" | "preferences.theme.dark" | "preferences.theme.system"> = {
  light: "preferences.theme.light",
  dark: "preferences.theme.dark",
  system: "preferences.theme.system",
};

export default function ThemeSwitcher() {
  const themeMode = useUiStore((state) => state.themeMode);
  const setThemeMode = useUiStore((state) => state.setThemeMode);
  const { t } = useTranslation();

  return (
    <div className="space-y-2">
      <div className="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        {t("preferences.theme")}
      </div>
      <ToggleGroup
        value={[themeMode]}
        onValueChange={(values) => {
          if (values.length > 0) setThemeMode(values[0] as ThemeMode);
        }}
        className="w-full"
      >
        {themeModes.map((mode) => {
          const Icon = themeIcons[mode];
          return (
            <ToggleGroupItem key={mode} value={mode} className="flex-1 gap-1.5" size="sm">
              <Icon className="h-3.5 w-3.5" />
              <span className="sr-only">{t(themeMessageKeys[mode])}</span>
            </ToggleGroupItem>
          );
        })}
      </ToggleGroup>
    </div>
  );
}