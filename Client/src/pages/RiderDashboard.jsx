import React, { useEffect, useRef, useState } from "react";
import mapboxgl from "mapbox-gl";
import axios from "axios";

mapboxgl.accessToken =
  "pk.eyJ1Ijoic3VieHViZXIiLCJhIjoiY204eXIxNGI5MDNvaTJsczgyYXJtenZyOCJ9.ba7NK0E5Ce_N3hHyFRzJTg";

const RiderDashboard = () => {
  const mapContainerRef = useRef(null);
  const mapRef = useRef(null);
  const [currentPosition, setCurrentPosition] = useState(null);
  const [selectedLocation, setSelectedLocation] = useState(null);
  const [rideInfo, setRideInfo] = useState(null);

  // Get user's location
  useEffect(() => {
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setCurrentPosition([
          position.coords.longitude,
          position.coords.latitude,
        ]);
      },
      (err) => console.error(err),
      { enableHighAccuracy: true }
    );
  }, []);

  // Initialize map + handle clicks
  useEffect(() => {
    if (!currentPosition) return;

    const map = new mapboxgl.Map({
      container: mapContainerRef.current,
      style: "mapbox://styles/mapbox/streets-v11",
      center: currentPosition,
      zoom: 13,
    });

    mapRef.current = map;

    const currentMarker = new mapboxgl.Marker({ color: "blue" })
      .setLngLat(currentPosition)
      .addTo(map);

    map.on("click", (e) => {
      const { lng, lat } = e.lngLat;
      setSelectedLocation({ longitude: lng, latitude: lat });
    });

    return () => map.remove();
  }, [currentPosition]);

  // Draw route when destination selected
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !currentPosition || !selectedLocation) return;

    const origin = currentPosition;
    const dest = [selectedLocation.longitude, selectedLocation.latitude];

    // Clear previous route
    if (map.getSource("route")) {
      map.removeLayer("route");
      map.removeSource("route");
    }

    // Add destination marker
    new mapboxgl.Marker({ color: "red" }).setLngLat(dest).addTo(map);

    const getRoute = async () => {
      const res = await fetch(
        `https://api.mapbox.com/directions/v5/mapbox/driving/${origin[0]},${origin[1]};${dest[0]},${dest[1]}?geometries=geojson&access_token=${mapboxgl.accessToken}`
      );
      const data = await res.json();
      const route = data.routes[0].geometry.coordinates;

      map.addSource("route", {
        type: "geojson",
        data: {
          type: "Feature",
          properties: {},
          geometry: {
            type: "LineString",
            coordinates: route,
          },
        },
      });

      map.addLayer({
        id: "route",
        type: "line",
        source: "route",
        layout: { "line-join": "round", "line-cap": "round" },
        paint: {
          "line-color": "#3b82f6",
          "line-width": 5,
        },
      });
    };

    getRoute();
  }, [selectedLocation]);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;

    const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

    ws.onopen = () => {
      console.log("✅ WebSocket connected");
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        console.log("📨 WebSocket message:", message);

        if (message.type === "ride_response") {
          const { status, driver_id, ride_id } = message.payload;
          alert(`🚕 Ride ${status} by driver ${driver_id}`);
          // Optional: Update rideInfo or show a banner/card
          setRideInfo((prev) => ({
            ...prev,
            status,
            driver_id,
          }));
        }
      } catch (error) {
        console.error("Invalid WebSocket message:", event.data);
      }
    };

    ws.onerror = (err) => {
      console.error("❌ WebSocket error:", err);
    };

    ws.onclose = () => {
      console.log("🔌 WebSocket disconnected");
    };

    return () => ws.close();
  }, []);

  const handleRequestRide = async () => {
    if (!currentPosition || !selectedLocation)
      return alert("Select a destination first");

    try {
      const token = localStorage.getItem("token");
      const res = await axios.post(
        "http://localhost:8080/rides/",
        {
          start_lat: currentPosition[1],
          start_lng: currentPosition[0],
          end_lat: selectedLocation.latitude,
          end_lng: selectedLocation.longitude,
          vehicle_type: "car", // Can add dropdown later
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      const { distance, duration, fare } = res.data;
      setRideInfo({ distance, duration, fare });
    } catch (err) {
      console.error(err);
      alert("Failed to request ride");
    }
  };

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-2xl font-semibold">Rider Dashboard</h1>

      <div ref={mapContainerRef} className="h-[500px] w-full rounded shadow" />

      {selectedLocation && (
        <button
          onClick={handleRequestRide}
          className="bg-blue-600 text-white px-4 py-2 rounded shadow"
        >
          Request Ride
        </button>
      )}

      {rideInfo && (
        <div className="mt-4 p-4 border rounded bg-gray-50 shadow-sm space-y-2">
          <h2 className="text-lg font-medium">Ride Info</h2>
          <p>🛣️ Distance: {rideInfo.distance} km</p>
          <p>⏱️ Duration: {rideInfo.duration} mins</p>
          <p>💰 Fare: ₹{rideInfo.fare}</p>
        </div>
      )}

      {rideInfo?.status && (
        <div className="mt-4 p-4 border rounded bg-green-50 shadow-sm space-y-2">
          <p className="text-green-700 font-medium">
            🚕 Ride has been <strong>{rideInfo.status}</strong> by Driver{" "}
            {rideInfo.driver_id}
          </p>
        </div>
      )}
    </div>
  );
};

export default RiderDashboard;
