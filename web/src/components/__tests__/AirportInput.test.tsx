import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import AirportInput from "../AirportInput";

afterEach(cleanup);

function renderInput(props: Partial<React.ComponentProps<typeof AirportInput>> = {}) {
  const defaultProps = {
    value: "",
    onChange: vi.fn(),
    placeholder: "Airport",
    ...props,
  };
  return { ...render(<AirportInput {...defaultProps} />), onChange: defaultProps.onChange };
}

describe("AirportInput", () => {
  it("renders with placeholder", () => {
    renderInput({ placeholder: "Origin" });
    expect(screen.getByPlaceholderText("Origin")).toBeDefined();
  });

  it("displays formatted value when not focused", () => {
    renderInput({ value: "GIG" });
    const input = screen.getByPlaceholderText("Airport") as HTMLInputElement;
    expect(input.value).toContain("GIG");
    expect(input.value).toContain("Rio de Janeiro");
  });

  it("shows empty string for empty value", () => {
    renderInput({ value: "" });
    const input = screen.getByPlaceholderText("Airport") as HTMLInputElement;
    expect(input.value).toBe("");
  });

  it("opens dropdown on typing city name", () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "miami" } });
    expect(screen.getByText("MIA")).toBeDefined();
  });

  it("calls onChange when 3-letter code is typed and input blurs", () => {
    const { onChange } = renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "GIG" } });
    fireEvent.blur(input, { relatedTarget: document.body });
    expect(onChange).toHaveBeenCalledWith("GIG");
  });

  it("selects airport from dropdown via mouseDown", () => {
    const { onChange } = renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "miami" } });
    fireEvent.mouseDown(screen.getByText("MIA").closest("li")!);
    expect(onChange).toHaveBeenCalledWith("MIA");
  });

  it("navigates with arrow keys and selects with Enter", () => {
    const { onChange } = renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "miami" } });

    fireEvent.keyDown(input, { key: "ArrowDown" });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(onChange).toHaveBeenCalledWith("MIA");
  });

  it("closes dropdown on Escape", () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "miami" } });
    expect(screen.getByText("MIA")).toBeDefined();

    fireEvent.keyDown(input, { key: "Escape" });
    expect(screen.queryByText("Miami")).toBeNull();
  });

  it("closes dropdown on blur outside", () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "miami" } });

    fireEvent.blur(input, { relatedTarget: document.body });
    expect(screen.queryByText("Miami")).toBeNull();
  });

  it("returns code as-is when not in airport list", () => {
    renderInput({ value: "XXX" });
    const input = screen.getByPlaceholderText("Airport") as HTMLInputElement;
    expect(input.value).toBe("XXX");
  });

  it("ArrowUp does not go below 0", () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: "miami" } });
    fireEvent.keyDown(input, { key: "ArrowUp" });
    // Should not crash
  });

  it("ignores keydown when dropdown is closed", () => {
    renderInput();
    const input = screen.getByPlaceholderText("Airport");
    fireEvent.keyDown(input, { key: "ArrowDown" });
    // Should not crash
  });
});
