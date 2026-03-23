import { renderHook, act, waitFor } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach, type Mock } from "vitest";
import { useRoutesModel } from "../model";
import type { Route } from "@/lib/types";

vi.mock("@/lib/api", () => ({
  getRoutes: vi.fn(),
  createRoute: vi.fn(),
  updateRoute: vi.fn(),
  deleteRoute: vi.fn(),
  pauseRoute: vi.fn(),
  resumeRoute: vi.fn(),
  searchFlights: vi.fn(),
}));

import {
  getRoutes,
  createRoute,
  updateRoute,
  deleteRoute,
  pauseRoute,
  resumeRoute,
} from "@/lib/api";

const mockGetRoutes = getRoutes as Mock;
const mockCreateRoute = createRoute as Mock;
const mockUpdateRoute = updateRoute as Mock;
const mockDeleteRoute = deleteRoute as Mock;
const mockPauseRoute = pauseRoute as Mock;
const mockResumeRoute = resumeRoute as Mock;

const sampleRoutes: Route[] = [
  {
    id: "r-1",
    origin: "GIG",
    destination: "SCL",
    alert_price: 500,
    check_frequency_minutes: 60,
    status: "active",
    created_at: "2026-03-01T00:00:00Z",
    updated_at: "2026-03-01T00:00:00Z",
    current_price: 350,
    last_check_at: "2026-03-20T10:00:00Z",
    price_trend: "down",
  },
  {
    id: "r-2",
    origin: "MIA",
    destination: "JFK",
    alert_price: 200,
    check_frequency_minutes: 30,
    status: "paused",
    created_at: "2026-03-02T00:00:00Z",
    updated_at: "2026-03-02T00:00:00Z",
    current_price: null,
  },
] as Route[];

beforeEach(() => {
  vi.clearAllMocks();
  mockGetRoutes.mockResolvedValue(sampleRoutes);
});

