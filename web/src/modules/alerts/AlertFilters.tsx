import type { Route } from "@/lib/types";

interface AlertFiltersProps {
  filter: "all" | "unread" | "read";
  onFilterChange: (v: string) => void;
  routeFilter: string;
  onRouteFilterChange: (v: string) => void;
  routes: Route[];
}

const selectClass =
  "rounded-md border border-border bg-white/5 px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-cyan-500/25 focus:border-cyan-500/50";

export default function AlertFilters({
  filter,
  onFilterChange,
  routeFilter,
  onRouteFilterChange,
  routes,
}: AlertFiltersProps) {
  return (
    <div className="flex items-center gap-3 mb-5">
      <select
        value={filter}
        onChange={(e) => onFilterChange(e.target.value)}
        className={selectClass}
      >
        <option value="all">All alerts</option>
        <option value="unread">Unread</option>
        <option value="read">Read</option>
      </select>
      <select
        value={routeFilter}
        onChange={(e) => onRouteFilterChange(e.target.value)}
        className={selectClass}
      >
        <option value="">All routes</option>
        {routes.map((r) => (
          <option key={r.id} value={r.id}>
            {r.origin} &rarr; {r.destination}
          </option>
        ))}
      </select>
    </div>
  );
}
