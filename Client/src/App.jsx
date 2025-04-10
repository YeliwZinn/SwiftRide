import React from "react";
import { Route, Routes } from "react-router-dom";
import Home from "./pages/Home";
import UserLogin from "./pages/UserLogin";

import UserSignup from "./pages/UserSignup";
import RiderDashboard from "./pages/RiderDashboard";
import DriverDashboard from "./pages/DriverDashboard";

const App = () => {
  return (
    <div>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<UserLogin />} />
        <Route path="/signup" element={<UserSignup />} />

        <Route path="/rider-dashboard" element={<RiderDashboard />}></Route>
        <Route path="/driver-dashboard" element={<DriverDashboard />}></Route>
      </Routes>
    </div>
  );
};

export default App;
