"use client";

import type { Alert, Route } from "@/lib/types";
import Card from "@/components/ui/Card";
import Badge from "@/components/ui/Badge";
import Button from "@/components/ui/Button";
import { motion } from "framer-motion";

interface AlertCardProps {
  alert: Alert;
  route?: Route;
  onMarkRead: () => void;
}

export default function AlertCard({ alert, route, onMarkRead }: AlertCardProps) {
  const savings = alert.alert_price - alert.triggered_price;

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <Card variant={alert.notified ? "muted" : "highlight"}>
        <Card.Body className="flex items-center justify-between">
          <div className="space-y-1">
            {route && (
              <span className="text-sm font-medium text-[var(--text-secondary)]">
                {route.origin} &rarr; {route.destination}
              </span>
            )}
            <div className="flex items-center gap-2">
              {!alert.notified && <Badge variant="success" dot>New</Badge>}
              <p className="font-medium text-[var(--text-primary)]">
                Price dropped to{" "}
                <span className="text-emerald-400 font-bold font-data">
                  ${alert.triggered_price.toFixed(0)}
                </span>
              </p>
            </div>
            <p className="text-sm text-[var(--text-secondary)]">
              Alert threshold: <span className="font-data">${alert.alert_price.toFixed(0)}</span>
              {savings > 0 && (
                <span className="text-emerald-400 ml-2 font-medium">
                  Save ${savings.toFixed(0)}
                </span>
              )}
            </p>
            <p className="text-xs text-[var(--text-tertiary)]">
              {new Date(alert.triggered_at).toLocaleString()}
            </p>
          </div>
          <div className="flex flex-col items-end gap-2">
            {route && (
              <a
                href={`https://www.kiwi.com/en/search/results/${route.origin}/${route.destination}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                <Button variant="primary" size="sm">
                  Book Now
                </Button>
              </a>
            )}
            {!alert.notified && (
              <Button variant="secondary" size="sm" onClick={onMarkRead}>
                Mark Read
              </Button>
            )}
          </div>
        </Card.Body>
      </Card>
    </motion.div>
  );
}
