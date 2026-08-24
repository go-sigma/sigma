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

import { toast } from 'sonner';

import { INotification } from "../../interfaces";

export default function (noti: INotification) {
  switch (noti.level) {
    case "success":
      toast.success(noti.title, { description: noti.message, duration: noti.autoClose || 3000 });
      break;
    case "warning":
      toast.warning(noti.title, { description: noti.message, duration: noti.autoClose || 3000 });
      break;
    case "info":
      toast.info(noti.title, { description: noti.message, duration: noti.autoClose || 3000 });
      break;
    default:
      toast(noti.title, { description: noti.message, duration: noti.autoClose || 3000 });
      break;
  }
}