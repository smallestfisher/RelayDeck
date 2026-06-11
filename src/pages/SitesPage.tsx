import { AlertTriangle, Plus, RefreshCw, ShieldCheck, Signal, XCircle } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { SearchInput, SelectControl } from '../components/ui/Controls';
import { Drawer } from '../components/ui/Drawer';
import { MetricCard } from '../components/ui/MetricCard';
import { adminApi } from '../lib/adminApi';
import { formatLatency, formatNumber } from '../lib/format';
import type {
  AccountCredentialStatus,
  UpstreamAccount,
  UpstreamAccountEvent,
  UpstreamAccountInput,
  UpstreamActionName,
  UpstreamAPIStatus,
  UpstreamBatchActionName,
  UpstreamCheckinStatus,
  UpstreamPlatformKind,
} from '../types';
import { SiteDrawer } from './sites/SiteDrawer';
import { SiteTable } from './sites/SiteTable';
import { accountStatusOptions, apiStatusOptions, checkinStatusOptions, platformOptions } from './sites/siteOptions';

type LatencyBand = 'all' | 'low' | 'medium' | 'high' | 'unknown';

const latencyOptions: Array<{ label: string; value: LatencyBand }> = [
  { label: '延迟：全部', value: 'all' },
  { label: '< 300ms', value: 'low' },
  { label: '300-1000ms', value: 'medium' },
  { label: '> 1000ms', value: 'high' },
  { label: '未知', value: 'unknown' },
];

