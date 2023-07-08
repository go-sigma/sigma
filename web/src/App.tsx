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

import { ToastContainer } from 'react-toastify';
import { Routes, Route, useNavigate } from 'react-router-dom';

import Home from "./pages/Home";
import Login from "./pages/Login";

import Namespace from "./pages/Namespace";
import Repository from "./pages/Repository";
import Tag from "./pages/Tag";
import Artifact from "./pages/Artifact";

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
      <ToastContainer position="top-right"
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
        <Route path="/login" element={<Login localServer={localServer} />} />
        <Route path="/home" element={<Home localServer={localServer} />} />
        <Route path="/namespace" element={<Namespace localServer={localServer} />} />
        <Route path="/namespace/:namespace/repository" element={<Repository localServer={localServer} />} />
        <Route path="/namespace/:namespace/artifact" element={<Artifact localServer={localServer} />} />
        <Route path="/namespace/:namespace/tag" element={<Tag localServer={localServer} />} />
        <Route path="/about" element={<About />} />
      </Routes>
    </>
  );
}
