import { describe, it, expect } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useDashboardModel } from "../model";
import type { Route } from "@/lib/types";

describe("useDashboardModel", () => {
  it("returns correct initial state", () => {
    const { result } = renderHook(() => useDashboardModel());

    expect(result.current.tab).toBe("routes");
    expect(result.current.chartRoute).toBeNull();
    expect(result.current.monitorRequest).toBeNull();
  });

  it("handleTabChange updates tab and clears chartRoute", () => {
    const { result } = renderHook(() => useDashboardModel());

    const fakeRoute = { id: 1, origin: "JFK", destination: "LAX" } as Route;
    act(() => {
      result.current.setChartRoute(fakeRoute);
    });
    expect(result.current.chartRoute).toEqual(fakeRoute);

    act(() => {
      result.current.setTab("alerts");
    });
    expect(result.current.tab).toBe("alerts");
    expect(result.current.chartRoute).toBeNull();
  });

  it("setChartRoute sets the route", () => {
    const { result } = renderHook(() => useDashboardModel());

    const fakeRoute = { id: 2, origin: "SFO", destination: "ORD" } as Route;
    act(() => {
      result.current.setChartRoute(fakeRoute);
    });
    expect(result.current.chartRoute).toEqual(fakeRoute);
  });

  it("handleMonitor creates request and switches to routes tab", () => {
    const { result } = renderHook(() => useDashboardModel());

    act(() => {
      result.current.setTab("search");
    });
    expect(result.current.tab).toBe("search");

    act(() => {
      result.current.handleMonitor("JFK", "LAX", 250);
    });
    expect(result.current.monitorRequest).toEqual({
      origin: "JFK",
      destination: "LAX",
      suggestedPrice: 250,
    });
    expect(result.current.tab).toBe("routes");
  });

  it("clearMonitorRequest resets monitorRequest to null", () => {
    const { result } = renderHook(() => useDashboardModel());

    act(() => {
      result.current.handleMonitor("SFO", "ORD", 180);
    });
    expect(result.current.monitorRequest).not.toBeNull();

    act(() => {
      result.current.clearMonitorRequest();
    });
    expect(result.current.monitorRequest).toBeNull();
  });
});
