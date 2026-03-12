"use client";

import { useRoutesViewModel } from "./viewmodel";
import type { Route } from "@/lib/types";
import type { MonitorRequest } from "@/modules/app/viewmodel";
import RouteCard from "./RouteCard";
import RouteEditForm from "./RouteEditForm";
import RouteCreateForm from "./RouteCreateForm";
import RouteSortBar from "./RouteSortBar";
import Button from "@/components/ui/Button";
import Spinner from "@/components/ui/Spinner";
import { motion } from "framer-motion";

interface RoutesViewProps {
  onViewHistory: (route: Route) => void;
  monitorRequest?: MonitorRequest | null;
  onMonitorRequestHandled?: () => void;
}

export default function RoutesView({
  onViewHistory,
  monitorRequest,
  onMonitorRequestHandled,
}: RoutesViewProps) {
  const vm = useRoutesViewModel(monitorRequest, onMonitorRequestHandled);

  if (vm.loading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h2 className="text-xl font-semibold text-[var(--text-primary)]">
          Monitored Routes
        </h2>
        <Button
          variant={vm.showForm ? "secondary" : "primary"}
          onClick={() => vm.setShowForm(!vm.showForm)}
        >
          {vm.showForm ? "Cancel" : "+ Add Route"}
        </Button>
      </div>

      {vm.error && (
        <p className="text-[var(--color-danger)] mb-4 text-sm">{vm.error}</p>
      )}

      {vm.showForm && (
        <RouteCreateForm
          origin={vm.origin}
          onOriginChange={vm.setOrigin}
          destination={vm.destination}
          onDestinationChange={vm.setDestination}
          departureDate={vm.departureDate}
          onDepartureDateChange={vm.setDepartureDate}
          returnDate={vm.returnDate}
          onReturnDateChange={vm.setReturnDate}
          alertPrice={vm.alertPrice}
          onAlertPriceChange={vm.setAlertPrice}
          frequency={vm.frequency}
          onFrequencyChange={vm.setFrequency}
          onSubmit={vm.handleCreate}
          estimateLoading={vm.estimateLoading}
          currentMarketPrice={vm.currentMarketPrice}
          savings={vm.savings}
        />
      )}

      {vm.routes.length === 0 ? (
        <p className="text-[var(--text-secondary)] text-center py-16">
          No routes being monitored. Add one to get started.
        </p>
      ) : (
        <>
          <RouteSortBar
            sortKey={vm.sortKey}
            sortDir={vm.sortDir}
            onSort={vm.handleSort}
            sortIndicator={vm.sortIndicator}
          />
          <motion.div
            className="space-y-3"
            initial="hidden"
            animate="show"
            variants={{
              hidden: {},
              show: { transition: { staggerChildren: 0.06 } },
            }}
          >
            {vm.routes.map((route) =>
              vm.editingId === route.id ? (
                <RouteEditForm
                  key={route.id}
                  route={route}
                  alertPrice={vm.editAlertPrice}
                  onAlertPriceChange={vm.setEditAlertPrice}
                  frequency={vm.editFrequency}
                  onFrequencyChange={vm.setEditFrequency}
                  onSave={vm.handleEditSave}
                  onCancel={vm.cancelEdit}
                />
              ) : (
                <RouteCard
                  key={route.id}
                  route={route}
                  onViewHistory={() => onViewHistory(route)}
                  onEdit={() => vm.startEdit(route)}
                  onToggle={() => vm.handleToggle(route)}
                  onDelete={() => vm.handleDelete(route.id)}
                />
              )
            )}
          </motion.div>
        </>
      )}
    </div>
  );
}
