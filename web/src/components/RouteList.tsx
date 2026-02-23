"use client";

import { useState, useEffect, useCallback } from "react";
import {
  getRoutes,
  createRoute,
  deleteRoute,
  pauseRoute,
  resumeRoute,
} from "@/lib/api";
import type { Route } from "@/lib/types";

interface Props {
  onViewHistory: (route: Route) => void;
}

export default function RouteList({ onViewHistory }: Props) {
  const [routes, setRoutes] = useState<Route[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Form state
  const [origin, setOrigin] = useState("");
  const [destination, setDestination] = useState("");
  const [alertPrice, setAlertPrice] = useState("");
  const [frequency, setFrequency] = useState("60");

  const loadRoutes = useCallback(async () => {
    try {
      const data = await getRoutes();
      setRoutes(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load routes");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadRoutes();
  }, [loadRoutes]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await createRoute({
        origin,
        destination,
        alert_price: parseFloat(alertPrice),
        check_frequency_minutes: parseInt(frequency),
      });
      setShowForm(false);
      setOrigin("");
      setDestination("");
      setAlertPrice("");
      setFrequency("60");
      loadRoutes();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create route");
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this route?")) return;
    try {
      await deleteRoute(id);
      loadRoutes();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete");
    }
  }

  async function handleToggle(route: Route) {
    try {
      if (route.status === "active") {
        await pauseRoute(route.id);
      } else {
        await resumeRoute(route.id);
      }
      loadRoutes();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to toggle");
    }
  }

  if (loading) return <p className="text-gray-500">Loading routes...</p>;

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold">Monitored Routes</h2>
        <button
          onClick={() => setShowForm(!showForm)}
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 text-sm"
        >
          {showForm ? "Cancel" : "+ Add Route"}
        </button>
      </div>

      {error && <p className="text-red-500 mb-4 text-sm">{error}</p>}

      {showForm && (
        <form
          onSubmit={handleCreate}
          className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-6 grid grid-cols-2 gap-3"
        >
          <input
            type="text"
            placeholder="Origin (e.g. GIG)"
            value={origin}
            onChange={(e) => setOrigin(e.target.value.toUpperCase())}
            maxLength={3}
            className="border border-gray-300 rounded px-3 py-2 uppercase bg-white text-gray-900"
            required
          />
          <input
            type="text"
            placeholder="Destination (e.g. SCL)"
            value={destination}
            onChange={(e) => setDestination(e.target.value.toUpperCase())}
            maxLength={3}
            className="border border-gray-300 rounded px-3 py-2 uppercase bg-white text-gray-900"
            required
          />
          <input
            type="number"
            placeholder="Alert price (USD)"
            value={alertPrice}
            onChange={(e) => setAlertPrice(e.target.value)}
            min="1"
            step="0.01"
            className="border border-gray-300 rounded px-3 py-2 bg-white text-gray-900"
            required
          />
          <select
            value={frequency}
            onChange={(e) => setFrequency(e.target.value)}
            className="border border-gray-300 rounded px-3 py-2 bg-white text-gray-900"
          >
            <option value="30">Every 30 min</option>
            <option value="60">Every 1 hour</option>
            <option value="120">Every 2 hours</option>
            <option value="360">Every 6 hours</option>
            <option value="720">Every 12 hours</option>
            <option value="1440">Every 24 hours</option>
          </select>
          <button
            type="submit"
            className="col-span-2 bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
          >
            Start Monitoring
          </button>
        </form>
      )}

      {routes.length === 0 ? (
        <p className="text-gray-500 text-center py-8">
          No routes being monitored. Add one to get started.
        </p>
      ) : (
        <div className="space-y-3">
          {routes.map((route) => (
            <div
              key={route.id}
              className="border border-gray-200 rounded-lg p-4 flex items-center justify-between hover:bg-gray-50"
            >
              <div className="flex items-center gap-4">
                <span
                  className={`w-2.5 h-2.5 rounded-full ${
                    route.status === "active" ? "bg-green-500" : "bg-gray-400"
                  }`}
                />
                <div>
                  <p className="font-semibold text-lg">
                    {route.origin} → {route.destination}
                  </p>
                  <p className="text-sm text-gray-500">
                    Alert at ${route.alert_price} · Every{" "}
                    {route.check_frequency_minutes >= 60
                      ? `${route.check_frequency_minutes / 60}h`
                      : `${route.check_frequency_minutes}m`}
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => onViewHistory(route)}
                  className="text-blue-600 hover:text-blue-800 text-sm px-3 py-1 border border-blue-200 rounded"
                >
                  History
                </button>
                <button
                  onClick={() => handleToggle(route)}
                  className={`text-sm px-3 py-1 border rounded ${
                    route.status === "active"
                      ? "text-orange-600 border-orange-200 hover:text-orange-800"
                      : "text-green-600 border-green-200 hover:text-green-800"
                  }`}
                >
                  {route.status === "active" ? "Pause" : "Resume"}
                </button>
                <button
                  onClick={() => handleDelete(route.id)}
                  className="text-red-500 hover:text-red-700 text-sm px-3 py-1 border border-red-200 rounded"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
