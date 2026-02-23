"use client";

import { useState } from "react";
import SearchFlights from "@/components/SearchFlights";
import RouteList from "@/components/RouteList";
import PriceChart from "@/components/PriceChart";
import AlertsList from "@/components/AlertsList";
import type { Route } from "@/lib/types";

type Tab = "search" | "routes" | "alerts";

export default function Home() {
  const [tab, setTab] = useState<Tab>("routes");
  const [chartRoute, setChartRoute] = useState<Route | null>(null);

  const tabs: { key: Tab; label: string }[] = [
    { key: "search", label: "Search Flights" },
    { key: "routes", label: "Monitored Routes" },
    { key: "alerts", label: "Alerts" },
  ];

  return (
    <div className="min-h-screen">
      <header className="border-b border-gray-200 bg-white">
        <div className="max-w-5xl mx-auto px-4 py-4">
          <h1 className="text-2xl font-bold">Flight Price Monitor</h1>
          <p className="text-sm text-gray-500 mt-1">
            Track flight prices and get alerts when they drop
          </p>
        </div>
      </header>

      <nav className="border-b border-gray-200 bg-white sticky top-0 z-10">
        <div className="max-w-5xl mx-auto px-4 flex gap-1">
          {tabs.map((t) => (
            <button
              key={t.key}
              onClick={() => {
                setTab(t.key);
                setChartRoute(null);
              }}
              className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                tab === t.key
                  ? "border-blue-600 text-blue-600"
                  : "border-transparent text-gray-500 hover:text-gray-700"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
      </nav>

      <main className="max-w-5xl mx-auto px-4 py-6">
        {chartRoute ? (
          <PriceChart route={chartRoute} onClose={() => setChartRoute(null)} />
        ) : (
          <>
            {tab === "search" && <SearchFlights />}
            {tab === "routes" && (
              <RouteList onViewHistory={(r) => setChartRoute(r)} />
            )}
            {tab === "alerts" && <AlertsList />}
          </>
        )}
      </main>
    </div>
  );
}
