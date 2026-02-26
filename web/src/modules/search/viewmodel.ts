import { useState } from "react";
import { useSearchModel } from "./model";

export function useSearchViewModel() {
  const model = useSearchModel();
  const [origin, setOrigin] = useState("");
  const [destination, setDestination] = useState("");
  const [date, setDate] = useState("");

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    await model.search(origin, destination, date || undefined);
  }

  return {
    origin,
    setOrigin,
    destination,
    setDestination,
    date,
    setDate,
    results: model.results,
    loading: model.loading,
    error: model.error,
    handleSearch,
  };
}