export function SitesPage() {
  const [accounts, setAccounts] = useState<UpstreamAccount[]>([]);
  const [events, setEvents] = useState<UpstreamAccountEvent[]>([]);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [busyIds, setBusyIds] = useState<string[]>([]);
  const [query, setQuery] = useState('');
  const [platform, setPlatform] = useState<UpstreamPlatformKind | 'all'>('all');
  const [apiStatus, setApiStatus] = useState<UpstreamAPIStatus | 'all'>('all');
  const [accountStatus, setAccountStatus] = useState<AccountCredentialStatus | 'all'>('all');
  const [checkinStatus, setCheckinStatus] = useState<UpstreamCheckinStatus | 'all'>('all');
  const [latencyBand, setLatencyBand] = useState<LatencyBand>('all');
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editingAccount, setEditingAccount] = useState<UpstreamAccount | null>(null);
  const [inspectAccount, setInspectAccount] = useState<UpstreamAccount | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testingDraft, setTestingDraft] = useState(false);
  const [error, setError] = useState('');
  const [drawerError, setDrawerError] = useState('');

  useEffect(() => {
    void loadAccounts();
  }, []);

  const filteredAccounts = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return accounts.filter((account) => {
      const matchesQuery =
        !keyword ||
        account.name.toLowerCase().includes(keyword) ||
        account.code.toLowerCase().includes(keyword) ||
        account.baseUrl.toLowerCase().includes(keyword) ||
        account.note.toLowerCase().includes(keyword);
      const matchesPlatform = platform === 'all' || account.platformKind === platform;
      const matchesAPIStatus = apiStatus === 'all' || account.status.apiStatus === apiStatus;
      const matchesAccountStatus = accountStatus === 'all' || account.status.accountStatus === accountStatus;
      const matchesCheckinStatus = checkinStatus === 'all' || account.status.checkinStatus === checkinStatus;
      const matchesLatency = latencyBand === 'all' || latencyMatches(account.status.latencyMs, latencyBand);
      return matchesQuery && matchesPlatform && matchesAPIStatus && matchesAccountStatus && matchesCheckinStatus && matchesLatency;
    });
  }, [accounts, accountStatus, apiStatus, checkinStatus, latencyBand, platform, query]);

  const metrics = useMemo(() => {
    const healthy = accounts.filter((account) => account.enabled && account.status.apiStatus === 'healthy').length;
    const warning = accounts.filter((account) => account.status.apiStatus === 'warning' || account.status.accountStatus === 'expired').length;
    const manual = accounts.filter(
      (account) =>
        account.status.accountStatus === 'action_required' ||
        account.status.checkinStatus === 'action_required' ||
        account.status.accountStatus === 'not_configured'
    ).length;
    return { total: accounts.length, healthy, warning, manual };
  }, [accounts]);

  async function loadAccounts() {
    setLoading(true);
    setError('');
    try {
      setAccounts(await adminApi.listUpstreamAccounts());
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载站点失败');
    } finally {
      setLoading(false);
    }
  }

  async function saveAccount(input: UpstreamAccountInput) {
    const validation = validateAccountInput(input, Boolean(editingAccount));
    if (validation) {
      setDrawerError(validation);
      return;
    }
    setSaving(true);
    setDrawerError('');
    try {
      if (editingAccount) {
        await adminApi.updateUpstreamAccount(editingAccount.id, input);
      } else {
        await adminApi.createUpstreamAccount(input);
      }
      setDrawerOpen(false);
      setEditingAccount(null);
      await loadAccounts();
    } catch (err) {
      setDrawerError(err instanceof Error ? err.message : '保存失败');
    } finally {
      setSaving(false);
    }
  }

  async function runDraftTest(input: UpstreamAccountInput) {
    if (editingAccount) {
      await runAction(editingAccount, 'test-api');
      return;
    }
    const validation = validateAccountInput(input, false);
    setDrawerError(validation || '请先保存站点后执行 API 测试');
    setTestingDraft(false);
  }

  async function runAction(account: UpstreamAccount, action: UpstreamActionName) {
    setBusyIds((current) => [...new Set([...current, account.id])]);
    try {
      await adminApi.runUpstreamAction(account.id, action);
      await loadAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : '操作失败');
    } finally {
      setBusyIds((current) => current.filter((id) => id !== account.id));
    }
  }

  async function runBatch(action: UpstreamBatchActionName) {
    if (selectedIds.length === 0) return;
    setBusyIds((current) => [...new Set([...current, ...selectedIds])]);
    try {
      await adminApi.runBatchUpstreamAction(selectedIds, action);
      await loadAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : '批量操作失败');
    } finally {
      setBusyIds((current) => current.filter((id) => !selectedIds.includes(id)));
    }
  }

  async function deleteAccount(account: UpstreamAccount) {
    if (!window.confirm(`删除站点 ${account.name}？`)) return;
    setBusyIds((current) => [...current, account.id]);
    try {
      await adminApi.deleteUpstreamAccount(account.id);
      setSelectedIds((current) => current.filter((id) => id !== account.id));
      await loadAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除失败');
    } finally {
      setBusyIds((current) => current.filter((id) => id !== account.id));
    }
  }

  async function inspect(account: UpstreamAccount) {
    setInspectAccount(account);
    try {
      setEvents(await adminApi.listUpstreamEvents(account.id));
    } catch {
      setEvents([]);
    }
  }

  function toggleSelected(id: string) {
    setSelectedIds((current) => (current.includes(id) ? current.filter((item) => item !== id) : [...current, id]));
  }

  function toggleAll() {
    const visibleIds = filteredAccounts.map((account) => account.id);
    const allSelected = visibleIds.length > 0 && visibleIds.every((id) => selectedIds.includes(id));
    setSelectedIds(allSelected ? selectedIds.filter((id) => !visibleIds.includes(id)) : [...new Set([...selectedIds, ...visibleIds])]);
  }

  function openCreateDrawer() {
    setEditingAccount(null);
    setDrawerError('');
    setDrawerOpen(true);
  }

  function openEditDrawer(account: UpstreamAccount) {
    setEditingAccount(account);
    setDrawerError('');
    setDrawerOpen(true);
  }

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-text">站点管理</h1>
          <p className="mt-1 text-sm text-muted">管理 New API / Sub2API 上游普通用户账号，统一模型、额度、健康状态与分发能力</p>
        </div>
        <Button variant="primary" icon={<Plus className="h-4 w-4" />} onClick={openCreateDrawer}>
          添加站点
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="站点总数" value={formatNumber(metrics.total)} detail="已接入账号" delta="" tone="info" icon={<Signal className="h-6 w-6" />} />
        <MetricCard label="健康 API" value={formatNumber(metrics.healthy)} detail="可用于模型调用" delta="" tone="success" icon={<ShieldCheck className="h-6 w-6" />} />
        <MetricCard label="异常警告" value={formatNumber(metrics.warning)} detail="需复查状态" delta="" tone="warning" icon={<AlertTriangle className="h-6 w-6" />} />
        <MetricCard label="待处理" value={formatNumber(metrics.manual)} detail="凭据或人工动作" delta="" tone="danger" icon={<XCircle className="h-6 w-6" />} />
      </div>

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <SearchInput className="w-[270px]" placeholder="搜索站点、代码、URL 或备注" value={query} onChange={(event) => setQuery(event.target.value)} />
          <SelectControl
            className="w-36"
            options={[{ label: '平台：全部', value: 'all' }, ...platformOptions]}
            value={platform}
            onChange={(event) => setPlatform(event.target.value as UpstreamPlatformKind | 'all')}
          />
          <SelectControl className="w-40" options={apiStatusOptions} value={apiStatus} onChange={(event) => setApiStatus(event.target.value as UpstreamAPIStatus | 'all')} />
          <SelectControl
            className="w-44"
            options={accountStatusOptions}
            value={accountStatus}
            onChange={(event) => setAccountStatus(event.target.value as AccountCredentialStatus | 'all')}
          />
          <SelectControl
            className="w-40"
            options={checkinStatusOptions}
            value={checkinStatus}
            onChange={(event) => setCheckinStatus(event.target.value as UpstreamCheckinStatus | 'all')}
          />
          <SelectControl className="w-36" options={latencyOptions} value={latencyBand} onChange={(event) => setLatencyBand(event.target.value as LatencyBand)} />
          <div className="ml-auto flex gap-2">
            <Button variant="secondary" icon={<RefreshCw className="h-4 w-4" />} onClick={loadAccounts} disabled={loading}>
              刷新
            </Button>
            <Button variant="secondary" onClick={() => runBatch('test-api')} disabled={selectedIds.length === 0}>
              批量检测
            </Button>
            <Button variant="secondary" onClick={() => runBatch('sync-models')} disabled={selectedIds.length === 0}>
              批量同步
            </Button>
          </div>
        </div>

        {error && <div className="mb-4 rounded-lg border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger">{error}</div>}

        {loading ? (
          <div className="rounded-lg border border-line bg-elevated/45 py-12 text-center text-sm text-muted">加载中...</div>
        ) : filteredAccounts.length === 0 ? (
          <div className="rounded-lg border border-line bg-elevated/45 py-12 text-center text-sm text-muted">暂无站点账号</div>
        ) : (
          <SiteTable
            accounts={filteredAccounts}
            selectedIds={selectedIds}
            busyIds={busyIds}
            onToggleSelected={toggleSelected}
            onToggleAll={toggleAll}
            onEdit={openEditDrawer}
            onDelete={deleteAccount}
            onAction={runAction}
            onInspect={inspect}
          />
        )}

        <div className="mt-4 flex items-center justify-between text-sm text-muted">
          <span>
            共 {filteredAccounts.length} 条记录，已选 {selectedIds.length} 条
          </span>
          <div className="flex items-center gap-2">
            <Button variant="secondary" className="h-9 w-9 p-0">
              1
            </Button>
          </div>
        </div>
      </Card>

      <SiteDrawer
        open={drawerOpen}
        account={editingAccount}
        saving={saving}
        testing={testingDraft}
        error={drawerError}
        onClose={() => setDrawerOpen(false)}
        onSave={saveAccount}
        onTestAPI={(input) => {
          setTestingDraft(true);
          void runDraftTest(input).finally(() => setTestingDraft(false));
        }}
      />

      <Drawer open={Boolean(inspectAccount)} title="站点历史" subtitle={inspectAccount?.name} onClose={() => setInspectAccount(null)}>
        <div className="space-y-3">
          {inspectAccount && (
            <div className="rounded-lg border border-line bg-elevated/45 p-4 text-sm text-muted">
              <div className="font-medium text-text">{inspectAccount.baseUrl}</div>
              <div className="mt-2 grid grid-cols-2 gap-2">
                <span>API：{inspectAccount.status.apiStatus}</span>
                <span>延迟：{formatLatency(inspectAccount.status.latencyMs || undefined)}</span>
                <span>模型：{formatNumber(inspectAccount.status.modelCount)}</span>
                <span>账号：{inspectAccount.status.accountStatus}</span>
              </div>
            </div>
          )}
          {events.length === 0 ? (
            <div className="rounded-lg border border-line bg-elevated/45 py-10 text-center text-sm text-muted">暂无操作历史</div>
          ) : (
            events.map((event) => (
              <div key={event.id} className="rounded-lg border border-line bg-elevated/45 p-4">
                <div className="flex items-center justify-between gap-3">
                  <span className="text-sm font-medium text-text">{event.operation}</span>
                  <span className="text-xs text-muted">{event.createdAt}</span>
                </div>
                <div className="mt-2 text-sm text-muted">{event.message || event.status}</div>
              </div>
            ))
          )}
        </div>
      </Drawer>
    </div>
  );
}

function latencyMatches(latencyMs: number, band: LatencyBand): boolean {
  if (band === 'unknown') return !latencyMs;
  if (!latencyMs) return false;
  if (band === 'low') return latencyMs < 300;
  if (band === 'medium') return latencyMs >= 300 && latencyMs <= 1000;
  if (band === 'high') return latencyMs > 1000;
  return true;
}

function validateAccountInput(input: UpstreamAccountInput, editing: boolean): string {
  if (!input.name.trim()) return '请填写站点名称';
  if (!input.code.trim()) return '请填写站点代码';
  if (!input.baseUrl.trim()) return '请填写 Base URL';
  if (!editing && !input.apiKey?.trim()) return '请填写 API Key';
  if (input.accountCredentialKind !== 'none' && !input.accountCredential?.trim()) return '请填写账号凭据或选择不配置';
  return '';
}
