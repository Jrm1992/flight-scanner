import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getRoutes,
  createRoute,
  updateRoute,
  deleteRoute,
  pauseRoute,
  resumeRoute,
  searchFlights,
  getAlerts,
  markAlertRead,
  getHistory,
  getExportUrl,
  setAuthToken,
  loadAuthToken,
  login,
  register,
} from "../api";

const mockFetch = vi.fn();
vi.stubGlobal("fetch", mockFetch);

function jsonResponse(data: unknown, status = 200) {
  return Promise.resolve({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(data),
  });
}

beforeEach(() => {
  mockFetch.mockReset();
  setAuthToken(null);
  localStorage.clear();
});

describe("getRoutes", () => {
  it("returns routes array", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ routes: [{ id: "r-1" }] }));
    const routes = await getRoutes();
    expect(routes).toHaveLength(1);
    expect(routes[0].id).toBe("r-1");
  });

  it("calls correct URL", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ routes: [] }));
    await getRoutes();
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/api/routes"),
      expect.anything()
    );
  });
});

describe("createRoute", () => {
  it("sends POST with body", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ id: "r-1" }));
    await createRoute({
      origin: "GIG",
      destination: "SCL",
      alert_price: 500,
      check_frequency_minutes: 60,
    });
    const [, opts] = mockFetch.mock.calls[0];
    expect(opts.method).toBe("POST");
    expect(JSON.parse(opts.body)).toMatchObject({ origin: "GIG" });
  });
});

describe("searchFlights", () => {
  it("returns search response", async () => {
    mockFetch.mockReturnValueOnce(
      jsonResponse({ results: [{ price: 299 }], count: 1 })
    );
    const res = await searchFlights("GIG", "SCL", "2026-05-01");
    expect(res.count).toBe(1);
    expect(res.results[0].price).toBe(299);
  });

  it("constructs correct URL", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ results: [], count: 0 }));
    await searchFlights("GIG", "SCL");
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/api/search/flights"),
      expect.anything()
    );
  });
});

describe("getAlerts", () => {
  it("returns alerts without route filter", async () => {
    mockFetch.mockReturnValueOnce(
      jsonResponse({ alerts: [{ id: "a-1" }], count: 1 })
    );
    const alerts = await getAlerts();
    expect(alerts).toHaveLength(1);
  });

  it("includes route_id param when provided", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ alerts: [], count: 0 }));
    await getAlerts("r-1");
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("route_id=r-1"),
      expect.anything()
    );
  });
});

describe("markAlertRead", () => {
  it("calls PATCH on correct URL", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ status: "read" }));
    await markAlertRead("a-1");
    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/alerts/a-1/mark-read");
    expect(opts.method).toBe("PATCH");
  });
});

describe("getHistory", () => {
  it("includes days param", async () => {
    mockFetch.mockReturnValueOnce(
      jsonResponse({ history: [], stats: {}, count: 0 })
    );
    await getHistory("r-1", 7);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("days=7"),
      expect.anything()
    );
  });
});

describe("updateRoute", () => {
  it("sends PUT with body", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ id: "r-1", alert_price: 400 }));
    await updateRoute("r-1", { alert_price: 400 });
    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/routes/r-1");
    expect(opts.method).toBe("PUT");
    expect(JSON.parse(opts.body)).toMatchObject({ alert_price: 400 });
  });
});

describe("deleteRoute", () => {
  it("sends DELETE", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ deleted: "r-1" }));
    await deleteRoute("r-1");
    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/routes/r-1");
    expect(opts.method).toBe("DELETE");
  });
});

describe("pauseRoute", () => {
  it("sends PATCH to pause endpoint", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ status: "paused" }));
    await pauseRoute("r-1");
    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/routes/r-1/pause");
    expect(opts.method).toBe("PATCH");
  });
});

describe("resumeRoute", () => {
  it("sends PATCH to resume endpoint", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ status: "active" }));
    await resumeRoute("r-1");
    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/routes/r-1/resume");
    expect(opts.method).toBe("PATCH");
  });
});

