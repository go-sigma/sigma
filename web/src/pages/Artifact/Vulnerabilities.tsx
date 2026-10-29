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

import { useTranslation } from "@/i18n/useTranslation";
import { IArtifactVulnerabilities, IArtifactVulnerability } from "@/interfaces";

// severityWeight orders the severities the way the table sorts them: the most severe first.
function severityWeight(severity: string) {
  switch (severity.toLowerCase()) {
    case "critical":
      return 5;
    case "high":
      return 4;
    case "medium":
      return 3;
    case "low":
      return 2;
    case "negligible":
      return 1;
    default:
      return 0;
  }
}

// severityClass colors a severity badge the way the tag list colors its counters.
function severityClass(severity: string) {
  switch (severity.toLowerCase()) {
    case "critical":
      return "bg-red-800 text-white dark:bg-red-900 dark:text-red-200";
    case "high":
      return "bg-red-300 text-gray-800 dark:bg-red-900 dark:text-red-300";
    case "medium":
      return "bg-amber-400 text-gray-800 dark:bg-amber-900 dark:text-amber-300";
    case "low":
      return "bg-amber-200 text-gray-800 dark:bg-amber-900/70 dark:text-amber-200";
    case "negligible":
      return "bg-gray-200 text-gray-700 dark:bg-gray-800 dark:text-gray-300";
    default:
      return "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400";
  }
}

// sortVulnerabilities orders the rows by severity first and, inside a severity, by the date the fix
// was published descending. Rows without a fix date come last, and the id keeps the order stable.
export function sortVulnerabilities(items: IArtifactVulnerability[]) {
  return items.toSorted((left, right) => {
    const bySeverity = severityWeight(right.severity) - severityWeight(left.severity);
    if (bySeverity !== 0) {
      return bySeverity;
    }
    if (left.fix_date !== right.fix_date) {
      return (right.fix_date ?? "").localeCompare(left.fix_date ?? "");
    }
    return left.id.localeCompare(right.id);
  });
}

type VulnerabilitiesProps = {
  report?: IArtifactVulnerabilities;
  // rows are the vulnerabilities of the current page. The page sorts and pages the report, so the
  // pagination can sit at the bottom of the page instead of inside this table.
  rows: IArtifactVulnerability[];
  // layerIndex maps a layer digest to its position in the manifest, so a row can name the layer that
  // introduced the vulnerable package.
  layerIndex: Map<string, number>;
};

