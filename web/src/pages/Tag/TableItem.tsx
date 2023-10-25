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

import dayjs from 'dayjs';
import humanFormat from "human-format";

import distros, { distroName } from '../../utils/distros';
import { IArtifact, IVuln, ISbom, IImageConfig } from "../../interfaces";

function skipManifest(raw: string) {
  let artifactObj = JSON.parse(raw);
  if (artifactObj["config"]["mediaType"] === "application/vnd.oci.image.config.v1+json") {
    if (artifactObj["layers"].length === 1 && artifactObj["layers"][0]["mediaType"] === "application/vnd.in-toto+json" && artifactObj["layers"][0]["annotations"]["in-toto.io/predicate-type"] !== "") {
      return true;
    }
  }
  return false;
}

export default function ({ namespace, repository, artifact, artifacts }: { namespace: string, repository: string, artifact: IArtifact, artifacts: IArtifact[] }) {
  const artifactObj = JSON.parse(artifact.raw);

  return (
    <tbody>
      {
        artifactObj.mediaType === "application/vnd.oci.image.manifest.v1+json" ||
          artifactObj.mediaType === "application/vnd.docker.distribution.manifest.v2+json" ||
          artifact.config_media_type == "application/vnd.cncf.helm.config.v1+json" ? (
          <DetailItem artifact={artifact} />
        ) : artifactObj.mediaType === "application/vnd.docker.distribution.manifest.list.v2+json" ||
          artifactObj.mediaType === "application/vnd.oci.image.index.v1+json" ? (
          artifacts.map((artifact: IArtifact, index: number) => {
            return (
              !skipManifest(artifact.raw) && (
                <DetailItem key={index} artifact={artifact} />
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

function DetailItem({ artifact }: { artifact: IArtifact }) {
  const cutDigest = (digest: string) => {
    if (digest === undefined) {
      return "";
    }
    if (digest.indexOf(":") < 0) {
      return "";
    }
    return digest.substring(digest.indexOf(":") + 1, digest.indexOf(":") + 13);
  }
  let sbomObj = JSON.parse(artifact.sbom === "" ? "{}" : artifact.sbom) as ISbom;
  let vulnerabilityObj = JSON.parse(artifact.vulnerability === "" ? "{}" : artifact.vulnerability) as IVuln;
  let imageConfigObj = JSON.parse(artifact.config_raw) as IImageConfig;
  return (
    <tr className="hover:bg-gray-50 cursor-pointer">
      <td className="text-left w-[180px]">
        <code className="text-xs underline underline-offset-1 text-blue-600 hover:text-blue-500">
          {cutDigest(artifact.digest)}
        </code>
      </td>
      <td className="text-left text-xs w-[180px]">
        Image
      </td>
      <td className="text-left text-xs w-[180px]">
        <div className='flex gap-1'>
          {distros(sbomObj.distro?.name) === "" ? "" : (
            <img src={"/distros/" + distros(sbomObj.distro.name)} alt={sbomObj.distro.name} className="w-4 h-4 inline relative" />
          )}
          <div className=''>
            {distroName(sbomObj.distro?.name) === "" ? "-" : distroName(sbomObj.distro.name) + " " + sbomObj.distro.version}
          </div>
        </div>
      </td>
      <td className="text-left text-xs w-[180px]">
        {
          imageConfigObj.os === undefined ||
            imageConfigObj.architecture === undefined ||
            imageConfigObj.os === "" ||
            imageConfigObj.architecture === "" ? "-" : (
            <span>{imageConfigObj.os}/{imageConfigObj.architecture}</span>
          )
        }
      </td>
      <td className="text-left text-xs w-[180px]">
        Verified
      </td>
      <td className="text-left text-xs w-[180px]">
        {(artifact.pull_times || 0) > 0 ? dayjs().to(dayjs(artifact.last_pull)) : "Never pulled"}
      </td>
      {/* <td className="text-right text-xs w-[180px]">
        {artifact.pull_times}
      </td> */}
      <td className="text-right text-xs w-[220px]">
        <span className="bg-red-800 text-white text-xs font-medium mr-1 px-2 py-0.5 dark:bg-red-900 dark:text-red-300"><span>{vulnerabilityObj.critical || 0}</span> C</span>
        <span className="bg-red-300 text-gray-800 text-xs font-medium mr-1 px-2 py-0.5 dark:bg-red-900 dark:text-red-300">{vulnerabilityObj.high || 0} H</span>
        <span className="bg-amber-400 text-gray-800 text-xs font-medium mr-1 px-2 py-0.5 dark:bg-red-900 dark:text-red-300">{vulnerabilityObj.medium || 0} M</span>
        <span className="bg-amber-200 text-gray-800 text-xs font-medium px-2 py-0.5 dark:bg-red-900 dark:text-red-300">{vulnerabilityObj.low || 0} L</span>
      </td>
      <td className="text-right text-xs w-[180px]">
        {humanFormat(artifact.blob_size || 0)}
      </td>
    </tr>
  );
}
