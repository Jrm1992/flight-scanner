import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import PriceChart from "@/modules/history";

vi.mock("recharts", () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  LineChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Line: () => null,
  XAxis: () => null,
  YAxis: () => null,
  CartesianGrid: () => null,
  Tooltip: () => null,
  ReferenceLine: () => null,
}));

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return {
    getHistory: vi.fn(),
    getExportUrl: actual.getExportUrl,
  };
});

import { getHistory } from "@/lib/api";

const mockGetHistory = vi.mocked(getHistory);

const mockOnClose = vi.fn();

const sampleRoute = {
  id: "r-1",
  origin: "GIG",
  destination: "SCL",
  alert_price: 500,
  check_frequency_minutes: 60,
  status: "active" as const,
  created_at: "2026-03-01T00:00:00Z",
  updated_at: "2026-03-01T00:00:00Z",
};

afterEach(() => {
  cleanup();
});

beforeEach(() => {
  mockGetHistory.mockReset();
  mockOnClose.mockReset();
});

describe("PriceChart", () => {
  it("shows loading state", () => {
    mockGetHistory.mockReturnValue(new Promise(() => {}));

    render(<PriceChart route={sampleRoute} onClose={mockOnClose} />);

    expect(screen.getByText("Loading chart...")).toBeDefined();
  });

  it("renders stats cards", async () => {
    mockGetHistory.mockResolvedValueOnce({
      route_id: "r-1",
      days: 30,
      history: [
        {
          id: "h-1",
          route_id: "r-1",
          min_price: 200,
          max_price: 600,
          avg_price: 400,
          airline: "LATAM",
          checked_at: "2026-03-15T12:00:00Z",
        },
      ],
      stats: {
        min_price: 200,
        max_price: 600,
        avg_price: 400,
        since: "2026-02-18T00:00:00Z",
      },
      count: 1,
    });

    render(<PriceChart route={sampleRoute} onClose={mockOnClose} />);

    await waitFor(() => {
      expect(screen.getByText("$200")).toBeDefined();
      expect(screen.getByText("$400")).toBeDefined();
      expect(screen.getByText("$600")).toBeDefined();
    });
  });

  it("renders period buttons", async () => {
    mockGetHistory.mockResolvedValueOnce({
      route_id: "r-1",
      days: 30,
      history: [],
      stats: { min_price: 0, max_price: 0, avg_price: 0, since: "" },
      count: 0,
    });

    render(<PriceChart route={sampleRoute} onClose={mockOnClose} />);

    expect(screen.getByText("7d")).toBeDefined();
    expect(screen.getByText("30d")).toBeDefined();
    expect(screen.getByText("90d")).toBeDefined();
  });

  it("renders export buttons", async () => {
    mockGetHistory.mockResolvedValueOnce({
      route_id: "r-1",
      days: 30,
      history: [],
      stats: { min_price: 0, max_price: 0, avg_price: 0, since: "" },
      count: 0,
    });

    render(<PriceChart route={sampleRoute} onClose={mockOnClose} />);

    expect(screen.getByText("Export CSV")).toBeDefined();
    expect(screen.getByText("Export JSON")).toBeDefined();
  });

  it("calls window.open on export click", async () => {
    mockGetHistory.mockResolvedValueOnce({
      route_id: "r-1",
      days: 30,
      history: [],
      stats: { min_price: 0, max_price: 0, avg_price: 0, since: "" },
      count: 0,
    });

    const mockOpen = vi.spyOn(window, "open").mockImplementation(() => null);

    render(<PriceChart route={sampleRoute} onClose={mockOnClose} />);

    fireEvent.click(screen.getByText("Export CSV"));

    expect(mockOpen).toHaveBeenCalledWith(
      expect.stringContaining("/api/routes/r-1/history/export")
    );
    expect(mockOpen.mock.calls[0][0]).toContain("format=csv");
    expect(mockOpen.mock.calls[0][0]).toContain("days=30");

    mockOpen.mockRestore();
  });
});
