import { describe, it, expect } from "vitest";
import {
  formatDuration,
  formatTime,
  formatDate,
  formatPrice,
  formatFrequency,
} from "../formatters";

describe("formatDuration", () => {
  it("formats 0 minutes", () => {
    expect(formatDuration(0)).toBe("0h00m");
  });

  it("formats minutes only", () => {
    expect(formatDuration(45)).toBe("0h45m");
  });

  it("formats exact hours", () => {
    expect(formatDuration(120)).toBe("2h00m");
  });

  it("formats hours and minutes", () => {
    expect(formatDuration(90)).toBe("1h30m");
  });

  it("pads single-digit minutes", () => {
    expect(formatDuration(65)).toBe("1h05m");
  });
});

describe("formatTime", () => {
  it("returns empty string for empty input", () => {
    expect(formatTime("")).toBe("");
  });

  it("formats ISO date to locale string", () => {
    const result = formatTime("2026-03-15T14:30:00Z");
    expect(result).toContain("Mar");
    expect(result).toContain("15");
  });
});

describe("formatDate", () => {
  it("returns empty string for empty input", () => {
    expect(formatDate("")).toBe("");
  });

  it("formats ISO date", () => {
    const result = formatDate("2026-03-15T00:00:00Z");
    expect(result).toContain("Mar");
    expect(result).toContain("2026");
  });
});

describe("formatPrice", () => {
  it("formats integer price", () => {
    expect(formatPrice(299)).toBe("$299");
  });

  it("rounds decimal price", () => {
    expect(formatPrice(299.7)).toBe("$300");
  });

  it("formats zero", () => {
    expect(formatPrice(0)).toBe("$0");
  });
});

describe("formatFrequency", () => {
  it("formats minutes below 60", () => {
    expect(formatFrequency(30)).toBe("30m");
  });

  it("formats exactly 60 as hours", () => {
    expect(formatFrequency(60)).toBe("1h");
  });

  it("formats multiple hours", () => {
    expect(formatFrequency(360)).toBe("6h");
  });
});
