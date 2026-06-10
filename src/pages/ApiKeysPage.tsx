import { Ban, Clock3, Copy, KeyRound, MoreHorizontal, Plus, RotateCw, ShieldCheck, ShieldEllipsis } from 'lucide-react';
import { useMemo, useState } from 'react';
import { apiKeyRecords } from '../data/mock';
import type { ApiKeyRecord, ApiKeyScope, ApiKeyStatus } from '../types';
import { cn, formatNumber, formatPercent } from '../lib/format';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { SearchInput, SelectControl } from '../components/ui/Controls';
import { DataTable, tableCellClass, tableHeadClass } from '../components/ui/DataTable';
import { Drawer } from '../components/ui/Drawer';
import { MetricCard } from '../components/ui/MetricCard';
import { StatusBadge } from '../components/ui/StatusBadge';

const statusOptions = [
  { label: '状态：全部', value: 'all' },
  { label: '活跃', value: 'active' },
  { label: '即将过期', value: 'expiring' },
  { label: '已撤销', value: 'blocked' },
  { label: '已过期', value: 'expired' },
  { label: '从未使用', value: 'unused' },
];

const scopeOptions = [
  { label: '权限：全部', value: 'all' },
  { label: 'Chat', value: 'Chat' },
  { label: 'Embedding', value: 'Embedding' },
  { label: 'Vision', value: 'Vision' },
  { label: 'Admin', value: 'Admin' },
];

const expiryOptions = [
  { label: '到期时间：全部', value: 'all' },
  { label: '7 天内到期', value: 'soon' },
  { label: '已过期', value: 'expired' },
  { label: '长期有效', value: 'open' },
];

function scopeClass(scope: ApiKeyScope): string {
  if (scope === 'Chat') return 'border-primary/25 bg-primary/10 text-primary';
  if (scope === 'Embedding') return 'border-success/25 bg-success/10 text-success';
  if (scope === 'Vision') return 'border-violet-400/30 bg-violet-500/12 text-violet-300';
  return 'border-warning/30 bg-warning/12 text-warning';
}

function matchesExpiry(key: ApiKeyRecord, expiry: string): boolean {
  if (expiry === 'all') return true;
  if (expiry === 'expired') return key.status === 'expired';
  if (expiry === 'soon') return key.status === 'expiring';
  return key.expiresAt === '-';
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-4 border-b border-line/70 py-2 text-sm last:border-b-0">
      <span className="text-muted">{label}</span>
      <span className="text-right font-medium text-text">{value}</span>
    </div>
  );
}

