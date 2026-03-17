import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { useHistoryModel, PERIODS } from "../model";

vi.mock("@/lib/api", () => ({
  getHistory: vi.fn(),
  getExportUrl: vi.fn(
    (routeId: string, days: number, format: string) =>
      `/api/routes/${routeId}/history/export?days=${days}&format=${format}`,
  ),
}));

import { getHistory, getExportUrl } from "@/lib/api";

const mockGetHistory = getHistory as ReturnType<typeof vi.fn>;
const mockGetExportUrl = getExportUrl as ReturnType<typeof vi.fn>;

const fakeStats = { min_price: 100, max_price: 300, avg_price: 200, since: "2026-02-20T00:00:00Z" };

const fakeHistory = [
  {
    id: "h1",
    route_id: "r1",
    min_price: 120,
    max_price: 280,
    avg_price: 190,
    airline: "AA",
    checked_at: "2026-03-20T10:00:00Z",
  },
  {
    id: "h2",
    route_id: "r1",
    min_price: 110,
    max_price: 260,
    avg_price: 180,
    airline: "UA",
    checked_at: "2026-03-21T14:30:00Z",
  },
];

describe("useHistoryModel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetHistory.mockResolvedValue({ history: fakeHistory, stats: fakeStats });
  });

  it("has correct initial state (loading true, days 30)", () => {
    mockGetHistory.mockReturnValue(new Promise(() => {})); // never resolves
    const { result } = renderHook(() => useHistoryModel("r1"));

    expect(result.current.loading).toBe(true);
    expect(result.current.days).toBe(30);
  });

  it("loads data on mount and sets history/stats", async () => {
    const { result } = renderHook(() => useHistoryModel("r1"));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(mockGetHistory).toHaveBeenCalledWith("r1", 30);
    expect(result.current.stats).toEqual(fakeStats);
    expect(result.current.chartData).toHaveLength(2);
  });

  it("transforms history into chartData correctly", async () => {
    const { result } = renderHook(() => useHistoryModel("r1"));

    await waitFor(() => expect(result.current.loading).toBe(false));

    const first = result.current.chartData[0];
    expect(first.min).toBe(120);
    expect(first.avg).toBe(190);
    expect(first.max).toBe(280);
    expect(first.time).toBeDefined();
    expect(typeof first.time).toBe("string");
  });

  it("changing days triggers reload", async () => {
    const { result } = renderHook(() => useHistoryModel("r1"));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(mockGetHistory).toHaveBeenCalledTimes(1);

    act(() => {
      result.current.setDays(7);
    });

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(mockGetHistory).toHaveBeenCalledWith("r1", 7);
    expect(mockGetHistory).toHaveBeenCalledTimes(2);
  });

  it("handles errors by setting empty history", async () => {
    mockGetHistory.mockRejectedValueOnce(new Error("Network error"));

    const { result } = renderHook(() => useHistoryModel("r1"));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.chartData).toEqual([]);
    expect(result.current.stats).toBeNull();
  });

  it("export URLs use correct params", async () => {
    const { result } = renderHook(() => useHistoryModel("r1"));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.exportCsvUrl).toBe(
      "/api/routes/r1/history/export?days=30&format=csv",
    );
    expect(result.current.exportJsonUrl).toBe(
      "/api/routes/r1/history/export?days=30&format=json",
    );
    expect(mockGetExportUrl).toHaveBeenCalledWith("r1", 30, "csv");
    expect(mockGetExportUrl).toHaveBeenCalledWith("r1", 30, "json");
  });

  it("PERIODS constant is correct", () => {
    expect(PERIODS).toEqual([
      { label: "7d", value: 7 },
      { label: "30d", value: 30 },
      { label: "90d", value: 90 },
    ]);
  });
});
