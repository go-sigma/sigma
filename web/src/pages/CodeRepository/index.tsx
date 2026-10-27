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

import axios from "axios";
import { KeyRound, RefreshCw, Trash2 } from "lucide-react";
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { useNavigate } from "react-router-dom";

import CredentialDialog from "@/components/CredentialDialog";
import Header from "@/components/Header";
import Menu from "@/components/Menu";
import Notification from "@/components/Notification";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useTranslation } from "@/i18n/useTranslation";
import {
  ICodeRepositoryProviderItem,
  ICodeRepositoryProviderList,
  IEndpoint,
  IHTTPError,
} from "@/interfaces";

function notifyError(error: any) {
  const errorcode = error?.response?.data as IHTTPError;
  Notification({
    level: "warning",
    title: errorcode?.title || "Request failed",
    message: errorcode?.description || "",
  });
}

export default function ({ localServer }: { localServer: string }) {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const [endpoint, setEndpoint] = useState("");
  const [providers, setProviders] = useState<ICodeRepositoryProviderItem[]>();
  const [refresh, setRefresh] = useState({});
  const [credentialProvider, setCredentialProvider] =
    useState<ICodeRepositoryProviderItem>();

  useEffect(() => {
    axios
      .get(`${localServer}/api/v1/systems/endpoint`)
      .then((response) => {
        if (response.status === 200) setEndpoint((response.data as IEndpoint).endpoint);
      })
      .catch(() => {});
  }, [localServer]);

  useEffect(() => {
    axios
      .get(`${localServer}/api/v1/coderepos/providers`)
      .then((response) => {
        if (response.status === 200)
          setProviders((response.data as ICodeRepositoryProviderList).items);
      })
      .catch(notifyError);
  }, [localServer, refresh]);

  const resync = (provider: string) => {
    axios
      .get(`${localServer}/api/v1/coderepos/${provider}/resync`)
      .then((response) => {
        if (response.status === 202) {
          Notification({
            level: "success",
            title: t("codeRepository.syncing"),
            message: "",
          });
          setTimeout(() => setRefresh({}), 300);
        }
      })
      .catch(notifyError);
  };

  const deleteCredential = (provider: string) => {
    axios
      .delete(`${localServer}/api/v1/coderepos/${provider}/credential`)
      .then((response) => {
        if (response.status === 204) {
          Notification({
            level: "success",
            title: t("codeRepository.credentialDeleted"),
            message: "",
          });
          setRefresh({});
        }
      })
      .catch(notifyError);
  };

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>{t("common.codeRepository")}</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-background">
        <Menu localServer={localServer} item="coderepos" />
        <div className="flex flex-col flex-1 max-h-screen">
          <main className="relative z-0 focus:outline-none" tabIndex={0}>
            <Header title={t("header.codeRepository")} />
          </main>
          <div className="px-4 py-4">
            <p className="text-sm text-muted-foreground">
              {t("codeRepository.description")}
            </p>
          </div>
          {providers?.length === 0 && (
            <div className="px-4 py-8 text-sm text-muted-foreground">
              {t("codeRepository.emptyProviders")}
            </div>
          )}
          <div className="px-4 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {providers?.map((provider) => (
              <div
                key={provider.provider}
                className="rounded-lg border bg-card p-4 flex flex-col gap-3"
              >
                <div className="flex items-center justify-between">
                  <div className="text-lg font-medium">{provider.display_name}</div>
                  <div className="flex items-center gap-2">
                    {provider.has_credential ? (
                      <Badge>{t("codeRepository.credentialConfigured")}</Badge>
                    ) : (
                      <Badge variant="secondary">
                        {t("codeRepository.credentialMissing")}
                      </Badge>
                    )}
                  </div>
                </div>
                <div className="text-xs text-muted-foreground break-all">
                  {provider.base_url}
                </div>
                <div className="text-xs flex flex-col gap-1">
                  <span>
                    {provider.linked
                      ? `${t("codeRepository.linkedAccount")}: ${provider.account_id || "-"}`
                      : t("codeRepository.notLinked")}
                  </span>
                  <span>
                    {t("codeRepository.lastSync")}:{" "}
                    {provider.cr_last_update_status || "-"}
                    {provider.cr_last_update_status === "Failed" &&
                    provider.cr_last_update_message
                      ? ` (${provider.cr_last_update_message})`
                      : ""}
                  </span>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <Button
                    size="sm"
                    variant={provider.has_credential ? "outline" : "default"}
                    onClick={() => setCredentialProvider(provider)}
                  >
                    <KeyRound data-icon="inline-start" />
                    {provider.has_credential
                      ? t("codeRepository.updateCredential")
                      : t("codeRepository.addCredential")}
                  </Button>
                  {provider.credential_type === "token" && (
                    <Tooltip>
                      <TooltipTrigger
                        render={
                          <Button size="sm" variant="outline" onClick={() => resync(provider.provider)} />
                        }
                      >
                        <RefreshCw data-icon="inline-start" />
                        {t("codeRepository.resync")}
                      </TooltipTrigger>
                      <TooltipContent>{t("codeRepository.resyncHint")}</TooltipContent>
                    </Tooltip>
                  )}
                  {provider.has_credential && (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => deleteCredential(provider.provider)}
                    >
                      <Trash2 data-icon="inline-start" />
                      {t("codeRepository.deleteCredential")}
                    </Button>
                  )}
                  {provider.linked && (
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => navigate(`/coderepos/${provider.provider}`)}
                    >
                      {t("codeRepository.viewRepositories")}
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
      {credentialProvider && (
        <CredentialDialog
          localServer={localServer}
          provider={credentialProvider}
          endpoint={endpoint}
          onClose={() => setCredentialProvider(undefined)}
          onSaved={() => {
            setCredentialProvider(undefined);
            setRefresh({});
          }}
        />
      )}
    </Fragment>
  );
}
