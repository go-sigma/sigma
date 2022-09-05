import { IconType } from "react-icons";
import { useRef, useState } from "react";
import { useClickAway } from 'react-use';
import { AiOutlineMore } from "react-icons/ai";
import { AiOutlineCopy, AiOutlineEdit, AiOutlineDelete, AiOutlineShareAlt } from "react-icons/ai";

interface IMenu {
  text: string;
  icon: IconType;
};

const menuList: IMenu[] = [
  { text: "Edit", icon: AiOutlineEdit },
  { text: "Duplicate", icon: AiOutlineCopy },
  { text: "Share", icon: AiOutlineShareAlt },
];

export default function TableItem({ name, description, updated }: { name: string, description: string, updated: string }) {
  let [show, setShow] = useState(false);

  const ref = useRef<HTMLDivElement>() as React.MutableRefObject<HTMLDivElement>;;
  useClickAway(ref, () => {
    if (show) {
      setShow(!show);
    }
  });

  return (
    <tr className="cursor-pointer">
      <td className="px-6 py-3 max-w-0 w-full whitespace-nowrap text-sm font-medium text-gray-900">
        <div className="flex items-center space-x-3 lg:pl-2">
          <div className="flex-shrink-0 w-2.5 h-2.5 rounded-full bg-pink-600"></div>
          <div className="cursor-pointer truncate hover:text-gray-600">
            <span>
              {name}
              <span className="text-gray-500 font-normal ml-4">{description}</span>
            </span>
          </div>
        </div>
      </td>
      <td className="hidden md:table-cell px-6 py-3 whitespace-nowrap text-sm text-gray-500 text-right">
        {updated}
      </td>
      <td className="pr-6">
        <div className="relative flex justify-end items-center" ref={ref}>
          <button className="w-6 h-6 bg-white inline-flex items-center justify-center text-gray-400 rounded-full hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500" onClick={() => { setShow(!show) }}>
            <span className="sr-only">Open options</span>
            <AiOutlineMore className="w-5 h-5" />
          </button>
          <div className={`mx-3 origin-top-right absolute right-7 top-0 w-48 mt-1 rounded-md shadow-lg z-10 bg-white ring-1 ring-black ring-opacity-5 divide-y divide-gray-200 ${show ? "" : "hidden"}`}>
            <div className="py-1">
              {
                menuList.map(m => {
                  return (
                    <div key={m.text} className="cursor-pointer group flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">
                      <m.icon className="mr-3 h-5 w-5 text-gray-400 group-hover:text-gray-500" />
                      {m.text}
                    </div>
                  );
                })
              }
            </div>
            <div className="py-1">
              <div className="cursor-pointer group flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900">
                <AiOutlineDelete className="mr-3 h-5 w-5 text-gray-400 group-hover:text-gray-500" />
                Delete
              </div>
            </div>
          </div>
        </div>
      </td>
    </tr>
  );
}
