import type { ChartPoint } from '../../types';

const colors = ['#22c55e', '#f59e0b', '#ef4444', '#3b82f6', '#8b5cf6', '#94a3b8'];

interface RingChartProps {
  data: ChartPoint[];
  centerLabel: string;
  centerValue: string;
  size?: number;
}

function segmentPath(cx: number, cy: number, r: number, start: number, end: number) {
  const startRad = ((start - 90) * Math.PI) / 180;
  const endRad = ((end - 90) * Math.PI) / 180;
  const x1 = cx + r * Math.cos(startRad);
  const y1 = cy + r * Math.sin(startRad);
  const x2 = cx + r * Math.cos(endRad);
  const y2 = cy + r * Math.sin(endRad);
  const largeArc = end - start > 180 ? 1 : 0;
  return `M ${x1} ${y1} A ${r} ${r} 0 ${largeArc} 1 ${x2} ${y2}`;
}

export function RingChart({ data, centerLabel, centerValue, size = 176 }: RingChartProps) {
  const total = data.reduce((sum, item) => sum + item.value, 0) || 1;
  let current = 0;
  const stroke = 24;
  const radius = size / 2 - stroke;

  return (
    <div className="flex items-center gap-6">
      <div className="relative shrink-0" style={{ width: size, height: size }}>
        <svg viewBox={`0 0 ${size} ${size}`} className="h-full w-full">
          <circle cx={size / 2} cy={size / 2} r={radius} fill="none" stroke="rgb(var(--color-line))" strokeWidth={stroke} />
          {data.map((item, index) => {
            const degrees = (item.value / total) * 360;
            const path = segmentPath(size / 2, size / 2, radius, current, current + degrees - 2);
            current += degrees;
            return (
              <path
                key={item.label}
                d={path}
                fill="none"
                stroke={colors[index % colors.length]}
                strokeLinecap="round"
                strokeWidth={stroke}
              />
            );
          })}
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center text-center">
          <div className="text-2xl font-semibold text-text">{centerValue}</div>
          <div className="mt-1 text-xs text-muted">{centerLabel}</div>
        </div>
      </div>
      <div className="min-w-0 flex-1 space-y-3">
        {data.map((item, index) => (
          <div key={item.label} className="flex items-center justify-between gap-3 text-sm">
            <span className="flex min-w-0 items-center gap-2 text-text">
              <span className="h-2.5 w-2.5 rounded-sm" style={{ backgroundColor: colors[index % colors.length] }} />
              <span className="truncate">{item.label}</span>
            </span>
            <span className="shrink-0 text-muted">{item.value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
