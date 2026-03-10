const colorSchemes = {
  green: "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20",
  blue: "bg-cyan-500/10 text-cyan-400 border border-cyan-500/20",
  red: "bg-red-500/10 text-red-400 border border-red-500/20",
  neutral: "bg-white/5 text-slate-300 border border-white/10",
} as const;

interface StatCardProps {
  label: string;
  value: string;
  colorScheme?: keyof typeof colorSchemes;
  className?: string;
}

export default function StatCard({
  label,
  value,
  colorScheme = "neutral",
  className = "",
}: StatCardProps) {
  return (
    <div
      className={`rounded-[var(--radius-lg)] backdrop-blur-sm p-4 text-center ${colorSchemes[colorScheme]} ${className}`}
    >
      <p className="text-xs font-medium opacity-70">{label}</p>
      <p className="text-xl font-bold mt-1 font-data">{value}</p>
    </div>
  );
}
