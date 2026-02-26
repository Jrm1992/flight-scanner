import type { Route } from "@/lib/types";

interface AlertFiltersProps {
  filter: "all" | "unread" | "read";
  onFilterChange: (v: string) => void;
  routeFilter: string;
  onRouteFilterChange: (v: string) => void;
  routes: Route[];
}

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
        className="rounded-[var(--radius-md)] border border-[var(--border-default)] bg-white px-3 py-2 text-sm text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--brand-500)]/25 focus:border-[var(--brand-500)]"
      >
        <option value="all">All alerts</option>
        <option value="unread">Unread</option>
        <option value="read">Read</option>
      </select>
      <select
        value={routeFilter}
        onChange={(e) => onRouteFilterChange(e.target.value)}
        className="rounded-[var(--radius-md)] border border-[var(--border-default)] bg-white px-3 py-2 text-sm text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--brand-500)]/25 focus:border-[var(--brand-500)]"
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
