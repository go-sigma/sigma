/**
 * Copyright 2025 sigma
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type ThemeMode = "light" | "dark" | "system";
export type ResolvedTheme = "light" | "dark";
export type Locale = "en-US" | "zh-CN";

export const uiPreferencesStorageKey = "sigma-ui-preferences";

function getDefaultLocale(): Locale {
  if (typeof navigator === "undefined") {
    return "en-US";
  }
  return navigator.language.toLowerCase().startsWith("zh") ? "zh-CN" : "en-US";
}

interface UIState {
  sidebarCollapsed: boolean;
  themeMode: ThemeMode;
  resolvedTheme: ResolvedTheme;
  locale: Locale;
  setSidebarCollapsed: (collapsed: boolean) => void;
  toggleSidebarCollapsed: () => void;
  setThemeMode: (mode: ThemeMode) => void;
  setResolvedTheme: (theme: ResolvedTheme) => void;
  setLocale: (locale: Locale) => void;
  toggleLocale: () => void;
}

export const useUiStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarCollapsed: false,
      themeMode: "system",
      resolvedTheme: "light",
      locale: getDefaultLocale(),
      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
      toggleSidebarCollapsed: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      setThemeMode: (mode) => set({ themeMode: mode }),
      setResolvedTheme: (theme) => set({ resolvedTheme: theme }),
      setLocale: (locale) => set({ locale }),
      toggleLocale: () => set((state) => ({ locale: state.locale === "en-US" ? "zh-CN" : "en-US" })),
    }),
    {
      name: uiPreferencesStorageKey,
      partialize: (state) => ({
        themeMode: state.themeMode,
        locale: state.locale,
      }),
    }
  )
);
