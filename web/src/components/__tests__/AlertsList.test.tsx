import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import AlertsList from "@/modules/alerts/AlertsView";

vi.mock("@/lib/api", () => ({
  getAlerts: vi.fn(),
  markAlertRead: vi.fn(),
  getRoutes: vi.fn(),
}));

import { getAlerts, markAlertRead, getRoutes } from "@/lib/api";

const mockGetAlerts = vi.mocked(getAlerts);
const mockMarkAlertRead = vi.mocked(markAlertRead);
const mockGetRoutes = vi.mocked(getRoutes);

afterEach(() => {
  cleanup();
});

beforeEach(() => {
  mockGetAlerts.mockReset();
  mockMarkAlertRead.mockReset();
  mockGetRoutes.mockReset();
  mockGetRoutes.mockResolvedValue([
    {
      id: "r-1",
      origin: "GIG",
      destination: "SCL",
      alert_price: 500,
      check_frequency_minutes: 60,
      status: "active" as const,
      created_at: "2026-03-01T00:00:00Z",
      updated_at: "2026-03-01T00:00:00Z",
    },
  ]);
});

describe("AlertsList", () => {
  it("renders empty state", async () => {
    mockGetAlerts.mockResolvedValueOnce([]);

    render(<AlertsList />);

    await waitFor(() => {
      expect(screen.getByText("No alerts yet.")).toBeDefined();
    });
  });

  it("renders alerts", async () => {
    mockGetAlerts.mockResolvedValueOnce([
      {
        id: "a-1",
        route_id: "r-1",
        alert_price: 500,
        triggered_price: 400,
        triggered_at: "2026-03-15T12:00:00Z",
        notified: false,
      },
    ]);

    render(<AlertsList />);

    await waitFor(() => {
      expect(screen.getByText(/\$400/)).toBeDefined();
      expect(screen.getByText("Mark Read")).toBeDefined();
    });
  });

  it("shows route info in alert card", async () => {
    mockGetAlerts.mockResolvedValueOnce([
      {
        id: "a-1",
        route_id: "r-1",
        alert_price: 500,
        triggered_price: 400,
        triggered_at: "2026-03-15T12:00:00Z",
        notified: false,
      },
    ]);

    render(<AlertsList />);

    await waitFor(() => {
      // GIG/SCL appear in both the route filter dropdown and the alert card
      expect(screen.getAllByText(/GIG/).length).toBeGreaterThanOrEqual(2);
      expect(screen.getAllByText(/SCL/).length).toBeGreaterThanOrEqual(2);
    });
  });

  it("calls markAlertRead and reloads", async () => {
    mockGetAlerts.mockResolvedValueOnce([
      {
        id: "a-1",
        route_id: "r-1",
        alert_price: 500,
        triggered_price: 400,
        triggered_at: "2026-03-15T12:00:00Z",
        notified: false,
      },
    ]);
    mockMarkAlertRead.mockResolvedValueOnce(undefined);
    mockGetAlerts.mockResolvedValueOnce([]);

    render(<AlertsList />);

    await waitFor(() => {
      expect(screen.getByText("Mark Read")).toBeDefined();
    });

    fireEvent.click(screen.getByText("Mark Read"));

    await waitFor(() => {
      expect(mockMarkAlertRead).toHaveBeenCalledWith("a-1");
    });
  });

  it("does not show Mark Read for already-read alerts", async () => {
    mockGetAlerts.mockResolvedValueOnce([
      {
        id: "a-1",
        route_id: "r-1",
        alert_price: 500,
        triggered_price: 400,
        triggered_at: "2026-03-15T12:00:00Z",
        notified: true,
        notified_at: "2026-03-15T13:00:00Z",
      },
    ]);

    render(<AlertsList />);

    await waitFor(() => {
      expect(screen.getByText(/\$400/)).toBeDefined();
    });

    expect(screen.queryByText("Mark Read")).toBeNull();
  });

  it("filters by read/unread", async () => {
    mockGetAlerts.mockResolvedValueOnce([
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
        notified_at: "2026-03-14T13:00:00Z",
      },
    ]);

    render(<AlertsList />);

    await waitFor(() => {
      expect(screen.getByText(/\$400/)).toBeDefined();
      expect(screen.getByText(/\$450/)).toBeDefined();
    });

    // Filter to read only
    const statusSelect = screen.getAllByRole("combobox")[0];
    fireEvent.change(statusSelect, { target: { value: "read" } });

    expect(screen.queryByText(/\$400/)).toBeNull();
    expect(screen.getByText(/\$450/)).toBeDefined();
  });
});
