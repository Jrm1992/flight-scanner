export function formatDuration(min: number): string {
  return `${Math.floor(min / 60)}h${String(min % 60).padStart(2, "0")}m`;
}

export function formatTime(iso: string): string {
  if (!iso) return "";
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatPrice(n: number): string {
  return `$${n.toFixed(0)}`;
}

export function formatFrequency(minutes: number): string {
  if (minutes >= 60) return `${minutes / 60}h`;
  return `${minutes}m`;
}
