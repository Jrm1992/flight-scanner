import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { useAlertsModel } from "../model";

vi.mock("@/lib/api", () => ({
  getAlerts: vi.fn(),
  markAlertRead: vi.fn(),
  getRoutes: vi.fn(),
}));

import { getAlerts, markAlertRead, getRoutes } from "@/lib/api";

const mockGetAlerts = getAlerts as ReturnType<typeof vi.fn>;
const mockMarkAlertRead = markAlertRead as ReturnType<typeof vi.fn>;
const mockGetRoutes = getRoutes as ReturnType<typeof vi.fn>;

const sampleRoutes = [
  {
    id: "r-1",
    origin: "GIG",
    destination: "SCL",
    alert_price: 500,
    check_frequency_minutes: 60,
    status: "active",
    created_at: "2026-03-01T00:00:00Z",
    updated_at: "2026-03-01T00:00:00Z",
  },
];

const sampleAlerts = [
  {
    id: "a-1",
    route_id: "r-1",
    alert_price: 500,
    triggered_price: 400,
    triggered_at: "2026-03-15T12:00:00Z",
    notified: false,
  },
  {
    id: "a-2",
    route_id: "r-1",
    alert_price: 500,
    triggered_price: 450,
    triggered_at: "2026-03-14T12:00:00Z",
    notified: true,
  },
];

describe("useAlertsModel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAlerts.mockResolvedValue(sampleAlerts);
    mockGetRoutes.mockResolvedValue(sampleRoutes);
    mockMarkAlertRead.mockResolvedValue(undefined);
  });

  it("loads alerts and routes on mount", async () => {
    const { result } = renderHook(() => useAlertsModel());

    expect(result.current.loading).toBe(true);

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(mockGetAlerts).toHaveBeenCalledTimes(1);
    expect(mockGetRoutes).toHaveBeenCalledTimes(1);
    expect(result.current.alerts).toHaveLength(2);
    expect(result.current.routes).toEqual(sampleRoutes);
  });

  it("routeMap maps route by id", async () => {
    const { result } = renderHook(() => useAlertsModel());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.routeMap.get("r-1")).toEqual(sampleRoutes[0]);
    expect(result.current.routeMap.size).toBe(1);
  });

  it("filter 'unread' hides notified alerts", async () => {
    const { result } = renderHook(() => useAlertsModel());

    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => {
      result.current.setFilter("unread");
    });

    expect(result.current.filter).toBe("unread");
    expect(result.current.alerts).toHaveLength(1);
    expect(result.current.alerts[0].id).toBe("a-1");
  });

  it("filter 'read' hides non-notified alerts", async () => {
    const { result } = renderHook(() => useAlertsModel());

    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => {
      result.current.setFilter("read");
    });

    expect(result.current.filter).toBe("read");
    expect(result.current.alerts).toHaveLength(1);
    expect(result.current.alerts[0].id).toBe("a-2");
  });

  it("filter 'all' shows everything", async () => {
    const { result } = renderHook(() => useAlertsModel());

    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => {
      result.current.setFilter("all");
    });

    expect(result.current.alerts).toHaveLength(2);
  });

  it("routeFilter filters by route_id", async () => {
    const extraAlerts = [
      ...sampleAlerts,
      {
        id: "a-3",
        route_id: "r-2",
        alert_price: 300,
        triggered_price: 250,
        triggered_at: "2026-03-16T12:00:00Z",
        notified: false,
      },
    ];
    mockGetAlerts.mockResolvedValue(extraAlerts);

    const { result } = renderHook(() => useAlertsModel());

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.alerts).toHaveLength(3);

    act(() => {
      result.current.setRouteFilter("r-1");
    });

    expect(result.current.routeFilter).toBe("r-1");
    expect(result.current.alerts).toHaveLength(2);
    expect(result.current.alerts.every((a) => a.route_id === "r-1")).toBe(true);
  });

  it("combined filters (read + route)", async () => {
    const extraAlerts = [
      ...sampleAlerts,
      {
        id: "a-3",
        route_id: "r-2",
        alert_price: 300,
        triggered_price: 250,
        triggered_at: "2026-03-16T12:00:00Z",
        notified: true,
      },
    ];
    mockGetAlerts.mockResolvedValue(extraAlerts);

    const { result } = renderHook(() => useAlertsModel());

    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => {
      result.current.setFilter("read");
      result.current.setRouteFilter("r-1");
    });

    expect(result.current.alerts).toHaveLength(1);
    expect(result.current.alerts[0].id).toBe("a-2");
  });

  it("handleMarkRead calls API and reloads alerts", async () => {
    const { result } = renderHook(() => useAlertsModel());

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(mockGetAlerts).toHaveBeenCalledTimes(1);

    await act(async () => {
      await result.current.handleMarkRead("a-1");
    });

    expect(mockMarkAlertRead).toHaveBeenCalledWith("a-1");
    expect(mockGetAlerts).toHaveBeenCalledTimes(2);
  });

  it("handles errors on load by setting empty arrays", async () => {
    mockGetAlerts.mockRejectedValueOnce(new Error("Network error"));
    mockGetRoutes.mockRejectedValueOnce(new Error("Network error"));

    const { result } = renderHook(() => useAlertsModel());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.alerts).toEqual([]);
    expect(result.current.routes).toEqual([]);
  });
});
