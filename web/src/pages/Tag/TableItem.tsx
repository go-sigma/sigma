/**
 * Copyright 2023 XImager
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
import dayjs from 'dayjs';
import humanFormat from "human-format";
import { useNavigate } from 'react-router-dom';
import { useState, Fragment, useEffect } from "react";
import relativeTime from 'dayjs/plugin/relativeTime';
import { Dialog, Menu, Transition } from "@headlessui/react";
import { EllipsisVerticalIcon } from "@heroicons/react/20/solid";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";

import Toast from "../../components/Notification";
import { ITag, IHTTPError, IArtifact } from "../../interfaces";

dayjs.extend(relativeTime);

export default function ({ localServer, namespace, repository, artifactDigest, artifact }: { localServer: string, namespace: string, repository: string, artifactDigest: string, artifact: string }) {
  const artifactObj = JSON.parse(artifact);

  return (
    <tbody>
      {
        artifactObj.mediaType === "application/vnd.oci.image.manifest.v1+json" || artifactObj.mediaType === "application/vnd.docker.distribution.manifest.v2+json" ? (
          <DetailItem localServer={localServer} namespace={namespace} repository={repository} digest={artifactDigest} mediaType={artifactObj.mediaType} />
        ) : artifactObj.mediaType === "application/vnd.docker.distribution.manifest.list.v2+json" || artifactObj.mediaType === "application/vnd.oci.image.index.v1+json" ? (
          artifactObj.manifests.map((manifest: { digest: string, annotations?: any }, index: number) => {
            return (
              manifest.annotations?.["vnd.docker.reference.type"] === "attestation-manifest" ? (
                <tr key={index} className="hidden"></tr>
              ) : (
                <DetailItem key={index} localServer={localServer} namespace={namespace} repository={repository} digest={manifest.digest} mediaType={artifactObj.mediaType} />
              )
            )
          })
        ) : (
          <tr></tr>
        )
      }
    </tbody >
  );
}

function DetailItem({ localServer, namespace, repository, digest, mediaType }: { localServer: string, namespace: string, repository: string, digest: string, mediaType: string }) {
  const [artifact, setArtifact] = useState<IArtifact>({} as IArtifact);
  const fetchArtifact = () => {
    let url = localServer + `/api/v1/namespaces/${namespace}/artifacts/${digest}?repository=${repository}`;
    axios.get(url).then(response => {
      if (response?.status === 200) {
        setArtifact(response.data as IArtifact)
      } else {
        const errorcode = response.data as IHTTPError;
        Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
      }
    }).catch(error => {
      const errorcode = error.response.data as IHTTPError;
      Toast({ level: "warning", title: errorcode.title, message: errorcode.description });
    });
  }
  useEffect(fetchArtifact, []);
  const cutDigest = (digest: string) => {
    if (digest === undefined) {
      return "";
    }
    if (digest.indexOf(":") < 0) {
      return "";
    }
    return digest.substring(digest.indexOf(":") + 1, digest.indexOf(":") + 13);
  }
  const configObj = JSON.parse(artifact.config_raw === undefined ? "{}" : artifact.config_raw)
  return (
    <tr className="hover:bg-gray-100">
      <td className="px-2 text-left cursor-pointer">
        <code className="text-xs underline underline-offset-1 text-blue-600 hover:text-blue-500">
          {cutDigest(artifact.digest)}
        </code>
      </td>
      <td className="text-right text-xs">
        {
          (mediaType === "application/vnd.oci.image.manifest.v1+json" || mediaType === "application/vnd.docker.distribution.manifest.v2+json") ? (
            <span>{configObj?.os}/{configObj?.architecture}</span>
          ) : (mediaType === "application/vnd.docker.distribution.manifest.list.v2+json" || mediaType === "application/vnd.oci.image.index.v1+json") ? (
            <span></span>
          ) : (
            <span></span>
          )
        }
      </td>
      <td className="text-right text-xs">
        {humanFormat(artifact.blob_size || 0)}
      </td>
      <td className="text-right text-xs">
        {(artifact.pull_times || 0) > 0 ? dayjs().to(dayjs(artifact.last_pull)) : "Never pulled"}
      </td>
      <td className="text-right text-xs">
        {artifact.pull_times}
      </td>
      <td className="text-right text-xs">
        {dayjs().to(dayjs(artifact.pushed_at))}
      </td>
      <td className="px-2 text-right text-xs">
        {humanFormat(artifact.blob_size || 0)}
      </td>
    </tr>
  );
}
