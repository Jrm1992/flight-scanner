"use client";

import { useRoutesModel } from "./model";
import type { MonitorRequest } from "./model";
import type { Route } from "@/lib/types";
import RoutesView from "./view";

export type { MonitorRequest };

interface RoutesProps {
  onViewHistory: (route: Route) => void;
  monitorRequest?: MonitorRequest | null;
  onMonitorRequestHandled?: () => void;
}

export default function Routes({
  onViewHistory,
  monitorRequest,
  onMonitorRequestHandled,
}: RoutesProps) {
  const model = useRoutesModel(monitorRequest, onMonitorRequestHandled);

  return (
    <RoutesView
      routes={model.routes}
      loading={model.loading}
      error={model.error}
      sortKey={model.sortKey}
      sortDir={model.sortDir}
      onSort={model.handleSort}
      sortIndicator={model.sortIndicator}
      showForm={model.showForm}
      onToggleForm={() => model.setShowForm(!model.showForm)}
      origin={model.origin}
      onOriginChange={model.setOrigin}
      destination={model.destination}
      onDestinationChange={model.setDestination}
      departureDate={model.departureDate}
      onDepartureDateChange={model.setDepartureDate}
      returnDate={model.returnDate}
      onReturnDateChange={model.setReturnDate}
      alertPrice={model.alertPrice}
      onAlertPriceChange={model.setAlertPrice}
      frequency={model.frequency}
      onFrequencyChange={model.setFrequency}
      onCreateSubmit={model.handleCreate}
      estimateLoading={model.estimateLoading}
      currentMarketPrice={model.currentMarketPrice}
      savings={model.savings}
      editingId={model.editingId}
      editAlertPrice={model.editAlertPrice}
      onEditAlertPriceChange={model.setEditAlertPrice}
      editFrequency={model.editFrequency}
      onEditFrequencyChange={model.setEditFrequency}
      onStartEdit={model.startEdit}
      onCancelEdit={model.cancelEdit}
      onEditSave={model.handleEditSave}
      onDelete={model.handleDelete}
      onToggle={model.handleToggle}
      onViewHistory={onViewHistory}
    />
  );
}
