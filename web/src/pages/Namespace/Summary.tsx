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

import "bytemd/dist/index.css";

import "./index.css";

import axios from "axios";
import gfm from "@bytemd/plugin-gfm";
import { Editor, Viewer } from "@bytemd/react";
import { Helmet, HelmetProvider } from "react-helmet-async";
import { useParams, useSearchParams } from "react-router-dom";
import { useEffect, useState } from "react";

import Header from "../../components/Header";
import IMenu from "../../components/Menu";
import NamespaceTabs from "../../components/NamespaceTabs";
import Notification from "../../components/Notification";
import { IHTTPError, INamespaceItem } from "../../interfaces";

import { Button } from "@/components/ui/button";

export default function ({ localServer }: { localServer: string }) {
  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repositoryId = parseInt(searchParams.get("repository_id") || "");
  const namespaceId = parseInt(searchParams.get("namespace_id") || "");

  const [, setNamespaceObj] = useState<INamespaceItem>({} as INamespaceItem);

  const [overview, setOverview] = useState("");
  const [overviewValid, setOverviewValid] = useState(true);
  useEffect(() => { setOverviewValid(overview?.length < 100000) }, [overview]);

  useEffect(() => {
    axios.get(localServer + `/api/v1/namespaces/${namespaceId}`).then(response => {
      if (response.status === 200) {
        const r = response.data as INamespaceItem;
        setNamespaceObj(r);
        setOverview(r.overview);
      } else if (response.status !== 404) {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
      }
    }).catch(error => {
      // A missing namespace means there is no overview yet, so stay silent.
      if (error.response?.status === 404) {
        return;
      }
      const errorcode = error.response?.data as IHTTPError;
      Notification({ level: "warning", title: errorcode?.title, message: errorcode?.description });
    });
  }, [namespace, repositoryId, localServer, namespaceId]);

  const [editorState, setEditorState] = useState(false);

  const updateNamespace = () => {
    if (!(overviewValid)) {
      Notification({ level: "warning", title: "Form validate failed", message: "Please check the field in the form." });
      return;
    }
    axios.put(localServer + `/api/v1/namespaces/${namespaceId}`, {
      overview: overview,
    } as INamespaceItem, {}).then(response => {
      if (response.status === 204) {
        Notification({ level: "info", title: "Success", message: "update overview success" });
        setEditorState(false);
      } else {
        const errorcode = response.data as IHTTPError;
        Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Notification({ level: "warning", title: errorcode.title, message: errorcode.description });
    })
  }

  return (
    <>
      <HelmetProvider>
        <Helmet>
          <title>sigma - Namespace Summary</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen max-h-screen flex overflow-hidden bg-white">
        <IMenu localServer={localServer} item="repositories" namespace={namespace} />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="relative z-0 focus:outline-none">
            <Header title="Repository"
              props={
                <NamespaceTabs namespace={namespace} namespaceId={Number.isNaN(namespaceId) ? "" : namespaceId.toString()} active="summary" />
              } />
          </main>
          <div className="flex flex-1 overflow-y-auto">
            <div className={(editorState ? "" : "pt-2 px-4") + " min-w-full min-h-full editor relative"} >
              {
                editorState ? (
                  <span></span>
                ) : (
                  <Button variant="outline" className="absolute right-4 top-2"
                    onClick={() => setEditorState(true)}
                  >Edit</Button>
                )
              }
              {
                editorState ? (
                  <Editor
                    placeholder='Write summary here with markdown'
                    value={overview}
                    plugins={[gfm()]}
                    onChange={e => setOverview(e)}
                  />
                ) : overview?.length === 0 ? (
                  <span className="text-gray-600">No description</span>
                ) : (
                  <Viewer plugins={[gfm()]} value={overview} />
                )
              }
            </div>
          </div>
          {
            editorState ? (
              <div className="flex items-center justify-end gap-2 border-t border-border bg-muted/40 px-4 py-3 sm:px-6">
                <Button variant="outline" onClick={() => setEditorState(false)}>Cancel</Button>
                <Button onClick={() => updateNamespace()}>Update</Button>
              </div>
            ) : (
              <div></div>
            )
          }
        </div>
      </div >
    </>
  )
}
