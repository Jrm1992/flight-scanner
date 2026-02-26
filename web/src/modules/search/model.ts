import { useState, useCallback } from "react";
import { searchFlights } from "@/lib/api";
import type { FlightResult } from "@/lib/types";

export function useSearchModel() {
  const [results, setResults] = useState<FlightResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const search = useCallback(
    async (origin: string, destination: string, date?: string) => {
      setError("");
      setLoading(true);
      try {
        const data = await searchFlights(origin, destination, date);
        setResults(data.results);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Search failed");
        setResults([]);
      } finally {
        setLoading(false);
      }
    },
    []
  );

  return { results, loading, error, search };
}