export function ApiKeysPage() {
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState('all');
  const [scope, setScope] = useState('all');
  const [owner, setOwner] = useState('all');
  const [expiry, setExpiry] = useState('all');
  const [selectedId, setSelectedId] = useState(apiKeyRecords[0].id);
  const [drawerOpen, setDrawerOpen] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);

  const selectedKey = apiKeyRecords.find((key) => key.id === selectedId) ?? apiKeyRecords[0];
  const ownerOptions = useMemo(() => {
    const owners = Array.from(new Set(apiKeyRecords.map((key) => key.ownerName)));
    return [{ label: '用户：全部', value: 'all' }, ...owners.map((name) => ({ label: name, value: name }))];
  }, []);

  const filteredKeys = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return apiKeyRecords.filter((key) => {
      const matchesQuery =
        !keyword ||
        key.name.toLowerCase().includes(keyword) ||
        key.ownerName.toLowerCase().includes(keyword) ||
        key.prefix.toLowerCase().includes(keyword);
      const matchesStatus = status === 'all' || key.status === (status as ApiKeyStatus);
      const matchesScope = scope === 'all' || key.scopes.includes(scope as ApiKeyScope);
      const matchesOwner = owner === 'all' || key.ownerName === owner;
      return matchesQuery && matchesStatus && matchesScope && matchesOwner && matchesExpiry(key, expiry);
    });
  }, [expiry, owner, query, scope, status]);

  const activeCount = apiKeyRecords.filter((key) => key.status === 'active').length;
  const expiringCount = apiKeyRecords.filter((key) => key.status === 'expiring').length;
  const blockedCount = apiKeyRecords.filter((key) => key.status === 'blocked').length;

  function openKey(key: ApiKeyRecord) {
    setSelectedId(key.id);
    setDrawerOpen(true);
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-text">API Key 管理</h1>
          <p className="mt-1 text-sm text-muted">创建与管理访问密钥、调用权限和安全策略</p>
        </div>
        <div className="flex gap-2">
          <Button variant="secondary" icon={<Ban className="h-4 w-4" />}>批量禁用</Button>
          <Button variant="secondary" icon={<ShieldEllipsis className="h-4 w-4" />}>访问策略</Button>
          <Button variant="primary" icon={<Plus className="h-4 w-4" />} onClick={() => setCreateOpen(true)}>创建 Key</Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="Key 总数" value={String(apiKeyRecords.length)} detail="较昨日" delta="+3" tone="info" icon={<KeyRound className="h-6 w-6" />} />
        <MetricCard label="活跃 Key" value={String(activeCount)} detail="较昨日" delta="+2" tone="success" icon={<ShieldCheck className="h-6 w-6" />} />
        <MetricCard label="即将过期" value={String(expiringCount)} detail="7 天内到期" delta="" tone="warning" icon={<Clock3 className="h-6 w-6" />} />
        <MetricCard label="已撤销" value={String(blockedCount)} detail="较昨日" delta="+1" tone="danger" icon={<Ban className="h-6 w-6" />} />
      </div>

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <SearchInput className="w-full sm:w-[280px]" placeholder="搜索 Key 名称、用户或前缀" value={query} onChange={(event) => setQuery(event.target.value)} />
          <SelectControl className="w-36" options={statusOptions} value={status} onChange={(event) => setStatus(event.target.value)} />
          <SelectControl className="w-36" options={scopeOptions} value={scope} onChange={(event) => setScope(event.target.value)} />
          <SelectControl className="w-36" options={ownerOptions} value={owner} onChange={(event) => setOwner(event.target.value)} />
          <SelectControl className="w-44" options={expiryOptions} value={expiry} onChange={(event) => setExpiry(event.target.value)} />
          <Button className="ml-auto" variant="icon" aria-label="刷新列表">
            <RotateCw className="h-4 w-4" />
          </Button>
        </div>

        <DataTable>
          <thead className={tableHeadClass}>
            <tr>
              <th className={tableCellClass}>Key 名称</th>
              <th className={tableCellClass}>所属用户</th>
              <th className={tableCellClass}>前缀</th>
              <th className={tableCellClass}>权限范围</th>
              <th className={tableCellClass}>限流策略</th>
              <th className={tableCellClass}>允许模型</th>
              <th className={tableCellClass}>创建时间</th>
              <th className={tableCellClass}>到期时间</th>
              <th className={tableCellClass}>最近使用</th>
              <th className={tableCellClass}>状态</th>
              <th className={tableCellClass}>操作</th>
            </tr>
          </thead>
          <tbody>
            {filteredKeys.map((key) => (
              <tr
                key={key.id}
                className={cn('cursor-pointer hover:bg-elevated/55', selectedId === key.id && 'bg-primary/10')}
                onClick={() => openKey(key)}
              >
                <td className={tableCellClass}>
                  <div className="font-medium text-text">{key.name}</div>
                  <div className="text-xs text-muted">{key.id}</div>
                </td>
                <td className={tableCellClass}>{key.ownerName}</td>
                <td className={tableCellClass}>
                  <div className="flex items-center gap-2">
                    <code className="rounded bg-elevated px-2 py-1 text-xs text-text">{key.prefix}</code>
                    <Copy className="h-3.5 w-3.5 text-muted" />
                  </div>
                </td>
                <td className={tableCellClass}>
                  <div className="flex flex-wrap gap-1.5">
                    {key.scopes.map((item) => (
                      <span key={item} className={cn('rounded-md border px-2 py-1 text-xs font-medium', scopeClass(item))}>{item}</span>
                    ))}
                  </div>
                </td>
                <td className={tableCellClass}>{key.rateLimit}</td>
                <td className={tableCellClass}>{key.allowedModels.length}</td>
                <td className={tableCellClass}>{key.createdAt}</td>
                <td className={tableCellClass}>{key.expiresAt}</td>
                <td className={tableCellClass}>{key.lastUsed}</td>
                <td className={tableCellClass}>
                  <StatusBadge status={key.status} />
                </td>
                <td className={tableCellClass}>
                  <Button variant="icon" className="h-8 w-8" onClick={(event) => { event.stopPropagation(); openKey(key); }}>
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </DataTable>

        <div className="mt-4 flex items-center justify-between text-sm text-muted">
          <span>共 {filteredKeys.length} 条记录</span>
          <div className="flex items-center gap-2">
            <Button variant="secondary" className="h-9 w-9 p-0">1</Button>
            <Button variant="ghost" className="h-9 w-9 p-0">2</Button>
            <Button variant="ghost" className="h-9 w-9 p-0">3</Button>
          </div>
        </div>
      </Card>

      <Drawer
        open={drawerOpen}
        title="Key 详情"
        subtitle={selectedKey.name}
        onClose={() => setDrawerOpen(false)}
        footer={
          <div className="grid grid-cols-3 gap-3">
            <Button variant="secondary" icon={<Copy className="h-4 w-4" />}>复制 Key</Button>
            <Button variant="secondary" icon={<RotateCw className="h-4 w-4" />}>重新生成</Button>
            <Button variant="danger" icon={<Ban className="h-4 w-4" />}>撤销</Button>
          </div>
        }
      >
        <div className="space-y-6">
          <div className="rounded-xl border border-line bg-elevated/45 p-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <div className="text-sm text-muted">Key 名称</div>
                <div className="mt-1 font-semibold text-text">{selectedKey.name}</div>
              </div>
              <StatusBadge status={selectedKey.status} />
            </div>
            <div className="flex items-center gap-2 rounded-lg border border-line bg-panel px-3 py-2">
              <code className="min-w-0 flex-1 truncate text-sm text-text">{selectedKey.maskedKey}</code>
              <Copy className="h-4 w-4 text-muted" />
            </div>
          </div>

          <div>
            <h3 className="text-sm font-semibold text-text">基本信息</h3>
            <div className="mt-3 rounded-lg border border-line bg-elevated/35 px-4 py-2">
              <DetailRow label="所属用户" value={selectedKey.ownerName} />
              <DetailRow label="用户邮箱" value={selectedKey.ownerEmail} />
              <DetailRow label="创建时间" value={selectedKey.createdAt} />
              <DetailRow label="到期时间" value={selectedKey.expiresAt} />
              <DetailRow label="最近使用" value={selectedKey.lastUsed} />
            </div>
          </div>

          <div>
            <h3 className="text-sm font-semibold text-text">权限范围</h3>
            <div className="mt-3 flex flex-wrap gap-2">
              {selectedKey.scopes.map((item) => (
                <span key={item} className={cn('rounded-md border px-2.5 py-1 text-xs font-medium', scopeClass(item))}>{item}</span>
              ))}
            </div>
            <div className="mt-4 flex flex-wrap gap-2">
              {selectedKey.allowedModels.map((item) => (
                <span key={item} className="rounded-md border border-line bg-elevated px-2.5 py-1 text-xs text-muted">{item}</span>
              ))}
            </div>
          </div>

          <div>
            <h3 className="text-sm font-semibold text-text">安全与限流</h3>
            <div className="mt-3 rounded-lg border border-line bg-elevated/35 px-4 py-2">
              <DetailRow label="IP 白名单" value={selectedKey.ipWhitelist} />
              <DetailRow label="请求速率" value={selectedKey.rateLimit} />
              <DetailRow label="每日额度" value={selectedKey.dailyLimit} />
            </div>
            <div className="mt-4 rounded-lg bg-elevated/55 p-4">
              <div className="mb-2 flex items-center justify-between text-sm">
                <span className="text-muted">已用 Tokens</span>
                <span className="font-medium text-text">{formatNumber(selectedKey.usageTokens)} ({formatPercent(selectedKey.usagePercent, 1)})</span>
              </div>
              <span className="block h-2 overflow-hidden rounded-full bg-line">
                <span className="block h-full rounded-full bg-primary" style={{ width: `${selectedKey.usagePercent}%` }} />
              </span>
            </div>
          </div>

          <div>
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-text">最近调用日志</h3>
              <button type="button" className="text-xs text-primary">查看全部</button>
            </div>
            <div className="mt-3 space-y-2">
              {selectedKey.recentCalls.length === 0 ? (
                <div className="rounded-lg border border-dashed border-line p-4 text-center text-sm text-muted">暂无调用记录</div>
              ) : (
                selectedKey.recentCalls.map((call) => (
                  <div key={call.id} className="flex items-center justify-between gap-3 rounded-lg border border-line bg-elevated/35 px-3 py-2 text-sm">
                    <span className="truncate text-text">{call.endpoint}</span>
                    <span className={call.statusCode < 300 ? 'text-success' : 'text-danger'}>{call.statusCode}</span>
                    <span className="text-xs text-muted">{call.time}</span>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </Drawer>

      {createOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/45 px-4">
          <section className="w-full max-w-3xl rounded-xl border border-line bg-panel p-6 shadow-2xl">
            <div className="mb-5 flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold text-text">创建新 Key</h2>
                <p className="mt-1 text-sm text-muted">配置所属用户、权限范围、模型白名单和限流规则</p>
              </div>
              <Button variant="ghost" className="h-9 w-9 p-0" onClick={() => setCreateOpen(false)}>×</Button>
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <label>
                <span className="text-sm font-medium text-text">Key 名称 *</span>
                <input className="mt-2 h-11 w-full rounded-lg border border-line bg-elevated px-3 text-sm outline-none focus:border-primary/55" placeholder="例如：生产环境 Key" />
              </label>
              <label>
                <span className="text-sm font-medium text-text">绑定用户 *</span>
                <SelectControl className="mt-2" options={ownerOptions.slice(1)} />
              </label>
              <div>
                <span className="text-sm font-medium text-text">权限范围 *</span>
                <div className="mt-2 flex flex-wrap gap-2">
                  {(['Chat', 'Embedding', 'Vision', 'Admin'] as ApiKeyScope[]).map((item) => (
                    <label key={item} className="flex items-center gap-2 rounded-lg border border-line bg-elevated px-3 py-2 text-sm text-text">
                      <input type="checkbox" defaultChecked={item !== 'Admin'} className="accent-primary" />
                      {item}
                    </label>
                  ))}
                </div>
              </div>
              <label>
                <span className="text-sm font-medium text-text">模型白名单 *</span>
                <SelectControl className="mt-2" options={[
                  { label: '全部可用模型', value: 'all' },
                  { label: 'Chat 模型', value: 'chat' },
                  { label: 'Embedding 模型', value: 'embedding' },
                ]} />
              </label>
              <label>
                <span className="text-sm font-medium text-text">IP 白名单</span>
                <input className="mt-2 h-11 w-full rounded-lg border border-line bg-elevated px-3 text-sm outline-none focus:border-primary/55" placeholder="多个网段用逗号分隔" />
              </label>
              <label>
                <span className="text-sm font-medium text-text">到期时间</span>
                <input className="mt-2 h-11 w-full rounded-lg border border-line bg-elevated px-3 text-sm outline-none focus:border-primary/55" placeholder="选择日期" />
              </label>
              <label>
                <span className="text-sm font-medium text-text">每分钟请求数（RPM）*</span>
                <input className="mt-2 h-11 w-full rounded-lg border border-line bg-elevated px-3 text-sm outline-none focus:border-primary/55" placeholder="100" />
              </label>
              <label>
                <span className="text-sm font-medium text-text">每日调用额度（Tokens）*</span>
                <input className="mt-2 h-11 w-full rounded-lg border border-line bg-elevated px-3 text-sm outline-none focus:border-primary/55" placeholder="1000000" />
              </label>
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <Button onClick={() => setCreateOpen(false)}>取消</Button>
              <Button variant="primary" icon={<KeyRound className="h-4 w-4" />} onClick={() => setCreateOpen(false)}>创建</Button>
            </div>
          </section>
        </div>
      )}
    </div>
  );
}
