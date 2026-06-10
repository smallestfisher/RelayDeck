import type { InputHTMLAttributes, SelectHTMLAttributes } from 'react';
import { ChevronDown, Search } from 'lucide-react';
import { cn } from '../../lib/format';

export function SearchInput({
  className,
  ...props
}: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <label className={cn('relative block', className)}>
      <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
      <input
        className="focus-ring h-10 w-full rounded-lg border border-line bg-elevated pl-10 pr-3 text-sm text-text outline-none transition placeholder:text-muted/70 focus:border-primary/55"
        {...props}
      />
    </label>
  );
}

interface SelectControlProps extends SelectHTMLAttributes<HTMLSelectElement> {
  options: Array<{ label: string; value: string }>;
}

export function SelectControl({ options, className, ...props }: SelectControlProps) {
  return (
    <label className={cn('relative block', className)}>
      <select
        className="focus-ring h-10 w-full appearance-none rounded-lg border border-line bg-elevated px-3 pr-9 text-sm text-text outline-none transition focus:border-primary/55"
        {...props}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <ChevronDown className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
    </label>
  );
}

interface ToggleSwitchProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: string;
  className?: string;
}

export function ToggleSwitch({ checked, onChange, label, className }: ToggleSwitchProps) {
  return (
    <button
      type="button"
      className={cn('flex items-center gap-3 text-sm text-text', className)}
      onClick={() => onChange(!checked)}
    >
      <span
        className={cn(
          'relative h-6 w-11 rounded-full border transition',
          checked ? 'border-primary bg-primary' : 'border-line bg-elevated'
        )}
      >
        <span
          className={cn(
            'absolute top-1 h-4 w-4 rounded-full bg-white shadow transition',
            checked ? 'left-6' : 'left-1'
          )}
        />
      </span>
      {label && <span>{label}</span>}
    </button>
  );
}

interface RangeSliderProps {
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
}

export function RangeSlider({ value, onChange, min = 0, max = 100 }: RangeSliderProps) {
  return (
    <input
      type="range"
      min={min}
      max={max}
      value={value}
      onChange={(event) => onChange(Number(event.target.value))}
      className="h-2 w-full accent-primary"
    />
  );
}
