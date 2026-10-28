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

import { Link } from "react-router-dom";

import { useTranslation } from "@/i18n/useTranslation";

// RepositoryTab names the section of a repository page that is currently shown.
export type RepositoryTab = "summary" | "builder" | "tags";

const activeTabClass = "z-10 inline-flex items-center border-b border-primary px-1 pt-1 text-sm font-medium text-foreground capitalize cursor-pointer";
const idleTabClass = "inline-flex items-center border-b border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 capitalize";

type RepositoryTabsProps = {
  namespace: string;
  namespaceId: string;
  repository: string;
  repositoryId: string;
  active: RepositoryTab;
  // showBuilder hides the builder tab when the builder daemon is disabled.
  showBuilder?: boolean;
};

// RepositoryTabs is the tab bar shared by the pages of a single repository. The builder and the
// run list are one page: it shows the setup form until a builder exists and the runs afterwards.
export default function RepositoryTabs({ namespace, namespaceId, repository, repositoryId, active, showBuilder = true }: RepositoryTabsProps) {
  const { t } = useTranslation();

  const base = `/namespaces/${namespace}/repository`;
  const query = `repository=${repository}&repository_id=${repositoryId}&namespace_id=${namespaceId}`;

  const tabs: { key: RepositoryTab; label: string; to: string }[] = [
    { key: "summary", label: t("common.summary"), to: `${base}/summary?${query}` },
    { key: "tags", label: t("common.tagList"), to: `${base}/tags?${query}` },
    ...(showBuilder ? [{
      key: "builder" as RepositoryTab,
      label: t("common.builder"),
      to: `${base}/builder?namespace=${namespace}&namespace_id=${namespaceId}&repository=${repository}&repository_id=${repositoryId}&namespace_stick=true&repository_stick=true`,
    }] : []),
  ];

  return (
    <div className="flex space-x-8">
      {tabs.map(tab => (
        tab.key === active ? (
          <span key={tab.key} className={activeTabClass}>{tab.label}</span>
        ) : (
          <Link key={tab.key} to={tab.to} className={idleTabClass}>{tab.label}</Link>
        )
      ))}
    </div>
  );
}
