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

import "./index.css";

import { ExclamationTriangleIcon, InboxIcon } from '@heroicons/react/24/outline';
import { toast } from 'react-toastify';

import { INotification } from "../../interfaces";

export default function (noti: INotification) {
  const id = toast('fake-toast', {
    position: "top-right",
    autoClose: noti.autoClose || 3000,
    hideProgressBar: true,
    closeOnClick: true,
    closeButton: false,
    pauseOnHover: false,
    draggable: false,
    theme: "light",
    className: "fake-toast",
  });

  const renderWithLevel = () => {
    switch (noti.level) {
      case "info":
        return <InboxIcon className="h-6 w-6 text-blue-400" aria-hidden="true" />
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
