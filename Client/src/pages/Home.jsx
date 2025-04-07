import React from "react";
import { Link } from "react-router-dom";

const Home = () => {
  return (
    <div>
      <div className="bg-cover bg-bottom bg-[url(https://images.unsplash.com/photo-1612867754336-c45d074c4f8e?q=80&w=1965&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D)] h-screen pt-8 w-screen flex justify-between flex-col">
        <img
          src="https://upload.wikimedia.org/wikipedia/commons/thumb/5/58/Uber_logo_2018.svg/1200px-Uber_logo_2018.svg.png"
          className="w-15 ml-8 bg-cover"
        />
        <div className="pb-7 bg-black text-white py-5 px-4">
          <h2 className="text-2xl font-bold">Get Started with RideCalling</h2>
          <Link
            to="/login"
            className="flex items-center justify-center w-full py-3 !text-black !text-xl !font-bold !bg-gray-300 rounded mt-6"
          >
            Continue
          </Link>
        </div>
      </div>
    </div>
  );
};

export default Home;
