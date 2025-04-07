import React, { useState } from "react";
import axios from "axios";
import { Link } from "react-router-dom";

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
      const { token } = response.data;

      localStorage.setItem("token", token);

      // Optionally update user state if you have one
      if (setUser) {
        setUser({ token });
      }

      // Redirect to dashboard or home page
      window.location.href = "/rider-dashboard";
    } catch (err) {
      setError(
        err.response?.data?.error || "An error occurred. Please try again."
      );
    }
  };

  return (
    <div className="min-h-screen text-black w-screen flex flex-col items-center justify-center bg-gray-100">
      <div className="bg-white p-8 rounded shadow-md w-full max-w-md mr-20 ml-20">
        <h2 className="text-2xl font-bold mb-6 text-center">Login</h2>
        {error && <div className="mb-4 text-red-500">{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-gray-700">Email</label>
            <input
              type="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              className="mt-1 p-2 w-full border border-gray-300 rounded"
              required
            />
          </div>
          <div className="mb-4">
            <label className="block text-gray-700">Password</label>
            <input
              type="password"
              name="password"
              value={formData.password}
              onChange={handleChange}
              className="mt-1 p-2 w-full border border-gray-300 rounded"
              required
            />
          </div>
          <div className="mb-4">
            <label className="block text-gray-700">Latitude</label>
            <input
              type="number"
              step="any"
              name="lat"
              value={formData.lat}
              onChange={handleChange}
              className="mt-1 p-2 w-full border border-gray-300 rounded"
              required
            />
          </div>
          <div className="mb-4">
            <label className="block text-gray-700">Longitude</label>
            <input
              type="number"
              step="any"
              name="lng"
              value={formData.lng}
              onChange={handleChange}
              className="mt-1 p-2 w-full border border-gray-300 rounded"
              required
            />
          </div>
          <button
            type="submit"
            className="w-full bg-blue-500 text-white p-2 rounded hover:bg-blue-600 transition"
          >
            Login
          </button>
        </form>
      </div>
      <Link to="/signup" className="mt-10 text-xl font-bold text-teal-600 hover:underline">
        Signup
      </Link>
    </div>
  );
}

export default UserLogin;
