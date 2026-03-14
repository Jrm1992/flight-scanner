"use client";

import { useState } from "react";
import { motion } from "framer-motion";
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
    <div className="min-h-screen bg-[#0a0e1a] flex items-center justify-center px-4 relative overflow-hidden">
      {/* Ambient glow orbs */}
      <div className="absolute top-1/4 -left-1/4 w-96 h-96 bg-cyan-500/10 rounded-full blur-3xl" />
      <div className="absolute bottom-1/4 -right-1/4 w-96 h-96 bg-amber-500/10 rounded-full blur-3xl" />

      <motion.div
        className="w-full max-w-sm relative z-10"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
      >
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-cyan-400 to-cyan-200 bg-clip-text text-transparent">
            Flight Price Monitor
          </h1>
          <p className="text-sm text-muted mt-2">
            Track flight prices and get alerts when they drop
          </p>
        </div>

        <Card>
          <Card.Header>
            <h2 className="text-lg font-semibold text-foreground">
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
      </motion.div>
    </div>
  );
}
