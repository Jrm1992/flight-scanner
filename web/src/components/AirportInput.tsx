"use client";

import { useState, useRef, useEffect } from "react";
import { searchAirports, airports, type Airport } from "@/lib/airports";

interface Props {
  value: string;
  onChange: (code: string) => void;
  placeholder?: string;
  required?: boolean;
  className?: string;
}

export default function AirportInput({
  value,
  onChange,
  placeholder = "Airport (e.g. GIG)",
  required,
  className = "",
}: Props) {
  const [query, setQuery] = useState(value);
  const [results, setResults] = useState<Airport[]>([]);
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (value) {
      const match = airports.find((a) => a.code === value);
      setQuery(match ? `${match.code} - ${match.city}` : value);
    } else {
      setQuery("");
    }
  }, [value]);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  function handleInput(val: string) {
    setQuery(val);
    const matches = searchAirports(val);
    setResults(matches);
    setOpen(matches.length > 0);
    setActiveIndex(-1);
    // If it looks like an IATA code (3 uppercase letters), send it directly
    const trimmed = val.trim().toUpperCase();
    if (/^[A-Z]{3}$/.test(trimmed)) {
      onChange(trimmed);
    }
  }

  function handleSelect(airport: Airport) {
    setQuery(`${airport.code} - ${airport.city}`);
    onChange(airport.code);
    setOpen(false);
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (!open) return;
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActiveIndex((i) => Math.min(i + 1, results.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActiveIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && activeIndex >= 0) {
      e.preventDefault();
      handleSelect(results[activeIndex]);
    } else if (e.key === "Escape") {
      setOpen(false);
    }
  }

  return (
    <div ref={wrapperRef} className="relative">
      <input
        type="text"
        value={query}
        onChange={(e) => handleInput(e.target.value)}
        onFocus={() => {
          if (value) {
            setQuery("");
            handleInput("");
          }
        }}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        required={required}
        className={`border border-gray-300 rounded px-3 py-2 bg-white text-gray-900 ${className}`}
      />
      {open && results.length > 0 && (
        <ul className="absolute z-20 top-full left-0 mt-1 w-64 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto">
          {results.map((a, i) => (
            <li
              key={a.code}
              onMouseDown={() => handleSelect(a)}
              className={`px-3 py-2 cursor-pointer text-sm flex justify-between ${
                i === activeIndex ? "bg-blue-50 text-blue-700" : "hover:bg-gray-50"
              }`}
            >
              <span>
                <span className="font-semibold">{a.code}</span>{" "}
                <span className="text-gray-500">{a.city}</span>
              </span>
              <span className="text-gray-400 text-xs">{a.country}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
