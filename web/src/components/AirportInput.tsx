"use client";

import { useState, useRef } from "react";
import { searchAirports, airports, type Airport } from "@/lib/airports";

interface Props {
  value: string;
  onChange: (code: string) => void;
  placeholder?: string;
  required?: boolean;
  className?: string;
}

function displayValue(code: string): string {
  if (!code) return "";
  const match = airports.find((a) => a.code === code);
  return match ? `${match.code} - ${match.city}` : code;
}

export default function AirportInput({
  value,
  onChange,
  placeholder = "Airport (e.g. GIG)",
  required,
  className = "",
}: Props) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Airport[]>([]);
  const [open, setOpen] = useState(false);
  const [focused, setFocused] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const wrapperRef = useRef<HTMLDivElement>(null);

  function handleInput(val: string) {
    setQuery(val);
    const matches = searchAirports(val);
    setResults(matches);
    setOpen(matches.length > 0);
    setActiveIndex(-1);
  }

  function handleSelect(airport: Airport) {
    setQuery("");
    onChange(airport.code);
    setOpen(false);
    setFocused(false);
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

  function handleBlur(e: React.FocusEvent) {
    if (!wrapperRef.current?.contains(e.relatedTarget as Node)) {
      const trimmed = query.trim().toUpperCase();
      if (/^[A-Z]{3}$/.test(trimmed) && airports.some((a) => a.code === trimmed)) {
        onChange(trimmed);
      }
      setOpen(false);
      setFocused(false);
      setQuery("");
    }
  }

  return (
    <div ref={wrapperRef} className="relative" tabIndex={-1} onBlur={handleBlur}>
      <input
        type="text"
        value={focused ? query : displayValue(value)}
        onChange={(e) => handleInput(e.target.value)}
        onFocus={() => {
          setFocused(true);
          setQuery("");
          handleInput("");
        }}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        required={required}
        className={`border border-border rounded-md px-3 py-2 bg-white/5 text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-cyan-500/25 focus:border-cyan-500/50 transition-colors text-sm ${className}`}
      />
      {open && results.length > 0 && (
        <ul className="absolute z-20 top-full left-0 mt-1 w-64 bg-[#1e293b] border border-white/10 rounded-lg shadow-[0_0_25px_rgba(0,0,0,0.5)] backdrop-blur-xl max-h-48 overflow-y-auto">
          {results.map((a, i) => (
            <li
              key={a.code}
              onMouseDown={() => handleSelect(a)}
              className={`px-3 py-2 cursor-pointer text-sm flex justify-between ${
                i === activeIndex ? "bg-cyan-500/15 text-cyan-400" : "hover:bg-white/10 text-foreground"
              }`}
            >
              <span>
                <span className="font-semibold">{a.code}</span>{" "}
                <span className="text-muted">{a.city}</span>
              </span>
              <span className="text-muted-foreground text-xs">{a.country}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
