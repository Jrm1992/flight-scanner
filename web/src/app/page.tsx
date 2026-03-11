"use client";

import { useAuth } from "@/modules/auth/AuthContext";
import AuthView from "@/modules/auth/AuthView";
import { useAppViewModel } from "@/modules/app/viewmodel";
import SearchView from "@/modules/search/SearchView";
import RoutesView from "@/modules/routes/RoutesView";
import HistoryView from "@/modules/history/HistoryView";
import AlertsView from "@/modules/alerts/AlertsView";
import Tabs from "@/components/ui/Tabs";
import Button from "@/components/ui/Button";
import { motion, AnimatePresence } from "framer-motion";

export default function Home() {
  const auth = useAuth();
  const vm = useAppViewModel();

  if (!auth.isAuthenticated) {
    return <AuthView onLogin={auth.login} onRegister={auth.register} />;
  }

  return (
    <div className="min-h-screen">
      <header className="bg-gradient-to-r from-[#0a0e1a] to-[#0f1629] border-b border-white/10">
        <div className="max-w-5xl mx-auto px-6 py-6 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-cyan-400 to-cyan-200 bg-clip-text text-transparent">
              Flight Price Monitor
            </h1>
            <p className="text-sm text-[var(--text-secondary)] mt-1">
              Track flight prices and get alerts when they drop
            </p>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-sm text-[var(--text-secondary)]">
              {auth.user?.name}
            </span>
            <Button variant="secondary" size="sm" onClick={auth.logout}>
              Sign Out
            </Button>
          </div>
        </div>
      </header>

      <nav className="bg-[#0a0e1a]/80 backdrop-blur-xl border-b border-white/5 sticky top-0 z-10">
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
        <AnimatePresence mode="wait">
          {vm.chartRoute ? (
            <motion.div
              key="history"
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.2 }}
            >
              <HistoryView
                route={vm.chartRoute}
                onClose={() => vm.setChartRoute(null)}
              />
            </motion.div>
          ) : (
            <motion.div
              key={vm.tab}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.2 }}
            >
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
            </motion.div>
          )}
        </AnimatePresence>
      </main>
    </div>
  );
}
