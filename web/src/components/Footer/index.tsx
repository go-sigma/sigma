/**
 * The MIT License (MIT)
 *
 * Copyright © 2023 Tosone <i@tosone.cn>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

import { AiOutlineTwitter, AiOutlineGithub } from "react-icons/ai";
import { Link } from "react-router-dom";

export default function Footer() {
  return (
    <footer className="bg-white">
      <div className="max-w-7xl mx-auto py-12 px-4 overflow-hidden sm:px-6 lg:px-8">
        <nav className="-mx-5 -my-2 flex flex-wrap justify-center">
          <div className="px-5 py-2">
            <Link to="" className="text-sm text-gray-500 hover:text-gray-900">
              About
            </Link>
          </div>

          <div className="px-5 py-2">
            <Link to="" className="text-sm text-gray-500 hover:text-gray-900">
              Blog
            </Link>
          </div>

          <div className="px-5 py-2">
            <Link to="" className="text-sm text-gray-500 hover:text-gray-900">
              Jobs
            </Link>
          </div>

          <div className="px-5 py-2">
            <Link to="" className="text-sm text-gray-500 hover:text-gray-900">
              Press
            </Link>
          </div>

          <div className="px-5 py-2">
            <Link to="" className="text-sm text-gray-500 hover:text-gray-900">
              Accessibility
            </Link>
          </div>

          <div className="px-5 py-2">
            <Link to="" className="text-sm text-gray-500 hover:text-gray-900">
              Partners
            </Link>
          </div>
        </nav>
        <div className="mt-8 flex justify-center space-x-6">
          <a href="https://twitter.com/itosone" target="_blank" rel="noreferrer" className="text-gray-400 hover:text-gray-500">
            <span className="sr-only">Twitter</span>
            <AiOutlineTwitter className="h-6 w-6" />
          </a>

          <a href="https://github.com/workerflow/" target="_blank" rel="noreferrer" className="text-gray-400 hover:text-gray-500">
            <span className="sr-only">GitHub</span>
            <AiOutlineGithub className="h-6 w-6" />
          </a>

        </div>
        <p className="mt-8 text-center text-sm text-gray-400">
          &copy; 2020 XImager, Inc. All rights reserved.
        </p>
      </div>
    </footer>
  );
}
