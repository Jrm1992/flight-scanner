"use client";

import Input from "@/components/ui/Input";
import Button from "@/components/ui/Button";
import { useRegisterViewModel } from "./viewmodel";
import type { RegisterRequest } from "@/lib/types";

interface Props {
  onRegister: (req: RegisterRequest) => Promise<void>;
  onSwitchToLogin: () => void;
}

export default function RegisterForm({ onRegister, onSwitchToLogin }: Props) {
  const vm = useRegisterViewModel(onRegister);

  return (
    <form onSubmit={vm.handleSubmit} className="space-y-4">
      <Input
        label="Name"
        type="text"
        value={vm.name}
        onChange={(e) => vm.setName(e.target.value)}
        placeholder="Your name"
        required
      />
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
        minLength={8}
        required
      />

      {vm.error && (
        <p className="text-sm text-[var(--color-danger)]">{vm.error}</p>
      )}

      <Button type="submit" loading={vm.loading} className="w-full">
        Create Account
      </Button>

      <p className="text-center text-sm text-muted">
        Already have an account?{" "}
        <button
          type="button"
          onClick={onSwitchToLogin}
          className="text-cyan-400 hover:text-cyan-300 font-medium"
        >
          Sign In
        </button>
      </p>
    </form>
  );
}
