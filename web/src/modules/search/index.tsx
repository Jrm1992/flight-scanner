"use client";

import { useSearchModel } from "./model";
import SearchView from "./view";

interface SearchProps {
  onMonitor?: (origin: string, destination: string, price: number) => void;
}

export default function Search({ onMonitor }: SearchProps) {
  const model = useSearchModel();

  return (
    <SearchView
      origin={model.origin}
      onOriginChange={model.setOrigin}
      destination={model.destination}
      onDestinationChange={model.setDestination}
      date={model.date}
      onDateChange={model.setDate}
      currency={model.currency}
      onCurrencyChange={model.setCurrency}
      onSubmit={model.handleSearch}
      loading={model.loading}
      error={model.error}
      results={model.results}
      onMonitor={onMonitor}
    />
  );
}
