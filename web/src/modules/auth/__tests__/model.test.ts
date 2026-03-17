import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";

vi.mock("@/lib/api", () => ({
  login: vi.fn(),
  register: vi.fn(),
  setAuthToken: vi.fn(),
  loadAuthToken: vi.fn(),
}));

import { useAuthModel, useAuthFormModel } from "../model";
import {
  login as apiLogin,
  register as apiRegister,
  setAuthToken,
  loadAuthToken,
} from "@/lib/api";
import type { User, AuthResponse } from "@/lib/types";

const mockedApiLogin = vi.mocked(apiLogin);
const mockedApiRegister = vi.mocked(apiRegister);
const mockedSetAuthToken = vi.mocked(setAuthToken);
const mockedLoadAuthToken = vi.mocked(loadAuthToken);

function makeJwt(exp: number) {
  const payload = btoa(JSON.stringify({ exp }));
  return `header.${payload}.signature`;
}

const fakeUser: User = {
  id: "u1",
  email: "test@example.com",
  name: "Test User",
  created_at: "2026-01-01T00:00:00Z",
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
});

describe("useAuthModel", () => {
  it("starts with no user and loading becomes false after mount", async () => {
    mockedLoadAuthToken.mockReturnValue(null);

    const { result } = renderHook(() => useAuthModel());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.user).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.error).toBe("");
  });

  it("restores user from valid non-expired JWT token in localStorage", async () => {
    const futureExp = Math.floor(Date.now() / 1000) + 3600;
    mockedLoadAuthToken.mockReturnValue(makeJwt(futureExp));
    localStorage.setItem("user", JSON.stringify(fakeUser));

    const { result } = renderHook(() => useAuthModel());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.user).toEqual(fakeUser);
    expect(result.current.isAuthenticated).toBe(true);
  });

  it("clears auth on expired token", async () => {
    const pastExp = Math.floor(Date.now() / 1000) - 3600;
    mockedLoadAuthToken.mockReturnValue(makeJwt(pastExp));
    localStorage.setItem("user", JSON.stringify(fakeUser));

    const { result } = renderHook(() => useAuthModel());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.user).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
    expect(mockedSetAuthToken).toHaveBeenCalledWith(null);
  });

  it("clears auth on invalid token format", async () => {
    mockedLoadAuthToken.mockReturnValue("not-a-valid-jwt");

    const { result } = renderHook(() => useAuthModel());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.user).toBeNull();
    expect(mockedSetAuthToken).toHaveBeenCalledWith(null);
  });

  it("login() calls API, sets token and user in localStorage", async () => {
    mockedLoadAuthToken.mockReturnValue(null);
    const authResp: AuthResponse = { token: "tok123", user: fakeUser };
    mockedApiLogin.mockResolvedValue(authResp);

    const { result } = renderHook(() => useAuthModel());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.login({ email: "test@example.com", password: "pass" });
    });

    expect(mockedApiLogin).toHaveBeenCalledWith({ email: "test@example.com", password: "pass" });
    expect(mockedSetAuthToken).toHaveBeenCalledWith("tok123");
    expect(localStorage.getItem("user")).toBe(JSON.stringify(fakeUser));
    expect(result.current.user).toEqual(fakeUser);
    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.error).toBe("");
  });

  it("register() calls API, sets token and user in localStorage", async () => {
    mockedLoadAuthToken.mockReturnValue(null);
    const authResp: AuthResponse = { token: "tok456", user: fakeUser };
    mockedApiRegister.mockResolvedValue(authResp);

    const { result } = renderHook(() => useAuthModel());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.register({ name: "Test User", email: "test@example.com", password: "pass" });
    });

    expect(mockedApiRegister).toHaveBeenCalledWith({
      name: "Test User",
      email: "test@example.com",
      password: "pass",
    });
    expect(mockedSetAuthToken).toHaveBeenCalledWith("tok456");
    expect(localStorage.getItem("user")).toBe(JSON.stringify(fakeUser));
    expect(result.current.user).toEqual(fakeUser);
    expect(result.current.isAuthenticated).toBe(true);
  });

  it("logout() clears everything", async () => {
    const futureExp = Math.floor(Date.now() / 1000) + 3600;
    mockedLoadAuthToken.mockReturnValue(makeJwt(futureExp));
    localStorage.setItem("user", JSON.stringify(fakeUser));

    const { result } = renderHook(() => useAuthModel());

    await waitFor(() => {
      expect(result.current.user).toEqual(fakeUser);
    });

    act(() => {
      result.current.logout();
    });

    expect(mockedSetAuthToken).toHaveBeenCalledWith(null);
    expect(localStorage.getItem("user")).toBeNull();
    expect(result.current.user).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it("auth:logout event clears user", async () => {
    const futureExp = Math.floor(Date.now() / 1000) + 3600;
    mockedLoadAuthToken.mockReturnValue(makeJwt(futureExp));
    localStorage.setItem("user", JSON.stringify(fakeUser));

    const { result } = renderHook(() => useAuthModel());

    await waitFor(() => {
      expect(result.current.user).toEqual(fakeUser);
    });

    act(() => {
      window.dispatchEvent(new Event("auth:logout"));
    });

    expect(result.current.user).toBeNull();
    expect(localStorage.getItem("user")).toBeNull();
  });
});

