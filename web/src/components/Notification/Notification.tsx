import "./index.css";
import { toast } from 'react-toastify';
import { InboxIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline';

import { INotification } from "../../interfaces/interfaces";

export default function Toast(noti: INotification) {
  const id = toast('fake-toast', {
    position: "top-right",
    autoClose: 5000,
    hideProgressBar: true,
    closeOnClick: true,
    closeButton: false,
    pauseOnHover: true,
    draggable: true,
    progress: 1,
    theme: "light",
    className: "fake-toast",
  });

  let renderWithLevel = () => {
    switch (noti.level) {
      case "info":
        return <InboxIcon className="h-6 w-6 text-red-400" aria-hidden="true" />
      case "success":
        return <InboxIcon className="h-6 w-6 text-gray-400" aria-hidden="true" />
      case "warning":
        return <ExclamationTriangleIcon className="h-6 w-6 text-red-400" aria-hidden="true" />
    }
  }

  toast.update(id, {
    render: (
      <div className="pointer-events-auto w-full max-w-sm overflow-hidden rounded-lg bg-white shadow-lg ring-1 ring-black ring-opacity-5">
        <div className="p-4">
          <div className="flex items-start">
            <div className="flex-shrink-0">
              {renderWithLevel()}
              {/* <InboxIcon className="h-6 w-6 text-gray-400" aria-hidden="true" /> */}
            </div>
            <div className="ml-3 w-0 flex-1 pt-0.5">
              <p className="text-sm font-medium text-gray-900">{noti.title}</p>
              <p className="mt-1 text-sm text-gray-500">
                {noti.message}
              </p>
            </div>
          </div>
        </div>
      </div>
    ),
  });
}
