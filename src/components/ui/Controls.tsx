import type { InputHTMLAttributes, ReactNode } from 'react';
import { useEffect, useId, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { Check, ChevronDown, Search } from 'lucide-react';
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

export interface SelectOption {
  label: string;
  value: string;
  icon?: ReactNode;
}

interface SelectChangeEvent {
  target: { value: string };
}

interface SelectControlProps {
  options: SelectOption[];
  value?: string;
  defaultValue?: string;
  onChange?: (event: SelectChangeEvent) => void;
  className?: string;
  disabled?: boolean;
  placeholder?: string;
  /** Show a search box inside the panel when there are many options. */
  searchable?: boolean;
  'aria-label'?: string;
}

export function SelectControl({
  options,
  value,
  defaultValue,
  onChange,
  className,
  disabled,
  placeholder = '请选择',
  searchable,
  'aria-label': ariaLabel,
}: SelectControlProps) {
  const isControlled = value !== undefined;
  const [internalValue, setInternalValue] = useState(defaultValue ?? options[0]?.value ?? '');
  const currentValue = isControlled ? value : internalValue;

  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [activeIndex, setActiveIndex] = useState(-1);
  const [position, setPosition] = useState<{ top: number; left: number; width: number; placement: 'top' | 'bottom' }>({
    top: 0,
    left: 0,
    width: 0,
    placement: 'bottom',
  });

  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);
  const listId = useId();

  const selectedOption = options.find((option) => option.value === currentValue);

  const filteredOptions = useMemo(() => {
    if (!searchable || !query.trim()) return options;
    const keyword = query.trim().toLowerCase();
    return options.filter((option) => option.label.toLowerCase().includes(keyword));
  }, [options, query, searchable]);

  function commit(next: string) {
    if (!isControlled) setInternalValue(next);
    onChange?.({ target: { value: next } });
    setOpen(false);
  }

  function updatePosition() {
    const rect = triggerRef.current?.getBoundingClientRect();
    if (!rect) return;
    const panelHeight = Math.min(320, filteredOptions.length * 40 + (searchable ? 52 : 0) + 16);
    const spaceBelow = window.innerHeight - rect.bottom;
    const placeAbove = spaceBelow < panelHeight + 12 && rect.top > spaceBelow;
    setPosition({
      top: placeAbove ? rect.top - 6 : rect.bottom + 6,
      left: rect.left,
      width: rect.width,
      placement: placeAbove ? 'top' : 'bottom',
    });
  }

  useLayoutEffect(() => {
    if (!open) return;
    updatePosition();
    window.addEventListener('resize', updatePosition);
    window.addEventListener('scroll', updatePosition, true);
    return () => {
      window.removeEventListener('resize', updatePosition);
      window.removeEventListener('scroll', updatePosition, true);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, filteredOptions.length]);

  useEffect(() => {
    if (!open) return;
    setActiveIndex(filteredOptions.findIndex((option) => option.value === currentValue));
    if (searchable) {
      const id = window.setTimeout(() => searchRef.current?.focus(), 0);
      return () => window.clearTimeout(id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  useEffect(() => {
    if (!open) return;

    function handlePointerDown(event: MouseEvent) {
      const target = event.target as Node;
      if (triggerRef.current?.contains(target) || panelRef.current?.contains(target)) return;
      setOpen(false);
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setOpen(false);
        triggerRef.current?.focus();
      }
    }

    document.addEventListener('mousedown', handlePointerDown);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [open]);

  function handleTriggerKeyDown(event: React.KeyboardEvent) {
    if (disabled) return;
    if (!open) {
      if (event.key === 'ArrowDown' || event.key === 'Enter' || event.key === ' ') {
        event.preventDefault();
        setOpen(true);
      }
      return;
    }
    if (event.key === 'ArrowDown') {
      event.preventDefault();
      setActiveIndex((index) => Math.min(filteredOptions.length - 1, index + 1));
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      setActiveIndex((index) => Math.max(0, index - 1));
    } else if (event.key === 'Enter') {
      event.preventDefault();
      const option = filteredOptions[activeIndex];
      if (option) commit(option.value);
    }
  }

  return (
    <div className={cn('relative', className)}>
      <button
        ref={triggerRef}
        type="button"
        disabled={disabled}
        onClick={() => {
          if (disabled) return;
          setQuery('');
          setOpen((value) => !value);
        }}
        onKeyDown={handleTriggerKeyDown}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-label={ariaLabel}
        className={cn(
          'focus-ring flex h-10 w-full items-center gap-2 rounded-lg border bg-elevated px-3 text-left text-sm text-text outline-none transition',
          open ? 'border-primary/60 ring-2 ring-primary/15' : 'border-line hover:border-primary/40',
          disabled && 'cursor-not-allowed opacity-50'
        )}
      >
        {selectedOption?.icon && <span className="flex h-5 w-5 shrink-0 items-center justify-center">{selectedOption.icon}</span>}
        <span className={cn('flex-1 truncate', !selectedOption && 'text-muted/70')}>
          {selectedOption?.label ?? placeholder}
        </span>
        <ChevronDown className={cn('h-4 w-4 shrink-0 text-muted transition-transform', open && 'rotate-180')} />
      </button>

      {open &&
        createPortal(
          <div
            ref={panelRef}
            role="listbox"
            id={listId}
            style={{
              position: 'fixed',
              top: position.placement === 'bottom' ? position.top : undefined,
              bottom: position.placement === 'top' ? window.innerHeight - position.top : undefined,
              left: position.left,
              width: position.width,
              zIndex: 1100,
            }}
            className="overflow-hidden rounded-xl border border-line bg-panel p-1.5 shadow-2xl shadow-slate-950/15 dark:shadow-black/40"
          >
            {searchable && (
              <div className="relative mb-1.5 px-1 pt-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted" />
                <input
                  ref={searchRef}
                  value={query}
                  onChange={(event) => {
                    setQuery(event.target.value);
                    setActiveIndex(0);
                  }}
                  placeholder="搜索..."
                  className="h-9 w-full rounded-lg border border-line bg-elevated pl-9 pr-3 text-sm text-text outline-none focus:border-primary/55"
                />
              </div>
            )}
            <div className="thin-scrollbar max-h-72 overflow-y-auto">
              {filteredOptions.length === 0 ? (
                <div className="px-3 py-6 text-center text-sm text-muted">无匹配项</div>
              ) : (
                filteredOptions.map((option, index) => {
                  const selected = option.value === currentValue;
                  const active = index === activeIndex;
                  return (
                    <button
                      key={option.value}
                      type="button"
                      role="option"
                      aria-selected={selected}
                      onClick={() => commit(option.value)}
                      onMouseEnter={() => setActiveIndex(index)}
                      className={cn(
                        'flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-sm transition',
                        selected ? 'bg-primary/10 text-primary' : 'text-text',
                        active && !selected && 'bg-elevated',
                      )}
                    >
                      {option.icon && <span className="flex h-5 w-5 shrink-0 items-center justify-center">{option.icon}</span>}
                      <span className="flex-1 truncate">{option.label}</span>
                      {selected && <Check className="h-4 w-4 shrink-0 text-primary" />}
                    </button>
                  );
                })
              )}
            </div>
          </div>,
          document.body
        )}
    </div>
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
