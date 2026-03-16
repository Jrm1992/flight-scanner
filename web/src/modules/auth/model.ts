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

  useEffect(() => {
    const token = loadAuthToken();
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split(".")[1]));
        const exp = payload.exp * 1000;
        if (Date.now() < exp) {
          const savedUser = localStorage.getItem("user");
          if (savedUser) {
            setUser(JSON.parse(savedUser));
          } else {
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

export function useAuthFormModel(
  onLogin: (req: LoginRequest) => Promise<void>,
  onRegister: (req: RegisterRequest) => Promise<void>,
) {
  const [mode, setMode] = useState<"login" | "register">("login");

  // Login form
  const [loginEmail, setLoginEmail] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState("");

  // Register form
  const [registerName, setRegisterName] = useState("");
  const [registerEmail, setRegisterEmail] = useState("");
  const [registerPassword, setRegisterPassword] = useState("");
  const [registerLoading, setRegisterLoading] = useState(false);
  const [registerError, setRegisterError] = useState("");

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setLoginError("");
    setLoginLoading(true);
    try {
      await onLogin({ email: loginEmail, password: loginPassword });
    } catch (err) {
      setLoginError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoginLoading(false);
    }
  }

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault();
    setRegisterError("");
    setRegisterLoading(true);
    try {
      await onRegister({ name: registerName, email: registerEmail, password: registerPassword });
    } catch (err) {
      setRegisterError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setRegisterLoading(false);
    }
  }

  return {
    mode,
    switchToLogin: () => setMode("login"),
    switchToRegister: () => setMode("register"),
    // Login
    loginEmail,
    setLoginEmail,
    loginPassword,
    setLoginPassword,
    loginLoading,
    loginError,
    handleLogin,
    // Register
    registerName,
    setRegisterName,
    registerEmail,
    setRegisterEmail,
    registerPassword,
    setRegisterPassword,
    registerLoading,
    registerError,
    handleRegister,
  };
}
