import { useEffect } from "react";

import { applyTheme, getSystemPrefersDark, resolveTheme } from "./theme";
import { useUiStore } from "@/stores";

export default function useThemeEffect() {
  const themeMode = useUiStore((state) => state.themeMode);
  const setResolvedTheme = useUiStore((state) => state.setResolvedTheme);

  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");

    const syncTheme = () => {
      const resolvedTheme = resolveTheme(themeMode, getSystemPrefersDark());
      setResolvedTheme(resolvedTheme);
      applyTheme(resolvedTheme);
    };

    syncTheme();
    mediaQuery.addEventListener("change", syncTheme);

    return () => {
      mediaQuery.removeEventListener("change", syncTheme);
    };
  }, [setResolvedTheme, themeMode]);
}
