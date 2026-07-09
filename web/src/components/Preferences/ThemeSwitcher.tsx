import { useUiStore, ThemeMode } from "../../stores";
import { useTranslation } from "../../i18n/useTranslation";

const themeModes: ThemeMode[] = ["light", "dark", "system"];

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
      <div className="px-1 text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
        {t("preferences.theme")}
      </div>
      <div className="grid grid-cols-3 rounded-lg bg-gray-100 p-1 text-xs font-medium text-gray-600 dark:bg-gray-800 dark:text-gray-300">
        {themeModes.map((mode) => (
          <button
            key={mode}
            type="button"
            className={`rounded-md px-2 py-1.5 transition ${themeMode === mode ? "bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-white" : "hover:text-gray-900 dark:hover:text-white"}`}
            onClick={() => setThemeMode(mode)}
          >
            {t(themeMessageKeys[mode])}
          </button>
        ))}
      </div>
    </div>
  );
}
