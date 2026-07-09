import { ResolvedTheme, ThemeMode, uiPreferencesStorageKey } from "../stores";

export function getSystemPrefersDark(): boolean {
  if (typeof window === "undefined" || typeof window.matchMedia !== "function") {
    return false;
  }
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export function resolveTheme(themeMode: ThemeMode, prefersDark: boolean): ResolvedTheme {
  if (themeMode === "system") {
    return prefersDark ? "dark" : "light";
  }
  return themeMode;
}

export function applyTheme(theme: ResolvedTheme) {
  if (typeof document === "undefined") {
    return;
  }

  document.documentElement.classList.toggle("dark", theme === "dark");
  document.documentElement.style.colorScheme = theme;
}

export function readStoredThemeMode(): ThemeMode {
  if (typeof window === "undefined") {
    return "system";
  }

  try {
    const stored = window.localStorage.getItem(uiPreferencesStorageKey);
    const parsed = stored ? JSON.parse(stored) : null;
    const themeMode = parsed?.state?.themeMode;
    return themeMode === "light" || themeMode === "dark" || themeMode === "system" ? themeMode : "system";
  } catch {
    return "system";
  }
}
