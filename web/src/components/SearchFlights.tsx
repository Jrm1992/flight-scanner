"use client";

import { useState } from "react";
import { searchFlights } from "@/lib/api";
import type { FlightResult } from "@/lib/types";

export default function SearchFlights() {
  const [origin, setOrigin] = useState("");
  const [destination, setDestination] = useState("");
  const [date, setDate] = useState("");
  const [results, setResults] = useState<FlightResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const data = await searchFlights(origin, destination, date || undefined);
      setResults(data.results);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Search failed");
      setResults([]);
    } finally {
      setLoading(false);
    }
  }

  function formatDuration(min: number) {
    return `${Math.floor(min / 60)}h${String(min % 60).padStart(2, "0")}m`;
  }

  function formatTime(iso: string) {
    if (!iso) return "";
    return new Date(iso).toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Search Flights</h2>

      <form onSubmit={handleSearch} className="flex flex-wrap gap-3 mb-6">
        <input
          type="text"
          placeholder="Origin (e.g. GIG)"
          value={origin}
          onChange={(e) => setOrigin(e.target.value.toUpperCase())}
          maxLength={3}
          className="border border-gray-300 rounded px-3 py-2 w-28 uppercase bg-white text-gray-900"
          required
        />
        <input
          type="text"
          placeholder="Dest (e.g. SCL)"
          value={destination}
          onChange={(e) => setDestination(e.target.value.toUpperCase())}
          maxLength={3}
          className="border border-gray-300 rounded px-3 py-2 w-28 uppercase bg-white text-gray-900"
          required
        />
        <input
          type="date"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          className="border border-gray-300 rounded px-3 py-2 bg-white text-gray-900"
        />
        <button
          type="submit"
          disabled={loading}
          className="bg-blue-600 text-white px-5 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? "Searching..." : "Search"}
        </button>
      </form>

      {error && <p className="text-red-500 mb-4">{error}</p>}

      {results.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="border-b border-gray-200 text-left text-gray-600">
                <th className="py-2 pr-4">Price</th>
                <th className="py-2 pr-4">Airline</th>
                <th className="py-2 pr-4">Flight</th>
                <th className="py-2 pr-4">Route</th>
                <th className="py-2 pr-4">Departure</th>
                <th className="py-2 pr-4">Duration</th>
                <th className="py-2">Stops</th>
              </tr>
            </thead>
            <tbody>
              {results.map((f, i) => (
                <tr key={i} className="border-b border-gray-100 hover:bg-gray-50">
                  <td className="py-2 pr-4 font-semibold text-green-700">
                    ${f.price}
                  </td>
                  <td className="py-2 pr-4">{f.airline}</td>
                  <td className="py-2 pr-4 text-gray-500">{f.flight_number}</td>
                  <td className="py-2 pr-4">
                    {f.departure_code} → {f.arrival_code}
                  </td>
                  <td className="py-2 pr-4">{formatTime(f.departure)}</td>
                  <td className="py-2 pr-4">{formatDuration(f.duration_minutes)}</td>
                  <td className="py-2">
                    {f.stops === 0 ? (
                      <span className="text-green-600">Direct</span>
                    ) : (
                      <span className="text-orange-500">{f.stops} stop{f.stops > 1 ? "s" : ""}</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
