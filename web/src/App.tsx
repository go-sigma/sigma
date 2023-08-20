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

import { ToastContainer } from 'react-toastify';
import { Routes, Route, useNavigate } from 'react-router-dom';

import Tag from "./pages/Tag";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Namespaces from "./pages/Namespace";
import Repositories from "./pages/Repository";
import LoginCallback from './pages/Login/Callback';
import Summary from './pages/Repository/Summary';
import NamespaceUsers from "./pages/Namespace/Users";
import NamespaceWebhooks from "./pages/Namespace/Webhook";
import NamespaceDaemonTasks from "./pages/Namespace/DaemonTask";

import CodeRepository from './pages/CodeRepository';

import { setupResponseInterceptor } from './utils/request'

const localServer = process.env.NODE_ENV === "development" ? "http://127.0.0.1:3000" : "";

function About() {
  return <h1>About</h1>;
}

export default function App() {
  const navigate = useNavigate();
  setupResponseInterceptor(navigate);

  return (
    <>
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
        <Route path="/" element={<Home localServer={localServer} />} />

        <Route path="/home" element={<Home localServer={localServer} />} />

        <Route path="/login" element={<Login localServer={localServer} />} />
        <Route path="/login/callback/:provider" element={<LoginCallback localServer={localServer} />} />

        <Route path="/namespaces" element={<Namespaces localServer={localServer} />} />
        <Route path="/namespaces/:namespace/namespace-users" element={<NamespaceUsers localServer={localServer} />} />
        <Route path="/namespaces/:namespace/namespace-webhooks" element={<NamespaceWebhooks localServer={localServer} />} />
        <Route path="/namespaces/:namespace/namespace-daemon-tasks" element={<NamespaceDaemonTasks localServer={localServer} />} />
        <Route path="/namespaces/:namespace/repositories" element={<Repositories localServer={localServer} />} />
        <Route path="/namespaces/:namespace/repository/summary" element={<Summary localServer={localServer} />} />
        <Route path="/namespaces/:namespace/repository/tags" element={<Tag localServer={localServer} />} />

        <Route path="/coderepos" element={<CodeRepository localServer={localServer} />} />

        <Route path="/about" element={<About />} />
      </Routes>
    </>
  );
}
