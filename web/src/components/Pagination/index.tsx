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

export default function ({ limit, page, total, setPage }: { limit: number, page: number, total: number, setPage: (page: number) => void }) {
  return (
    <div
      className="flex flex-2 items-center justify-between border-gray-200 px-4 py-3 sm:px-6 border-t-0 bg-slate-100"
      aria-label="Pagination"
    >
      <div>
        <p className="text-sm text-gray-700">
          Showing <span className="font-medium">{(page - 1) * limit + 1 > total ? total : (page - 1) * limit + 1}</span> to <span className="font-medium">{total > page * limit ? page * limit : total}</span> of{' '}
          <span className="font-medium">{total}</span> results
        </p>
      </div>
      <div className="flex flex-1 justify-between sm:justify-end">
        <button
          className="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          onClick={() => {
            if (page <= 1) {
              return;
            } else {
              setPage(page - 1)
            }
          }}
        >
          Previous
        </button>
        <button
          className="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          onClick={() => {
            if (total / limit < page) {
              return;
            } else {
              setPage(page + 1)
            }
          }}
        >
          Next
        </button>
      </div>
    </div>
  )
}
