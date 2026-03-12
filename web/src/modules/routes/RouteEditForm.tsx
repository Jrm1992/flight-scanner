import type { Route } from "@/lib/types";
import Card from "@/components/ui/Card";
import Button from "@/components/ui/Button";

interface RouteEditFormProps {
  route: Route;
  alertPrice: string;
  onAlertPriceChange: (v: string) => void;
  frequency: string;
  onFrequencyChange: (v: string) => void;
  onSave: (e: React.FormEvent) => void;
  onCancel: () => void;
}

const inlineInputClass =
  "rounded-[var(--radius-md)] border border-[var(--border-default)] bg-white/5 px-3 py-1.5 text-sm text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-cyan-500/25 focus:border-cyan-500/50";

export default function RouteEditForm({
  route,
  alertPrice,
  onAlertPriceChange,
  frequency,
  onFrequencyChange,
  onSave,
  onCancel,
}: RouteEditFormProps) {
  return (
    <Card>
      <Card.Body>
        <form onSubmit={onSave} className="flex items-center gap-3">
          <span className="font-semibold text-[var(--text-primary)]">
            {route.origin} &rarr; {route.destination}
          </span>
          <input
            type="number"
            value={alertPrice}
            onChange={(e) => onAlertPriceChange(e.target.value)}
            min="1"
            step="0.01"
            className={`${inlineInputClass} w-28`}
            placeholder="Alert price"
          />
          <select
            value={frequency}
            onChange={(e) => onFrequencyChange(e.target.value)}
            className={inlineInputClass}
          >
            <option value="30">30m</option>
            <option value="60">1h</option>
            <option value="120">2h</option>
            <option value="360">6h</option>
            <option value="720">12h</option>
            <option value="1440">24h</option>
          </select>
          <Button variant="success" size="sm" type="submit">
            Save
          </Button>
          <Button variant="secondary" size="sm" type="button" onClick={onCancel}>
            Cancel
          </Button>
        </form>
      </Card.Body>
    </Card>
  );
}
