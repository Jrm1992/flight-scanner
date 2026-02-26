import { useState, useEffect, useRef } from "react";
import { useRoutesModel } from "./model";
import { searchFlights } from "@/lib/api";
import type { Route } from "@/lib/types";
import type { MonitorRequest } from "@/modules/app/viewmodel";

export type SortKey = "origin" | "destination" | "current_price" | "alert_price" | "status";
export type SortDir = "asc" | "desc";

export function useRoutesViewModel(monitorRequest?: MonitorRequest | null, onMonitorRequestHandled?: () => void) {
  const model = useRoutesModel();

  // Sort
  const [sortKey, setSortKey] = useState<SortKey>("status");
  const [sortDir, setSortDir] = useState<SortDir>("asc");

  // Create form
  const [showForm, setShowForm] = useState(false);
  const [origin, setOrigin] = useState("");
  const [destination, setDestination] = useState("");
  const [alertPrice, setAlertPrice] = useState("");
  const [frequency, setFrequency] = useState("60");

  // Savings estimate
  const [estimateLoading, setEstimateLoading] = useState(false);
  const [currentMarketPrice, setCurrentMarketPrice] = useState<number | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  // Edit
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editAlertPrice, setEditAlertPrice] = useState("");
  const [editFrequency, setEditFrequency] = useState("");

  // Handle monitor request from search
  useEffect(() => {
    if (monitorRequest) {
      setOrigin(monitorRequest.origin);
      setDestination(monitorRequest.destination);
      setAlertPrice(String(Math.floor(monitorRequest.suggestedPrice)));
      setShowForm(true);
      onMonitorRequestHandled?.();
    }
  }, [monitorRequest, onMonitorRequestHandled]);

  // Fetch savings estimate
  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    setCurrentMarketPrice(null);

    if (origin.length === 3 && destination.length === 3) {
      debounceRef.current = setTimeout(async () => {
        setEstimateLoading(true);
        try {
          const data = await searchFlights(origin, destination);
          if (data.results.length > 0) {
            setCurrentMarketPrice(Math.min(...data.results.map((r) => r.price)));
          }
        } catch {
          // silently fail
        } finally {
          setEstimateLoading(false);
        }
      }, 800);
    }

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [origin, destination]);

  function handleSort(key: SortKey) {
    if (sortKey === key) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortKey(key);
      setSortDir("asc");
    }
  }

  function sortIndicator(key: SortKey): string {
    return sortKey === key ? (sortDir === "asc" ? " \u2191" : " \u2193") : "";
  }

  const sortedRoutes = [...model.routes].sort((a, b) => {
    const dir = sortDir === "asc" ? 1 : -1;
    switch (sortKey) {
      case "origin":
        return dir * a.origin.localeCompare(b.origin);
      case "destination":
        return dir * a.destination.localeCompare(b.destination);
      case "current_price":
        return dir * ((a.current_price ?? Infinity) - (b.current_price ?? Infinity));
      case "alert_price":
        return dir * (a.alert_price - b.alert_price);
      case "status":
        return dir * a.status.localeCompare(b.status);
      default:
        return 0;
    }
  });

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    model.setError("");
    try {
      await model.create({
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
      setCurrentMarketPrice(null);
    } catch (err) {
      model.setError(err instanceof Error ? err.message : "Failed to create route");
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this route?")) return;
    try {
      await model.remove(id);
    } catch (err) {
      model.setError(err instanceof Error ? err.message : "Failed to delete");
    }
  }

  async function handleToggle(route: Route) {
    try {
      if (route.status === "active") {
        await model.pause(route.id);
      } else {
        await model.resume(route.id);
      }
    } catch (err) {
      model.setError(err instanceof Error ? err.message : "Failed to toggle");
    }
  }

  function startEdit(route: Route) {
    setEditingId(route.id);
    setEditAlertPrice(String(route.alert_price));
    setEditFrequency(String(route.check_frequency_minutes));
  }

  function cancelEdit() {
    setEditingId(null);
  }

  async function handleEditSave(e: React.FormEvent) {
    e.preventDefault();
    if (!editingId) return;
    model.setError("");
    try {
      await model.update(editingId, {
        alert_price: parseFloat(editAlertPrice),
        check_frequency_minutes: parseInt(editFrequency),
      });
      setEditingId(null);
    } catch (err) {
      model.setError(err instanceof Error ? err.message : "Failed to update route");
    }
  }

  const savings =
    currentMarketPrice != null && alertPrice
      ? currentMarketPrice - parseFloat(alertPrice)
      : null;

  return {
    routes: sortedRoutes,
    loading: model.loading,
    error: model.error,
    // Sort
    sortKey,
    sortDir,
    handleSort,
    sortIndicator,
    // Create form
    showForm,
    setShowForm,
    origin,
    setOrigin,
    destination,
    setDestination,
    alertPrice,
    setAlertPrice,
    frequency,
    setFrequency,
    estimateLoading,
    currentMarketPrice,
    savings,
    handleCreate,
    // Edit
    editingId,
    editAlertPrice,
    setEditAlertPrice,
    editFrequency,
    setEditFrequency,
    startEdit,
    cancelEdit,
    handleEditSave,
    // Actions
    handleDelete,
    handleToggle,
  };
}
