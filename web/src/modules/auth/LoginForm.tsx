import Input from "@/components/ui/Input";
import Button from "@/components/ui/Button";

interface LoginFormProps {
  email: string;
  onEmailChange: (v: string) => void;
  password: string;
  onPasswordChange: (v: string) => void;
  loading: boolean;
  error: string;
  onSubmit: (e: React.FormEvent) => void;
  onSwitchToRegister: () => void;
}

export default function LoginForm({
  email,
  onEmailChange,
  password,
  onPasswordChange,
  loading,
  error,
  onSubmit,
  onSwitchToRegister,
}: LoginFormProps) {
  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <Input
        label="Email"
        type="email"
        value={email}
        onChange={(e) => onEmailChange(e.target.value)}
        placeholder="you@example.com"
        required
      />
      <Input
        label="Password"
        type="password"
        value={password}
        onChange={(e) => onPasswordChange(e.target.value)}
        placeholder="At least 8 characters"
        required
      />

      {error && (
        <p className="text-sm text-[var(--color-danger)]">{error}</p>
      )}

      <Button type="submit" loading={loading} className="w-full">
        Sign In
      </Button>

      <p className="text-center text-sm text-muted">
        Don&apos;t have an account?{" "}
        <button
          type="button"
          onClick={onSwitchToRegister}
          className="text-cyan-400 hover:text-cyan-300 font-medium"
        >
          Sign Up
        </button>
      </p>
    </form>
  );
}
