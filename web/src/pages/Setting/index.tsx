/**
 * Copyright 2023 sigma
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

import { Fragment } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';

import Header from "../../components/Header";
import Menu from "../../components/Menu";
import { useTranslation } from "../../i18n/useTranslation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function ({ localServer }: { localServer: string }) {
  const { t } = useTranslation();

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>{t("header.setting")}</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background">
        <Menu localServer={localServer} item="settings" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title={t("header.setting")} />
          </main>
          <div className="flex flex-1 overflow-y-auto p-6">
            <Card className="w-full">
              <CardHeader>
                <CardTitle>{t("header.setting")}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">Select a setting from the sidebar.</p>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </Fragment >
  );
}