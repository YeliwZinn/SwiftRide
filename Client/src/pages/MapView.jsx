// components/MapView.jsx
import { MapContainer, TileLayer, Marker, Popup, useMap } from "react-leaflet";
import { useEffect, useState, useRef } from "react";
import "leaflet/dist/leaflet.css";
import L from "leaflet";

// Fix default icon bug in Leaflet (React Leaflet doesn't auto handle this)
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl:
    "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png",
  iconUrl:
    "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png",
  shadowUrl:
    "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png",
});

// Component to programmatically change map center
const RecenterMap = ({ lat, lng }) => {
  const map = useMap();
  useEffect(() => {
    if (lat && lng) {
      map.setView([lat, lng]);
    }
  }, [lat, lng, map]);
  return null;
};

const MapView = () => {
  const [position, setPosition] = useState(null);

  useEffect(() => {
    // if (!navigator.geolocation) {
    //   console.error("Geolocation is not supported by your browser");
    //   return;
    // }
    // navigator.geolocation.getCurrentPosition(
    //   (pos) => {
    //     console.log("✅ Success! Your location is:", pos.coords);
    //   },
    //   (err) => {
    //     console.error("❌ Geolocation error:", err);
    //   },
    //   {
    //     enableHighAccuracy: false,
    //     timeout: 30000, // 30 seconds
    //     maximumAge: 0,
    //   }
    // );

    const watchId = navigator.geolocation.watchPosition(
      (pos) => {
        const { latitude, longitude } = pos.coords;
        setPosition([latitude, longitude]);
      },
      (err) => {
        console.error("Geolocation error:", err.message || err);
      },
      {
        enableHighAccuracy: true,
        maximumAge: 0,
        timeout: 20000,
      }
    );

    return () => {
      navigator.geolocation.clearWatch(watchId);
    };
  }, []);

  return (
    <div>
      {!position ? (
        <p>Fetching your live location...</p>
      ) : (
        <MapContainer
          center={position}
          zoom={20}
          scrollWheelZoom={true}
          style={{ height: "400px", width: "100%", borderRadius: "12px" }}
        >
          <TileLayer
            attribution='&copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors'
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          />
          <Marker position={position}>
            <Popup>You are here 📍</Popup>
          </Marker>
          <RecenterMap lat={position[0]} lng={position[1]} />
        </MapContainer>
      )}
    </div>
  );
};

export default MapView;

// const MapView = ({ latitude, longitude }) => {
//   const position = [latitude || 20.3555, longitude || 85.8161];
//   return (
//     <MapContainer
//       center={position} // KIIT as default
//       zoom={20}
//       scrollWheelZoom={false}
//       style={{ height: "400px", width: "100%", borderRadius: "12px" }}
//     >
//       <TileLayer
//         attribution='&copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors'
//         url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
//       />
//       <Marker position={position}>
//         <Popup>You are here 📍</Popup>
//       </Marker>
//     </MapContainer>
//   );
// };

// export default MapView;
