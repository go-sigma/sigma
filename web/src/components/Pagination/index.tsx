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

import { useTranslation } from "../../i18n/useTranslation";
import {
  Pagination as ShadPagination,
  PaginationContent,
  PaginationItem,
  PaginationPrevious,
  PaginationNext,
} from "@/components/ui/pagination";

export default function ({ limit, page, total, setPage }: { limit: number, page: number, total: number, setPage: (page: number) => void }) {
  const { t } = useTranslation();
  const start = (page - 1) * limit + 1 > total ? total : (page - 1) * limit + 1;
  const end = total > page * limit ? page * limit : total;

  return (
    <div
      className="flex items-center justify-between border-gray-200 px-4 py-3 sm:px-6 border-t-0 bg-slate-100 dark:border-gray-800 dark:bg-gray-900"
      aria-label="Pagination"
    >
      <div>
        <p className="text-sm text-gray-700 dark:text-gray-200">
          {t("pagination.summary", { start, end, total })}
        </p>
      </div>
      <ShadPagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious
              text={t("common.previous")}
              onClick={(e: React.MouseEvent) => {
                e.preventDefault();
                if (page > 1) setPage(page - 1);
              }}
              className={page <= 1 ? "pointer-events-none opacity-50" : ""}
            />
          </PaginationItem>
          <PaginationItem>
            <PaginationNext
              text={t("common.next")}
              onClick={(e: React.MouseEvent) => {
                e.preventDefault();
                if (total / limit > page) setPage(page + 1);
              }}
              className={total / limit <= page ? "pointer-events-none opacity-50" : ""}
            />
          </PaginationItem>
        </PaginationContent>
      </ShadPagination>
    </div>
  )
}