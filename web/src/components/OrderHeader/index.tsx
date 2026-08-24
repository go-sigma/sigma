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

import { ArrowUp, ArrowDown, ArrowUpDown } from "lucide-react";
import { IOrder } from "../../interfaces";

export default function ({ text, orderStatus, setOrder }: { text: string, orderStatus: IOrder, setOrder: (order: IOrder) => void }) {
  return (
    <span className="whitespace-nowrap relative cursor-pointer select-none inline-flex items-center gap-1" onClick={() => {
      switch (orderStatus) {
        case IOrder.Asc:
          setOrder(IOrder.Desc);
          break;
        case IOrder.Desc:
          setOrder(IOrder.None);
          break;
        default:
          setOrder(IOrder.Asc);
          break;
      }
    }}>
      {text}
      {orderStatus === IOrder.Asc ? (
        <ArrowUp className="h-3 w-3 text-gray-500" />
      ) : orderStatus === IOrder.Desc ? (
        <ArrowDown className="h-3 w-3 text-gray-500" />
      ) : (
        <ArrowUpDown className="h-3 w-3 text-gray-300" />
      )}
    </span>
  );
}