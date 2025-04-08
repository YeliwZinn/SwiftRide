import { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet-routing-machine";
import { useMap } from "react-leaflet";

const Routing = ({ from, to }) => {
  const map = useMap();
  const routingControlRef = useRef(null);

  useEffect(() => {
    // Avoid triggering before map and coords are ready
    if (!map || !from || !to) return;

    // If a previous route exists, remove it safely
    if (routingControlRef.current) {
      try {
        map.removeControl(routingControlRef.current);
      } catch (e) {
        console.warn("Error removing routing control:", e.message);
      }
      routingControlRef.current = null;
    }

    // Add a new routing control
    const control = L.Routing.control({
      waypoints: [L.latLng(from[0], from[1]), L.latLng(to[0], to[1])],
      routeWhileDragging: false,
      show: false,
      addWaypoints: false,
      draggableWaypoints: false,
      fitSelectedRoutes: true,
      lineOptions: {
        styles: [{ color: "#0074D9", weight: 5, opacity: 0.8 }],
      },
      createMarker: () => null, // don't add extra markers
    });

    try {
      control.addTo(map);
      routingControlRef.current = control;
    } catch (e) {
      console.warn("Error adding routing control:", e.message);
    }

    return () => {
      // Clean up on unmount
      if (routingControlRef.current) {
        try {
          map.removeControl(routingControlRef.current);
        } catch (e) {
          console.warn("Cleanup failed:", e.message);
        }
        routingControlRef.current = null;
      }
    };
  }, [from, to, map]);

  return null;
};

export default Routing;
