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

import { Check, Copy, ExternalLink, X } from "lucide-react";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useTranslation } from "@/i18n/useTranslation";

import CardTitleWithHint from "./CardTitleWithHint";
import { mockOverview } from "./mock";

const copyResetDelay = 1500;

// Remembers that the user dismissed the card so it stays hidden across reloads.
const dismissedStorageKey = "home.quickstart.dismissed";

// QuickStart shows the three commands a user needs to push a first image, with
// the registry address derived from the API base URL of this instance.
export default function QuickStart({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState<string>();
  const [dismissed, setDismissed] = useState(() => localStorage.getItem(dismissedStorageKey) === "1");

  const endpoint = useMemo(() => {
    const base = localServer === "" ? window.location.origin : localServer;
    return base.replace(/^https?:\/\//, "").replace(/\/+$/, "");
  }, [localServer]);

  const steps = useMemo(() => {
    const image = `${endpoint}/library/myapp:latest`;
    return [
      { key: "login", label: t("home.quickStartLogin"), command: `docker login ${endpoint}` },
      { key: "tag", label: t("home.quickStartTag"), command: `docker tag myapp:latest ${image}` },
      { key: "push", label: t("home.quickStartPush"), command: `docker push ${image}` },
    ];
  }, [endpoint, t]);

  const copy = async (key: string, command: string) => {
    try {
      await navigator.clipboard.writeText(command);
    } catch {
      // Clipboard access is denied outside a secure context; the command stays readable.
      return;
    }
    setCopied(key);
    window.setTimeout(() => {
      setCopied(current => (current === key ? undefined : current));
    }, copyResetDelay);
  };

  const dismiss = () => {
    localStorage.setItem(dismissedStorageKey, "1");
    setDismissed(true);
  };

  if (dismissed) {
    return null;
  }

  return (
    <Card className="relative gap-0 overflow-hidden py-0">
      <Button
        variant="ghost"
        size="icon-xs"
        className="absolute right-2 top-2 z-10 text-muted-foreground"
        aria-label={t("home.quickStartDismiss")}
        onClick={dismiss}
      >
        <X />
      </Button>
      <div className="grid lg:grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)]">
        <div className="min-w-0 p-5 sm:p-6">
          <CardTitleWithHint title={t("home.quickStart")} hint={t("home.quickStartDescription")} />

          <ol className="mt-[0.8rem] space-y-2">
            {steps.map((step, index) => (
              <li
                key={step.key}
                className="group flex items-center gap-3 rounded-lg bg-muted/50 px-3 py-2"
              >
                <span className="flex size-5 shrink-0 items-center justify-center rounded-full bg-background text-[10px] font-medium text-muted-foreground ring-1 ring-foreground/10">
                  {index + 1}
                </span>
                <code className="min-w-0 flex-1 overflow-x-auto whitespace-nowrap font-mono text-xs text-foreground">
                  {step.command}
                </code>
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100"
                  aria-label={`${t("home.quickStartCopy")}: ${step.label}`}
                  onClick={() => copy(step.key, step.command)}
                >
                  {copied === step.key ? (
                    <Check className="text-green-600 dark:text-green-500" />
                  ) : (
                    <Copy />
                  )}
                </Button>
              </li>
            ))}
          </ol>

          <p className="mt-3 text-xs text-muted-foreground">{t("home.quickStartHint")}</p>
        </div>

        <div className="flex flex-col justify-between gap-5 border-t border-border bg-muted/30 p-5 sm:p-6 lg:border-l lg:border-t-0">
          <dl className="space-y-3">
            <div>
              <dt className="text-xs text-muted-foreground">{t("home.version")}</dt>
              <dd className="mt-1 flex items-baseline gap-2">
                <span className="text-sm font-medium tabular-nums text-foreground">
                  {mockOverview.version}
                </span>
                <code className="font-mono text-xs text-muted-foreground">{mockOverview.commit}</code>
              </dd>
            </div>
            <div className="min-w-0">
              <dt className="text-xs text-muted-foreground">{t("home.endpoint")}</dt>
              <dd className="mt-1 truncate font-mono text-xs text-foreground">{endpoint}</dd>
            </div>
          </dl>

          <nav className="flex flex-col gap-0.5 text-sm">
            <a
              className="group inline-flex items-center justify-between gap-2 rounded-md py-1 text-muted-foreground transition-colors hover:text-foreground"
              href="https://docs.sigma.tosone.cn/quickstart"
              target="_blank"
              rel="noreferrer"
            >
              {t("home.docs")}
              <ExternalLink className="size-3.5 opacity-60 transition-opacity group-hover:opacity-100" />
            </a>
            <a
              className="group inline-flex items-center justify-between gap-2 rounded-md py-1 text-muted-foreground transition-colors hover:text-foreground"
              href="https://github.com/go-sigma/sigma"
              target="_blank"
              rel="noreferrer"
            >
              {t("home.source")}
              <ExternalLink className="size-3.5 opacity-60 transition-opacity group-hover:opacity-100" />
            </a>
          </nav>
        </div>
      </div>
    </Card>
  );
}
