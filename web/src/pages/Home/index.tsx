/**
 * Copyright 2023 XImager
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

import { Fragment, useCallback } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import Tags from '@yaireo/tagify/dist/react.tagify';
import "@yaireo/tagify/dist/tagify.css";

import Menu from "../../components/Menu";
import Header from "../../components/Header";

import { ScaleIcon } from '@heroicons/react/24/outline'

import "./index.css";

const cards = [
  { name: 'Account balance1', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
  { name: 'Account balance2', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
  { name: 'Account balance3', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
  { name: 'Account balance4', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
];

const settings = {
  pattern: /^[a-z_-]{1,15}$/,
  maxTags: 10,
  templates: {}
}

export default function Home({ localServer }: { localServer: string }) {
  const onChange = useCallback((e: any) => {
    console.log("CHANGED:"
      , e.detail.tagify.value // Array where each tag includes tagify's (needed) extra properties
      , e.detail.tagify.getCleanValue() // Same as above, without the extra properties
      , e.detail.value // a string representing the tags
    )
  }, [])
  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>XImager - Home</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white min-w-1600">
        <Menu item="Home" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="flex-1 relative z-0 focus:outline-none" tabIndex={0}>
            <Header title="Home" />
            <div className="py-3 px-3 sm:px-6 lg:px-8">
              <div className="flex flex-wrap justify-around mt-2 gap-5">
                {cards.map((card) => (
                  <div key={card.name} className="overflow-hidden rounded-lg bg-white shadow w-1/5">
                    <div className="p-5">
                      <div className="flex items-center">
                        <div className="flex-shrink-0">
                          <card.icon className="h-6 w-6 text-gray-400" aria-hidden="true" />
                        </div>
                        <div className="ml-5 w-0 flex-1">
                          <dl>
                            <dt className="truncate text-sm font-medium text-gray-500">{card.name}</dt>
                            <dd>
                              <div className="text-lg font-medium text-gray-900">{card.amount}</div>
                            </dd>
                          </dl>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            <Tags
              settings={settings}
              defaultValue="a,b,c"
              autoFocus={true}
              placeholder="please type your tags"
              onChange={onChange}
            // className="mt-1 py-2 px-3 w-full border-2 border-purple-300 rounded-2xl outline-none  invalid:text-pink-700 invalid:focus:ring-pink-700 peer  dark:text-gray-200"
            />
          </main>
        </div>
      </div>
    </Fragment >
  )
}
