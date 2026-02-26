import type {
  Route,
  CreateRouteRequest,
  UpdateRouteRequest,
  SearchResponse,
  HistoryResponse,
  Alert,
} from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: { "Content-Type": "application/json", ...options?.headers },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }

  return res.json();
}

// Routes
export async function getRoutes(): Promise<Route[]> {
  const data = await request<{ routes: Route[] }>("/api/routes");
  return data.routes;
}

export async function createRoute(req: CreateRouteRequest): Promise<Route> {
  return request<Route>("/api/routes", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function updateRoute(id: string, req: UpdateRouteRequest): Promise<Route> {
  return request<Route>(`/api/routes/${id}`, {
    method: "PUT",
    body: JSON.stringify(req),
  });
}

export async function deleteRoute(id: string): Promise<void> {
  await request(`/api/routes/${id}`, { method: "DELETE" });
}

export async function pauseRoute(id: string): Promise<void> {
  await request(`/api/routes/${id}/pause`, { method: "PATCH" });
}

export async function resumeRoute(id: string): Promise<void> {
  await request(`/api/routes/${id}/resume`, { method: "PATCH" });
}

// Search
export async function searchFlights(
  origin: string,
  destination: string,
  date?: string
): Promise<SearchResponse> {
  return request<SearchResponse>("/api/search/flights", {
    method: "POST",
    body: JSON.stringify({ origin, destination, date }),
  });
}

// History
export async function getHistory(routeId: string, days = 30): Promise<HistoryResponse> {
  return request<HistoryResponse>(`/api/routes/${routeId}/history?days=${days}`);
}

// Alerts
export async function getAlerts(routeId?: string): Promise<Alert[]> {
  const params = routeId ? `?route_id=${routeId}` : "";
  const data = await request<{ alerts: Alert[]; count: number }>(`/api/alerts${params}`);
  return data.alerts;
}

export async function markAlertRead(id: string): Promise<void> {
  await request(`/api/alerts/${id}/mark-read`, { method: "PATCH" });
}

// Export
export function getExportUrl(routeId: string, days: number, format: "csv" | "json"): string {
  return `${API_URL}/api/routes/${routeId}/history/export?days=${days}&format=${format}`;
}
