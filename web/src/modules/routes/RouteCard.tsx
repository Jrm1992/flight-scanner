import type { Route } from "@/lib/types";
import { formatFrequency } from "@/lib/formatters";
import Card from "@/components/ui/Card";
import Badge from "@/components/ui/Badge";
import Button from "@/components/ui/Button";

interface RouteCardProps {
  route: Route;
  onViewHistory: () => void;
  onEdit: () => void;
  onToggle: () => void;
  onDelete: () => void;
}

export default function RouteCard({
  route,
  onViewHistory,
  onEdit,
  onToggle,
  onDelete,
}: RouteCardProps) {
  const priceBelow =
    route.current_price != null && route.current_price < route.alert_price;

  return (
    <Card variant={priceBelow ? "highlight" : "default"}>
      <Card.Body className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <span
            className={`w-2.5 h-2.5 rounded-full flex-shrink-0 ${
              route.status === "active" ? "bg-emerald-500" : "bg-slate-400"
            }`}
          />
          <div>
            <p className="font-semibold text-lg text-[var(--text-primary)] flex items-center gap-2">
              {route.origin} &rarr; {route.destination}
              {priceBelow && (
                <Badge variant="success" dot>
                  Below alert!
                </Badge>
              )}
            </p>
            <p className="text-sm text-[var(--text-secondary)]">
              {route.departure_date}
              {route.return_date && ` → ${route.return_date}`}
              {" "}&middot; Alert at ${route.alert_price} &middot; Every{" "}
              {formatFrequency(route.check_frequency_minutes)}
            </p>
          </div>
        </div>

        {route.current_price != null && (
          <div className="text-right mr-4">
            <p className="text-lg font-bold text-[var(--text-primary)]">
              ${route.current_price.toFixed(0)}
              {route.price_trend === "down" && (
                <span className="text-emerald-600 ml-1">&darr;</span>
              )}
              {route.price_trend === "up" && (
                <span className="text-red-500 ml-1">&uarr;</span>
              )}
              {route.price_trend === "stable" && (
                <span className="text-slate-400 ml-1">&rarr;</span>
              )}
            </p>
            {route.last_check_at && (
              <p className="text-xs text-[var(--text-tertiary)]">
                {new Date(route.last_check_at).toLocaleString("en-US", {
                  month: "short",
                  day: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </p>
            )}
          </div>
        )}

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={onViewHistory}>
            History
          </Button>
          <Button variant="secondary" size="sm" onClick={onEdit}>
            Edit
          </Button>
          <Button
            variant={route.status === "active" ? "secondary" : "success"}
            size="sm"
            onClick={onToggle}
          >
            {route.status === "active" ? "Pause" : "Resume"}
          </Button>
          <Button variant="danger" size="sm" onClick={onDelete}>
            Delete
          </Button>
        </div>
      </Card.Body>
    </Card>
  );
}
