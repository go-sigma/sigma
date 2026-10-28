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

import { useEffect, useRef } from "react";
import { Terminal } from "xterm";
import { AttachAddon } from "xterm-addon-attach";
import { FitAddon } from "xterm-addon-fit";
import "xterm/css/xterm.css";

// RunnerLogView renders the log of one builder runner in a terminal. A finished runner reads the
// stored log, a running one streams the container output over the log websocket.
export default function RunnerLogView({
  localServer,
  namespaceId,
  repositoryId,
  builderId,
  runnerId,
  status,
  className,
}: {
  localServer: string;
  namespaceId: string;
  repositoryId: string;
  builderId: string;
  runnerId: string;
  status: string;
  className?: string;
}) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (container === null || builderId === "" || runnerId === "") {
      return;
    }

    const term = new Terminal({
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      fontSize: 13,
      disableStdin: true,
      cursorInactiveStyle: "none",
      cursorStyle: "bar",
      convertEol: true,
    });
    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(container);

    // The dialog animates open, so the container has no size when the effect runs.
    const refit = () => {
      if (container.clientWidth === 0 || container.clientHeight === 0) {
        return;
      }
      fitAddon.fit();
    };
    refit();
    const observer = new ResizeObserver(refit);
    observer.observe(container);

    let ws: WebSocket | undefined;
    if (status === "Pending") {
      term.writeln("Pending...");
    } else if (status === "Building" || status === "Success" || status === "Failed" || status === "Stopped") {
      const base = localServer === "" ? window.location.origin : localServer;
      const url = new URL(
        `/api/v1/namespaces/${namespaceId}/repositories/${repositoryId}/builders/${builderId}/runners/${runnerId}/log`,
        base,
      );
      ws = new WebSocket(url.toString().replace(/^http/, "ws"));
      term.loadAddon(new AttachAddon(ws));
    }

    return () => {
      observer.disconnect();
      ws?.close();
      term.dispose();
    };
  }, [localServer, namespaceId, repositoryId, builderId, runnerId, status]);

  return <div ref={containerRef} className={className} />;
}
