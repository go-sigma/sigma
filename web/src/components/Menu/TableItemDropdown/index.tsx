/**
 * Copyright 2024 sigma
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

import "./index.css";

import clsx from 'clsx';
import { Fragment } from 'react';
import { EllipsisVerticalIcon } from "@heroicons/react/20/solid";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/react";

interface ITableItemDropdownProps {
  index?: number;
  items?: ITableItemDropdownItem[];
}

interface ITableItemDropdownItem {
  name: string;
  onClick: () => void;
  warn?: boolean;
  disable?: boolean;
}

export default function ({ index, items }: ITableItemDropdownProps) {
  return items && items.length > 0 && !items.every(item => item.disable) ? (
    <Menu as="div" className="relative flex-none" onClick={e => e.stopPropagation()}>
      <MenuButton className="mx-auto -m-2.5 block p-2.5 text-gray-500 hover:text-gray-900 margin">
        <span className="sr-only">Open options</span>
        <EllipsisVerticalIcon className="h-5 w-5" aria-hidden="true" />
      </MenuButton>
      <MenuItems
        className={clsx(
          index || 0 > 10 ? "menu-action-top" : "mt-2",
          " text-left absolute right-0 z-10 w-20 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none")
        }
      >
        {
          items.map((item, index) => (
            item.disable ? null :
              <MenuItem as={Fragment} key={index}>
                {({ focus }) => (
                  <div
                    className={clsx(
                      focus ? 'bg-gray-100' : '',
                      'block px-4 py-1.5 text-sm leading-6 text-gray-900 cursor-pointer',
                      item.warn ? 'hover:text-white hover:bg-red-600' : ''
                    )}
                    onClick={item.onClick}
                  >
                    {item.name}
                  </div>
                )}
              </MenuItem>
          ))
        }
      </MenuItems>
    </Menu >
  ) : null;
}
