// MapView.js
import React, { useEffect, useRef } from "react";
import mapboxgl from "mapbox-gl";

mapboxgl.accessToken = "YOUR_MAPBOX_ACCESS_TOKEN"; // Use your actual token here

function MapView({ latitude, longitude, destination }) {
  const mapContainer = useRef(null);
  const map = useRef(null);

  useEffect(() => {
    if (!latitude || !longitude || !destination) return;

    const origin = [longitude, latitude];
    const dest = [destination.longitude, destination.latitude];

    if (map.current) {
      map.current.remove(); // Reset map if already created
    }

    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: "mapbox://styles/mapbox/streets-v11",
      center: origin,
      zoom: 13,
    });

    new mapboxgl.Marker({ color: "blue" }).setLngLat(origin).addTo(map.current);

    new mapboxgl.Marker({ color: "red" }).setLngLat(dest).addTo(map.current);

    // Fetch directions from Mapbox Directions API
    const fetchRoute = async () => {
      const res = await fetch(
        `https://api.mapbox.com/directions/v5/mapbox/driving/${origin[0]},${origin[1]};${dest[0]},${dest[1]}?geometries=geojson&access_token=${mapboxgl.accessToken}`
      );
      const data = await res.json();
      const route = data.routes[0].geometry.coordinates;

      map.current.addSource("route", {
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

      map.current.addLayer({
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

    map.current.on("load", fetchRoute);
  }, [latitude, longitude, destination]);

  return <div ref={mapContainer} className="h-[500px] w-full rounded shadow" />;
}

export default MapView;
