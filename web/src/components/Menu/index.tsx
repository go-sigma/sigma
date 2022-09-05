import { useRef, useState } from "react";
import { Link } from "react-router-dom";
import { AiOutlineHome, AiOutlineMenu, AiOutlineClockCircle } from "react-icons/ai";
import { IconType } from "react-icons";
import { useClickAway } from 'react-use';

interface IMenu {
  text: string;
  icon: IconType;
};

const menuList: IMenu[] = [
  { text: "Home", icon: AiOutlineHome },
  { text: "My tasks", icon: AiOutlineMenu },
  { text: "Recent", icon: AiOutlineClockCircle },
];

export default function Menu() {
  let [showProfileMenu, setShowProfileMenu] = useState(false);
  let [menuActive, setMenuActive] = useState(menuList[0].text);

  const ref = useRef<HTMLDivElement>() as React.MutableRefObject<HTMLDivElement>;;
  useClickAway(ref, () => {
    if (showProfileMenu) {
      setShowProfileMenu(!showProfileMenu);
    }
  });

  return (
    <div className="hidden lg:flex lg:flex-shrink-0">
      <div className="flex flex-col w-64 border-r border-gray-200 pt-5 pb-4 bg-gray-100">
        <div className="flex items-center flex-shrink-0 px-6">
          <img className="h-8 w-auto" src="https://tailwindui.com/img/logos/workflow-logo-purple-500-mark-gray-700-text.svg" alt="Workflow" />
        </div>
        <div className="h-0 flex-1 flex flex-col overflow-y-auto">
          <div className="px-3 mt-6 relative inline-block text-left" ref={ref}>
            <div>
              <button type="button" className="group w-full bg-gray-100 rounded-md px-3.5 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-100 focus:ring-purple-500" onClick={() => { setShowProfileMenu(!showProfileMenu) }}>
                <span className="flex w-full justify-between items-center">
                  <span className="flex min-w-0 items-center justify-between space-x-3">
                    <img className="w-10 h-10 bg-gray-300 rounded-full flex-shrink-0" src="https://images.unsplash.com/photo-1502685104226-ee32379fefbe?ixlib=rb-1.2.1&amp;ixid=eyJhcHBfaWQiOjEyMDd9&amp;auto=format&amp;fit=facearea&amp;facepad=3&amp;w=256&amp;h=256&amp;q=80" alt="" />
                    <span className="flex min-w-0 flex-col">
                      <span className="text-gray-900 text-sm font-medium truncate">Tosone</span>
                    </span>
                  </span>
                  <svg className="flex-shrink-0 h-5 w-5 text-gray-400 group-hover:text-gray-500" x-description="Heroicon name: solid/selector" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                    <path fillRule="evenodd" d="M10 3a1 1 0 01.707.293l3 3a1 1 0 01-1.414 1.414L10 5.414 7.707 7.707a1 1 0 01-1.414-1.414l3-3A1 1 0 0110 3zm-3.707 9.293a1 1 0 011.414 0L10 14.586l2.293-2.293a1 1 0 011.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd"></path>
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
                    <Link to="" key={m.text} className={`text-gray-900 group flex items-center px-2 py-2 text-sm font-medium rounded-md ${menuActive === m.text ? "bg-gray-200" : "text-gray-700"}`} onClick={() => {
                      setMenuActive(m.text);
                    }}>
                      <m.icon className="text-gray-500 mr-3 h-6 w-6" />
                      {m.text}
                    </Link>
                  )
                })
              }
            </div>
            <div className="mt-8">
              <h3 className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider" id="teams-headline">
                Teams
              </h3>
              <div className="mt-1 space-y-1" role="group" aria-labelledby="teams-headline">
                <div className="group flex items-center px-3 py-2 text-sm font-medium text-gray-700 rounded-md hover:text-gray-900 hover:bg-gray-50 cursor-pointer">
                  <span className="w-2.5 h-2.5 mr-4 bg-indigo-500 rounded-full" aria-hidden="true"></span>
                  <span className="truncate">
                    Engineering
                  </span>
                </div>

                <div className="group flex items-center px-3 py-2 text-sm font-medium text-gray-700 rounded-md hover:text-gray-900 hover:bg-gray-50 cursor-pointer">
                  <span className="w-2.5 h-2.5 mr-4 bg-green-500 rounded-full" aria-hidden="true"></span>
                  <span className="truncate">
                    Human Resources
                  </span>
                </div>

                <div className="group flex items-center px-3 py-2 text-sm font-medium text-gray-700 rounded-md hover:text-gray-900 hover:bg-gray-50 cursor-pointer">
                  <span className="w-2.5 h-2.5 mr-4 bg-yellow-500 rounded-full" aria-hidden="true"></span>
                  <span className="truncate">
                    Customer Success
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
