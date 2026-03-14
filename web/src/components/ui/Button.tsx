import Spinner from "./Spinner";

const variantStyles = {
  primary:
    "bg-gradient-to-r from-cyan-500 to-cyan-600 text-white hover:from-cyan-400 hover:to-cyan-500 hover:shadow-[0_0_20px_rgba(6,182,212,0.3)] focus-visible:ring-cyan-500",
  secondary:
    "bg-white/5 text-muted border border-border hover:border-[var(--border-hover)] hover:text-foreground hover:bg-white/10 focus-visible:ring-cyan-500",
  ghost:
    "text-muted hover:bg-white/5 hover:text-foreground",
  danger:
    "text-red-400 border border-red-500/20 hover:bg-red-500/10 hover:border-red-500/40 focus-visible:ring-red-500",
  success:
    "bg-emerald-500/90 text-white hover:bg-emerald-500 hover:shadow-[0_0_20px_rgba(16,185,129,0.3)] focus-visible:ring-emerald-500",
} as const;

const sizeStyles = {
  sm: "px-3 py-1.5 text-xs",
  md: "px-4 py-2 text-sm",
} as const;

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof variantStyles;
  size?: keyof typeof sizeStyles;
  loading?: boolean;
}

export default function Button({
  variant = "primary",
  size = "md",
  loading,
  disabled,
  children,
  className = "",
  ...props
}: ButtonProps) {
  return (
    <button
      disabled={disabled || loading}
      className={`inline-flex items-center justify-center gap-2 rounded-md font-medium transition-all duration-[var(--transition-fast)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0a0e1a] disabled:opacity-50 disabled:pointer-events-none ${variantStyles[variant]} ${sizeStyles[size]} ${className}`}
      {...props}
    >
      {loading && <Spinner size="sm" />}
      {children}
    </button>
  );
}
