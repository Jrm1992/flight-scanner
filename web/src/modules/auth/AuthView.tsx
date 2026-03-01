"use client";

import { useState } from "react";
import Card from "@/components/ui/Card";
import LoginForm from "./LoginForm";
import RegisterForm from "./RegisterForm";
import type { LoginRequest, RegisterRequest } from "@/lib/types";

interface Props {
  onLogin: (req: LoginRequest) => Promise<void>;
  onRegister: (req: RegisterRequest) => Promise<void>;
}

export default function AuthView({ onLogin, onRegister }: Props) {
  const [mode, setMode] = useState<"login" | "register">("login");

  return (
    <div className="min-h-screen bg-[var(--surface-secondary)] flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold tracking-tight text-[var(--text-primary)]">
            Flight Price Monitor
          </h1>
          <p className="text-sm text-[var(--text-secondary)] mt-2">
            Track flight prices and get alerts when they drop
          </p>
        </div>

        <Card>
          <Card.Header>
            <h2 className="text-lg font-semibold text-[var(--text-primary)]">
              {mode === "login" ? "Welcome back" : "Create an account"}
            </h2>
          </Card.Header>
          <Card.Body>
            {mode === "login" ? (
              <LoginForm
                onLogin={onLogin}
                onSwitchToRegister={() => setMode("register")}
              />
            ) : (
              <RegisterForm
                onRegister={onRegister}
                onSwitchToLogin={() => setMode("login")}
              />
            )}
          </Card.Body>
        </Card>
      </div>
    </div>
  );
}
