import type { FlightResult } from "@/lib/types";
import { formatDuration, formatTime } from "@/lib/formatters";
import Button from "@/components/ui/Button";

interface FlightResultsTableProps {
  results: FlightResult[];
  onMonitor?: (origin: string, destination: string, price: number) => void;
}

export default function FlightResultsTable({
  results,
  onMonitor,
}: FlightResultsTableProps) {
  if (results.length === 0) return null;

  return (
    <div className="overflow-x-auto rounded-[var(--radius-lg)] border border-[var(--border-default)] bg-white">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--border-default)] text-left text-[var(--text-secondary)]">
            <th className="py-3 px-4 font-medium">Price</th>
            <th className="py-3 px-4 font-medium">Airline</th>
            <th className="py-3 px-4 font-medium">Flight</th>
            <th className="py-3 px-4 font-medium">Route</th>
            <th className="py-3 px-4 font-medium">Departure</th>
            <th className="py-3 px-4 font-medium">Duration</th>
            <th className="py-3 px-4 font-medium">Stops</th>
            {onMonitor && <th className="py-3 px-4 font-medium">Action</th>}
          </tr>
        </thead>
        <tbody>
          {results.map((f, i) => (
            <tr
              key={i}
              className="border-b border-[var(--border-default)] last:border-0 hover:bg-slate-50/50 transition-colors"
            >
              <td className="py-3 px-4 font-semibold text-emerald-700">
                ${f.price}
              </td>
              <td className="py-3 px-4 text-[var(--text-primary)]">{f.airline}</td>
              <td className="py-3 px-4 text-[var(--text-tertiary)]">
                {f.flight_number}
              </td>
              <td className="py-3 px-4 text-[var(--text-primary)]">
                {f.departure_code} &rarr; {f.arrival_code}
              </td>
              <td className="py-3 px-4 text-[var(--text-primary)]">
                {formatTime(f.departure)}
              </td>
              <td className="py-3 px-4 text-[var(--text-primary)]">
                {formatDuration(f.duration_minutes)}
              </td>
              <td className="py-3 px-4">
                {f.stops === 0 ? (
                  <span className="text-emerald-600 font-medium">Direct</span>
                ) : (
                  <span className="text-amber-600">
                    {f.stops} stop{f.stops > 1 ? "s" : ""}
                  </span>
                )}
              </td>
              {onMonitor && (
                <td className="py-3 px-4">
                  <Button
                    variant="success"
                    size="sm"
                    onClick={() =>
                      onMonitor(f.departure_code, f.arrival_code, f.price)
                    }
                  >
                    Monitor
                  </Button>
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
