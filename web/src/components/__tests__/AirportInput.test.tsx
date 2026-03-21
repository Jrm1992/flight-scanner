import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup, waitFor, act } from "@testing-library/react";
import AirportInput from "../AirportInput";

vi.mock("@/lib/api", () => ({
  searchAirportsAPI: vi.fn(),
}));

import { searchAirportsAPI } from "@/lib/api";

const mockSearchAirports = vi.mocked(searchAirportsAPI);

const mockResults = [
  { code: "MIA", name: "Miami Intl", city: "Miami" },
  { code: "MIA2", name: "Opa Locka", city: "Miami" },
];

afterEach(cleanup);

beforeEach(() => {
  vi.clearAllMocks();
  vi.useFakeTimers();
  mockSearchAirports.mockResolvedValue(mockResults);
});

afterEach(() => {
  vi.useRealTimers();
});

function renderInput(props: Partial<React.ComponentProps<typeof AirportInput>> = {}) {
  const defaultProps = {
    value: "",
    onChange: vi.fn(),
    placeholder: "Airport",
    ...props,
  };
  return { ...render(<AirportInput {...defaultProps} />), onChange: defaultProps.onChange };
}

async function typeAndWait(input: HTMLElement, value: string) {
  fireEvent.focus(input);
  fireEvent.change(input, { target: { value } });
  await act(async () => {
    vi.advanceTimersByTime(350);
  });
}

describe("AirportInput", () => {
  it("renders with placeholder", () => {
    renderInput({ placeholder: "Origin" });
    expect(screen.getByPlaceholderText("Origin")).toBeDefined();
  });

  it("shows empty string for empty value", () => {
    renderInput({ value: "" });
    const input = screen.getByPlaceholderText("Airport") as HTMLInputElement;
    expect(input.value).toBe("");
  });

  it("shows value as-is when no display label", () => {
    renderInput({ value: "GIG" });
    const input = screen.getByPlaceholderText("Airport") as HTMLInputElement;
    expect(input.value).toBe("GIG");
  });

  it("opens dropdown after typing with debounce", async () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    await typeAndWait(input, "miami");

    expect(mockSearchAirports).toHaveBeenCalledWith("miami");
    expect(screen.getByText("MIA")).toBeDefined();
  });

  it("does not call API for single character", async () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    await typeAndWait(input, "m");

    expect(mockSearchAirports).not.toHaveBeenCalled();
  });

  it("calls onChange when 3-letter code is typed and input blurs", async () => {
    const { onChange } = renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "GIG" } });
    fireEvent.blur(input, { relatedTarget: document.body });
    expect(onChange).toHaveBeenCalledWith("GIG");
  });

  it("selects airport from dropdown via mouseDown", async () => {
    const { onChange } = renderInput();
    const input = screen.getByPlaceholderText("Airport");
    await typeAndWait(input, "miami");

    fireEvent.mouseDown(screen.getByText("MIA").closest("li")!);
    expect(onChange).toHaveBeenCalledWith("MIA");
  });

  it("navigates with arrow keys and selects with Enter", async () => {
    const { onChange } = renderInput();
    const input = screen.getByPlaceholderText("Airport");
    await typeAndWait(input, "miami");

    fireEvent.keyDown(input, { key: "ArrowDown" });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(onChange).toHaveBeenCalledWith("MIA");
  });

  it("closes dropdown on Escape", async () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    await typeAndWait(input, "miami");

    expect(screen.getByText("MIA")).toBeDefined();
    fireEvent.keyDown(input, { key: "Escape" });
    expect(screen.queryByText("Miami")).toBeNull();
  });

  it("closes dropdown on blur outside", async () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    await typeAndWait(input, "miami");

    fireEvent.blur(input, { relatedTarget: document.body });
    expect(screen.queryByText("Miami")).toBeNull();
  });

  it("ArrowUp does not go below 0", async () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    await typeAndWait(input, "miami");
    fireEvent.keyDown(input, { key: "ArrowUp" });
    // Should not crash
  });

  it("ignores keydown when dropdown is closed", () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.keyDown(input, { key: "ArrowDown" });
    // Should not crash
  });

  it("shows display label after selection", async () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport") as HTMLInputElement;
    await typeAndWait(input, "miami");

    fireEvent.mouseDown(screen.getByText("MIA").closest("li")!);
    expect(input.value).toBe("MIA - Miami");
  });
});
