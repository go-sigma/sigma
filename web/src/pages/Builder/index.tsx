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
import { lazy, Suspense, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

import Notification from "@/components/Notification";
import { IHTTPError, IRepositoryItem } from "@/interfaces";

const RunnerList = lazy(() => import("@/pages/Builder/RunnerList"));
const Setup = lazy(() => import("@/pages/Builder/Setup"));

// RepositoryBuilder is the builder tab of a repository. A repository without a builder has
// nothing to run yet, so the tab opens the setup form; once a builder exists the tab shows its
// runs instead of making the user pick one of two similar tabs.
export default function RepositoryBuilder({ localServer }: { localServer: string }) {
  const [searchParams] = useSearchParams();
  const repositoryId = searchParams.get("repository_id") || "";
  const namespaceId = searchParams.get("namespace_id") || "";

  const [hasBuilder, setHasBuilder] = useState<boolean>();

  useEffect(() => {
    if (repositoryId === "" || namespaceId === "") {
      return;
    }
    axios.get(`${localServer}/api/v1/namespaces/${namespaceId}/repositories/${repositoryId}`).then(response => {
      if (response?.status === 200) {
        const repository = response.data as IRepositoryItem;
        setHasBuilder(Boolean(repository.builder?.id));
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  }, [localServer, namespaceId, repositoryId]);

  if (hasBuilder === undefined) {
    // Without a repository in the URL there is nothing to look up yet: the form picks one.
    return repositoryId === "" || namespaceId === ""
      ? <Suspense fallback={null}><Setup localServer={localServer} embedded /></Suspense>
      : null;
  }

  return (
    <Suspense fallback={null}>
      {hasBuilder
        ? <RunnerList localServer={localServer} />
        : <Setup localServer={localServer} embedded onConfigured={() => setHasBuilder(true)} />}
    </Suspense>
  );
}
