const colorSchemes = {
  green: "bg-emerald-50 text-emerald-700",
  blue: "bg-blue-50 text-blue-700",
  red: "bg-red-50 text-red-700",
  neutral: "bg-slate-50 text-slate-700",
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
      className={`rounded-[var(--radius-lg)] p-4 text-center ${colorSchemes[colorScheme]} ${className}`}
    >
      <p className="text-xs font-medium opacity-70">{label}</p>
      <p className="text-xl font-bold mt-1">{value}</p>
    </div>
  );
}
