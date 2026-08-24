/**
 * Copyright 2024 sigma
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

import { EllipsisVerticalIcon } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

interface ITableItemDropdownProps {
  index?: number;
  items?: ITableItemDropdownItem[];
}

interface ITableItemDropdownItem {
  name: string;
  onClick: () => void;
  warn?: boolean;
  disable?: boolean;
}

export default function ({ index, items }: ITableItemDropdownProps) {
  return items && items.length > 0 && !items.every(item => item.disable) ? (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button variant="ghost" size="icon" className="h-8 w-8" />}>
          <span className="sr-only">Open options</span>
          <EllipsisVerticalIcon className="h-4 w-4" />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-20">
        {items.map((item, idx) =>
          item.disable ? null : (
            <DropdownMenuItem
              key={idx}
              variant={item.warn ? "destructive" : "default"}
              onClick={item.onClick}
            >
              {item.name}
            </DropdownMenuItem>
          )
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  ) : null;
}