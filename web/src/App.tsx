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

import { Routes, Route, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { lazy, Suspense, useLayoutEffect, type ComponentProps } from 'react';

import Tag from "./pages/Tag";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Namespaces from "./pages/Namespace";
import Repositories from "./pages/Repository";
import LoginCallback from './pages/Login/Callback';
import NamespaceMember from "./pages/Namespace/Member";

import NamespaceWebhooks from "./pages/Webhook/List";
import NamespaceWebhookLogs from "./pages/Webhook/Log";

import DaemonTasks from "./pages/DaemonTask/Tasks";
import DaemonTaskRunners from "./pages/DaemonTask/Runners";
import DaemonTaskRecords from "./pages/DaemonTask/Records";

import CodeRepositoryHome from './pages/CodeRepository';
import CodeRepositoryList from './pages/CodeRepository/List';

import BuilderRunnerList from './pages/Builder/RunnerList';

import Setting from './pages/Setting';
import SettingUsers from './pages/Setting/Users';

import { apiBaseURL } from './app/config';
import { setupAxiosInterceptor } from './utils/request'
import Menu, { AppLayoutMenuContext } from './components/Menu';

const Summary = lazy(() => import('./pages/Repository/Summary'));
const NamespaceSummary = lazy(() => import('./pages/Namespace/Summary'));
const BuildersSetup = lazy(() => import('./pages/Builder/Setup'));
const BuilderRunnerLog = lazy(() => import('./pages/Builder/RunnerLog'));

type AppMenuProps = ComponentProps<typeof Menu>;

function getAppMenuProps(pathname: string, search: string): Omit<AppMenuProps, "localServer" | "persistent"> {
  const params = new URLSearchParams(search);
  const segments = pathname.split("/").filter(Boolean).map(segment => decodeURIComponent(segment));
  const [section, namespace, detail, resource] = segments;

  if (section === "home") {
    return { item: "home" };
  }

  if (section === "coderepos" || section === "builders") {
    return { item: "coderepos" };
  }

  if (section === "settings") {
    if (detail === "webhooks" || namespace === "webhooks") {
      return { item: "webhooks" };
    }
    if (detail === "daemon-tasks" || namespace === "daemon-tasks") {
      return { item: "daemon-tasks" };
    }
    if (namespace === "users") {
      return { item: "users" };
    }
    return { item: "settings" };
  }

  if (section === "namespaces" && namespace) {
    if (detail === "repository" && resource === "tags") {
      return {
        item: "tags",
        namespace,
        namespace_id: params.get("namespace_id") || "",
        repository: params.get("repository") || "",
        repository_id: params.get("repository_id") || "",
      };
    }

    return {
      item: "repositories",
      namespace,
      namespace_id: params.get("namespace_id") || "",
      repository: params.get("repository") || "",
      repository_id: params.get("repository_id") || "",
      selfClick: detail === "repository" && (resource === "runners" || resource === "runner-logs"),
    };
  }

  return { item: "namespaces" };
}

function AppLayout() {
  const location = useLocation();
  const menuProps = getAppMenuProps(location.pathname, location.search);

  return (
    <AppLayoutMenuContext.Provider value={true}>
      <div className="min-h-screen flex overflow-hidden bg-white dark:bg-gray-950">
        <Menu localServer={apiBaseURL} persistent {...menuProps} />
        <div className="flex min-w-0 flex-1 [&>*]:w-full">
          <Outlet />
        </div>
      </div>
    </AppLayoutMenuContext.Provider>
  );
}

export default function App() {
  const navigate = useNavigate();

  useLayoutEffect(() => {
    setupAxiosInterceptor(navigate);
  }, [navigate]);

  return (
    <>
      <Suspense fallback={null}>
        <Routes>
          <Route path="/login" element={<Login localServer={apiBaseURL} />} />
          <Route path="/login/callback/:provider" element={<LoginCallback localServer={apiBaseURL} />} />

          <Route element={<AppLayout />}>
            <Route path="/" element={<Namespaces localServer={apiBaseURL} />} />

            <Route path="/home" element={<Home localServer={apiBaseURL} />} />

            <Route path="/namespaces" element={<Namespaces localServer={apiBaseURL} />} />
            <Route path="/namespaces/:namespace/namespace-summary" element={<NamespaceSummary localServer={apiBaseURL} />} />
            <Route path="/namespaces/:namespace/members" element={<NamespaceMember localServer={apiBaseURL} />} />
            <Route path="/namespaces/:namespace/webhooks" element={<NamespaceWebhooks localServer={apiBaseURL} />} />
            <Route path="/namespaces/:namespace/webhooks/:webhook_id" element={<NamespaceWebhookLogs localServer={apiBaseURL} />} />

            <Route path="/namespaces/:namespace/daemon-tasks" element={<DaemonTasks localServer={apiBaseURL} />} />
            <Route path="/namespaces/:namespace/daemon-tasks/:resource" element={<DaemonTaskRunners localServer={apiBaseURL} />} />
            <Route path="/namespaces/:namespace/daemon-tasks/:resource/:runner_id/records" element={<DaemonTaskRecords localServer={apiBaseURL} />} />

            <Route path="/namespaces/:namespace/repositories" element={<Repositories localServer={apiBaseURL} />} />
            <Route path="/namespaces/:namespace/repository/summary" element={<Summary localServer={apiBaseURL} />} />
            <Route path="/namespaces/:namespace/repository/tags" element={<Tag localServer={apiBaseURL} />} />

            <Route path="/namespaces/:namespace/repository/runners" element={<BuilderRunnerList localServer={apiBaseURL} />} />
            <Route path="/namespaces/:namespace/repository/runner-logs/:runner_id" element={<BuilderRunnerLog localServer={apiBaseURL} />} />

            <Route path="/coderepos" element={<CodeRepositoryHome localServer={apiBaseURL} />} />
            <Route path="/coderepos/:provider" element={<CodeRepositoryList localServer={apiBaseURL} />} />

            <Route path="/builders/setup" element={<BuildersSetup localServer={apiBaseURL} />} />
            <Route path="/builders/setup/:id" element={<BuildersSetup localServer={apiBaseURL} />} />

            <Route path="/settings" element={<Setting localServer={apiBaseURL} />} />
            <Route path="/settings/users" element={<SettingUsers localServer={apiBaseURL} />} />
            <Route path="/settings/webhooks" element={<NamespaceWebhooks localServer={apiBaseURL} />} />
            <Route path="/settings/webhooks/:webhook_id" element={<NamespaceWebhookLogs localServer={apiBaseURL} />} />
            <Route path="/settings/daemon-tasks" element={<DaemonTasks localServer={apiBaseURL} />} />
            <Route path="/settings/daemon-tasks/:resource" element={<DaemonTaskRunners localServer={apiBaseURL} />} />
            <Route path="/settings/daemon-tasks/:resource/:runner_id/records" element={<DaemonTaskRecords localServer={apiBaseURL} />} />
          </Route>

        </Routes>
      </Suspense>
    </>
  );
}
