import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import PriceInsightsCard from "../PriceInsightsCard";
import type { PriceInsights } from "@/lib/types";

afterEach(cleanup);

const baseInsights: PriceInsights = {
  lowest_price: 299,
  price_level: "low",
  typical_price_range: [250, 500],
  price_history: [
    [1700000000, 300],
    [1700100000, 280],
    [1700200000, 299],
  ],
};

describe("PriceInsightsCard", () => {
  it("shows price level badge", () => {
    render(<PriceInsightsCard insights={baseInsights} currency="BRL" />);
    expect(screen.getByText("Low")).toBeDefined();
  });

  it("shows lowest price with currency symbol", () => {
    render(<PriceInsightsCard insights={baseInsights} currency="BRL" />);
    const lowest = screen.getByText(/Lowest/).textContent;
    expect(lowest).toContain("R$");
    expect(lowest).toContain("299");
  });

  it("shows typical price range", () => {
    render(<PriceInsightsCard insights={baseInsights} currency="USD" />);
    expect(screen.getByText(/Typical range/).textContent).toContain("$ 250");
    expect(screen.getByText(/Typical range/).textContent).toContain("$ 500");
  });

  it("shows high level badge", () => {
    render(
      <PriceInsightsCard
        insights={{ ...baseInsights, price_level: "high" }}
        currency="BRL"
      />
    );
    expect(screen.getByText("High")).toBeDefined();
  });

  it("shows typical level badge", () => {
    render(
      <PriceInsightsCard
        insights={{ ...baseInsights, price_level: "typical" }}
        currency="BRL"
      />
    );
    expect(screen.getByText("Typical")).toBeDefined();
  });

  it("falls back to typical for unknown level", () => {
    render(
      <PriceInsightsCard
        insights={{ ...baseInsights, price_level: "unknown" }}
        currency="BRL"
      />
    );
    expect(screen.getByText("Typical")).toBeDefined();
  });

  it("falls back to currency code for unknown currency", () => {
    render(<PriceInsightsCard insights={baseInsights} currency="JPY" />);
    const lowest = screen.getByText(/Lowest/).textContent;
    expect(lowest).toContain("JPY");
    expect(lowest).toContain("299");
  });

  it("hides typical range when values are zero", () => {
    render(
      <PriceInsightsCard
        insights={{ ...baseInsights, typical_price_range: [0, 0] }}
        currency="BRL"
      />
    );
    expect(screen.queryByText(/Typical range/)).toBeNull();
  });

  it("does not render chart when price history has 2 or fewer points", () => {
    const { container } = render(
      <PriceInsightsCard
        insights={{ ...baseInsights, price_history: [[1700000000, 300], [1700100000, 280]] }}
        currency="BRL"
      />
    );
    expect(container.querySelector(".recharts-responsive-container")).toBeNull();
  });

  it("renders chart when price history has more than 2 points", () => {
    const { container } = render(
      <PriceInsightsCard insights={baseInsights} currency="BRL" />
    );
    expect(container.querySelector(".recharts-responsive-container")).toBeDefined();
  });

  it("handles missing price_history gracefully", () => {
    const insights = { ...baseInsights, price_history: undefined as unknown as [number, number][] };
    render(<PriceInsightsCard insights={insights} currency="BRL" />);
    expect(screen.getByText("Low")).toBeDefined();
  });
});
