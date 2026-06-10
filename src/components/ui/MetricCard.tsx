import type { ReactNode } from 'react';
import type { AlertSeverity } from '../../types';
import { cn } from '../../lib/format';
import { Card } from './Card';

interface MetricCardProps {
  label: string;
  value: string;
  delta?: string;
  icon: ReactNode;
  tone?: AlertSeverity;
  detail?: string;
  className?: string;
}

const toneClass: Record<AlertSeverity, string> = {
  info: 'bg-info/12 text-info',
  success: 'bg-success/12 text-success',
  warning: 'bg-warning/14 text-warning',
  danger: 'bg-danger/12 text-danger',
};

export function MetricCard({ label, value, delta, icon, tone = 'info', detail, className }: MetricCardProps) {
  return (
    <Card className={cn('p-4', className)}>
      <div className="flex items-center gap-4">
        <div className={cn('flex h-12 w-12 shrink-0 items-center justify-center rounded-xl', toneClass[tone])}>
          {icon}
        </div>
        <div className="min-w-0">
          <div className="text-sm text-muted">{label}</div>
          <div className="mt-1 text-2xl font-semibold leading-none text-text">{value}</div>
          {(delta || detail) && (
            <div className="mt-2 flex items-center gap-2 text-xs text-muted">
              {detail && <span>{detail}</span>}
              {delta && <span className={cn(tone === 'danger' ? 'text-danger' : 'text-success')}>{delta}</span>}
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}
