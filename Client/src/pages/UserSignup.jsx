import React, { useState } from "react";
import axios from "axios";

function SignupPage() {
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    phone: "",
    password: "",
    role: "rider",
    vehicle_type: "",
    license_number: "",
    car_plate: "",
    lat: "",
    lng: "",
  });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    // Prepare the payload
    const payload = {
      name: formData.name,
      email: formData.email,
      phone: formData.phone,
      password: formData.password,
      role: formData.role,
      lat: parseFloat(formData.lat),
      lng: parseFloat(formData.lng),
    };

    // Include driver-specific fields if role is 'driver'
    if (formData.role === "driver") {
      payload.vehicle_type = formData.vehicle_type;
      payload.license_number = formData.license_number;
      payload.car_plate = formData.car_plate;
    }

    try {
      const response = await axios.post(
        "http://localhost:8080/signup",
        payload
      );
      setSuccess(response.data.message || "Signup successful!");
      // Optionally redirect or clear form fields here
    } catch (err) {
      console.error(err);
      setError(
        err.response?.data?.error || "An error occurred. Please try again."
      );
    }
  };

  return (
    <div className="min-h-screen w-screen flex items-center justify-center bg-gray-100 p-4 text-black">
      <div className="bg-white p-8 rounded shadow-md w-full max-w-lg">
        <h2 className="text-3xl font-bold mb-6 text-center">Sign Up</h2>
        {error && <div className="mb-4 text-red-500">{error}</div>}
        {success && <div className="mb-4 text-green-500">{success}</div>}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-gray-700">Name</label>
            <input
              type="text"
              name="name"
              value={formData.name}
              onChange={handleChange}
              className="mt-1 p-2 border border-gray-300 rounded w-full"
              required
            />
          </div>
          <div>
            <label className="block text-gray-700">Email</label>
            <input
              type="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              className="mt-1 p-2 border border-gray-300 rounded w-full"
              required
            />
          </div>
          <div>
            <label className="block text-gray-700">Phone</label>
            <input
              type="text"
              name="phone"
              value={formData.phone}
              onChange={handleChange}
              className="mt-1 p-2 border border-gray-300 rounded w-full"
              required
            />
          </div>
          <div>
            <label className="block text-gray-700">Password</label>
            <input
              type="password"
              name="password"
              value={formData.password}
              onChange={handleChange}
              className="mt-1 p-2 border border-gray-300 rounded w-full"
              required
            />
          </div>
          <div>
            <label className="block text-gray-700">Role</label>
            <select
              name="role"
              value={formData.role}
              onChange={handleChange}
              className="mt-1 p-2 border border-gray-300 rounded w-full"
              required
            >
              <option value="rider">Rider</option>
              <option value="driver">Driver</option>
            </select>
          </div>
          {/* Driver-specific fields */}
          {formData.role === "driver" && (
            <>
              <div>
                <label className="block text-gray-700">Vehicle Type</label>
                <select
                  name="vehicle_type"
                  value={formData.vehicle_type}
                  onChange={handleChange}
                  className="mt-1 p-2 border border-gray-300 rounded w-full"
                  required
                >
                  <option value="">Select Vehicle Type</option>
                  <option value="two_wheeler">Two Wheeler</option>
                  <option value="three_wheeler">Three Wheeler</option>
                  <option value="car">Car</option>
                  <option value="premium_car">Premium Car</option>
                </select>
              </div>
              <div>
                <label className="block text-gray-700">License Number</label>
                <input
                  type="text"
                  name="license_number"
                  value={formData.license_number}
                  onChange={handleChange}
                  className="mt-1 p-2 border border-gray-300 rounded w-full"
                  required
                />
              </div>
              <div>
                <label className="block text-gray-700">Car Plate</label>
                <input
                  type="text"
                  name="car_plate"
                  value={formData.car_plate}
                  onChange={handleChange}
                  className="mt-1 p-2 border border-gray-300 rounded w-full"
                  required
                />
              </div>
            </>
          )}
          <div>
            <label className="block text-gray-700">Latitude</label>
            <input
              type="number"
              name="lat"
              step="any"
              value={formData.lat}
              onChange={handleChange}
              className="mt-1 p-2 border border-gray-300 rounded w-full"
              required
            />
          </div>
          <div>
            <label className="block text-gray-700">Longitude</label>
            <input
              type="number"
              name="lng"
              step="any"
              value={formData.lng}
              onChange={handleChange}
              className="mt-1 p-2 border border-gray-300 rounded w-full"
              required
            />
          </div>
          <button
            type="submit"
            className="w-full bg-blue-500 text-white py-2 rounded hover:bg-blue-600 transition"
          >
            Sign Up
          </button>
        </form>
      </div>
    </div>
  );
}

export default SignupPage;
