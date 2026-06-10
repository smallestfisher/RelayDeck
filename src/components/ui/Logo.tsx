import { cn } from '../../lib/format';

interface LogoProps {
  compact?: boolean;
  className?: string;
}

export function Logo({ compact = false, className }: LogoProps) {
  return (
    <div className={cn('flex items-center gap-3', className)}>
      <div className="relative h-11 w-9 shrink-0">
        <div className="absolute left-0 top-1 h-8 w-4 rounded-md bg-gradient-to-b from-sky-400 to-blue-700 shadow-[0_0_24px_rgb(37_99_235/0.55)] -skew-x-12" />
        <div className="absolute bottom-0 right-0 h-8 w-4 rounded-md bg-gradient-to-b from-blue-400 to-cyan-500 shadow-[0_0_28px_rgb(14_165_233/0.5)] -skew-x-12" />
      </div>
      {!compact && (
        <div className="min-w-0">
          <div className="text-xl font-semibold leading-tight text-text">RelayDeck</div>
          <div className="mt-0.5 text-sm text-muted">模型线路中控台</div>
        </div>
      )}
    </div>
  );
}
