import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import RouteList from "@/modules/routes";

vi.mock("@/lib/api", () => ({
  getRoutes: vi.fn(),
  createRoute: vi.fn(),
  deleteRoute: vi.fn(),
  pauseRoute: vi.fn(),
  resumeRoute: vi.fn(),
  updateRoute: vi.fn(),
  searchFlights: vi.fn().mockResolvedValue({ results: [] }),
}));

import {
  getRoutes,
  createRoute,
  deleteRoute,
  pauseRoute,
  resumeRoute,
  updateRoute,
} from "@/lib/api";

const mockGetRoutes = vi.mocked(getRoutes);
const mockCreateRoute = vi.mocked(createRoute);
const mockDeleteRoute = vi.mocked(deleteRoute);
const mockPauseRoute = vi.mocked(pauseRoute);
const mockResumeRoute = vi.mocked(resumeRoute);
const mockUpdateRoute = vi.mocked(updateRoute);

const mockOnViewHistory = vi.fn();

afterEach(() => {
  cleanup();
});

beforeEach(() => {
  mockGetRoutes.mockReset();
  mockCreateRoute.mockReset();
  mockDeleteRoute.mockReset();
  mockPauseRoute.mockReset();
  mockResumeRoute.mockReset();
  mockUpdateRoute.mockReset();
  mockOnViewHistory.mockReset();
});

const sampleRoute = {
  id: "r-1",
  origin: "GIG",
  destination: "SCL",
  currency: "BRL",
  alert_price: 500,
  check_frequency_minutes: 60,
  status: "active" as const,
  created_at: "2026-03-01T00:00:00Z",
  updated_at: "2026-03-01T00:00:00Z",
  current_price: 350,
  last_check_at: "2026-03-20T10:00:00Z",
  price_trend: "down",
};

describe("RouteList", () => {
  it("renders empty state", async () => {
    mockGetRoutes.mockResolvedValueOnce([]);

    render(<RouteList onViewHistory={mockOnViewHistory} />);

    await waitFor(() => {
      expect(screen.getByText("No routes being monitored. Add one to get started.")).toBeDefined();
    });
  });

  it("renders routes with current price", async () => {
    mockGetRoutes.mockResolvedValueOnce([sampleRoute]);

    render(<RouteList onViewHistory={mockOnViewHistory} />);

    await waitFor(() => {
      expect(screen.getByText(/GIG/)).toBeDefined();
      expect(screen.getByText(/SCL/)).toBeDefined();
      expect(screen.getByText(/R\$\s*350/)).toBeDefined();
    });
  });

  it("opens create form and submits", async () => {
    mockGetRoutes.mockResolvedValueOnce([]);
    mockCreateRoute.mockResolvedValueOnce(sampleRoute);
    mockGetRoutes.mockResolvedValueOnce([sampleRoute]);

    render(<RouteList onViewHistory={mockOnViewHistory} />);

    await waitFor(() => {
      expect(screen.getByText("+ Add Route")).toBeDefined();
    });

    fireEvent.click(screen.getByText("+ Add Route"));

    const originInput = screen.getByPlaceholderText(/Origin/);
    const destInput = screen.getByPlaceholderText(/Dest/);
    fireEvent.focus(originInput);
    fireEvent.change(originInput, { target: { value: "GIG" } });
    fireEvent.blur(originInput, { relatedTarget: document.body });
    fireEvent.focus(destInput);
    fireEvent.change(destInput, { target: { value: "SCL" } });
    fireEvent.blur(destInput, { relatedTarget: document.body });
    fireEvent.change(screen.getByPlaceholderText(/Alert price/), {
      target: { value: "500" },
    });

    const form = screen.getByText("Start Monitoring").closest("form")!;
    const departureDateInput = form.querySelector('input[type="date"]')!;
    fireEvent.change(departureDateInput, {
      target: { value: "2026-08-21" },
    });

    fireEvent.click(screen.getByText("Start Monitoring"));

    await waitFor(() => {
      expect(mockCreateRoute).toHaveBeenCalledWith({
        origin: "GIG",
        destination: "SCL",
        departure_date: "2026-08-21",
        return_date: undefined,
        currency: "BRL",
        alert_price: 500,
        check_frequency_minutes: 60,
      });
    });
  });

  it("deletes a route", async () => {
    mockGetRoutes.mockResolvedValueOnce([sampleRoute]);
    mockDeleteRoute.mockResolvedValueOnce(undefined);
    mockGetRoutes.mockResolvedValueOnce([]);

    vi.spyOn(window, "confirm").mockReturnValueOnce(true);

    render(<RouteList onViewHistory={mockOnViewHistory} />);

    await waitFor(() => {
      expect(screen.getByText("Delete")).toBeDefined();
    });

    fireEvent.click(screen.getByText("Delete"));

    await waitFor(() => {
      expect(mockDeleteRoute).toHaveBeenCalledWith("r-1");
    });
  });

  it("toggles pause/resume", async () => {
    mockGetRoutes.mockResolvedValueOnce([sampleRoute]);
    mockPauseRoute.mockResolvedValueOnce(undefined);
    mockGetRoutes.mockResolvedValueOnce([{ ...sampleRoute, status: "paused" as const }]);

    render(<RouteList onViewHistory={mockOnViewHistory} />);

    await waitFor(() => {
      expect(screen.getByText("Pause")).toBeDefined();
    });

    fireEvent.click(screen.getByText("Pause"));

    await waitFor(() => {
      expect(mockPauseRoute).toHaveBeenCalledWith("r-1");
    });
  });

  it("edit flow", async () => {
    mockGetRoutes.mockResolvedValueOnce([sampleRoute]);
    mockUpdateRoute.mockResolvedValueOnce({ ...sampleRoute, alert_price: 400 });
    mockGetRoutes.mockResolvedValueOnce([{ ...sampleRoute, alert_price: 400 }]);

    render(<RouteList onViewHistory={mockOnViewHistory} />);

    await waitFor(() => {
      expect(screen.getByText("Edit")).toBeDefined();
    });

    fireEvent.click(screen.getByText("Edit"));

    const alertInput = screen.getByPlaceholderText("Alert price");
    fireEvent.change(alertInput, { target: { value: "400" } });

    fireEvent.click(screen.getByText("Save"));

    await waitFor(() => {
      expect(mockUpdateRoute).toHaveBeenCalledWith("r-1", {
        alert_price: 400,
        check_frequency_minutes: 60,
      });
    });
  });
});
