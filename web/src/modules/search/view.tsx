import SearchForm from "./SearchForm";
import FlightResultsTable from "./FlightResultsTable";
import type { FlightResult } from "@/lib/types";

interface SearchViewProps {
  origin: string;
  onOriginChange: (v: string) => void;
  destination: string;
  onDestinationChange: (v: string) => void;
  date: string;
  onDateChange: (v: string) => void;
  onSubmit: (e: React.FormEvent) => void;
  loading: boolean;
  error: string;
  results: FlightResult[];
  onMonitor?: (origin: string, destination: string, price: number) => void;
}

export default function SearchView({
  origin,
  onOriginChange,
  destination,
  onDestinationChange,
  date,
  onDateChange,
  onSubmit,
  loading,
  error,
  results,
  onMonitor,
}: SearchViewProps) {
  return (
    <div>
      <h2 className="text-xl font-semibold text-foreground mb-5">
        Search Flights
      </h2>

      <SearchForm
        origin={origin}
        onOriginChange={onOriginChange}
        destination={destination}
        onDestinationChange={onDestinationChange}
        date={date}
        onDateChange={onDateChange}
        onSubmit={onSubmit}
        loading={loading}
      />

      {error && (
        <p className="text-[var(--color-danger)] mb-4 text-sm">{error}</p>
      )}

      <FlightResultsTable results={results} onMonitor={onMonitor} />
    </div>
  );
}
