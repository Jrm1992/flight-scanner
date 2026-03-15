"use client";

import { useHistoryModel } from "./model";
import type { Route } from "@/lib/types";
import HistoryView from "./view";

interface HistoryProps {
  route: Route;
  onClose: () => void;
}

export default function History({ route, onClose }: HistoryProps) {
  const model = useHistoryModel(route.id);

  const routeLabel = `${route.origin} \u2192 ${route.destination}`;

  return (
    <HistoryView
      routeLabel={routeLabel}
      alertPrice={route.alert_price}
      days={model.days}
      onDaysChange={model.setDays}
      periods={model.periods}
      exportCsvUrl={model.exportCsvUrl}
      exportJsonUrl={model.exportJsonUrl}
      onClose={onClose}
      stats={model.stats}
      loading={model.loading}
      chartData={model.chartData}
    />
  );
}
