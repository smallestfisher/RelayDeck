interface MiniTrendProps {
  data: number[];
  className?: string;
}

export function MiniTrend({ data, className }: MiniTrendProps) {
  const width = 94;
  const height = 36;
  const min = Math.min(...data);
  const max = Math.max(...data);
  const range = max - min || 1;
  const step = width / Math.max(data.length - 1, 1);
  const path = data
    .map((value, index) => {
      const x = index * step;
      const y = height - ((value - min) / range) * (height - 8) - 4;
      return `${index === 0 ? 'M' : 'L'} ${x} ${y}`;
    })
    .join(' ');

  return (
    <svg viewBox={`0 0 ${width} ${height}`} className={className ?? 'h-9 w-24'}>
      <path d={`${path} L ${width} ${height} L 0 ${height} Z`} fill="rgb(34 197 94 / 0.16)" />
      <path d={path} fill="none" stroke="#22c55e" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" />
    </svg>
  );
}
