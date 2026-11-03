/**
 * Copyright 2026 sigma
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

import { Info } from "lucide-react";

import { CardTitle } from "@/components/ui/card";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

type CardTitleWithHintProps = {
  title: string;
  hint: string;
};

// CardTitleWithHint renders a card title with an info icon next to it: the explanation of the
// card lives in the tooltip, which keeps a dashboard full of cards free of description lines.
export default function CardTitleWithHint({ title, hint }: CardTitleWithHintProps) {
  return (
    <div className="flex items-center gap-1.5">
      <CardTitle>{title}</CardTitle>
      <Tooltip>
        <TooltipTrigger className="inline-flex shrink-0 text-muted-foreground transition-colors hover:text-foreground">
          <Info className="size-3.5" aria-hidden="true" />
          <span className="sr-only">{hint}</span>
        </TooltipTrigger>
        <TooltipContent>{hint}</TooltipContent>
      </Tooltip>
    </div>
  );
}
