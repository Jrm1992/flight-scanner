const variantStyles = {
  success: "bg-emerald-500/15 text-emerald-400 border-emerald-500/25",
  warning: "bg-amber-500/15 text-amber-400 border-amber-500/25",
  danger: "bg-red-500/15 text-red-400 border-red-500/25",
  neutral: "bg-white/10 text-slate-300 border-white/10",
  info: "bg-cyan-500/15 text-cyan-400 border-cyan-500/25",
} as const;

interface BadgeProps {
  variant?: keyof typeof variantStyles;
  dot?: boolean;
  children: React.ReactNode;
  className?: string;
}

export default function Badge({
  variant = "neutral",
  dot,
  children,
  className = "",
}: BadgeProps) {
  const dotColors: Record<string, string> = {
    success: "bg-emerald-400",
    warning: "bg-amber-400",
    danger: "bg-red-400",
    neutral: "bg-slate-400",
    info: "bg-cyan-400",
  };

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium border ${variantStyles[variant]} ${className}`}
    >
      {dot && (
        <span className={`w-1.5 h-1.5 rounded-full ${dotColors[variant]}`} />
      )}
      {children}
    </span>
  );
}
