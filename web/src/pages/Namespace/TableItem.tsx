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

import dayjs from 'dayjs';
import { Fragment } from 'react';
import { useRef, useState } from "react";
import { useClickAway } from 'react-use';
import { useNavigate } from 'react-router-dom';
import relativeTime from 'dayjs/plugin/relativeTime';
import { Menu, Transition } from '@headlessui/react'
import { EllipsisVerticalIcon } from '@heroicons/react/20/solid'

dayjs.extend(relativeTime);

export default function TableItem({ index, name, description, size, repository_count, tag_count, created_at, updated_at }: { index: number, name: string, description: string, size: number, repository_count: number, tag_count: number, created_at: string, updated_at: string }) {
  const navigate = useNavigate();
  const [show, setShow] = useState(false);

  const ref = useRef<HTMLDivElement>() as React.MutableRefObject<HTMLDivElement>;
  useClickAway(ref, () => {
    if (show) {
      setShow(!show);
    }
  });

  console.log(description)

  return (
    <tr className="cursor-pointer"
      onClick={() => {
        navigate(`/namespace/${name}/repository`);
      }}
    >
      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-medium text-gray-900">
        <div className="items-center space-x-3 lg:pl-2">
          <div className="truncate hover:text-gray-600">
            <span>
              {name}
              <span className="text-gray-500 font-normal ml-4">{description}</span>
            </span>
          </div>
        </div>
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {size}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {repository_count}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {tag_count}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs().to(dayjs(created_at))}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs().to(dayjs(updated_at))}
      </td>
      <td className="pr-3 whitespace-nowrap text-center" onClick={e => {
        e.stopPropagation();
      }}>
        <Menu as="div" className="relative flex-none" onClick={e => {
          e.stopPropagation();
        }}>
          <Menu.Button className="mx-auto -m-2.5 block p-2.5 text-gray-500 hover:text-gray-900 margin">
            <span className="sr-only">Open options</span>
            <EllipsisVerticalIcon className="h-5 w-5" aria-hidden="true" />
          </Menu.Button>
          <Transition
            as={Fragment}
            enter="transition ease-out duration-100"
            enterFrom="transform opacity-0 scale-95"
            enterTo="transform opacity-100 scale-100"
            leave="transition ease-in duration-75"
            leaveFrom="transform opacity-100 scale-100"
            leaveTo="transform opacity-0 scale-95"
          >
            <Menu.Items className={(index > 10 ? "menu-action-top" : "mt-2") + " text-left absolute right-0 z-10 w-20 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"} >
              <Menu.Item>
                {({ active }) => (
                  <div
                    className={
                      (active ? 'bg-gray-100' : '') +
                      ' block px-3 py-1 text-sm leading-6 text-gray-900'
                    }
                  >
                    Update
                  </div>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <div
                    className={
                      (active ? 'bg-gray-50' : '') + ' block px-3 py-1 text-sm leading-6 text-gray-900 hover:text-white hover:bg-red-600'
                    }
                  >
                    Delete
                  </div>
                )}
              </Menu.Item>
            </Menu.Items>
          </Transition>
        </Menu>
      </td>
    </tr >
  );
}
