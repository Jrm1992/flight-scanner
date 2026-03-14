import AirportInput from "@/components/AirportInput";
import Button from "@/components/ui/Button";

interface RouteCreateFormProps {
  origin: string;
  onOriginChange: (v: string) => void;
  destination: string;
  onDestinationChange: (v: string) => void;
  departureDate: string;
  onDepartureDateChange: (v: string) => void;
  returnDate: string;
  onReturnDateChange: (v: string) => void;
  alertPrice: string;
  onAlertPriceChange: (v: string) => void;
  frequency: string;
  onFrequencyChange: (v: string) => void;
  onSubmit: (e: React.FormEvent) => void;
  estimateLoading: boolean;
  currentMarketPrice: number | null;
  savings: number | null;
}

const inputClass =
  "rounded-md border border-border bg-white/5 px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-cyan-500/25 focus:border-cyan-500/50";

function getTomorrow() {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  return d.toISOString().split("T")[0];
}

export default function RouteCreateForm({
  origin,
  onOriginChange,
  destination,
  onDestinationChange,
  departureDate,
  onDepartureDateChange,
  returnDate,
  onReturnDateChange,
  alertPrice,
  onAlertPriceChange,
  frequency,
  onFrequencyChange,
  onSubmit,
  estimateLoading,
  currentMarketPrice,
  savings,
}: RouteCreateFormProps) {
  const minDate = getTomorrow();

  return (
    <form
      onSubmit={onSubmit}
      className="bg-white/5 backdrop-blur-xl border border-white/10 rounded-lg p-5 mb-6 grid grid-cols-2 gap-4 shadow-[var(--shadow-sm)]"
    >
      <AirportInput
        value={origin}
        onChange={onOriginChange}
        placeholder="Origin (e.g. GIG)"
        required
      />
      <AirportInput
        value={destination}
        onChange={onDestinationChange}
        placeholder="Destination (e.g. SCL)"
        required
      />
      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted">
          Departure date
        </label>
        <input
          type="date"
          value={departureDate}
          onChange={(e) => onDepartureDateChange(e.target.value)}
          min={minDate}
          className={inputClass}
          required
        />
      </div>
      <div className="flex flex-col gap-1.5">
        <label className="text-xs font-medium text-muted">
          Return date (optional)
        </label>
        <input
          type="date"
          value={returnDate}
          onChange={(e) => onReturnDateChange(e.target.value)}
          min={departureDate || minDate}
          className={inputClass}
        />
      </div>
      <input
        type="number"
        placeholder="Alert price (USD)"
        value={alertPrice}
        onChange={(e) => onAlertPriceChange(e.target.value)}
        min="1"
        step="0.01"
        className={inputClass}
        required
      />
      <select
        value={frequency}
        onChange={(e) => onFrequencyChange(e.target.value)}
        className={inputClass}
      >
        <option value="30">Every 30 min</option>
        <option value="60">Every 1 hour</option>
        <option value="120">Every 2 hours</option>
        <option value="360">Every 6 hours</option>
        <option value="720">Every 12 hours</option>
        <option value="1440">Every 24 hours</option>
      </select>

      {(estimateLoading || currentMarketPrice != null) && (
        <div className="col-span-2 bg-cyan-500/10 border border-cyan-500/20 rounded-md px-4 py-3 text-sm">
          {estimateLoading ? (
            <span className="text-cyan-400">Fetching current prices...</span>
          ) : currentMarketPrice != null ? (
            <div className="flex items-center gap-4 flex-wrap">
              <span className="text-muted">
                Current best price:{" "}
                <span className="font-bold text-foreground font-data">
                  ${currentMarketPrice.toFixed(0)}
                </span>
              </span>
              {alertPrice && (
                <>
                  <span className="text-muted-foreground">|</span>
                  <span className="text-muted">
                    Your alert:{" "}
                    <span className="font-bold font-data">
                      ${parseFloat(alertPrice).toFixed(0)}
                    </span>
                  </span>
                  <span className="text-muted-foreground">|</span>
                  {savings != null && savings > 0 ? (
                    <span className="text-emerald-400 font-semibold">
                      Potential savings: ${savings.toFixed(0)}
                    </span>
                  ) : (
                    <span className="text-amber-400">
                      Price is already below your alert threshold
                    </span>
                  )}
                </>
              )}
            </div>
          ) : null}
        </div>
      )}

      <Button variant="success" type="submit" className="col-span-2">
        Start Monitoring
      </Button>
    </form>
  );
}
