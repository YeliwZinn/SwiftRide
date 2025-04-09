import React, { useEffect, useRef, useState } from "react";
import mapboxgl from "mapbox-gl";
import axios from "axios";
import DestinationSelector from "./DestinationSelector";

mapboxgl.accessToken =
  "pk.eyJ1Ijoic3VieHViZXIiLCJhIjoiY204eXIxNGI5MDNvaTJsczgyYXJtenZyOCJ9.ba7NK0E5Ce_N3hHyFRzJTg";

const RiderDashboard = () => {
  const mapContainerRef = useRef(null);
  const mapRef = useRef(null);
  const [currentPosition, setCurrentPosition] = useState(null);
  const [selectedLocation, setSelectedLocation] = useState(null);
  const [rideInfo, setRideInfo] = useState(null);
  const [profileData, setProfileData] = useState(null);
  const [showProfile, setShowProfile] = useState(false);

  useEffect(() => {
    const fetchProfile = async () => {
      const token = localStorage.getItem("riderToken");
      if (!token) return;

      try {
        const res = await axios.get("http://localhost:8080/profile", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        setProfileData(res.data);
      } catch (err) {
        console.error("Failed to fetch profile:", err);
      }
    };

    fetchProfile();
  }, []);

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

  useEffect(() => {
    if (!currentPosition) return;

    const map = new mapboxgl.Map({
      container: mapContainerRef.current,
      style: "mapbox://styles/mapbox/streets-v11",
      center: currentPosition,
      zoom: 13,
    });

    mapRef.current = map;

    new mapboxgl.Marker({ color: "blue" })
      .setLngLat(currentPosition)
      .addTo(map);

    map.on("click", (e) => {
      const { lng, lat } = e.lngLat;
      setSelectedLocation({ longitude: lng, latitude: lat });
    });

    return () => map.remove();
  }, [currentPosition]);

  useEffect(() => {
    const map = mapRef.current;
    if (!map || !currentPosition || !selectedLocation) return;

    const origin = currentPosition;
    const dest = [selectedLocation.longitude, selectedLocation.latitude];

    if (map.getSource("route")) {
      map.removeLayer("route");
      map.removeSource("route");
    }

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
    const token = localStorage.getItem("riderToken");
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
          const { status, driver_id, driver_name, ride_id } = message.payload;
          alert(`🚕 Ride ${status} by driver ${driver_name}`);
          setRideInfo((prev) => ({
            ...prev,
            status,
            driver_id,
            driver_name,
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
      const token = localStorage.getItem("riderToken");
      const res = await axios.post(
        "http://localhost:8080/rides/",
        {
          start_lat: currentPosition[1],
          start_lng: currentPosition[0],
          end_lat: selectedLocation.latitude,
          end_lng: selectedLocation.longitude,
          vehicle_type: "car",
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

  const handleLogout = () => {
    localStorage.removeItem("riderToken");
    window.location.href = "/";
  };

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Navbar */}
      <nav className="bg-white shadow px-6 py-4 flex items-center justify-between">
        <a href="/" className="text-black text-2xl font-semibold">
          SwiftRide
        </a>

        {profileData && (
          <div
            className="relative flex items-center gap-4"
            onMouseEnter={() => setShowProfile(true)}
            onMouseLeave={() => setShowProfile(false)}
          >
            <div className="w-10 h-10 rounded-full bg-blue-500 text-white flex items-center justify-center cursor-pointer">
              {profileData.name[0]?.toUpperCase()}
            </div>

            {showProfile && (
              <div className="absolute right-16 mt-2 w-60 bg-white border shadow-lg rounded p-4 space-y-1 z-10">
                <p className="text-sm font-semibold text-gray-700">
                  {profileData.name}
                </p>
                <p className="text-sm text-gray-600">{profileData.email}</p>
                <p className="text-sm text-gray-600">{profileData.phone}</p>
              </div>
            )}

            <button
              onClick={handleLogout}
              className="text-sm bg-red-500 hover:bg-red-600 text-white px-3 py-1 rounded"
            >
              Logout
            </button>
          </div>
        )}
      </nav>

      {/* Dashboard */}
      <div className="grid grid-cols-5 gap-4 p-6">
        {/* Left panel (2/5) */}
        <div className="col-span-2 bg-white rounded shadow p-4 space-y-4">
          <DestinationSelector onSelect={setSelectedLocation} />

          {selectedLocation && (
            <button
              onClick={handleRequestRide}
              className="w-full bg-blue-600 text-white px-4 py-2 rounded shadow mt-4"
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
                {rideInfo.driver_name}
              </p>
            </div>
          )}
        </div>

        {/* Map panel (3/5) */}
        <div className="col-span-3">
          <div
            ref={mapContainerRef}
            className="h-[600px] w-full rounded shadow"
          />
        </div>
      </div>
    </div>
  );
};

export default RiderDashboard;
