import type { PriceStats } from "@/lib/types";
import PriceStatsBar from "./PriceStatsBar";
import PriceChartGraph from "./PriceChartGraph";
import PeriodSelector from "./PeriodSelector";
import Spinner from "@/components/ui/Spinner";

interface ChartDataPoint {
  time: string;
  min: number;
  avg: number;
  max: number;
}

interface HistoryViewProps {
  routeLabel: string;
  alertPrice: number;
  days: number;
  onDaysChange: (d: number) => void;
  periods: { label: string; value: number }[];
  exportCsvUrl: string;
  exportJsonUrl: string;
  onClose: () => void;
  stats: PriceStats | null;
  loading: boolean;
  chartData: ChartDataPoint[];
}

export default function HistoryView({
  routeLabel,
  alertPrice,
  days,
  onDaysChange,
  periods,
  exportCsvUrl,
  exportJsonUrl,
  onClose,
  stats,
  loading,
  chartData,
}: HistoryViewProps) {
  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-xl font-semibold text-foreground">
            {routeLabel}
          </h2>
          <p className="text-sm text-muted">Price History</p>
        </div>
        <PeriodSelector
          days={days}
          onDaysChange={onDaysChange}
          periods={periods}
          exportCsvUrl={exportCsvUrl}
          exportJsonUrl={exportJsonUrl}
          onClose={onClose}
        />
      </div>

      {stats && <PriceStatsBar stats={stats} />}

      {loading ? (
        <div className="flex justify-center py-12">
          <Spinner />
          <span className="ml-2 text-muted">Loading chart...</span>
        </div>
      ) : (
        <PriceChartGraph data={chartData} alertPrice={alertPrice} />
      )}
    </div>
  );
}