// Vulnerabilities lists the vulnerabilities grype reported for an artifact. The API answers without
// pagination, so the caller sorts and pages the rows.
export default function Vulnerabilities({ report, rows, layerIndex }: VulnerabilitiesProps) {
  const { t } = useTranslation();

  if (report === undefined) {
    return null;
  }

  return (
    <div>
      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
          <span className="font-medium text-foreground">{t("artifact.scanStatus")}:</span>
          <span>{report.status === "" ? "-" : report.status}</span>
          {report.scanned_at === "" ? null : (
            <span>
              {t("artifact.scannedAt")} {report.scanned_at}
            </span>
          )}
          {report.db_built_at === "" ? null : (
            <span>
              {t("artifact.dbBuiltAt")} {report.db_built_at}
            </span>
          )}
        </div>
        <div className="flex flex-wrap items-center gap-1">
          <span className={`px-2 py-0.5 text-xs font-medium ${severityClass("critical")}`}>
            {report.critical} {t("artifact.criticalShort")}
          </span>
          <span className={`px-2 py-0.5 text-xs font-medium ${severityClass("high")}`}>
            {report.high} {t("artifact.highShort")}
          </span>
          <span className={`px-2 py-0.5 text-xs font-medium ${severityClass("medium")}`}>
            {report.medium} {t("artifact.mediumShort")}
          </span>
          <span className={`px-2 py-0.5 text-xs font-medium ${severityClass("low")}`}>
            {report.low} {t("artifact.lowShort")}
          </span>
          {report.negligible === 0 ? null : (
            <span className={`px-2 py-0.5 text-xs font-medium ${severityClass("negligible")}`}>
              {report.negligible} {t("artifact.negligibleShort")}
            </span>
          )}
          {report.unknown === 0 ? null : (
            <span className={`px-2 py-0.5 text-xs font-medium ${severityClass("unknown")}`}>
              {report.unknown} {t("artifact.unknownShort")}
            </span>
          )}
        </div>
      </div>

      {report.status !== "Success" ? (
        <p className="rounded-lg border border-border bg-card px-4 py-6 text-sm text-muted-foreground">
          {t("artifact.scanPending")}
        </p>
      ) : report.total === 0 ? (
        <p className="rounded-lg border border-border bg-card px-4 py-6 text-sm text-muted-foreground">
          {t("artifact.noVulnerabilities")}
        </p>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-border">
          <table className="min-w-full table-fixed">
            <thead>
              <tr className="bg-muted/40">
                <th className="w-[110px] px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {t("artifact.severity")}
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {t("artifact.vulnerability")}
                </th>
                <th className="w-[40px] px-2 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {t("artifact.package")}
                </th>
                <th className="w-[40px] px-2 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {t("artifact.fix")}
                </th>
                <th className="w-[88px] px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {t("artifact.cvss")}
                </th>
                <th className="w-[104px] px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {t("artifact.fixDate")}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {rows.map(vulnerability => (
                <VulnerabilityRow key={`${vulnerability.id}-${vulnerability.package}-${vulnerability.package_version}`} vulnerability={vulnerability} layerIndex={layerIndex} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function VulnerabilityRow({ vulnerability, layerIndex }: { vulnerability: IArtifactVulnerability; layerIndex: Map<string, number> }) {
  const { t } = useTranslation();

  const link = vulnerability.urls?.find(url => url !== "");
  const layer = (vulnerability.layers ?? [])
    .map(digest => layerIndex.get(digest))
    .find(index => index !== undefined);

  return (
    <tr className="align-top">
      <td className="px-4 py-3">
        <span className={`px-2 py-0.5 text-xs font-medium ${severityClass(vulnerability.severity)}`}>
          {vulnerability.severity === "" ? "Unknown" : vulnerability.severity}
        </span>
        {vulnerability.known_exploited === "" ? null : (
          <span className="mt-1 block text-[11px] font-medium text-red-600 dark:text-red-400">
            {t("artifact.knownExploited")}
          </span>
        )}
      </td>
      <td className="px-4 py-3">
        {link === undefined ? (
          <span className="font-mono text-xs">{vulnerability.id}</span>
        ) : (
          <a className="font-mono text-xs text-primary underline underline-offset-2 hover:text-primary/80" href={link} target="_blank" rel="noreferrer">
            {vulnerability.id}
          </a>
        )}
        {vulnerability.description === "" ? null : (
          <div className="mt-1 text-xs text-muted-foreground line-clamp-2" title={vulnerability.description}>
            {vulnerability.description}
          </div>
        )}
        {vulnerability.cwes.length === 0 ? null : (
          <div className="mt-1 text-[11px] text-muted-foreground">{vulnerability.cwes.join(", ")}</div>
        )}
      </td>
      <td className="min-w-0 px-2 py-3">
        <div className="truncate font-mono text-xs" title={vulnerability.package}>{vulnerability.package}</div>
        <div className="truncate font-mono text-[11px] text-muted-foreground" title={vulnerability.package_version}>
          {vulnerability.package_version}
        </div>
        <div className="truncate text-[11px] text-muted-foreground">
          {vulnerability.package_type === "" ? vulnerability.package_language : vulnerability.package_type}
        </div>
        {layer === undefined ? null : (
          <div className="mt-1 truncate text-[11px] text-muted-foreground" title={t("artifact.layerIntroduced", { layer })}>
            {t("artifact.layerIntroduced", { layer })}
          </div>
        )}
      </td>
      <td className="min-w-0 px-2 py-3">
        {vulnerability.fix_versions.length === 0 ? (
          <div className="truncate text-xs text-muted-foreground" title={t("artifact.notFixed")}>{t("artifact.notFixed")}</div>
        ) : (
          <>
            <div className="truncate text-xs font-medium text-green-600 dark:text-green-500" title={t("artifact.fixedVersion")}>
              {t("artifact.fixedVersion")}
            </div>
            <div className="truncate font-mono text-xs" title={vulnerability.fix_versions.join(", ")}>
              {vulnerability.fix_versions.join(", ")}
            </div>
          </>
        )}
        {vulnerability.suggested_version === "" ? null : (
          <div className="mt-1 truncate font-mono text-[11px] text-muted-foreground" title={vulnerability.suggested_version}>
            {vulnerability.suggested_version}
          </div>
        )}
      </td>
      <td className="px-4 py-3 text-right">
        <div className="text-xs tabular-nums">{vulnerability.cvss_score === 0 ? "-" : vulnerability.cvss_score.toFixed(1)}</div>
        {vulnerability.epss_score === 0 ? null : (
          <div className="text-[11px] text-muted-foreground" title={t("artifact.epss")}>
            {t("artifact.epss")} {(vulnerability.epss_score * 100).toFixed(1)}%
          </div>
        )}
      </td>
      <td className="whitespace-nowrap px-4 py-3 text-right text-xs tabular-nums">
        {vulnerability.fix_date === "" ? "-" : vulnerability.fix_date}
      </td>
    </tr>
  );
}
