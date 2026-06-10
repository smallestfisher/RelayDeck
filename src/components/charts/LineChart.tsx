import type { ChartPoint } from '../../types';

interface LineChartProps {
  data: ChartPoint[];
  height?: number;
  valuePrefix?: string;
}

export function LineChart({ data, height = 210, valuePrefix = '' }: LineChartProps) {
  const width = 640;
  const padding = 34;
  const values = data.map((point) => point.value);
  const min = Math.min(...values, 0);
  const max = Math.max(...values);
  const range = max - min || 1;
  const step = (width - padding * 2) / Math.max(data.length - 1, 1);
  const points = data.map((point, index) => {
    const x = padding + index * step;
    const y = height - padding - ((point.value - min) / range) * (height - padding * 2);
    return { ...point, x, y };
  });
  const path = points.map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x} ${point.y}`).join(' ');
  const area = `${path} L ${points[points.length - 1]?.x ?? padding} ${height - padding} L ${padding} ${height - padding} Z`;

  return (
    <div className="w-full">
      <svg viewBox={`0 0 ${width} ${height}`} className="h-auto w-full overflow-visible">
        <defs>
          <linearGradient id="line-area" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="rgb(var(--color-primary))" stopOpacity="0.35" />
            <stop offset="100%" stopColor="rgb(var(--color-primary))" stopOpacity="0.02" />
          </linearGradient>
        </defs>
        {[0, 1, 2, 3].map((line) => {
          const y = padding + line * ((height - padding * 2) / 3);
          return <line key={line} x1={padding} x2={width - padding} y1={y} y2={y} stroke="rgb(var(--color-line))" strokeDasharray="4 4" />;
        })}
        <path d={area} fill="url(#line-area)" />
        <path d={path} fill="none" stroke="rgb(var(--color-primary))" strokeLinecap="round" strokeLinejoin="round" strokeWidth="3" />
        {points.map((point) => (
          <g key={point.label}>
            <circle cx={point.x} cy={point.y} r="4" fill="rgb(var(--color-primary))" stroke="rgb(var(--color-panel))" strokeWidth="2" />
            <text x={point.x} y={height - 8} textAnchor="middle" className="fill-muted text-[12px]">
              {point.label}
            </text>
          </g>
        ))}
      </svg>
      <div className="mt-2 text-right text-xs text-muted">
        峰值 {valuePrefix}
        {max.toLocaleString('en-US')}
      </div>
    </div>
  );
}
