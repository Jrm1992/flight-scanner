"use client";

import { useState, useEffect, useCallback } from "react";
import { getAlerts, markAlertRead } from "@/lib/api";
import type { Alert } from "@/lib/types";

export default function AlertsList() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);

  const loadAlerts = useCallback(async () => {
    try {
      const data = await getAlerts();
      setAlerts(data);
    } catch {
      setAlerts([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadAlerts();
  }, [loadAlerts]);

  async function handleMarkRead(id: string) {
    try {
      await markAlertRead(id);
      loadAlerts();
    } catch {
      // ignore
    }
  }

  if (loading) return <p className="text-gray-500">Loading alerts...</p>;

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Alerts</h2>

      {alerts.length === 0 ? (
        <p className="text-gray-500 text-center py-8">No alerts yet.</p>
      ) : (
        <div className="space-y-3">
          {alerts.map((alert) => {
            const savings = alert.alert_price - alert.triggered_price;
            return (
              <div
                key={alert.id}
                className={`border rounded-lg p-4 flex items-center justify-between ${
                  alert.notified
                    ? "border-gray-200 bg-gray-50"
                    : "border-green-300 bg-green-50"
                }`}
              >
                <div>
                  <div className="flex items-center gap-2">
                    {!alert.notified && (
                      <span className="w-2 h-2 rounded-full bg-green-500" />
                    )}
                    <p className="font-medium">
                      Price dropped to{" "}
                      <span className="text-green-700 font-bold">
                        ${alert.triggered_price.toFixed(0)}
                      </span>
                    </p>
                  </div>
                  <p className="text-sm text-gray-500 mt-1">
                    Alert threshold: ${alert.alert_price.toFixed(0)}
                    {savings > 0 && (
                      <span className="text-green-600 ml-2">
                        Save ${savings.toFixed(0)}
                      </span>
                    )}
                  </p>
                  <p className="text-xs text-gray-400 mt-1">
                    {new Date(alert.triggered_at).toLocaleString()}
                  </p>
                </div>
                {!alert.notified && (
                  <button
                    onClick={() => handleMarkRead(alert.id)}
                    className="text-sm text-blue-600 hover:text-blue-800 px-3 py-1 border border-blue-200 rounded"
                  >
                    Mark Read
                  </button>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
