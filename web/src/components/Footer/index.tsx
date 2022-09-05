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
          &copy; 2020 Workflow, Inc. All rights reserved.
        </p>
      </div>
    </footer>
  );
}
