"use client";

import Input from "@/components/ui/Input";
import Button from "@/components/ui/Button";
import { useLoginViewModel } from "./viewmodel";
import type { LoginRequest } from "@/lib/types";

interface Props {
  onLogin: (req: LoginRequest) => Promise<void>;
  onSwitchToRegister: () => void;
}

export default function LoginForm({ onLogin, onSwitchToRegister }: Props) {
  const vm = useLoginViewModel(onLogin);

  return (
    <form onSubmit={vm.handleSubmit} className="space-y-4">
      <Input
        label="Email"
        type="email"
        value={vm.email}
        onChange={(e) => vm.setEmail(e.target.value)}
        placeholder="you@example.com"
        required
      />
      <Input
        label="Password"
        type="password"
        value={vm.password}
        onChange={(e) => vm.setPassword(e.target.value)}
        placeholder="At least 8 characters"
        required
      />

      {vm.error && (
        <p className="text-sm text-[var(--color-danger)]">{vm.error}</p>
      )}

      <Button type="submit" loading={vm.loading} className="w-full">
        Sign In
      </Button>

      <p className="text-center text-sm text-[var(--text-secondary)]">
        Don&apos;t have an account?{" "}
        <button
          type="button"
          onClick={onSwitchToRegister}
          className="text-[var(--brand-600)] hover:text-[var(--brand-700)] font-medium"
        >
          Sign Up
        </button>
      </p>
    </form>
  );
}
