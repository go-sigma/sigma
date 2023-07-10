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



import { useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { IconType } from "react-icons";
import { useClickAway } from 'react-use';

import { AiOutlineHome, AiOutlineMenu, AiOutlineClockCircle } from "react-icons/ai";

interface IMenu {
  text: string;
  icon: any;
}

const menuList: IMenu[] = [
  {
    text: "Home", icon: (
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 576 512" fill="currentColor" className="h-6 w-6"><path d="M575.8 255.5c0 18-15 32.1-32 32.1h-32l.7 160.2c0 2.7-.2 5.4-.5 8.1V472c0 22.1-17.9 40-40 40H456c-1.1 0-2.2 0-3.3-.1c-1.4 .1-2.8 .1-4.2 .1H416 392c-22.1 0-40-17.9-40-40V448 384c0-17.7-14.3-32-32-32H256c-17.7 0-32 14.3-32 32v64 24c0 22.1-17.9 40-40 40H160 128.1c-1.5 0-3-.1-4.5-.2c-1.2 .1-2.4 .2-3.6 .2H104c-22.1 0-40-17.9-40-40V360c0-.9 0-1.9 .1-2.8V287.6H32c-18 0-32-14-32-32.1c0-9 3-17 10-24L266.4 8c7-7 15-8 22-8s15 2 21 7L564.8 231.5c8 7 12 15 11 24z" /></svg>
    )
  },
  {
    text: "Namespace", icon: (
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512" fill="currentColor" className="h-6 w-6"><path d="M0 96C0 78.3 14.3 64 32 64H416c17.7 0 32 14.3 32 32s-14.3 32-32 32H32C14.3 128 0 113.7 0 96zM0 256c0-17.7 14.3-32 32-32H416c17.7 0 32 14.3 32 32s-14.3 32-32 32H32c-17.7 0-32-14.3-32-32zM448 416c0 17.7-14.3 32-32 32H32c-17.7 0-32-14.3-32-32s14.3-32 32-32H416c17.7 0 32 14.3 32 32z" /></svg>
    )
  },
  {
    text: "Settings", icon: (
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" fill="currentColor" className="h-6 w-6"><path d="M495.9 166.6c3.2 8.7 .5 18.4-6.4 24.6l-43.3 39.4c1.1 8.3 1.7 16.8 1.7 25.4s-.6 17.1-1.7 25.4l43.3 39.4c6.9 6.2 9.6 15.9 6.4 24.6c-4.4 11.9-9.7 23.3-15.8 34.3l-4.7 8.1c-6.6 11-14 21.4-22.1 31.2c-5.9 7.2-15.7 9.6-24.5 6.8l-55.7-17.7c-13.4 10.3-28.2 18.9-44 25.4l-12.5 57.1c-2 9.1-9 16.3-18.2 17.8c-13.8 2.3-28 3.5-42.5 3.5s-28.7-1.2-42.5-3.5c-9.2-1.5-16.2-8.7-18.2-17.8l-12.5-57.1c-15.8-6.5-30.6-15.1-44-25.4L83.1 425.9c-8.8 2.8-18.6 .3-24.5-6.8c-8.1-9.8-15.5-20.2-22.1-31.2l-4.7-8.1c-6.1-11-11.4-22.4-15.8-34.3c-3.2-8.7-.5-18.4 6.4-24.6l43.3-39.4C64.6 273.1 64 264.6 64 256s.6-17.1 1.7-25.4L22.4 191.2c-6.9-6.2-9.6-15.9-6.4-24.6c4.4-11.9 9.7-23.3 15.8-34.3l4.7-8.1c6.6-11 14-21.4 22.1-31.2c5.9-7.2 15.7-9.6 24.5-6.8l55.7 17.7c13.4-10.3 28.2-18.9 44-25.4l12.5-57.1c2-9.1 9-16.3 18.2-17.8C227.3 1.2 241.5 0 256 0s28.7 1.2 42.5 3.5c9.2 1.5 16.2 8.7 18.2 17.8l12.5 57.1c15.8 6.5 30.6 15.1 44 25.4l55.7-17.7c8.8-2.8 18.6-.3 24.5 6.8c8.1 9.8 15.5 20.2 22.1 31.2l4.7 8.1c6.1 11 11.4 22.4 15.8 34.3zM256 336a80 80 0 1 0 0-160 80 80 0 1 0 0 160z" /></svg>
    )
  },
];

export default function Menu({ item }: { item: string }) {
  const [showProfileMenu, setShowProfileMenu] = useState(false);
  const [menuActive, setMenuActive] = useState(item === "" ? menuList[0].text : item);

  const ref = useRef<HTMLDivElement>() as React.MutableRefObject<HTMLDivElement>;
  useClickAway(ref, () => {
    if (showProfileMenu) {
      setShowProfileMenu(!showProfileMenu);
    }
  });

  return (
    <div className="flex flex-shrink-0">
      <div className="flex flex-col w-64 border-r border-gray-200 pt-5 pb-4 bg-white">
        <div className="flex items-center flex-shrink-0 px-6">
          <img className="h-10 w-auto" src="/title.svg" alt="Workflow" />
        </div>
        <div className="h-0 flex-1 flex flex-col overflow-y-auto">
          <div className="px-3 mt-6 relative inline-block text-left" ref={ref}>
            <div>
              <button type="button" className="group w-full bg-gray-100 rounded-md px-3.5 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-100 focus:ring-purple-500" onClick={() => { setShowProfileMenu(!showProfileMenu) }}>
                <span className="flex w-full justify-between items-center">
                  <span className="flex min-w-0 items-center justify-between space-x-3">
                    <span className="flex min-w-0 flex-col">
                      <span className="text-gray-700 text-sm font-medium truncate">Tosone</span>
                    </span>
                  </span>
                  <svg className="flex-shrink-0 text-xs h-4 w-4 text-gray-400 group-hover:text-gray-500" xmlns="http://www.w3.org/2000/svg" fill="currentColor" aria-hidden="true" viewBox="0 0 512 512">
                    <path d="M233.4 406.6c12.5 12.5 32.8 12.5 45.3 0l192-192c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0L256 338.7 86.6 169.4c-12.5-12.5-32.8-12.5-45.3 0s-12.5 32.8 0 45.3l192 192z" />
                  </svg>
                </span>
              </button>
            </div>
            <div className={showProfileMenu ? "" : "hidden"}>
              <div className="z-10 mx-3 origin-top absolute right-0 left-0 mt-1 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 divide-y divide-gray-200">
                <div className="py-1">
                  <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">View profile</div>
                  <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">Settings</div>
                  <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">Notifications</div>
                </div>
                <div className="py-1">
                  <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">Get desktop app</div>
                  <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">Support</div>
                </div>
                <div className="py-1">
                  <div className="cursor-pointer block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">Logout</div>
                </div>
              </div>
            </div>
          </div>
          <nav className="px-3 mt-6">
            <div className="space-y-1">
              {
                menuList.map(m => {
                  return (
                    <Link to={`/${m.text.toLowerCase()}`} key={m.text} className={`text-gray-700 group flex items-center px-2 py-2 text-sm font-medium rounded-md ${menuActive === m.text ? "bg-gray-100" : "hover:bg-gray-100 text-gray-700"}`} onClick={() => {
                      setMenuActive(m.text);
                    }}>
                      <span className="text-gray-400 mr-3 h-6 w-6">
                        {m.icon}
                      </span>
                      {m.text}
                    </Link>
                  )
                })
              }
            </div>
            <div className="mt-8">
              <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider" id="teams-headline">
                Hot namespace
              </h3>
              <div className="mt-1 space-y-1" role="group" aria-labelledby="teams-headline">
                <div className="group flex items-center px-3 py-2 text-sm font-medium text-gray-700 rounded-md hover:text-gray-900 hover:bg-gray-50 cursor-pointer">
                  <span className="w-2.5 h-2.5 mr-4 bg-indigo-500 rounded-full" aria-hidden="true"></span>
                  <span className="truncate">
                    public
                  </span>
                </div>

                <div className="group flex items-center px-3 py-2 text-sm font-medium text-gray-700 rounded-md hover:text-gray-900 hover:bg-gray-50 cursor-pointer">
                  <span className="w-2.5 h-2.5 mr-4 bg-green-500 rounded-full" aria-hidden="true"></span>
                  <span className="truncate">
                    library
                  </span>
                </div>

                <div className="group flex items-center px-3 py-2 text-sm font-medium text-gray-700 rounded-md hover:text-gray-900 hover:bg-gray-50 cursor-pointer">
                  <span className="w-2.5 h-2.5 mr-4 bg-yellow-500 rounded-full" aria-hidden="true"></span>
                  <span className="truncate">
                    test
                  </span>
                </div>
              </div>
            </div>
          </nav>
        </div>
      </div>
    </div >
  );
}
