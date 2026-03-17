import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useSearchModel } from "../model";
import type { FlightResult } from "@/lib/types";

vi.mock("@/lib/api", () => ({
  searchFlights: vi.fn(),
}));

import { searchFlights } from "@/lib/api";

const mockSearchFlights = vi.mocked(searchFlights);

const mockResults: FlightResult[] = [
  {
    flight_id: "abc123",
    price: 150,
    currency: "USD",
    airline: "Test Air",
    departure_time: "2026-04-01T08:00:00Z",
    arrival_time: "2026-04-01T12:00:00Z",
    origin: "JFK",
    destination: "LAX",
    stops: 0,
    deep_link: "https://example.com/book",
  },
];

describe("useSearchModel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns correct initial state", () => {
    const { result } = renderHook(() => useSearchModel());

    expect(result.current.results).toEqual([]);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe("");
    expect(result.current.origin).toBe("");
    expect(result.current.destination).toBe("");
    expect(result.current.date).toBe("");
  });

  it("updates origin via setOrigin", () => {
    const { result } = renderHook(() => useSearchModel());

    act(() => {
      result.current.setOrigin("JFK");
    });

    expect(result.current.origin).toBe("JFK");
  });

  it("updates destination via setDestination", () => {
    const { result } = renderHook(() => useSearchModel());

    act(() => {
      result.current.setDestination("LAX");
    });

    expect(result.current.destination).toBe("LAX");
  });

  it("updates date via setDate", () => {
    const { result } = renderHook(() => useSearchModel());

    act(() => {
      result.current.setDate("2026-04-01");
    });

    expect(result.current.date).toBe("2026-04-01");
  });

  it("calls preventDefault on the form event", async () => {
    mockSearchFlights.mockResolvedValue({ results: [] });
    const { result } = renderHook(() => useSearchModel());
    const mockEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;

    await act(async () => {
      await result.current.handleSearch(mockEvent);
    });

    expect(mockEvent.preventDefault).toHaveBeenCalledOnce();
  });

  it("sets loading during search and populates results on success", async () => {
    let resolveSearch!: (value: { results: FlightResult[] }) => void;
    mockSearchFlights.mockImplementation(
      () => new Promise((resolve) => { resolveSearch = resolve; })
    );

    const { result } = renderHook(() => useSearchModel());

    act(() => {
      result.current.setOrigin("JFK");
      result.current.setDestination("LAX");
      result.current.setDate("2026-04-01");
    });

    const mockEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;

    let searchPromise: Promise<void>;
    act(() => {
      searchPromise = result.current.handleSearch(mockEvent);
    });

    expect(result.current.loading).toBe(true);
    expect(result.current.error).toBe("");

    await act(async () => {
      resolveSearch({ results: mockResults });
      await searchPromise!;
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.results).toEqual(mockResults);
    expect(result.current.error).toBe("");
    expect(mockSearchFlights).toHaveBeenCalledWith("JFK", "LAX", "2026-04-01");
  });

  it("passes undefined for date when date is empty", async () => {
    mockSearchFlights.mockResolvedValue({ results: [] });
    const { result } = renderHook(() => useSearchModel());

    act(() => {
      result.current.setOrigin("JFK");
      result.current.setDestination("LAX");
    });

    const mockEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;

    await act(async () => {
      await result.current.handleSearch(mockEvent);
    });

    expect(mockSearchFlights).toHaveBeenCalledWith("JFK", "LAX", undefined);
  });

  it("sets error and clears results on failed search", async () => {
    mockSearchFlights.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useSearchModel());

    act(() => {
      result.current.setOrigin("JFK");
      result.current.setDestination("LAX");
    });

    const mockEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;

    await act(async () => {
      await result.current.handleSearch(mockEvent);
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe("Network error");
    expect(result.current.results).toEqual([]);
  });

  it("sets generic error message for non-Error exceptions", async () => {
    mockSearchFlights.mockRejectedValue("something unexpected");

    const { result } = renderHook(() => useSearchModel());
    const mockEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;

    await act(async () => {
      await result.current.handleSearch(mockEvent);
    });

    expect(result.current.error).toBe("Search failed");
    expect(result.current.results).toEqual([]);
  });

  it("clears previous error on new search", async () => {
    mockSearchFlights.mockRejectedValueOnce(new Error("First failure"));
    mockSearchFlights.mockResolvedValueOnce({ results: mockResults });

    const { result } = renderHook(() => useSearchModel());
    const mockEvent = { preventDefault: vi.fn() } as unknown as React.FormEvent;

    await act(async () => {
      await result.current.handleSearch(mockEvent);
    });

    expect(result.current.error).toBe("First failure");

    await act(async () => {
      await result.current.handleSearch(mockEvent);
    });

    expect(result.current.error).toBe("");
    expect(result.current.results).toEqual(mockResults);
  });
});