describe("getExportUrl", () => {
  it("returns correct CSV URL", () => {
    const url = getExportUrl("r-1", 30, "csv");
    expect(url).toContain("/api/routes/r-1/history/export");
    expect(url).toContain("days=30");
    expect(url).toContain("format=csv");
  });

  it("returns correct JSON URL", () => {
    const url = getExportUrl("r-1", 7, "json");
    expect(url).toContain("days=7");
    expect(url).toContain("format=json");
  });
});

describe("error handling", () => {
  it("throws on non-ok response", async () => {
    mockFetch.mockReturnValueOnce(
      Promise.resolve({
        ok: false,
        status: 500,
        json: () => Promise.resolve({ error: "server error" }),
      })
    );
    await expect(getRoutes()).rejects.toThrow("server error");
  });

  it("handles 401 by clearing token and dispatching logout event", async () => {
    setAuthToken("some-token");
    const dispatchSpy = vi.spyOn(window, "dispatchEvent");

    mockFetch.mockReturnValueOnce(
      Promise.resolve({
        ok: false,
        status: 401,
        json: () => Promise.resolve({}),
      })
    );

    await expect(getRoutes()).rejects.toThrow("Session expired");
    expect(localStorage.getItem("token")).toBeNull();
    expect(dispatchSpy).toHaveBeenCalledWith(expect.objectContaining({ type: "auth:logout" }));
    dispatchSpy.mockRestore();
  });

  it("handles non-ok response without error body", async () => {
    mockFetch.mockReturnValueOnce(
      Promise.resolve({
        ok: false,
        status: 503,
        json: () => Promise.reject(new Error("invalid json")),
      })
    );
    await expect(getRoutes()).rejects.toThrow("Request failed: 503");
  });
});

describe("token management", () => {
  it("setAuthToken stores token in localStorage", () => {
    setAuthToken("test-token");
    expect(localStorage.getItem("token")).toBe("test-token");
  });

  it("setAuthToken(null) removes token from localStorage", () => {
    setAuthToken("test-token");
    setAuthToken(null);
    expect(localStorage.getItem("token")).toBeNull();
  });

  it("loadAuthToken reads from localStorage", () => {
    localStorage.setItem("token", "stored-token");
    const token = loadAuthToken();
    expect(token).toBe("stored-token");
  });

  it("sends Authorization header when token is set", async () => {
    setAuthToken("my-jwt");
    mockFetch.mockReturnValueOnce(jsonResponse({ routes: [] }));
    await getRoutes();

    const [, opts] = mockFetch.mock.calls[0];
    expect(opts.headers["Authorization"]).toBe("Bearer my-jwt");
  });

  it("does not send Authorization header when no token", async () => {
    mockFetch.mockReturnValueOnce(jsonResponse({ routes: [] }));
    await getRoutes();

    const [, opts] = mockFetch.mock.calls[0];
    expect(opts.headers["Authorization"]).toBeUndefined();
  });
});

describe("auth endpoints", () => {
  it("login sends POST to /api/auth/login", async () => {
    mockFetch.mockReturnValueOnce(
      jsonResponse({ token: "jwt", user: { id: "u-1", name: "Test", email: "t@t.com" } })
    );
    const res = await login({ email: "t@t.com", password: "pass" });

    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/auth/login");
    expect(opts.method).toBe("POST");
    expect(JSON.parse(opts.body)).toMatchObject({ email: "t@t.com", password: "pass" });
    expect(res.token).toBe("jwt");
  });

  it("register sends POST to /api/auth/register", async () => {
    mockFetch.mockReturnValueOnce(
      jsonResponse({ token: "jwt", user: { id: "u-1", name: "Test", email: "t@t.com" } })
    );
    const res = await register({ name: "Test", email: "t@t.com", password: "pass" });

    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/auth/register");
    expect(opts.method).toBe("POST");
    expect(JSON.parse(opts.body)).toMatchObject({ name: "Test" });
    expect(res.user.name).toBe("Test");
  });
});