describe("useRoutesModel", () => {
  // ── Loading & CRUD ──────────────────────────────────────────────

  it("loads routes on mount", async () => {
    const { result } = renderHook(() => useRoutesModel());

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(mockGetRoutes).toHaveBeenCalledTimes(1);
    expect(result.current.routes).toHaveLength(2);
  });

  it("handleCreate submits form data and reloads", async () => {
    mockCreateRoute.mockResolvedValue({});

    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    // Fill form fields
    act(() => {
      result.current.setOrigin("GRU");
      result.current.setDestination("EZE");
      result.current.setDepartureDate("2026-06-01");
      result.current.setReturnDate("2026-06-15");
      result.current.setAlertPrice("300");
      result.current.setFrequency("120");
      result.current.setShowForm(true);
    });

    const fakeEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;
    await act(() => result.current.handleCreate(fakeEvent));

    expect(fakeEvent.preventDefault).toHaveBeenCalled();
    expect(mockCreateRoute).toHaveBeenCalledWith({
      origin: "GRU",
      destination: "EZE",
      departure_date: "2026-06-01",
      return_date: "2026-06-15",
      currency: "BRL",
      alert_price: 300,
      check_frequency_minutes: 120,
    });
    // loadRoutes called again after create (initial + reload)
    expect(mockGetRoutes).toHaveBeenCalledTimes(2);
  });

  it("handleDelete with confirm=true deletes and reloads", async () => {
    vi.spyOn(window, "confirm").mockReturnValue(true);
    mockDeleteRoute.mockResolvedValue(undefined);

    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() => result.current.handleDelete("r-1"));

    expect(window.confirm).toHaveBeenCalledWith("Delete this route?");
    expect(mockDeleteRoute).toHaveBeenCalledWith("r-1");
    expect(mockGetRoutes).toHaveBeenCalledTimes(2);
  });

  it("handleDelete with confirm=false does nothing", async () => {
    vi.spyOn(window, "confirm").mockReturnValue(false);

    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() => result.current.handleDelete("r-1"));

    expect(mockDeleteRoute).not.toHaveBeenCalled();
  });

  it("handleToggle pauses active route and resumes paused route", async () => {
    mockPauseRoute.mockResolvedValue(undefined);
    mockResumeRoute.mockResolvedValue(undefined);

    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    // Active route → pause
    await act(() => result.current.handleToggle(sampleRoutes[0]));
    expect(mockPauseRoute).toHaveBeenCalledWith("r-1");

    // Paused route → resume
    await act(() => result.current.handleToggle(sampleRoutes[1]));
    expect(mockResumeRoute).toHaveBeenCalledWith("r-2");

    // loadRoutes called: initial + after pause + after resume
    expect(mockGetRoutes).toHaveBeenCalledTimes(3);
  });

  it("handleEditSave updates route and reloads", async () => {
    mockUpdateRoute.mockResolvedValue({});

    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    // Start editing
    act(() => result.current.startEdit(sampleRoutes[0]));
    expect(result.current.editingId).toBe("r-1");

    // Change edit values
    act(() => {
      result.current.setEditAlertPrice("450");
      result.current.setEditFrequency("30");
    });

    const fakeEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;
    await act(() => result.current.handleEditSave(fakeEvent));

    expect(mockUpdateRoute).toHaveBeenCalledWith("r-1", {
      alert_price: 450,
      check_frequency_minutes: 30,
    });
    expect(result.current.editingId).toBeNull();
    expect(mockGetRoutes).toHaveBeenCalledTimes(2);
  });

  // ── Sort ────────────────────────────────────────────────────────

  it("default sort by status ascending", async () => {
    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.sortKey).toBe("status");
    expect(result.current.sortDir).toBe("asc");
    // "active" < "paused" alphabetically → r-1 first
    expect(result.current.routes[0].id).toBe("r-1");
    expect(result.current.routes[1].id).toBe("r-2");
  });

  it("handleSort toggles direction on same key", async () => {
    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.handleSort("status"));
    expect(result.current.sortDir).toBe("desc");

    act(() => result.current.handleSort("status"));
    expect(result.current.sortDir).toBe("asc");
  });

  it("handleSort changes key and resets to asc", async () => {
    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    // Toggle status to desc first
    act(() => result.current.handleSort("status"));
    expect(result.current.sortDir).toBe("desc");

    // Switch to a different key
    act(() => result.current.handleSort("origin"));
    expect(result.current.sortKey).toBe("origin");
    expect(result.current.sortDir).toBe("asc");
  });

  it("sortIndicator returns arrow for active key, empty for others", async () => {
    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    // Default: status asc
    expect(result.current.sortIndicator("status")).toBe(" \u2191");
    expect(result.current.sortIndicator("origin")).toBe("");

    // Toggle to desc
    act(() => result.current.handleSort("status"));
    expect(result.current.sortIndicator("status")).toBe(" \u2193");
  });

  it("sorts by current_price with null handled as Infinity", async () => {
    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.handleSort("current_price"));

    // asc: 350 (r-1) < Infinity/null (r-2)
    expect(result.current.routes[0].id).toBe("r-1");
    expect(result.current.routes[1].id).toBe("r-2");

    // desc: null (Infinity, r-2) first
    act(() => result.current.handleSort("current_price"));
    expect(result.current.routes[0].id).toBe("r-2");
    expect(result.current.routes[1].id).toBe("r-1");
  });

  // ── Form state ──────────────────────────────────────────────────

  it("showForm toggle", async () => {
    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.showForm).toBe(false);

    act(() => result.current.setShowForm(true));
    expect(result.current.showForm).toBe(true);

    act(() => result.current.setShowForm(false));
    expect(result.current.showForm).toBe(false);
  });

  it("create form resets after successful create", async () => {
    mockCreateRoute.mockResolvedValue({});

    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => {
      result.current.setOrigin("GRU");
      result.current.setDestination("EZE");
      result.current.setAlertPrice("400");
      result.current.setShowForm(true);
    });

    const fakeEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;
    await act(() => result.current.handleCreate(fakeEvent));

    expect(result.current.showForm).toBe(false);
    expect(result.current.origin).toBe("");
    expect(result.current.destination).toBe("");
    expect(result.current.alertPrice).toBe("");
    expect(result.current.frequency).toBe("60");
  });

  // ── Monitor request ─────────────────────────────────────────────

  it("monitorRequest prefills form and shows it", async () => {
    const onHandled = vi.fn();
    const monitorReq = { origin: "LAX", destination: "SFO", suggestedPrice: 149.9 };

    const { result } = renderHook(() => useRoutesModel(monitorReq, onHandled));
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.origin).toBe("LAX");
    expect(result.current.destination).toBe("SFO");
    expect(result.current.alertPrice).toBe("149");
    expect(result.current.showForm).toBe(true);
    expect(onHandled).toHaveBeenCalled();
  });

  // ── Error handling ──────────────────────────────────────────────

  it("error on create sets error message", async () => {
    mockCreateRoute.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useRoutesModel());
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => {
      result.current.setOrigin("GRU");
      result.current.setDestination("EZE");
      result.current.setAlertPrice("300");
    });

    const fakeEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;
    await act(() => result.current.handleCreate(fakeEvent));

    expect(result.current.error).toBe("Network error");
  });

  it("error on load sets error message", async () => {
    mockGetRoutes.mockRejectedValue(new Error("Server down"));

    const { result } = renderHook(() => useRoutesModel());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe("Server down");
    expect(result.current.routes).toHaveLength(0);
  });
});
