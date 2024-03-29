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

import { ReactNode } from 'react';

export default function Header({ title, breadcrumb, props }: { title: string, breadcrumb?: ReactNode, props?: ReactNode }) {
  return (
    <div className={breadcrumb === undefined ? "" : "border-gray-200 border-b"}>
      <div className="border-gray-200 border-b px-4 py-0 sm:flex sm:items-center sm:justify-between sm:px-6 lg:px-8 h-16">
        <div className="flex-1 min-w-0 my-4">
          <h1 className="text-lg font-medium leading-6 text-gray-900 sm:truncate mr-8">{title}</h1>
        </div>
        <div className="flex h-16">
          {props}
        </div>
      </div>
      <div className={breadcrumb === undefined ? "" : "px-4 py-3"}>
        {breadcrumb}
      </div>
    </div>
  );
}
