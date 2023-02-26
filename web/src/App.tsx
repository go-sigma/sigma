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

import { Fragment } from 'react';
import { Routes, Route, HashRouter as Router } from 'react-router-dom';
import { ToastContainer } from 'react-toastify';

import Home from "./pages/Home";
import Namespace from "./pages/Namespace";
import Repository from "./pages/Repository";
import Artifact from "./pages/Artifact";
import Tag from "./pages/Tag";

const localServer = process.env.NODE_ENV === "development" ? "http://127.0.0.1:3000" : "";

function About() {
  return <h1>About</h1>;
}

export default function App() {
  return (
    <Fragment>
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
      <Router>
        <Routes>
          <Route path="/" element={<Home localServer={localServer} />} />
          <Route path="/home" element={<Home localServer={localServer} />} />
          <Route path="/namespace" element={<Namespace localServer={localServer} />} />
          <Route path="/namespace/:namespace/repository" element={<Repository localServer={localServer} />} />
          <Route path="/namespace/:namespace/artifact" element={<Artifact localServer={localServer} />} />
          <Route path="/namespace/:namespace/tag" element={<Tag localServer={localServer} />} />
          <Route path="/about" element={<About />} />
        </Routes>
      </Router>
    </Fragment>
  );
}
