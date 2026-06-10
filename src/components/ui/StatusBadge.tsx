import { cn, statusText } from '../../lib/format';

interface StatusBadgeProps {
  status: string;
  label?: string;
  className?: string;
}

function statusClass(status: string): string {
  if (['normal', 'success', 'checked', 'closed'].includes(status)) {
    return 'border-success/25 bg-success/10 text-success';
  }
  if (['warning', 'partial', 'maintenance', 'cooldown'].includes(status)) {
    return 'border-warning/30 bg-warning/12 text-warning';
  }
  if (['failed', 'offline', 'unavailable', 'unchecked', 'open'].includes(status)) {
    return 'border-danger/30 bg-danger/12 text-danger';
  }
  if (status === 'disabled') {
    return 'border-line bg-elevated text-muted';
  }
  return 'border-info/25 bg-info/10 text-info';
}

export function StatusBadge({ status, label, className }: StatusBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex h-7 min-w-0 items-center justify-center rounded-md border px-2.5 text-xs font-medium',
        statusClass(status),
        className
      )}
    >
      {label ?? statusText(status)}
    </span>
  );
}
