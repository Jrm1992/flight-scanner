"use client";

import { useAlertsViewModel } from "./viewmodel";
import AlertCard from "./AlertCard";
import AlertFilters from "./AlertFilters";
import Spinner from "@/components/ui/Spinner";
import { motion } from "framer-motion";

export default function AlertsView() {
  const vm = useAlertsViewModel();

  if (vm.loading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner />
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-5">
        Alerts
      </h2>

      <AlertFilters
        filter={vm.filter}
        onFilterChange={(v) => vm.setFilter(v as "all" | "unread" | "read")}
        routeFilter={vm.routeFilter}
        onRouteFilterChange={vm.setRouteFilter}
        routes={vm.routes}
      />

      {vm.alerts.length === 0 ? (
        <p className="text-[var(--text-secondary)] text-center py-16">
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
          {vm.alerts.map((alert) => (
            <AlertCard
              key={alert.id}
              alert={alert}
              route={vm.routeMap.get(alert.route_id)}
              onMarkRead={() => vm.handleMarkRead(alert.id)}
            />
          ))}
        </motion.div>
      )}
    </div>
  );
}
