"use client";

import { useHistoryViewModel } from "./viewmodel";
import type { Route } from "@/lib/types";
import PriceStatsBar from "./PriceStatsBar";
import PriceChartGraph from "./PriceChartGraph";
import PeriodSelector from "./PeriodSelector";
import Spinner from "@/components/ui/Spinner";

interface HistoryViewProps {
  route: Route;
  onClose: () => void;
}

export default function HistoryView({ route, onClose }: HistoryViewProps) {
  const vm = useHistoryViewModel(route);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-xl font-semibold text-foreground">
            {route.origin} &rarr; {route.destination}
          </h2>
          <p className="text-sm text-muted">Price History</p>
        </div>
        <PeriodSelector
          days={vm.days}
          onDaysChange={vm.setDays}
          periods={vm.periods}
          exportCsvUrl={vm.exportCsvUrl}
          exportJsonUrl={vm.exportJsonUrl}
          onClose={onClose}
        />
      </div>

      {vm.stats && <PriceStatsBar stats={vm.stats} />}

      {vm.loading ? (
        <div className="flex justify-center py-12">
          <Spinner />
          <span className="ml-2 text-muted">Loading chart...</span>
        </div>
      ) : (
        <PriceChartGraph data={vm.chartData} alertPrice={route.alert_price} />
      )}
    </div>
  );
}
