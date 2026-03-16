import type { Route } from "@/lib/types";
import type { Tab, MonitorRequest } from "./model";
import Search from "@/modules/search";
import Routes from "@/modules/routes";
import History from "@/modules/history";
import Alerts from "@/modules/alerts";
import Tabs from "@/components/ui/Tabs";
import Button from "@/components/ui/Button";
import { motion, AnimatePresence } from "framer-motion";

interface DashboardViewProps {
  userName: string;
  onLogout: () => void;
  tab: Tab;
  onTabChange: (t: Tab) => void;
  chartRoute: Route | null;
  onCloseHistory: () => void;
  onViewHistory: (route: Route) => void;
  onMonitor: (origin: string, destination: string, price: number) => void;
  monitorRequest: MonitorRequest | null;
  onMonitorRequestHandled: () => void;
}

export default function DashboardView({
  userName,
  onLogout,
  tab,
  onTabChange,
  chartRoute,
  onCloseHistory,
  onViewHistory,
  onMonitor,
  monitorRequest,
  onMonitorRequestHandled,
}: DashboardViewProps) {
  return (
    <div className="min-h-screen">
      <header className="bg-gradient-to-r from-[#0a0e1a] to-[#0f1629] border-b border-white/10">
        <div className="max-w-5xl mx-auto px-6 py-6 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-cyan-400 to-cyan-200 bg-clip-text text-transparent">
              Flight Price Monitor
            </h1>
            <p className="text-sm text-muted mt-1">
              Track flight prices and get alerts when they drop
            </p>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-sm text-muted">
              {userName}
            </span>
            <Button variant="secondary" size="sm" onClick={onLogout}>
              Sign Out
            </Button>
          </div>
        </div>
      </header>

      <nav className="bg-[#0a0e1a]/80 backdrop-blur-xl border-b border-white/5 sticky top-0 z-10">
        <div className="max-w-5xl mx-auto px-6 py-3">
          <Tabs
            value={tab}
            onValueChange={(v) => onTabChange(v as Tab)}
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
          {chartRoute ? (
            <motion.div
              key="history"
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.2 }}
            >
              <History
                route={chartRoute}
                onClose={onCloseHistory}
              />
            </motion.div>
          ) : (
            <motion.div
              key={tab}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.2 }}
            >
              {tab === "search" && (
                <Search onMonitor={onMonitor} />
              )}
              {tab === "routes" && (
                <Routes
                  onViewHistory={onViewHistory}
                  monitorRequest={monitorRequest}
                  onMonitorRequestHandled={onMonitorRequestHandled}
                />
              )}
              {tab === "alerts" && <Alerts />}
            </motion.div>
          )}
        </AnimatePresence>
      </main>
    </div>
  );
}
