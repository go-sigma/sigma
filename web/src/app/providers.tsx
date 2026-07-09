/**
 * Copyright 2026 sigma
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

import { ReactNode } from 'react';
import { HelmetProvider } from 'react-helmet-async';
import { Toaster } from 'react-hot-toast';
import { ToastContainer } from 'react-toastify';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { useUiStore } from '../stores';
import useThemeEffect from './useThemeEffect';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30 * 1000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

export default function AppProviders({ children }: { children: ReactNode }) {
  useThemeEffect();

  const resolvedTheme = useUiStore((state) => state.resolvedTheme);

  return (
    <QueryClientProvider client={queryClient}>
      <HelmetProvider>
        {children}
        <Toaster
          toastOptions={{
            style: {
              maxWidth: "500px",
              fontSize: "0.875rem",
              lineHeight: "1.25rem",
            }
          }}
        />
        <ToastContainer
          position="top-right"
          autoClose={5000}
          hideProgressBar={false}
          newestOnTop={false}
          closeOnClick
          rtl={false}
          pauseOnFocusLoss
          draggable
          pauseOnHover
          theme={resolvedTheme}
          style={{ top: "3rem" }}
        />
      </HelmetProvider>
    </QueryClientProvider>
  );
}
