import { useState, useEffect, useCallback } from "react";
import {
  login as apiLogin,
  register as apiRegister,
  setAuthToken,
  loadAuthToken,
} from "@/lib/api";
import type { User, LoginRequest, RegisterRequest } from "@/lib/types";

export function useAuthModel() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const isAuthenticated = !!user;

  // Initialize: load token from localStorage
  useEffect(() => {
    const token = loadAuthToken();
    if (token) {
      // Decode user from JWT payload (base64)
      try {
        const payload = JSON.parse(atob(token.split(".")[1]));
        const exp = payload.exp * 1000;
        if (Date.now() < exp) {
          // Token is still valid, restore user from localStorage
          const savedUser = localStorage.getItem("user");
          if (savedUser) {
            setUser(JSON.parse(savedUser));
          } else {
            // Token exists but no user data — force logout
            setAuthToken(null);
          }
        } else {
          setAuthToken(null);
        }
      } catch {
        setAuthToken(null);
      }
    }
    setLoading(false);
  }, []);

  // Listen for auth:logout events (triggered by 401 responses)
  useEffect(() => {
    function handleLogout() {
      setUser(null);
      localStorage.removeItem("user");
    }
    window.addEventListener("auth:logout", handleLogout);
    return () => window.removeEventListener("auth:logout", handleLogout);
  }, []);

  const login = useCallback(async (req: LoginRequest) => {
    setError("");
    try {
      const resp = await apiLogin(req);
      setAuthToken(resp.token);
      localStorage.setItem("user", JSON.stringify(resp.user));
      setUser(resp.user);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Login failed";
      setError(msg);
      throw err;
    }
  }, []);

  const registerUser = useCallback(async (req: RegisterRequest) => {
    setError("");
    try {
      const resp = await apiRegister(req);
      setAuthToken(resp.token);
      localStorage.setItem("user", JSON.stringify(resp.user));
      setUser(resp.user);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Registration failed";
      setError(msg);
      throw err;
    }
  }, []);

  const logout = useCallback(() => {
    setAuthToken(null);
    localStorage.removeItem("user");
    setUser(null);
  }, []);

  return { user, isAuthenticated, loading, error, login, register: registerUser, logout };
}
