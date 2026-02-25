interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
}

export default function Select({
  label,
  className = "",
  id,
  children,
  ...props
}: SelectProps) {
  const selectId = id || label?.toLowerCase().replace(/\s+/g, "-");
  return (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label
          htmlFor={selectId}
          className="text-xs font-medium text-[var(--text-secondary)]"
        >
          {label}
        </label>
      )}
      <select
        id={selectId}
        className={`rounded-[var(--radius-md)] border border-[var(--border-default)] bg-white px-3 py-2 text-sm text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--brand-500)]/25 focus:border-[var(--brand-500)] transition-colors duration-[var(--transition-fast)] ${className}`}
        {...props}
      >
        {children}
      </select>
    </div>
  );
}
