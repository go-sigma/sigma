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

import { Toaster } from 'react-hot-toast';
import { ToastContainer } from 'react-toastify';
import { Routes, Route, useNavigate } from 'react-router-dom';

import Tag from "./pages/Tag";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Namespaces from "./pages/Namespace";
import Repositories from "./pages/Repository";
import Summary from './pages/Repository/Summary';
import LoginCallback from './pages/Login/Callback';
import NamespaceMember from "./pages/Namespace/Member";
import NamespaceSummary from "./pages/Namespace/Summary";
import NamespaceWebhooks from "./pages/Namespace/Webhook";

import DaemonTasks from "./pages/DaemonTask/Tasks";
import DaemonTaskRunners from "./pages/DaemonTask/Runners";
import DaemonTaskRecords from "./pages/DaemonTask/Records";

import CodeRepositoryHome from './pages/CodeRepository';
import CodeRepositoryList from './pages/CodeRepository/List';

import BuildersSetup from './pages/Builder/Setup';
import BuilderRunnerList from './pages/Builder/RunnerList';
import BuilderRunnerLog from './pages/Builder/RunnerLog';

import Setting from './pages/Setting';
import SettingUsers from './pages/Setting/Users';

import { setupResponseInterceptor } from './utils/request'

const localServer = process.env.NODE_ENV === "development" ? "http://127.0.0.1:3000" : "";
// const localServer = process.env.NODE_ENV === "development" ? "https://sigma.tosone.cn" : "";

export default function App() {
  const navigate = useNavigate();
  setupResponseInterceptor(navigate);

  return (
    <>
      <Toaster
        toastOptions={{
          style: {
            maxWidth: "500px",
            fontSize: "0.875rem",
            lineHeight: "1.25rem",
          }
        }}
      />
      <ToastContainer
        position="top-right"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop={false}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
        theme="light"
        style={{ top: "3rem" }}
      />
      <Routes>
        <Route path="/" element={<Namespaces localServer={localServer} />} />

        <Route path="/home" element={<Home localServer={localServer} />} />

        <Route path="/login" element={<Login localServer={localServer} />} />
        <Route path="/login/callback/:provider" element={<LoginCallback localServer={localServer} />} />

        <Route path="/namespaces" element={<Namespaces localServer={localServer} />} />
        <Route path="/namespaces/:namespace/namespace-summary" element={<NamespaceSummary localServer={localServer} />} />
        <Route path="/namespaces/:namespace/members" element={<NamespaceMember localServer={localServer} />} />
        <Route path="/namespaces/:namespace/namespace-webhooks" element={<NamespaceWebhooks localServer={localServer} />} />

        <Route path="/namespaces/:namespace/daemon-tasks" element={<DaemonTasks localServer={localServer} />} />
        <Route path="/namespaces/:namespace/daemon-tasks/:resource" element={<DaemonTaskRunners localServer={localServer} />} />
        <Route path="/namespaces/:namespace/daemon-tasks/:resource/:runner_id/records" element={<DaemonTaskRecords localServer={localServer} />} />

        <Route path="/namespaces/:namespace/repositories" element={<Repositories localServer={localServer} />} />
        <Route path="/namespaces/:namespace/repository/summary" element={<Summary localServer={localServer} />} />
        <Route path="/namespaces/:namespace/repository/tags" element={<Tag localServer={localServer} />} />

        <Route path="/namespaces/:namespace/repository/runners" element={<BuilderRunnerList localServer={localServer} />} />
        <Route path="/namespaces/:namespace/repository/runner-logs/:runner_id" element={<BuilderRunnerLog localServer={localServer} />} />

        <Route path="/coderepos" element={<CodeRepositoryHome localServer={localServer} />} />
        <Route path="/coderepos/:provider" element={<CodeRepositoryList localServer={localServer} />} />

        <Route path="/builders/setup" element={<BuildersSetup localServer={localServer} />} />
        <Route path="/builders/setup/:id" element={<BuildersSetup localServer={localServer} />} />

        <Route path="/settings" element={<Setting localServer={localServer} />} />
        <Route path="/settings/users" element={<SettingUsers localServer={localServer} />} />
        <Route path="/settings/daemon-tasks" element={<DaemonTasks localServer={localServer} />} />
        <Route path="/settings/daemon-tasks/:resource" element={<DaemonTaskRunners localServer={localServer} />} />
        <Route path="/settings/daemon-tasks/:resource/:runner_id/records" element={<DaemonTaskRecords localServer={localServer} />} />

      </Routes>
    </>
  );
}
