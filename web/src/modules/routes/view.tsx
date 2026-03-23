import type { Route } from "@/lib/types";
import type { SortKey, SortDir } from "./model";
import RouteCard from "./RouteCard";
import RouteEditForm from "./RouteEditForm";
import RouteCreateForm from "./RouteCreateForm";
import RouteSortBar from "./RouteSortBar";
import Button from "@/components/ui/Button";
import Spinner from "@/components/ui/Spinner";
import { motion } from "framer-motion";

interface RoutesViewProps {
  routes: Route[];
  loading: boolean;
  error: string;
  // Sort
  sortKey: SortKey;
  sortDir: SortDir;
  onSort: (key: SortKey) => void;
  sortIndicator: (key: SortKey) => string;
  // Create form
  showForm: boolean;
  onToggleForm: () => void;
  origin: string;
  onOriginChange: (v: string) => void;
  destination: string;
  onDestinationChange: (v: string) => void;
  departureDate: string;
  onDepartureDateChange: (v: string) => void;
  returnDate: string;
  onReturnDateChange: (v: string) => void;
  currency: string;
  onCurrencyChange: (v: string) => void;
  alertPrice: string;
  onAlertPriceChange: (v: string) => void;
  frequency: string;
  onFrequencyChange: (v: string) => void;
  onCreateSubmit: (e: React.FormEvent) => void;
  estimateLoading: boolean;
  currentMarketPrice: number | null;
  savings: number | null;
  // Edit
  editingId: string | null;
  editAlertPrice: string;
  onEditAlertPriceChange: (v: string) => void;
  editFrequency: string;
  onEditFrequencyChange: (v: string) => void;
  onStartEdit: (route: Route) => void;
  onCancelEdit: () => void;
  onEditSave: (e: React.FormEvent) => void;
  // Actions
  onDelete: (id: string) => void;
  onToggle: (route: Route) => void;
  onViewHistory: (route: Route) => void;
}

export default function RoutesView({
  routes,
  loading,
  error,
  sortKey,
  sortDir,
  onSort,
  sortIndicator,
  showForm,
  onToggleForm,
  origin,
  onOriginChange,
  destination,
  onDestinationChange,
  departureDate,
  onDepartureDateChange,
  returnDate,
  onReturnDateChange,
  currency,
  onCurrencyChange,
  alertPrice,
  onAlertPriceChange,
  frequency,
  onFrequencyChange,
  onCreateSubmit,
  estimateLoading,
  currentMarketPrice,
  savings,
  editingId,
  editAlertPrice,
  onEditAlertPriceChange,
  editFrequency,
  onEditFrequencyChange,
  onStartEdit,
  onCancelEdit,
  onEditSave,
  onDelete,
  onToggle,
  onViewHistory,
}: RoutesViewProps) {
  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h2 className="text-xl font-semibold text-foreground">
          Monitored Routes
        </h2>
        <Button
          variant={showForm ? "secondary" : "primary"}
          onClick={onToggleForm}
        >
          {showForm ? "Cancel" : "+ Add Route"}
        </Button>
      </div>

      {error && (
        <p className="text-[var(--color-danger)] mb-4 text-sm">{error}</p>
      )}

      {showForm && (
        <RouteCreateForm
          origin={origin}
          onOriginChange={onOriginChange}
          destination={destination}
          onDestinationChange={onDestinationChange}
          departureDate={departureDate}
          onDepartureDateChange={onDepartureDateChange}
          returnDate={returnDate}
          onReturnDateChange={onReturnDateChange}
          currency={currency}
          onCurrencyChange={onCurrencyChange}
          alertPrice={alertPrice}
          onAlertPriceChange={onAlertPriceChange}
          frequency={frequency}
          onFrequencyChange={onFrequencyChange}
          onSubmit={onCreateSubmit}
          estimateLoading={estimateLoading}
          currentMarketPrice={currentMarketPrice}
          savings={savings}
        />
      )}

      {routes.length === 0 ? (
        <p className="text-muted text-center py-16">
          No routes being monitored. Add one to get started.
        </p>
      ) : (
        <>
          <RouteSortBar
            sortKey={sortKey}
            sortDir={sortDir}
            onSort={onSort}
            sortIndicator={sortIndicator}
          />
          <motion.div
            className="space-y-3"
            initial="hidden"
            animate="show"
            variants={{
              hidden: {},
              show: { transition: { staggerChildren: 0.06 } },
            }}
          >
            {routes.map((route) =>
              editingId === route.id ? (
                <RouteEditForm
                  key={route.id}
                  route={route}
                  alertPrice={editAlertPrice}
                  onAlertPriceChange={onEditAlertPriceChange}
                  frequency={editFrequency}
                  onFrequencyChange={onEditFrequencyChange}
                  onSave={onEditSave}
                  onCancel={onCancelEdit}
                />
              ) : (
                <RouteCard
                  key={route.id}
                  route={route}
                  onViewHistory={() => onViewHistory(route)}
                  onEdit={() => onStartEdit(route)}
                  onToggle={() => onToggle(route)}
                  onDelete={() => onDelete(route.id)}
                />
              )
            )}
          </motion.div>
        </>
      )}
    </div>
  );
}
