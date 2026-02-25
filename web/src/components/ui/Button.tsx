import Spinner from "./Spinner";

const variantStyles = {
  primary:
    "bg-[var(--brand-600)] text-white hover:bg-[var(--brand-700)] focus-visible:ring-[var(--brand-500)]",
  secondary:
    "bg-white text-[var(--text-secondary)] border border-[var(--border-default)] hover:border-[var(--border-hover)] hover:text-[var(--text-primary)] focus-visible:ring-[var(--brand-500)]",
  ghost:
    "text-[var(--text-secondary)] hover:bg-[var(--surface-secondary)] hover:text-[var(--text-primary)]",
  danger:
    "text-[var(--color-danger)] border border-red-200 hover:bg-red-50 focus-visible:ring-red-500",
  success:
    "bg-[var(--color-success)] text-white hover:bg-emerald-700 focus-visible:ring-emerald-500",
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
      className={`inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] font-medium transition-all duration-[var(--transition-fast)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none ${variantStyles[variant]} ${sizeStyles[size]} ${className}`}
      {...props}
    >
      {loading && <Spinner size="sm" />}
      {children}
    </button>
  );
}
