import React, { useState, useEffect } from "react";
import axios from "axios";
import MapView from "./MapView";

function RiderDashboard() {
  const [user, setUser] = useState(null);
  const [error, setError] = useState("");

  useEffect(() => {
    async function fetchUser() {
      try {
        const token = localStorage.getItem("token");

        const response = await axios.get("http://localhost:8080/profile", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        setUser(response.data);
        // console.log("Fetched user data:", response.data);
      } catch (err) {
        console.error("Error fetching user details:", err);
        setError(err.response?.data?.error || "Error fetching user details");
      }
    }

    fetchUser();
  }, []);

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-red-500">{error}</div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div>Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 p-8 text-black">
      <div className="max-w-2xl mx-auto bg-white p-8 rounded shadow">
        <h2 className="text-3xl font-bold mb-6">Rider Dashboard</h2>
        <div className="mb-4">
          <span className="font-semibold">Name:</span> {user.name}
        </div>
        <div className="mb-4">
          <span className="font-semibold">Email:</span> {user.email}
        </div>
        <div className="mb-4">
          <span className="font-semibold">Phone:</span> {user.phone}
        </div>
        <div className="mb-4">
          <span className="font-semibold">Role:</span> {user.role}
        </div>
        <div className="mb-4">
          <span className="font-semibold">Latitude:</span> {user.latitude}
        </div>
        <div className="mb-4">
          <span className="font-semibold">Longitude:</span> {user.longitude}
        </div>

        <MapView latitude={user.latitude} longitude={user.longitude} />
      </div>
    </div>
  );
}

export default RiderDashboard;
