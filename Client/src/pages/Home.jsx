import React, { useState } from "react";
import { Link } from "react-router-dom";

import { GiAbstract109 } from "react-icons/gi";
const Home = () => {
  const [showVehicles, setShowVehicles] = useState(false);

  const vehicles = [
    {
      name: "Bike",
      img: "https://cdn.pixabay.com/photo/2018/05/03/23/22/motorcycle-3372805_1280.png",
    },
    {
      name: "Car",
      img: "https://cdn.pixabay.com/photo/2013/07/12/12/45/car-146185_1280.png",
    },
    {
      name: "Premium Car",
      img: "https://cdn.pixabay.com/photo/2015/09/12/21/31/car-937414_1280.png",
    },
  ];

  return (
    <div className="h-screen w-screen bg-black overflow-hidden">
      <div
        className="relative h-full w-full bg-center bg-cover bg-no-repeat 
        bg-[url('https://images.unsplash.com/photo-1681524415449-14c88372eb08?q=80&w=1924&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D')]"
      >
        <div
          className="absolute top-10 left-1/2 transform -translate-x-1/2 w-4/5 h-[90%] 
            border-t-4 border-l-4 border-r-4 border-white 
            rounded-2xl shadow-2xl backdrop-blur-md bg-black/20 p-6"
        >
          {/* Navbar */}
          <nav className="flex justify-between items-center border-b border-white pb-4">
            <div className="flex items-center gap-2">
              <GiAbstract109 className="text-amber text-3xl bg-white mr-2" />
              {/* <div className="w-8 h-8 bg-white rounded-full" /> */}
              <span className="text-white text-2xl font-semibold">
                SwiftRide
              </span>
            </div>

            <div className="flex items-center gap-4">
              <Link
                to={"/login"}
                className="bg-black text-white px-6 py-2 rounded-full shadow-lg hover:bg-gray-800 transition duration-200 hover:underline"
              >
                Login
              </Link>
              <Link
                to={"/signup"}
                className="bg-white text-black px-4 py-2 rounded-full shadow-lg hover:bg-gray-400 transition duration-200"
              >
                Sign Up
              </Link>
            </div>
          </nav>

          {/* Hero Section */}
          <div className="flex flex-col items-center justify-center text-center text-white h-full mt-2">
            <h1 className="text-5xl font-extrabold tracking-tight mb-4 drop-shadow-md">
              Your Ride, Your Way.
            </h1>
            <p className="text-lg text-white/80 mt-4 mb-6 max-w-xl">
              SwiftRide connects you to fast, reliable, and affordable
              rides—anytime, anywhere.
            </p>

            {/* Show Vehicles Button */}
            <button
              onClick={() => setShowVehicles((prev) => !prev)}
              className="bg-white text-black px-6 py-2 rounded-full shadow-md hover:bg-gray-300 transition duration-200"
            >
              {showVehicles ? "Hide Vehicles" : "Show Vehicles"}
            </button>

            {/* Vehicles Carousel */}
            {showVehicles && (
              <div className="mt-10 flex justify-center gap-8 flex-wrap">
                {vehicles.map((vehicle, idx) => (
                  <div
                    key={idx}
                    className="w-78 bg-white/30 rounded-xl shadow-md p-3 text-center hover:scale-105 transition-transform"
                  >
                    <img
                      src={vehicle.img}
                      alt={vehicle.name}
                      className="w-full h-36 object-cover rounded-md mb-2"
                    />
                    <p className="font-bold text-lg text-gray-800">
                      {vehicle.name}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Home;
