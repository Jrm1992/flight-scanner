import Input from "@/components/ui/Input";
import Button from "@/components/ui/Button";

interface RegisterFormProps {
  name: string;
  onNameChange: (v: string) => void;
  email: string;
  onEmailChange: (v: string) => void;
  password: string;
  onPasswordChange: (v: string) => void;
  loading: boolean;
  error: string;
  onSubmit: (e: React.FormEvent) => void;
  onSwitchToLogin: () => void;
}

export default function RegisterForm({
  name,
  onNameChange,
  email,
  onEmailChange,
  password,
  onPasswordChange,
  loading,
  error,
  onSubmit,
  onSwitchToLogin,
}: RegisterFormProps) {
  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <Input
        label="Name"
        type="text"
        value={name}
        onChange={(e) => onNameChange(e.target.value)}
        placeholder="Your name"
        required
      />
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
        minLength={8}
        required
      />

      {error && (
        <p className="text-sm text-[var(--color-danger)]">{error}</p>
      )}

      <Button type="submit" loading={loading} className="w-full">
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
