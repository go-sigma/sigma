import { Fragment } from 'react';
import { Routes, Route, HashRouter as Router } from 'react-router-dom';

import Home from "./pages/Home";

const localServer = process.env.NODE_ENV === "development" ? "http://127.0.0.1:3000" : "";

function About() {
  return <h1>About</h1>;
}

export default function App() {
  return (
    <Fragment>
      <Router>
        <Routes>
          <Route path="/" element={<Home localServer={localServer} />} />
          <Route path="/about" element={<About />} />
        </Routes>
      </Router>
    </Fragment>
  );
}
