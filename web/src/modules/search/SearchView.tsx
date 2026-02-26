"use client";

import { useSearchViewModel } from "./viewmodel";
import SearchForm from "./SearchForm";
import FlightResultsTable from "./FlightResultsTable";

interface SearchViewProps {
  onMonitor?: (origin: string, destination: string, price: number) => void;
}

export default function SearchView({ onMonitor }: SearchViewProps) {
  const vm = useSearchViewModel();

  return (
    <div>
      <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-5">
        Search Flights
      </h2>

      <SearchForm
        origin={vm.origin}
        onOriginChange={vm.setOrigin}
        destination={vm.destination}
        onDestinationChange={vm.setDestination}
        date={vm.date}
        onDateChange={vm.setDate}
        onSubmit={vm.handleSearch}
        loading={vm.loading}
      />

      {vm.error && (
        <p className="text-[var(--color-danger)] mb-4 text-sm">{vm.error}</p>
      )}

      <FlightResultsTable results={vm.results} onMonitor={onMonitor} />
    </div>
  );
}
