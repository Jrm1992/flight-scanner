import { describe, it, expect } from "vitest";
import { searchAirports } from "../airports";

describe("searchAirports", () => {
  it("returns empty for empty query", () => {
    expect(searchAirports("")).toEqual([]);
  });

  it("matches by airport code prefix", () => {
    const results = searchAirports("GIG");
    expect(results.length).toBeGreaterThan(0);
    expect(results[0].code).toBe("GIG");
  });

  it("is case insensitive", () => {
    const results = searchAirports("gig");
    expect(results[0].code).toBe("GIG");
  });

  it("matches by city name substring", () => {
    const results = searchAirports("miami");
    expect(results.some((a) => a.code === "MIA")).toBe(true);
  });

  it("matches by country code", () => {
    const results = searchAirports("JP");
    expect(results.length).toBeGreaterThan(0);
    results.forEach((a) => expect(a.country).toBe("JP"));
  });

  it("limits results to 8", () => {
    const results = searchAirports("a");
    expect(results.length).toBeLessThanOrEqual(8);
  });

  it("matches partial code prefix", () => {
    const results = searchAirports("JF");
    expect(results.some((a) => a.code === "JFK")).toBe(true);
  });
});
