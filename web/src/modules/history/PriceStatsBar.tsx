import type { PriceStats } from "@/lib/types";
import StatCard from "@/components/ui/StatCard";
import { formatPrice } from "@/lib/formatters";

interface PriceStatsBarProps {
  stats: PriceStats;
}

export default function PriceStatsBar({ stats }: PriceStatsBarProps) {
  return (
    <div className="grid grid-cols-3 gap-4 mb-6">
      <StatCard label="Min Price" value={formatPrice(stats.min_price)} colorScheme="green" />
      <StatCard label="Avg Price" value={formatPrice(stats.avg_price)} colorScheme="blue" />
      <StatCard label="Max Price" value={formatPrice(stats.max_price)} colorScheme="red" />
    </div>
  );
}
