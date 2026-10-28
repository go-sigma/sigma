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
import { useState } from "react";

import Notification from "@/components/Notification";
import { Button } from "@/components/ui/button";
import {
  Dialog as ShadDialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useTranslation } from "@/i18n/useTranslation";
import {
  ICodeRepositoryProviderItem,
  IHTTPError,
  IOauth2ClientID,
} from "@/interfaces";

function notifyError(error: any) {
  const errorcode = error?.response?.data as IHTTPError;
  Notification({
    level: "warning",
    title: errorcode?.title || "Request failed",
    message: errorcode?.description || "",
  });
}

export default function ({
  localServer,
  provider,
  endpoint,
  onClose,
  onSaved,
}: {
  localServer: string;
  provider: ICodeRepositoryProviderItem;
  endpoint: string;
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useTranslation();

  const [token, setToken] = useState("");
  const [saving, setSaving] = useState(false);
  const [clientID, setClientID] = useState("");

  const submit = () => {
    setSaving(true);
    axios
      .put(
        `${localServer}/api/v1/coderepos/${provider.provider}/credential`,
        // Listing repositories goes through the provider API, and only a token credential
        // can authenticate those calls, so this dialog only stores tokens.
        { type: "token", token },
      )
      .then((response) => {
        if (response.status === 200) {
          Notification({
            level: "success",
            title: t("codeRepository.credentialSaved"),
            message: "",
          });
          onSaved();
        }
      })
      .catch(notifyError)
      .finally(() => setSaving(false));
  };

  // OAuth binding is an alternative to a personal access token: it links the account, but
  // the issued token is not kept, so a token credential is still required for syncing.
  const loadClientID = () => {
    if (clientID !== "") return;
    axios
      .get(`${localServer}/api/v1/oauth2/${provider.provider}/client_id`)
      .then((response) => {
        if (response.status === 200)
          setClientID((response.data as IOauth2ClientID).client_id);
      })
      .catch(notifyError);
  };

  return (
    <ShadDialog
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {t("codeRepository.credentialTitle")} · {provider.display_name}
          </DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label>
              <span className="text-destructive">*</span>
              {t("codeRepository.credentialToken")}
            </Label>
            <Input
              type="password"
              placeholder={t("codeRepository.tokenPlaceholder")}
              value={token}
              onChange={(e) => setToken(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              {t("codeRepository.tokenHint")}
            </p>
          </div>

          {provider.oauth_enabled && !provider.linked && (
            <div className="rounded-md border p-3 text-xs flex flex-col gap-2">
              <span>{t("codeRepository.oauthLinkHint")}</span>
              <a
                className="underline"
                href={`${
                  provider.provider === "github"
                    ? "https://github.com/login/oauth/authorize"
                    : `${provider.base_url}/oauth/authorize`
                }?client_id=${clientID}&redirect_uri=${encodeURIComponent(
                  `${endpoint}/api/v1/oauth2/${provider.provider}/redirect_callback?endpoint=${encodeURIComponent(
                    location.protocol + "//" + location.host,
                  )}`,
                )}&scope=repo`}
                onMouseEnter={loadClientID}
                onClick={loadClientID}
              >
                {t("codeRepository.oauthLink")}
              </a>
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {t("common.cancel")}
          </Button>
          <Button disabled={saving} onClick={submit}>
            {t("common.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </ShadDialog>
  );
}
