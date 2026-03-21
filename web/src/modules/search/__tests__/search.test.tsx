import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import SearchFlights from "@/modules/search";

vi.mock("@/lib/api", () => ({
  searchFlights: vi.fn(),
}));

import { searchFlights } from "@/lib/api";

const mockSearchFlights = vi.mocked(searchFlights);

afterEach(() => {
  cleanup();
});

beforeEach(() => {
  mockSearchFlights.mockReset();
});

describe("SearchFlights", () => {
  it("renders the search form", () => {
    render(<SearchFlights />);
    expect(screen.getByPlaceholderText(/Origin/)).toBeDefined();
    expect(screen.getByPlaceholderText(/Dest/)).toBeDefined();
    expect(screen.getByRole("button", { name: /Search/ })).toBeDefined();
  });

  it("submits search and displays results", async () => {
    mockSearchFlights.mockResolvedValueOnce({
      origin: "GIG",
      destination: "SCL",
      date: "2026-05-01",
      currency: "USD",
      results: [
        {
          price: 299,
          airline: "LATAM",
          flight_number: "LA800",
          departure_code: "GIG",
          arrival_code: "SCL",
          departure: "2026-05-01T10:00:00Z",
          arrival: "2026-05-01T16:00:00Z",
          duration_minutes: 360,
          stops: 0,
        },
      ],
      count: 1,
    });

    render(<SearchFlights />);

    const originInput = screen.getByPlaceholderText(/Origin/);
    const destInput = screen.getByPlaceholderText(/Dest/);
    fireEvent.focus(originInput);
    fireEvent.change(originInput, { target: { value: "GIG" } });
    fireEvent.blur(originInput, { relatedTarget: document.body });
    fireEvent.focus(destInput);
    fireEvent.change(destInput, { target: { value: "SCL" } });
    fireEvent.blur(destInput, { relatedTarget: document.body });
    fireEvent.click(screen.getByRole("button", { name: /Search/ }));

    await waitFor(() => {
      expect(screen.getByText(/R\$\s*299/)).toBeDefined();
      expect(screen.getByText("LATAM")).toBeDefined();
      expect(screen.getByText("Direct")).toBeDefined();
    });
  });

  it("shows error on search failure", async () => {
    mockSearchFlights.mockRejectedValueOnce(new Error("API down"));

    render(<SearchFlights />);

    const originInput = screen.getByPlaceholderText(/Origin/);
    const destInput = screen.getByPlaceholderText(/Dest/);
    fireEvent.focus(originInput);
    fireEvent.change(originInput, { target: { value: "GIG" } });
    fireEvent.blur(originInput, { relatedTarget: document.body });
    fireEvent.focus(destInput);
    fireEvent.change(destInput, { target: { value: "SCL" } });
    fireEvent.blur(destInput, { relatedTarget: document.body });
    fireEvent.click(screen.getByRole("button", { name: /Search/ }));

    await waitFor(() => {
      expect(screen.getByText("API down")).toBeDefined();
    });
  });
});
