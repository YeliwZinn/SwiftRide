import React, { useState } from "react";
import axios from "axios";
import { Link } from "react-router-dom";
import { GiAbstract109 } from "react-icons/gi";

function UserLogin({ setUser }) {
  const [formData, setFormData] = useState({
    email: "",
    password: "",
    lat: "",
    lng: "",
  });
  const [error, setError] = useState("");

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    try {
      // Prepare the payload with lat and lng as numbers
      const payload = {
        ...formData,
        lat: parseFloat(formData.lat),
        lng: parseFloat(formData.lng),
      };

      const response = await axios.post("http://localhost:8080/login", payload);
      const { token, role } = response.data;

      // localStorage.setItem("token", token);
      // localStorage.setItem("role", role);

      // Optionally update user state if you have one
      if (setUser) {
        setUser({ token, role });
      }
      console.log(role);

      if (role === "rider") {
        localStorage.setItem("riderToken", token);
        window.location.href = "/rider-dashboard";
      } else if (role === "driver") {
        localStorage.setItem("driverToken", token);
        window.location.href = "/driver-dashboard";
      } else {
        setError("Unknown role. Cannot redirect.");
      }
    } catch (err) {
      setError(
        err.response?.data?.error || "An error occurred. Please try again."
      );
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white p-2 ">
      <nav className="flex justify-between items-center">
        <div className="flex items-center gap-2">
          <GiAbstract109 className="text-black text-3xl bg-white mr-1" />
          {/* <div className="w-8 h-8 bg-white rounded-full" /> */}
          <Link to="/" className="text-white text-2xl font-semibold">
            SwiftRide
          </Link>
        </div>
      </nav>
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center p-6">
        <div className="max-w-md w-full bg-gray-800 px-8 py-10 rounded shadow-lg">
          <p className="text-2xl text-gray-200 text-center mb-12 font-semibold">
            Welcome back! Please login to continue.
          </p>
          {/* <h2 className="text-2xl font-semibold mb-4 text-center">Login</h2> */}
          {error && <p className="text-red-400 text-sm mb-4">{error}</p>}
          <form onSubmit={handleSubmit} className="space-y-6">
            {[
              { name: "email", type: "email" },
              { name: "password", type: "password" },
              { name: "lat", type: "number" },
              { name: "lng", type: "number" },
            ].map(({ name, type }) => (
              <input
                key={name}
                name={name}
                type={type}
                step={type === "number" ? "any" : undefined}
                placeholder={name.charAt(0).toUpperCase() + name.slice(1)}
                value={formData[name]}
                onChange={handleChange}
                className="w-full p-3 rounded bg-gray-700 text-white focus:outline-none text-sm"
                required
              />
            ))}
            <button
              type="submit"
              className="w-full bg-purple-600 hover:bg-purple-700 transition rounded py-4 text-base font-bold mt-6"
            >
              Login
            </button>
          </form>
          <div className="text-center mt-6">
            <Link
              to="/signup"
              className="block text-sm text-gray-400 hover:text-white text-center mt-4"
            >
              Don't have an account? Sign up
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}

export default UserLogin;
