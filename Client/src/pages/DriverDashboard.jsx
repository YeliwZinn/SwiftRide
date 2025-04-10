import React, { useEffect, useState } from "react";
import axios from "axios";
import RideNotification from "./RideNotification";

function DriverDashboard() {
  const [notifications, setNotifications] = useState([]);

  useEffect(() => {
    const token = localStorage.getItem("driverToken");
    if (!token) return;

    const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

    ws.onopen = () => {
      console.log("✅ WebSocket connected");
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "ping" }));
      }
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        console.log("📩 Message from server:", message);

        if (message.type === "ride_request") {
          setNotifications((prev) => [...prev, message.payload]);
        }

        if (message.type === "pong") {
          console.log("📡 Server is alive");
        }
      } catch (error) {
        console.error("❌ Failed to parse WebSocket message:", event.data);
      }
    };

    ws.onerror = (error) => {
      console.error("❌ WebSocket error:", error);
    };

    ws.onclose = () => {
      console.log("🔌 WebSocket disconnected");
    };

    return () => ws.close();
  }, []);

  const handleAccept = async (rideId) => {
    console.log(rideId);
    try {
      const res = await axios.post(
        `http://localhost:8080/rides/${rideId}/respond`,
        {
          // ride_id: rideId,
          accept: true,
        },
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("driverToken")}`,
          },
        }
      );

      console.log("✅ Ride accepted:", res.data.message);

      // Remove the accepted notification from the UI
      setNotifications((prev) => prev.filter((n) => n.ride_id !== rideId));
    } catch (err) {
      console.error(
        "❌ Error accepting ride:",
        err.response?.data?.error || err.message
      );
    }
  };

  const handleReject = async (rideId) => {
    try {
      const res = await axios.post(
        `http://localhost:8080/rides/${rideId}/respond`,
        {
          // ride_id: rideId,
          accept: false,
        },
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("driverToken")}`,
          },
        }
      );

      console.log("❌ Ride rejected:", res.data.message);

      // Remove the rejected notification from the UI
      setNotifications((prev) => prev.filter((n) => n.ride_id !== rideId));
    } catch (err) {
      console.error(
        "❌ Error rejecting ride:",
        err.response?.data?.error || err.message
      );
    }
  };

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <h2 className="text-2xl font-bold mb-4 text-gray-800">
        Driver Dashboard
      </h2>
      {notifications.map((notif, idx) => (
        <RideNotification
          key={idx}
          data={notif}
          onAccept={handleAccept}
          onReject={handleReject}
        />
      ))}
    </div>
  );
}

export default DriverDashboard;
