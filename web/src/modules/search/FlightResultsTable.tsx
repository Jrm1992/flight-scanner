"use client";

import type { FlightResult } from "@/lib/types";
import { formatDuration, formatTime } from "@/lib/formatters";
import Button from "@/components/ui/Button";
import { motion } from "framer-motion";

interface FlightResultsTableProps {
  results: FlightResult[];
  onMonitor?: (origin: string, destination: string, price: number) => void;
}

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

const rowVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.3 } },
};

export default function FlightResultsTable({
  results,
  onMonitor,
}: FlightResultsTableProps) {
  if (results.length === 0) return null;

  return (
    <div className="overflow-x-auto rounded-lg border border-border bg-white/5 backdrop-blur-xl">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border text-left text-muted">
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
        <motion.tbody
          variants={containerVariants}
          initial="hidden"
          animate="show"
        >
          {results.map((f, i) => (
            <motion.tr
              key={i}
              variants={rowVariants}
              className="border-b border-border last:border-0 hover:bg-white/5 transition-colors duration-150"
            >
              <td className="py-3 px-4 font-semibold text-emerald-400 font-data">
                ${f.price}
              </td>
              <td className="py-3 px-4 text-foreground">{f.airline}</td>
              <td className="py-3 px-4 text-muted-foreground">
                {f.flight_number}
              </td>
              <td className="py-3 px-4 text-foreground">
                {f.departure_code} &rarr; {f.arrival_code}
              </td>
              <td className="py-3 px-4 text-foreground">
                {formatTime(f.departure)}
              </td>
              <td className="py-3 px-4 text-foreground">
                {formatDuration(f.duration_minutes)}
              </td>
              <td className="py-3 px-4">
                {f.stops === 0 ? (
                  <span className="text-emerald-400 font-medium">Direct</span>
                ) : (
                  <span className="text-amber-400">
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
            </motion.tr>
          ))}
        </motion.tbody>
      </table>
    </div>
  );
}
