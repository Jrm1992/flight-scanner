"use client";

import { useAppViewModel } from "@/modules/app/viewmodel";
import SearchView from "@/modules/search/SearchView";
import RoutesView from "@/modules/routes/RoutesView";
import HistoryView from "@/modules/history/HistoryView";
import AlertsView from "@/modules/alerts/AlertsView";
import Tabs from "@/components/ui/Tabs";

export default function Home() {
  const vm = useAppViewModel();

  return (
    <div className="min-h-screen bg-[var(--surface-secondary)]">
      <header className="bg-white border-b border-[var(--border-default)]">
        <div className="max-w-5xl mx-auto px-6 py-6">
          <h1 className="text-3xl font-bold tracking-tight text-[var(--text-primary)]">
            Flight Price Monitor
          </h1>
          <p className="text-sm text-[var(--text-secondary)] mt-1">
            Track flight prices and get alerts when they drop
          </p>
        </div>
      </header>

      <nav className="bg-white border-b border-[var(--border-default)] sticky top-0 z-10">
        <div className="max-w-5xl mx-auto px-6 py-3">
          <Tabs
            value={vm.tab}
            onValueChange={(v) => {
              vm.setTab(v as "search" | "routes" | "alerts");
            }}
          >
            <Tabs.List>
              <Tabs.Tab value="search">Search Flights</Tabs.Tab>
              <Tabs.Tab value="routes">Monitored Routes</Tabs.Tab>
              <Tabs.Tab value="alerts">Alerts</Tabs.Tab>
            </Tabs.List>
          </Tabs>
        </div>
      </nav>

      <main className="max-w-5xl mx-auto px-6 py-8">
        {vm.chartRoute ? (
          <HistoryView
            route={vm.chartRoute}
            onClose={() => vm.setChartRoute(null)}
          />
        ) : (
          <>
            {vm.tab === "search" && (
              <SearchView onMonitor={vm.handleMonitor} />
            )}
            {vm.tab === "routes" && (
              <RoutesView
                onViewHistory={(r) => vm.setChartRoute(r)}
                monitorRequest={vm.monitorRequest}
                onMonitorRequestHandled={vm.clearMonitorRequest}
              />
            )}
            {vm.tab === "alerts" && <AlertsView />}
          </>
        )}
      </main>
    </div>
  );
}
