import { Activity, BarChart3, CheckCircle2, Edit, KeyRound, RefreshCw, RotateCcw, Trash2 } from 'lucide-react';
import { Button } from '../../components/ui/Button';
import { DataTable, tableCellClass, tableHeadClass } from '../../components/ui/DataTable';
import { StatusBadge } from '../../components/ui/StatusBadge';
import { formatCurrency, formatLatency, formatNumber } from '../../lib/format';
import type { UpstreamAccount, UpstreamActionName } from '../../types';
import { platformLabels } from './siteOptions';

interface SiteTableProps {
  accounts: UpstreamAccount[];
  selectedIds: string[];
  busyIds: string[];
  onToggleSelected: (id: string) => void;
  onToggleAll: () => void;
  onEdit: (account: UpstreamAccount) => void;
  onDelete: (account: UpstreamAccount) => void;
  onAction: (account: UpstreamAccount, action: UpstreamActionName) => void;
  onInspect: (account: UpstreamAccount) => void;
  onTestCall: (account: UpstreamAccount) => void;
}

export function SiteTable({
  accounts,
  selectedIds,
  busyIds,
  onToggleSelected,
  onToggleAll,
  onEdit,
  onDelete,
  onAction,
  onInspect,
  onTestCall,
}: SiteTableProps) {
  const allSelected = accounts.length > 0 && accounts.every((account) => selectedIds.includes(account.id));

  return (
    <DataTable>
      <thead className={tableHeadClass}>
        <tr>
          <th className={tableCellClass}>
            <input type="checkbox" className="accent-primary" checked={allSelected} onChange={onToggleAll} aria-label="选择全部站点" />
          </th>
          <th className={tableCellClass}>站点名称</th>
          <th className={tableCellClass}>平台</th>
          <th className={tableCellClass}>API 状态</th>
          <th className={tableCellClass}>账号凭据</th>
          <th className={tableCellClass}>模型</th>
          <th className={tableCellClass}>延迟</th>
          <th className={tableCellClass}>额度</th>
          <th className={tableCellClass}>操作</th>
        </tr>
      </thead>
      <tbody>
        {accounts.map((account) => {
          const busy = busyIds.includes(account.id);
          return (
            <tr key={account.id} className="hover:bg-elevated/55">
              <td className={tableCellClass}>
                <input
                  type="checkbox"
                  className="accent-primary"
                  checked={selectedIds.includes(account.id)}
                  onChange={() => onToggleSelected(account.id)}
                  aria-label={`选择 ${account.name}`}
                />
              </td>
              <td className={tableCellClass}>
                <div className="flex items-center gap-3">
                  <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-xs font-semibold text-white">{account.code.slice(0, 3).toUpperCase()}</span>
                  <div>
                    <div className="font-medium text-text">{account.name}</div>
                    <div className="max-w-[280px] truncate text-xs text-muted">{account.baseUrl}</div>
                    {account.note && <div className="mt-0.5 max-w-[280px] truncate text-xs text-muted/80">{account.note}</div>}
                  </div>
                </div>
              </td>
              <td className={tableCellClass}>
                <span className="rounded-md border border-line bg-elevated px-2 py-1 text-xs text-muted">{platformLabels[account.platformKind]}</span>
              </td>
              <td className={tableCellClass}>
                <StatusBadge status={account.enabled ? account.status.apiStatus : 'disabled'} />
              </td>
              <td className={tableCellClass}>
                <StatusBadge status={account.status.accountStatus} />
              </td>
              <td className={tableCellClass}>{formatNumber(account.status.modelCount)}</td>
              <td className={tableCellClass}>
                <span
                  className={
                    account.status.latencyMs == null
                      ? ''
                      : account.status.latencyMs < 200
                        ? 'text-green-500'
                        : account.status.latencyMs < 500
                          ? 'text-yellow-500'
                          : 'text-red-500'
                  }
                >
                  {formatLatency(account.status.latencyMs || undefined)}
                </span>
              </td>
              <td className={tableCellClass}>
                {formatBalance(account.status.balanceAmount, account.status.balanceUnit)}
              </td>
              <td className={tableCellClass}>
                <div className="flex items-center gap-1.5">
                  <IconButton label="历史" onClick={() => onInspect(account)} icon={<BarChart3 className="h-4 w-4" />} />
                  <IconButton label="编辑" onClick={() => onEdit(account)} icon={<Edit className="h-4 w-4" />} />
                  <IconButton label="测试 API" disabled={busy} onClick={() => onTestCall(account)} icon={<Activity className="h-4 w-4" />} />
                  <IconButton label="全量刷新" disabled={busy} onClick={() => onAction(account, 'refresh-all')} icon={<RefreshCw className="h-4 w-4" />} />
                  <IconButton label="删除" disabled={busy} onClick={() => onDelete(account)} icon={<Trash2 className="h-4 w-4" />} danger />
                </div>
              </td>
            </tr>
          );
        })}
      </tbody>
    </DataTable>
  );
}

function formatBalance(amount: number, unit: string): string {
  if (!unit) {
    return '-';
  }
  if (unit.toLowerCase() === 'usd') {
    return formatCurrency(amount);
  }
  return `${formatNumber(amount)} ${unit}`;
}

function IconButton({
  label,
  icon,
  onClick,
  disabled,
  danger,
}: {
  label: string;
  icon: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
  danger?: boolean;
}) {
  return (
    <Button variant={danger ? 'danger' : 'icon'} icon={icon} onClick={onClick} disabled={disabled} title={label} aria-label={label} />
  );
}
