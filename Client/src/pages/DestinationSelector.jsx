import {
  Combobox,
  ComboboxInput,
  ComboboxOptions,
  ComboboxOption,
} from "@headlessui/react";
import { useState } from "react";
import { predefinedLocations } from "../data/locations"; // your existing locations

const DestinationSelector = ({ onSelect }) => {
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState(null);

  const filteredLocations =
    query === ""
      ? predefinedLocations
      : predefinedLocations.filter((loc) =>
          loc.name.toLowerCase().includes(query.toLowerCase())
        );

  const handleSelect = (location) => {
    console.log(location);
    setSelected(location);
    onSelect({
      latitude: location.latitude,
      longitude: location.longitude,
    });
  };

  return (
    <div className="space-y-2">
      <h2 className="text-xl font-semibold">Enter your destination</h2>
      <Combobox value={selected} onChange={handleSelect}>
        <div className="relative">
          <ComboboxInput
            className="w-full px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
            displayValue={(loc) => loc?.name || ""}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search locations..."
          />
          <ComboboxOptions className="absolute z-10 mt-1 w-full bg-white border border-gray-200 rounded shadow-lg max-h-60 overflow-y-auto">
            {filteredLocations.length === 0 ? (
              <div className="p-2 text-gray-500">No results found</div>
            ) : (
              filteredLocations.map((loc, index) => (
                <ComboboxOption
                  key={index}
                  value={loc}
                  className={({ active }) =>
                    `cursor-pointer px-4 py-2 ${
                      active ? "bg-blue-100 text-blue-900" : "text-gray-700"
                    }`
                  }
                >
                  📍 {loc.name}
                </ComboboxOption>
              ))
            )}
          </ComboboxOptions>
        </div>
      </Combobox>
    </div>
  );
};

export default DestinationSelector;
