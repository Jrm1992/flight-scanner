import type { Alert, Route } from "@/lib/types";
import AlertCard from "./AlertCard";
import AlertFilters from "./AlertFilters";
import Spinner from "@/components/ui/Spinner";
import { motion } from "framer-motion";

interface AlertsViewProps {
  alerts: Alert[];
  routes: Route[];
  routeMap: Map<string, Route>;
  loading: boolean;
  filter: "all" | "unread" | "read";
  onFilterChange: (v: string) => void;
  routeFilter: string;
  onRouteFilterChange: (v: string) => void;
  onMarkRead: (id: string) => void;
}

export default function AlertsView({
  alerts,
  routes,
  routeMap,
  loading,
  filter,
  onFilterChange,
  routeFilter,
  onRouteFilterChange,
  onMarkRead,
}: AlertsViewProps) {
  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner />
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-xl font-semibold text-foreground mb-5">
        Alerts
      </h2>

      <AlertFilters
        filter={filter}
        onFilterChange={onFilterChange}
        routeFilter={routeFilter}
        onRouteFilterChange={onRouteFilterChange}
        routes={routes}
      />

      {alerts.length === 0 ? (
        <p className="text-muted text-center py-16">
          No alerts yet.
        </p>
      ) : (
        <motion.div
          className="space-y-3"
          initial="hidden"
          animate="show"
          variants={{
            hidden: {},
            show: { transition: { staggerChildren: 0.06 } },
          }}
        >
          {alerts.map((alert) => (
            <AlertCard
              key={alert.id}
              alert={alert}
              route={routeMap.get(alert.route_id)}
              onMarkRead={() => onMarkRead(alert.id)}
            />
          ))}
        </motion.div>
      )}
    </div>
  );
}
