import Button from "@/components/ui/Button";

interface PeriodSelectorProps {
  days: number;
  onDaysChange: (d: number) => void;
  periods: { label: string; value: number }[];
  exportCsvUrl: string;
  exportJsonUrl: string;
  onClose: () => void;
}

export default function PeriodSelector({
  days,
  onDaysChange,
  periods,
  exportCsvUrl,
  exportJsonUrl,
  onClose,
}: PeriodSelectorProps) {
  return (
    <div className="flex items-center gap-2">
      {periods.map((p) => (
        <Button
          key={p.value}
          variant={days === p.value ? "primary" : "secondary"}
          size="sm"
          onClick={() => onDaysChange(p.value)}
        >
          {p.label}
        </Button>
      ))}
      <Button
        variant="secondary"
        size="sm"
        onClick={() => window.open(exportCsvUrl)}
      >
        Export CSV
      </Button>
      <Button
        variant="secondary"
        size="sm"
        onClick={() => window.open(exportJsonUrl)}
      >
        Export JSON
      </Button>
      <Button variant="ghost" size="sm" onClick={onClose} className="ml-2">
        Close
      </Button>
    </div>
  );
}
