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
import { ScaleIcon } from 'lucide-react';

import Header from "../../components/Header";
import Menu from "../../components/Menu";
import { useTranslation } from "../../i18n/useTranslation";
import { Card, CardContent } from "@/components/ui/card";

const cards = [
  { name: 'Account balance1', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
  { name: 'Account balance2', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
  { name: 'Account balance3', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
  { name: 'Account balance4', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
];

export default function Home({ localServer }: { localServer: string }) {
  const { t } = useTranslation();

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>{t("header.home")}</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background min-w-[1600px]">
        <Menu item="home" localServer={localServer} />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="flex-1 relative z-0 focus:outline-none" tabIndex={0}>
            <Header title={t("header.home")} />
            <div className="py-3 px-3 sm:px-6 lg:px-8">
              <div className="flex flex-wrap justify-around mt-2 gap-5">
                {cards.map((card) => (
                  <Card key={card.name} className="w-1/5">
                    <CardContent className="p-5">
                      <div className="flex items-center">
                        <div className="flex-shrink-0">
                          <card.icon className="h-6 w-6 text-muted-foreground" aria-hidden="true" />
                        </div>
                        <div className="ml-5 w-0 flex-1">
                          <dl>
                            <dt className="truncate text-sm font-medium text-muted-foreground">{card.name}</dt>
                            <dd>
                              <div className="text-lg font-medium text-foreground">{card.amount}</div>
                            </dd>
                          </dl>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </div>
          </main>
        </div>
      </div>
    </Fragment>
  )
}