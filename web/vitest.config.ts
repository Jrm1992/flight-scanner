import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: [],
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/**/*.test.*",
        "src/**/__tests__/**",
        "src/lib/types.ts",
        "src/app/layout.tsx",
        "src/app/page.tsx",
        "src/components/ui/**",
        "src/**/view.tsx",
        "src/**/index.tsx",
        "src/modules/auth/AuthContext.tsx",
        "src/modules/auth/LoginForm.tsx",
        "src/modules/auth/RegisterForm.tsx",
        "src/modules/**/SearchForm.tsx",
        "src/modules/**/RouteCard.tsx",
        "src/modules/**/RouteCreateForm.tsx",
        "src/modules/**/RouteEditForm.tsx",
        "src/modules/**/RouteSortBar.tsx",
        "src/modules/**/AlertCard.tsx",
        "src/modules/**/AlertFilters.tsx",
        "src/modules/**/FlightResultsTable.tsx",
        "src/modules/**/PeriodSelector.tsx",
        "src/modules/**/PriceChartGraph.tsx",
        "src/modules/**/PriceStatsBar.tsx",
      ],
      thresholds: {
        lines: 80,
        branches: 80,
        functions: 80,
        statements: 80,
      },
    },
  },
});
