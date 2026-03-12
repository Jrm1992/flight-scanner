import type { SortKey } from "./viewmodel";

interface RouteSortBarProps {
  sortKey: SortKey;
  sortDir: "asc" | "desc";
  onSort: (key: SortKey) => void;
  sortIndicator: (key: SortKey) => string;
}

const sortOptions: [SortKey, string][] = [
  ["status", "Status"],
  ["origin", "Origin"],
  ["destination", "Dest"],
  ["current_price", "Price"],
  ["alert_price", "Alert"],
];

export default function RouteSortBar({
  sortKey,
  onSort,
  sortIndicator,
}: RouteSortBarProps) {
  return (
    <div className="flex items-center gap-2 mb-4 text-xs text-[var(--text-tertiary)]">
      <span className="font-medium">Sort by:</span>
      {sortOptions.map(([key, label]) => (
        <button
          key={key}
          onClick={() => onSort(key)}
          className={`px-2.5 py-1 rounded-[var(--radius-md)] transition-colors duration-[var(--transition-fast)] ${
            sortKey === key
              ? "bg-cyan-500/15 text-cyan-400 font-medium"
              : "hover:bg-white/5 text-[var(--text-secondary)]"
          }`}
        >
          {label}
          {sortIndicator(key)}
        </button>
      ))}
    </div>
  );
}
