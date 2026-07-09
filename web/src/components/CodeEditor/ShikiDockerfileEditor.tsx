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

import { ChangeEvent, KeyboardEvent, UIEvent, useEffect, useRef, useState } from 'react';
import { createHighlighterCore, type HighlighterCore } from '@shikijs/core';
import { createJavaScriptRegexEngine } from '@shikijs/engine-javascript';
import dockerfile from '@shikijs/langs/dockerfile';
import githubDark from '@shikijs/themes/github-dark';

export default function ShikiDockerfileEditor({
  value,
  onChange,
  height = "40vh",
}: {
  value: string;
  onChange: (value: string) => void;
  height?: string;
}) {
  const [html, setHtml] = useState("");
  const [highlighter, setHighlighter] = useState<HighlighterCore>();
  const highlightRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let cancelled = false;

    createHighlighterCore({
      engine: createJavaScriptRegexEngine(),
      langs: [dockerfile],
      themes: [githubDark],
    }).then(nextHighlighter => {
      if (!cancelled) {
        setHighlighter(nextHighlighter);
      }
    });

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (highlighter === undefined) {
      return;
    }

    setHtml(highlighter.codeToHtml(value || " ", {
      lang: "dockerfile",
      theme: "github-dark",
    }));
  }, [highlighter, value]);

  const handleScroll = (event: UIEvent<HTMLTextAreaElement>) => {
    if (highlightRef.current === null) {
      return;
    }
    highlightRef.current.scrollTop = event.currentTarget.scrollTop;
    highlightRef.current.scrollLeft = event.currentTarget.scrollLeft;
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key !== "Tab") {
      return;
    }
    event.preventDefault();

    const target = event.currentTarget;
    const start = target.selectionStart;
    const end = target.selectionEnd;
    const nextValue = `${value.slice(0, start)}  ${value.slice(end)}`;
    onChange(nextValue);

    requestAnimationFrame(() => {
      target.selectionStart = start + 2;
      target.selectionEnd = start + 2;
    });
  };

  const handleChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    onChange(event.target.value);
  };

  return (
    <div
      className="relative w-full overflow-hidden rounded-md bg-[#24292e] shadow-sm ring-1 ring-inset ring-gray-700"
      style={{ height }}
    >
      <div
        ref={highlightRef}
        className="pointer-events-none absolute inset-0 overflow-auto p-3 font-mono text-sm leading-6 [&_pre]:!m-0 [&_pre]:!bg-transparent [&_pre]:!p-0 [&_pre]:font-mono [&_pre]:text-sm [&_pre]:leading-6"
        dangerouslySetInnerHTML={{ __html: html }}
      />
      <textarea
        aria-label="Dockerfile editor"
        value={value}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        onScroll={handleScroll}
        wrap="off"
        spellCheck={false}
        className="absolute inset-0 h-full w-full resize-none overflow-auto whitespace-pre border-0 bg-transparent p-3 font-mono text-sm leading-6 text-transparent caret-white outline-none selection:bg-indigo-500/40 focus:ring-0"
      />
    </div>
  );
}
