import { Fragment } from 'react';
import { Routes, Route, HashRouter as Router } from 'react-router-dom';

import Home from "./pages/Home";
// import Login from "./pages/User/Login";
// import Signup from "./pages/User/Signup";
// import Editor from "./pages/Editor";

function About() {
  return <h1>About</h1>;
}

export default function App() {
  return (
    <Fragment>
      <Router>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/about" element={<About />} />
          {/* <Route path="/editor" element={<Editor />} />
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} /> */}
        </Routes>
      </Router>
    </Fragment>
  );
}
