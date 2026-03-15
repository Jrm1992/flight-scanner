import { useState, useCallback } from "react";
import { searchFlights } from "@/lib/api";
import type { FlightResult } from "@/lib/types";

export function useSearchModel() {
  const [results, setResults] = useState<FlightResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [origin, setOrigin] = useState("");
  const [destination, setDestination] = useState("");
  const [date, setDate] = useState("");

  const handleSearch = useCallback(
    async (e: React.FormEvent) => {
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
    },
    [origin, destination, date]
  );

  return {
    results,
    loading,
    error,
    origin,
    setOrigin,
    destination,
    setDestination,
    date,
    setDate,
    handleSearch,
  };
}
