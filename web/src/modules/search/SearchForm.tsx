import AirportInput from "@/components/AirportInput";
import Button from "@/components/ui/Button";

interface SearchFormProps {
  origin: string;
  onOriginChange: (v: string) => void;
  destination: string;
  onDestinationChange: (v: string) => void;
  date: string;
  onDateChange: (v: string) => void;
  onSubmit: (e: React.FormEvent) => void;
  loading: boolean;
}

export default function SearchForm({
  origin,
  onOriginChange,
  destination,
  onDestinationChange,
  date,
  onDateChange,
  onSubmit,
  loading,
}: SearchFormProps) {
  return (
    <form onSubmit={onSubmit} className="flex flex-wrap items-end gap-3 mb-6">
      <AirportInput
        value={origin}
        onChange={onOriginChange}
        placeholder="Origin (e.g. GIG)"
        required
        className="w-28"
      />
      <AirportInput
        value={destination}
        onChange={onDestinationChange}
        placeholder="Dest (e.g. SCL)"
        required
        className="w-28"
      />
      <input
        type="date"
        value={date}
        onChange={(e) => onDateChange(e.target.value)}
        className="rounded-[var(--radius-md)] border border-[var(--border-default)] bg-white px-3 py-2 text-sm text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--brand-500)]/25 focus:border-[var(--brand-500)]"
      />
      <Button type="submit" loading={loading}>
        {loading ? "Searching..." : "Search"}
      </Button>
    </form>
  );
}
