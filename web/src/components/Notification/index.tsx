/**
 * The MIT License (MIT)
 *
 * Copyright © 2023 Tosone <i@tosone.cn>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

import { toast } from 'react-toastify';
import { InboxIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline';

import { INotification } from "../../interfaces/interfaces";

import "./index.css";

export default function (noti: INotification) {
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
