export function cn(...classes: Array<string | false | null | undefined>): string {
  return classes.filter(Boolean).join(' ');
}

export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat('en-US').format(value);
}

export function formatPercent(value: number, digits = 1): string {
  return `${value.toFixed(digits)}%`;
}

export function formatLatency(value?: number): string {
  return typeof value === 'number' ? `${value} ms` : '-';
}

export function statusText(status: string): string {
  const map: Record<string, string> = {
    normal: '正常',
    warning: '部分异常',
    failed: '连接失败',
    maintenance: '维护中',
    offline: '离线',
    partial: '部分可用',
    unavailable: '不可用',
    checked: '已签到',
    unchecked: '未签到',
    disabled: '已禁用',
    success: '成功',
  };

  return map[status] ?? status;
}
