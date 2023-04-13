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

import { Fragment } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';

import Menu from "../../components/Menu";
import Header from "../../components/Header";

import { ScaleIcon } from '@heroicons/react/24/outline'

import "./index.css";

const cards = [
  { name: 'Account balance1', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
  { name: 'Account balance2', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
  { name: 'Account balance3', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
  { name: 'Account balance4', href: '#', icon: ScaleIcon, amount: '$30,659.45' },
]

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import { faker } from '@faker-js/faker';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
);

export const options = {
  responsive: true,
  interaction: {
    mode: 'index' as const,
    intersect: false,
  },
  cubicInterpolationMode: 'monotone',
  stacked: false,
  plugins: {
    title: {
      display: true,
      text: 'Chart.js Line Chart - Multi Axis',
    },
  },
  scales: {
    y: {
      type: 'linear' as const,
      display: true,
      position: 'left' as const,
    },
    y1: {
      type: 'linear' as const,
      display: true,
      position: 'right' as const,
      grid: {
        drawOnChartArea: false,
      },
    },
  },
};

const labels = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

export const data = {
  labels,
  datasets: [
    {
      label: 'Dataset 1',
      data: labels.map(() => faker.datatype.number({ min: 1, max: 1000 })),
      borderColor: 'rgb(255, 99, 132)',
      backgroundColor: 'rgba(255, 99, 132, 0.5)',
      yAxisID: 'y',
    },
    {
      label: 'Dataset 2',
      data: labels.map(() => faker.datatype.number({ min: 1, max: 1000 })),
      borderColor: 'rgb(53, 162, 235)',
      backgroundColor: 'rgba(53, 162, 235, 0.5)',
      yAxisID: 'y1',
    },
  ],
};

export default function Home({ localServer }: { localServer: string }) {
  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>XImager - Home</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
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
            <div className="flex justify-around mt-7 px-3 gap-1">
              <div className="px-4 py-5 sm:px-6 w-1/2 bg-slate-50 rounded-md">
                <h3 className="text-base font-semibold leading-6 text-gray-900">Storage</h3>
                <Line className="px-3" options={options} data={data} />
              </div>
              <div className="px-4 py-5 sm:px-6 w-1/2 bg-slate-50 rounded-md">
                <h3 className="text-base font-semibold leading-6 text-gray-900">Pull & Push</h3>
                <Line className="px-3" options={options} data={data} />
              </div>
            </div>
          </main>
        </div>
      </div>
    </Fragment >
  )
}
