import { useRef, useState } from "react";
import { useClickAway } from 'react-use';
import { useNavigate } from 'react-router-dom';
import relativeTime from 'dayjs/plugin/relativeTime';
import dayjs from 'dayjs';

dayjs.extend(relativeTime);

export default function TableItem({ id, namespace, name, artifact_count, created_at, updated_at }: { id: number, namespace: string | undefined, name: string, artifact_count: number, created_at: string, updated_at: string }) {
  const navigate = useNavigate();
  let [show, setShow] = useState(false);

  const ref = useRef<HTMLDivElement>() as React.MutableRefObject<HTMLDivElement>;;
  useClickAway(ref, () => {
    if (show) {
      setShow(!show);
    }
  });

  return (
    <tr className="cursor-pointer" onClick={() => {
      navigate(`/namespace/${namespace}/artifact?repository=${name}`);
    }}>
      <td className="px-6 py-4 max-w-0 w-full whitespace-nowrap text-sm font-medium text-gray-900">
        <div className="flex items-center space-x-3 lg:pl-2">
          <div className="cursor-pointer truncate hover:text-gray-600">
            {name}
          </div>
        </div>
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {artifact_count}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs().to(dayjs(created_at))}
      </td>
      <td className="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
        {dayjs().to(dayjs(updated_at))}
      </td>
      <td className="pr-3 whitespace-nowrap">
        <button
          type="button"
          className=" w-1/2  rounded-md border border-transparent bg-white font-medium text-indigo-600 hover:text-indigo-500  mr-5"
        >
          Remove
        </button>
      </td>
    </tr>
  );
}
