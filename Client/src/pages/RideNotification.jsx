// RideNotification.js
import React from "react";

function RideNotification({ data, onAccept, onReject }) {
  const { ride_id, rider_id, fare, distance, pickup } = data;

  return (
    <div className="bg-white p-4 rounded-xl shadow-md mb-4 border border-gray-200">
      <h3 className="text-xl font-semibold mb-2 text-gray-800">🚕 New Ride Request</h3>
      <p><strong>Rider ID:</strong> {rider_id}</p>
      <p><strong>Pickup Location:</strong> {pickup?.join(", ")}</p>
      <p><strong>Distance:</strong> {distance.toFixed(2)} km</p>
      <p><strong>Fare:</strong> ₹{fare.toFixed(2)}</p>
      <div className="mt-4 flex gap-3">
        <button
          onClick={() => onAccept(ride_id)}
          className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
        >
          Accept
        </button>
        <button
          onClick={() => onReject(ride_id)}
          className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
        >
          Reject
        </button>
      </div>
    </div>
  );
}

export default RideNotification;
