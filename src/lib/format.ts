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
    healthy: '健康',
    warning: '部分异常',
    failed: '连接失败',
    maintenance: '维护中',
    offline: '离线',
    unknown: '未知',
    valid: '有效',
    not_configured: '未配置',
    action_required: '需人工处理',
    unsupported: '不支持',
    partial: '部分可用',
    unavailable: '不可用',
    checked: '已签到',
    unchecked: '未签到',
    disabled: '已禁用',
    success: '成功',
    active: '活跃',
    inactive: '待激活',
    blocked: '已停用',
    expiring: '即将过期',
    expired: '已过期',
    unused: '从未使用',
  };

  return map[status] ?? status;
}
