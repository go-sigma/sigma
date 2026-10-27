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
import { useEffect, useState } from "react";
import { useParams, useSearchParams } from "react-router-dom";

import { INamespaceList } from "@/interfaces";

// namespacesPerPage is the page size used to look a namespace up by name; the API caps at 100.
const namespacesPerPage = 100;

// useNamespaceId resolves the namespace of the current route.
//
// The namespace_id query parameter is preferred. When it is missing — a bookmarked or hand
// written URL such as /namespaces/library/members — the id is looked up from the namespace name
// in the route, or from fallbackNamespace for routes that carry the name only in the query.
// It returns an empty string until the id is known; callers must skip requests while it is empty.
export function useNamespaceId(localServer: string, fallbackNamespace = ""): string {
  const params = useParams<{ namespace?: string }>();
  const namespace = params.namespace || fallbackNamespace;
  const [searchParams] = useSearchParams();
  const fromQuery = searchParams.get("namespace_id") || "";
  const [lookedUp, setLookedUp] = useState<{ namespace: string; id: string }>({ namespace: "", id: "" });

  useEffect(() => {
    if (fromQuery !== "" || !namespace) {
      return;
    }
    axios
      .get(`${localServer}/api/v1/namespaces/?limit=${namespacesPerPage}&name=${namespace}`)
      .then((response) => {
        if (response.status !== 200) {
          return;
        }
        const match = (response.data as INamespaceList).items?.find(item => item.name === namespace);
        if (match) {
          setLookedUp({ namespace, id: match.id });
        }
      })
      .catch(() => {});
  }, [fromQuery, namespace, localServer]);

  if (fromQuery !== "") {
    return fromQuery;
  }
  // Ignore a result that belongs to another namespace than the one in the route.
  return lookedUp.namespace === namespace ? lookedUp.id : "";
}
