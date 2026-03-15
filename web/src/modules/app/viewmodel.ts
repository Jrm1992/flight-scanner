import { useState } from "react";
import type { Route } from "@/lib/types";
import type { MonitorRequest } from "@/modules/routes/model";

export type Tab = "search" | "routes" | "alerts";
export type { MonitorRequest };

export function useAppViewModel() {
  const [tab, setTab] = useState<Tab>("routes");
  const [chartRoute, setChartRoute] = useState<Route | null>(null);
  const [monitorRequest, setMonitorRequest] = useState<MonitorRequest | null>(null);

  function handleMonitor(origin: string, destination: string, price: number) {
    setMonitorRequest({ origin, destination, suggestedPrice: price });
    setTab("routes");
  }

  function clearMonitorRequest() {
    setMonitorRequest(null);
  }

  function handleTabChange(t: Tab) {
    setTab(t);
    setChartRoute(null);
  }

  return {
    tab,
    setTab: handleTabChange,
    chartRoute,
    setChartRoute,
    monitorRequest,
    handleMonitor,
    clearMonitorRequest,
  };
}
