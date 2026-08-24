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

import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";

export default function Footer() {
  return (
    <footer className="bg-background">
      <div className="max-w-7xl mx-auto py-12 px-4 overflow-hidden sm:px-6 lg:px-8">
        <nav className="-mx-5 -my-2 flex flex-wrap justify-center">
          <div className="px-5 py-2">
            <Link to="" className="text-sm text-muted-foreground hover:text-foreground">
              About
            </Link>
          </div>
          <div className="px-5 py-2">
            <Link to="" className="text-sm text-muted-foreground hover:text-foreground">
              Blog
            </Link>
          </div>
          <div className="px-5 py-2">
            <Link to="" className="text-sm text-muted-foreground hover:text-foreground">
              Jobs
            </Link>
          </div>
          <div className="px-5 py-2">
            <Link to="" className="text-sm text-muted-foreground hover:text-foreground">
              Press
            </Link>
          </div>
          <div className="px-5 py-2">
            <Link to="" className="text-sm text-muted-foreground hover:text-foreground">
              Accessibility
            </Link>
          </div>
          <div className="px-5 py-2">
            <Link to="" className="text-sm text-muted-foreground hover:text-foreground">
              Partners
            </Link>
          </div>
        </nav>
        <div className="mt-8 flex justify-center gap-6">
          <Button variant="ghost" size="icon" render={<a href="https://twitter.com/itosone" target="_blank" rel="noreferrer" />}>
            <span className="sr-only">Twitter</span>
            <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg>
          </Button>
          <Button variant="ghost" size="icon" render={<a href="https://github.com/workerflow/" target="_blank" rel="noreferrer" />}>
            <span className="sr-only">GitHub</span>
            <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
          </Button>
        </div>
        <p className="mt-8 text-center text-sm text-muted-foreground">
          &copy; 2020 sigma, Inc. All rights reserved.
        </p>
      </div>
    </footer>
  );
}