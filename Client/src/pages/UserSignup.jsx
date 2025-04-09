import React, { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import axios from "axios";
import { useNavigate, Link } from "react-router-dom";

const riderImage =
  "https://images.unsplash.com/photo-1612867754336-c45d074c4f8e?q=80&w=1965&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D";
const driverImage =
  "https://images.unsplash.com/photo-1581266454053-6432a63ff315?q=80&w=1944&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D";

const ToggleTabs = ({ role, setRole }) => (
  <div className="absolute top-[-54px] left-1/2 transform -translate-x-1/2 z-10">
    <div className="bg-gray-800 p-1 rounded-full flex w-fit">
      {["rider", "driver"].map((type) => (
        <button
          key={type}
          className={`px-6 py-2 rounded-full text-sm font-semibold tracking-wide transition-all duration-300 ${
            role === type
              ? "bg-purple-600 text-white"
              : "text-gray-400 hover:text-white"
          }`}
          onClick={() => setRole(type)}
        >
          {type.charAt(0).toUpperCase() + type.slice(1)}
        </button>
      ))}
    </div>
  </div>
);

const SignupPage = () => {
  const navigate = useNavigate();

  const [role, setRole] = useState("rider");
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

    const payload = {
      name: formData.name,
      email: formData.email,
      phone: formData.phone,
      password: formData.password,
      role,
      lat: parseFloat(formData.lat),
      lng: parseFloat(formData.lng),
    };

    if (role === "driver") {
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
      setTimeout(() => {
        navigate("/login");
      }, 1000);
    } catch (err) {
      console.error(err);
      setError(
        err.response?.data?.error || "An error occurred. Please try again."
      );
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center p-6">
      <div className="max-w-3xl w-full flex flex-col items-center">
        <div className="relative w-full h-[580px]">
          <ToggleTabs
            role={role}
            setRole={(newRole) => {
              setRole(newRole);
              setFormData({ ...formData, role: newRole });
            }}
          />
          <AnimatePresence mode="wait">
            {role === "rider" ? (
              <motion.div
                key="rider"
                initial={{ opacity: 0, x: -100 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -100 }}
                transition={{ duration: 0.5 }}
                className="absolute w-full h-full grid grid-cols-2"
              >
                <div className="bg-gray-800 px-8 py-10 h-full">
                  <h2 className="text-2xl font-semibold mb-4">
                    Sign up as Rider
                  </h2>
                  {error && <p className="text-red-400 text-sm">{error}</p>}
                  {success && (
                    <p className="text-green-400 text-sm">{success}</p>
                  )}
                  <form onSubmit={handleSubmit} className="space-y-6 mt-2">
                    {["name", "email", "phone", "password", "lat", "lng"].map(
                      (field) => (
                        <input
                          key={field}
                          name={field}
                          type={field === "password" ? "password" : "text"}
                          placeholder={
                            field.charAt(0).toUpperCase() + field.slice(1)
                          }
                          value={formData[field]}
                          onChange={handleChange}
                          className="w-full p-2 rounded bg-gray-700 text-white focus:outline-none text-sm"
                          required
                        />
                      )
                    )}
                    <button
                      type="submit"
                      className="w-2/5 bg-purple-600 hover:bg-purple-700 transition rounded py-2 text-sm mt-6"
                    >
                      Sign Up
                    </button>
                    <Link
                      to="/login"
                      className="block text-sm text-gray-400 hover:text-white text-center mt-4"
                    >
                      Already have an account? Login
                    </Link>
                  </form>
                </div>
                <div>
                  <img
                    src={riderImage}
                    className="h-full w-full object-cover"
                    alt="Rider"
                  />
                </div>
              </motion.div>
            ) : (
              <motion.div
                key="driver"
                initial={{ opacity: 0, x: 100 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 100 }}
                transition={{ duration: 0.5 }}
                className="absolute w-full h-full grid grid-cols-2"
              >
                <div>
                  <img
                    src={driverImage}
                    className="h-full w-full object-cover"
                    alt="Driver"
                  />
                </div>
                <div className="bg-gray-800 px-8 py-10 h-full">
                  <h2 className="text-2xl font-semibold mb-4">
                    Sign up as Driver
                  </h2>
                  {error && <p className="text-red-400 text-sm">{error}</p>}
                  {success && (
                    <p className="text-green-400 text-sm">{success}</p>
                  )}
                  <form onSubmit={handleSubmit} className="space-y-3 mt-2">
                    {["name", "email", "phone", "password"].map((field) => (
                      <input
                        key={field}
                        name={field}
                        type={field === "password" ? "password" : "text"}
                        placeholder={
                          field.charAt(0).toUpperCase() + field.slice(1)
                        }
                        value={formData[field]}
                        onChange={handleChange}
                        className="w-full p-3 rounded bg-gray-700 text-white focus:outline-none text-sm"
                        required
                      />
                    ))}

                    <select
                      name="vehicle_type"
                      value={formData.vehicle_type}
                      onChange={handleChange}
                      className="w-full p-3 rounded bg-gray-700 text-white text-sm"
                      required
                    >
                      <option value="">Select Vehicle Type</option>
                      <option value="two_wheeler">Two Wheeler</option>
                      <option value="car">Car</option>
                      <option value="premium_car">Premium Car</option>
                    </select>

                    {/* license_number and car_plate on the same line */}
                    <div className="flex gap-4">
                      <input
                        name="license_number"
                        type="text"
                        placeholder="License Number"
                        value={formData.license_number}
                        onChange={handleChange}
                        className="w-1/2 p-2 rounded bg-gray-700 text-white focus:outline-none text-sm"
                        required
                      />
                      <input
                        name="car_plate"
                        type="text"
                        placeholder="Car Plate"
                        value={formData.car_plate}
                        onChange={handleChange}
                        className="w-1/2 p-3 rounded bg-gray-700 text-white focus:outline-none text-sm"
                        required
                      />
                    </div>

                    {/* lat and lng on the same line */}
                    <div className="flex gap-4">
                      <input
                        name="lat"
                        type="text"
                        placeholder="Lat"
                        value={formData.lat}
                        onChange={handleChange}
                        className="w-1/2 p-2 rounded bg-gray-700 text-white focus:outline-none text-sm"
                      />
                      <input
                        name="lng"
                        type="text"
                        placeholder="Lng"
                        value={formData.lng}
                        onChange={handleChange}
                        className="w-1/2 p-2 rounded bg-gray-700 text-white focus:outline-none text-sm"
                      />
                    </div>

                    <button
                      type="submit"
                      className="w-2/5 bg-purple-600 hover:bg-purple-700 transition rounded py-2 text-sm mt-4"
                    >
                      Sign Up
                    </button>
                    <Link
                      to="/login"
                      className="block text-sm text-gray-400 hover:text-white text-center mt-4"
                    >
                      Already have an account? Login
                    </Link>
                  </form>
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
};

export default SignupPage;
