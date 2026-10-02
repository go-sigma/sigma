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

import { Link } from "react-router-dom";

import { useTranslation } from "../../i18n/useTranslation";
import { MessageKey } from "../../i18n/types";

import { cn } from "@/lib/utils";

export type NamespaceSection = "summary" | "repositories" | "members" | "daemon-tasks" | "webhooks";

const sections: { key: NamespaceSection; path: string; labelKey: MessageKey }[] = [
  { key: "summary", path: "namespace-summary", labelKey: "common.summary" },
  { key: "repositories", path: "repositories", labelKey: "common.repositoryList" },
  { key: "members", path: "members", labelKey: "common.members" },
  { key: "daemon-tasks", path: "daemon-tasks", labelKey: "common.daemonTask" },
  { key: "webhooks", path: "webhooks", labelKey: "common.webhook" },
];

export default function NamespaceTabs({ namespace, namespaceId, active }: {
  namespace?: string;
  namespaceId?: string;
  active: NamespaceSection;
}) {
  const { t } = useTranslation();
  const query = namespaceId ? `?namespace_id=${namespaceId}` : "";

  return (
    <nav className="flex h-full items-stretch gap-1 overflow-x-auto" aria-label={t("common.namespaces")}>
      {sections.map(section => {
        const isActive = section.key === active;
        const className = cn(
          "relative inline-flex cursor-pointer items-center whitespace-nowrap border-b-2 px-3 text-sm transition-colors",
          isActive
            ? "border-foreground text-foreground"
            : "border-transparent text-muted-foreground hover:border-border hover:text-foreground",
        );

        if (isActive) {
          return (
            <span key={section.key} className={className} aria-current="page">
              {t(section.labelKey)}
            </span>
          );
        }

        return (
          <Link key={section.key} to={`/namespaces/${namespace}/${section.path}${query}`} className={className}>
            {t(section.labelKey)}
          </Link>
        );
      })}
    </nav>
  );
}
