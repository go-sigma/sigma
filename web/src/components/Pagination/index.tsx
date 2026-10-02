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
  const pageCount = Math.max(1, Math.ceil(total / limit));
  const hasPrevious = page > 1;
  const hasNext = total / limit > page;

  return (
    <div
      className="flex flex-wrap items-center justify-between gap-3 border-t border-border bg-muted/40 px-4 py-3"
      aria-label="Pagination"
    >
      <p className="text-sm text-muted-foreground">
        {t("pagination.summary", { start, end, total })}
      </p>
      <ShadPagination className="mx-0 w-auto">
        <PaginationContent className="gap-2">
          <PaginationItem>
            <PaginationPrevious
              text={t("common.previous")}
              aria-disabled={!hasPrevious}
              onClick={(e: React.MouseEvent) => {
                e.preventDefault();
                if (hasPrevious) setPage(page - 1);
              }}
              className={hasPrevious ? "cursor-pointer" : "cursor-not-allowed opacity-50 hover:bg-transparent"}
            />
          </PaginationItem>
          {total > 0 && (
            <PaginationItem>
              <span className="min-w-14 text-center text-sm tabular-nums text-muted-foreground">
                {page} / {pageCount}
              </span>
            </PaginationItem>
          )}
          <PaginationItem>
            <PaginationNext
              text={t("common.next")}
              aria-disabled={!hasNext}
              onClick={(e: React.MouseEvent) => {
                e.preventDefault();
                if (hasNext) setPage(page + 1);
              }}
              className={hasNext ? "cursor-pointer" : "cursor-not-allowed opacity-50 hover:bg-transparent"}
            />
          </PaginationItem>
        </PaginationContent>
      </ShadPagination>
    </div>
  )
}