describe("useAuthFormModel", () => {
  const mockOnLogin = vi.fn();
  const mockOnRegister = vi.fn();

  beforeEach(() => {
    mockOnLogin.mockReset();
    mockOnRegister.mockReset();
  });

  it("initial mode is login", () => {
    const { result } = renderHook(() => useAuthFormModel(mockOnLogin, mockOnRegister));

    expect(result.current.mode).toBe("login");
  });

  it("switchToRegister and switchToLogin toggle mode", () => {
    const { result } = renderHook(() => useAuthFormModel(mockOnLogin, mockOnRegister));

    act(() => {
      result.current.switchToRegister();
    });
    expect(result.current.mode).toBe("register");

    act(() => {
      result.current.switchToLogin();
    });
    expect(result.current.mode).toBe("login");
  });

  it("handleLogin calls onLogin with email/password and manages loading", async () => {
    mockOnLogin.mockResolvedValue(undefined);

    const { result } = renderHook(() => useAuthFormModel(mockOnLogin, mockOnRegister));

    act(() => {
      result.current.setLoginEmail("user@test.com");
      result.current.setLoginPassword("secret");
    });

    await act(async () => {
      await result.current.handleLogin({ preventDefault: vi.fn() } as unknown as React.FormEvent);
    });

    expect(mockOnLogin).toHaveBeenCalledWith({ email: "user@test.com", password: "secret" });
    expect(result.current.loginLoading).toBe(false);
    expect(result.current.loginError).toBe("");
  });

  it("handleRegister calls onRegister with name/email/password and manages loading", async () => {
    mockOnRegister.mockResolvedValue(undefined);

    const { result } = renderHook(() => useAuthFormModel(mockOnLogin, mockOnRegister));

    act(() => {
      result.current.setRegisterName("New User");
      result.current.setRegisterEmail("new@test.com");
      result.current.setRegisterPassword("pass123");
    });

    await act(async () => {
      await result.current.handleRegister({ preventDefault: vi.fn() } as unknown as React.FormEvent);
    });

    expect(mockOnRegister).toHaveBeenCalledWith({
      name: "New User",
      email: "new@test.com",
      password: "pass123",
    });
    expect(result.current.registerLoading).toBe(false);
    expect(result.current.registerError).toBe("");
  });

  it("sets loginError on failed login", async () => {
    mockOnLogin.mockRejectedValue(new Error("Invalid credentials"));

    const { result } = renderHook(() => useAuthFormModel(mockOnLogin, mockOnRegister));

    act(() => {
      result.current.setLoginEmail("user@test.com");
      result.current.setLoginPassword("wrong");
    });

    await act(async () => {
      await result.current.handleLogin({ preventDefault: vi.fn() } as unknown as React.FormEvent);
    });

    expect(result.current.loginError).toBe("Invalid credentials");
    expect(result.current.loginLoading).toBe(false);
  });

  it("sets registerError on failed register", async () => {
    mockOnRegister.mockRejectedValue(new Error("Email taken"));

    const { result } = renderHook(() => useAuthFormModel(mockOnLogin, mockOnRegister));

    act(() => {
      result.current.setRegisterName("User");
      result.current.setRegisterEmail("taken@test.com");
      result.current.setRegisterPassword("pass");
    });

    await act(async () => {
      await result.current.handleRegister({ preventDefault: vi.fn() } as unknown as React.FormEvent);
    });

    expect(result.current.registerError).toBe("Email taken");
    expect(result.current.registerLoading).toBe(false);
  });
});
