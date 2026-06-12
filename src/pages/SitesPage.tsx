import { Activity, AlertTriangle, ChevronLeft, ChevronRight, Plus, RefreshCw, ShieldCheck, Signal, Trash2, XCircle } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { SearchInput, SelectControl } from '../components/ui/Controls';
import { Drawer } from '../components/ui/Drawer';
import { MetricCard } from '../components/ui/MetricCard';
import { TestCallModal } from '../components/modals/TestCallModal';
import { adminApi } from '../lib/adminApi';
import { formatLatency, formatNumber } from '../lib/format';
import type {
  AccountCredentialStatus,
  UpstreamAccount,
  UpstreamAccountEvent,
  UpstreamAccountInput,
  UpstreamAccountMetrics,
  UpstreamActionResult,
  UpstreamActionName,
  UpstreamAPIStatus,
  UpstreamBatchActionName,
  UpstreamPlatformKind,
} from '../types';
import { SiteDrawer } from './sites/SiteDrawer';
import { SiteTable } from './sites/SiteTable';
import { accountStatusOptions, apiStatusOptions, platformOptions } from './sites/siteOptions';

type LatencyBand = 'all' | 'low' | 'medium' | 'high' | 'unknown';

const latencyOptions: Array<{ label: string; value: LatencyBand }> = [
  { label: '延迟：全部', value: 'all' },
  { label: '< 300ms', value: 'low' },
  { label: '300-1000ms', value: 'medium' },
  { label: '> 1000ms', value: 'high' },
  { label: '未知', value: 'unknown' },
];

const pageSizeOptions = [
  { label: '25 / 页', value: '25' },
  { label: '50 / 页', value: '50' },
  { label: '100 / 页', value: '100' },
];

const emptyMetrics: UpstreamAccountMetrics = { total: 0, healthy: 0, warning: 0, manual: 0 };

const SEARCH_DEBOUNCE_MS = 300;

export function SitesPage() {
  const [accounts, setAccounts] = useState<UpstreamAccount[]>([]);
  const [events, setEvents] = useState<UpstreamAccountEvent[]>([]);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [busyIds, setBusyIds] = useState<string[]>([]);
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [platform, setPlatform] = useState<UpstreamPlatformKind | 'all'>('all');
  const [apiStatus, setApiStatus] = useState<UpstreamAPIStatus | 'all'>('all');
  const [accountStatus, setAccountStatus] = useState<AccountCredentialStatus | 'all'>('all');
  const [latencyBand, setLatencyBand] = useState<LatencyBand>('all');
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editingAccount, setEditingAccount] = useState<UpstreamAccount | null>(null);
  const [inspectAccount, setInspectAccount] = useState<UpstreamAccount | null>(null);
  const [testCallAccount, setTestCallAccount] = useState<UpstreamAccount | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<UpstreamAccount | null>(null);
  const [batchDeleteConfirm, setBatchDeleteConfirm] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState('');
  const [batchNotice, setBatchNotice] = useState('');
  const [drawerError, setDrawerError] = useState('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [total, setTotal] = useState(0);
  const [metrics, setMetrics] = useState<UpstreamAccountMetrics>(emptyMetrics);
  const requestSeq = useRef(0);

  // Debounce the search input so each keystroke doesn't hit the backend.
  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query.trim()), SEARCH_DEBOUNCE_MS);
    return () => window.clearTimeout(timer);
  }, [query]);

  // Reset to first page and drop the (now possibly invalid) selection
  // whenever any filter dimension changes.
  useEffect(() => {
    setPage(1);
    setSelectedIds([]);
  }, [accountStatus, apiStatus, latencyBand, platform, debouncedQuery]);

  // Drop selection when the visible page changes — selected ids on a hidden
  // page would otherwise be acted on by batch operations the user can't see.
  useEffect(() => {
    setSelectedIds([]);
  }, [page, pageSize]);

  useEffect(() => {
    void loadAccounts();
    // loadAccounts reads the latest filter state via closure; depend on the
    // primitive filter values so we refetch on any change.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, debouncedQuery, platform, apiStatus, accountStatus, latencyBand]);

  async function loadAccounts() {
    const seq = ++requestSeq.current;
    setLoading(true);
    setError('');
    setBatchNotice('');
    try {
      const offset = (page - 1) * pageSize;
      const payload = await adminApi.listUpstreamAccounts({
        limit: pageSize,
        offset,
        q: debouncedQuery || undefined,
        platformKind: platform === 'all' ? undefined : platform,
        apiStatus: apiStatus === 'all' ? undefined : apiStatus,
        accountStatus: accountStatus === 'all' ? undefined : accountStatus,
        latencyBand: latencyBand === 'all' ? undefined : latencyBand,
      });
      // A newer request started while we were waiting — drop this stale response.
      if (seq !== requestSeq.current) return;
      // Result page is empty but the filter still has matches: caller is past
      // the last page, walk back.
      if (payload.items.length === 0 && payload.total > 0 && offset >= payload.total && page > 1) {
        setTotal(payload.total);
        setMetrics(payload.metrics ?? { ...emptyMetrics, total: payload.total });
        setPage(Math.ceil(payload.total / pageSize));
        return;
      }
      setAccounts(payload.items);
      setTotal(payload.total);
      setMetrics(payload.metrics ?? { ...emptyMetrics, total: payload.total });
    } catch (err) {
      if (seq !== requestSeq.current) return;
      setError(err instanceof Error ? err.message : '加载站点失败');
    } finally {
      if (seq === requestSeq.current) setLoading(false);
    }
  }

  async function saveAccount(input: UpstreamAccountInput) {
    const validation = validateAccountInput(input, editingAccount);
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

  async function runAction(account: UpstreamAccount, action: UpstreamActionName) {
    setBusyIds((current) => [...new Set([...current, account.id])]);
    setError('');
    setBatchNotice('');
    try {
      const result = await adminApi.runUpstreamAction(account.id, action);
      applyActionResults([result]);
      if (result.status !== 'success') {
        setError(result.message || '操作失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '操作失败');
    } finally {
      setBusyIds((current) => current.filter((id) => id !== account.id));
    }
  }

  async function runBatch(action: UpstreamBatchActionName) {
    if (selectedIds.length === 0) return;
    const targetIds = selectedIds;
    setBusyIds((current) => [...new Set([...current, ...targetIds])]);
    setError('');
    setBatchNotice('');
    try {
      const results = await adminApi.runBatchUpstreamAction(targetIds, action);
      applyActionResults(results);
      setBatchNotice(batchSummary(results));
    } catch (err) {
      setError(err instanceof Error ? err.message : '批量操作失败');
    } finally {
      setBusyIds((current) => current.filter((id) => !targetIds.includes(id)));
    }
  }

  function deleteAccount(account: UpstreamAccount) {
    setDeleteTarget(account);
  }

  async function confirmDeleteAccount() {
    const target = deleteTarget;
    if (!target) return;
    setDeleting(true);
    setBusyIds((current) => [...current, target.id]);
    try {
      await adminApi.deleteUpstreamAccount(target.id);
      setSelectedIds((current) => current.filter((id) => id !== target.id));
      setDeleteTarget(null);
      await loadAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除失败');
    } finally {
      setBusyIds((current) => current.filter((id) => id !== target.id));
      setDeleting(false);
    }
  }

  async function confirmBatchDelete() {
    if (selectedIds.length === 0) return;
    const targetIds = selectedIds;
    setDeleting(true);
    setBusyIds((current) => [...new Set([...current, ...targetIds])]);
    setError('');
    setBatchNotice('');
    try {
      const results = await adminApi.batchDeleteUpstreamAccounts(targetIds);
      const deletedIds = new Set(
        results.filter((item) => item.status === 'success' || item.status === 'not_found').map((item) => item.id)
      );
      setSelectedIds((current) => current.filter((id) => !deletedIds.has(id)));
      const failed = results.filter((item) => item.status === 'failed');
      if (failed.length > 0) {
        setError(`批量删除完成：成功 ${deletedIds.size}，失败 ${failed.length}`);
      } else {
        setBatchNotice(`批量删除完成：成功 ${deletedIds.size}`);
        setBatchDeleteConfirm(false);
      }
      await loadAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : '批量删除失败');
    } finally {
      setBusyIds((current) => current.filter((id) => !targetIds.includes(id)));
      setDeleting(false);
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

  // toggleAll only operates on the currently visible page. This is intentional:
  // selection no longer leaks across pages or filter changes (effects above
  // clear it), so "select all" should match what the user can see.
  function toggleAll() {
    const visibleIds = accounts.map((account) => account.id);
    const allSelected = visibleIds.length > 0 && visibleIds.every((id) => selectedIds.includes(id));
    setSelectedIds(allSelected ? selectedIds.filter((id) => !visibleIds.includes(id)) : [...new Set([...selectedIds, ...visibleIds])]);
  }

  function applyActionResults(results: UpstreamActionResult[]) {
    setAccounts((current) =>
      current.map((account) => {
        const result = results.find((item) => item.id === account.id);
        return result?.accountStatus ? { ...account, status: result.accountStatus } : account;
      })
    );
  }

  function applyStatusUpdate(accountId: string, status?: UpstreamAccount['status']) {
    setAccounts((current) => applyStatusUpdateToAccounts(current, accountId, status));
    setTestCallAccount((current) => (current && current.id === accountId && status ? { ...current, status: { ...current.status, ...status } } : current));
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

  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const pageStart = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const pageEnd = Math.min(total, page * pageSize);
  const selectedActionDisabled = selectedIds.length === 0 || busyIds.length > 0;

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
        <MetricCard label="站点总数" value={formatNumber(metrics.total)} detail={metricsDetail(metrics.total, debouncedQuery, platform, apiStatus, accountStatus, latencyBand)} delta="" tone="info" icon={<Signal className="h-6 w-6" />} />
        <MetricCard label="健康 API" value={formatNumber(metrics.healthy)} detail="可用于模型调用" delta="" tone="success" icon={<ShieldCheck className="h-6 w-6" />} />
        <MetricCard label="异常警告" value={formatNumber(metrics.warning)} detail="需复查状态" delta="" tone="warning" icon={<AlertTriangle className="h-6 w-6" />} />
        <MetricCard label="待处理" value={formatNumber(metrics.manual)} detail="需人工处理" delta="" tone="danger" icon={<XCircle className="h-6 w-6" />} />
      </div>

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <SearchInput className="w-[270px]" placeholder="搜索站点、URL 或备注" value={query} onChange={(event) => setQuery(event.target.value)} />
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
          <SelectControl className="w-36" options={latencyOptions} value={latencyBand} onChange={(event) => setLatencyBand(event.target.value as LatencyBand)} />
          <div className="ml-auto flex flex-wrap justify-end gap-2">
            <Button variant="secondary" icon={<RefreshCw className="h-4 w-4" />} onClick={() => void loadAccounts()} disabled={loading}>
              刷新
            </Button>
            <Button variant="secondary" icon={<RefreshCw className="h-4 w-4" />} onClick={() => runBatch('refresh-all')} disabled={selectedActionDisabled}>
              批量全量刷新
            </Button>
            <Button variant="danger" icon={<Trash2 className="h-4 w-4" />} onClick={() => setBatchDeleteConfirm(true)} disabled={selectedActionDisabled}>
              批量删除
            </Button>
          </div>
        </div>

        {error && <div className="mb-4 rounded-lg border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger">{error}</div>}
        {batchNotice && <div className="mb-4 rounded-lg border border-success/30 bg-success/10 px-3 py-2 text-sm text-success">{batchNotice}</div>}

        {loading ? (
          <div className="rounded-lg border border-line bg-elevated/45 py-12 text-center text-sm text-muted">加载中...</div>
        ) : accounts.length === 0 ? (
          <div className="rounded-lg border border-line bg-elevated/45 py-12 text-center text-sm text-muted">暂无站点账号</div>
        ) : (
          <SiteTable
            accounts={accounts}
            selectedIds={selectedIds}
            busyIds={busyIds}
            onToggleSelected={toggleSelected}
            onToggleAll={toggleAll}
            onEdit={openEditDrawer}
            onDelete={deleteAccount}
            onAction={runAction}
            onInspect={inspect}
            onTestCall={setTestCallAccount}
          />
        )}

        <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm text-muted">
          <span>
            共 {formatNumber(total)} 条记录，当前 {formatNumber(pageStart)}-{formatNumber(pageEnd)}，已选 {selectedIds.length} 条
          </span>
          <div className="flex flex-wrap items-center justify-end gap-2">
            <SelectControl
              className="w-28"
              options={pageSizeOptions}
              value={String(pageSize)}
              onChange={(event) => {
                setPageSize(Number(event.target.value));
                setPage(1);
              }}
            />
            <Button variant="icon" icon={<ChevronLeft className="h-4 w-4" />} disabled={page <= 1 || loading} onClick={() => setPage((current) => Math.max(1, current - 1))} aria-label="上一页" />
            <span className="min-w-20 text-center text-sm text-muted">
              {page} / {totalPages}
            </span>
            <Button variant="icon" icon={<ChevronRight className="h-4 w-4" />} disabled={page >= totalPages || loading} onClick={() => setPage((current) => Math.min(totalPages, current + 1))} aria-label="下一页" />
          </div>
        </div>
      </Card>

      <SiteDrawer
        open={drawerOpen}
        variant="modal"
        account={editingAccount}
        saving={saving}
        error={drawerError}
        onClose={() => setDrawerOpen(false)}
        onSave={saveAccount}
      />

      <Drawer
        open={Boolean(deleteTarget)}
        variant="modal"
        title="删除站点"
        subtitle={deleteTarget?.name}
        onClose={() => {
          if (!deleting) setDeleteTarget(null);
        }}
        footer={
          <div className="flex justify-end gap-3">
            <Button onClick={() => setDeleteTarget(null)} disabled={deleting}>
              取消
            </Button>
            <Button variant="danger" onClick={() => void confirmDeleteAccount()} disabled={deleting}>
              {deleting ? '删除中' : '确认删除'}
            </Button>
          </div>
        }
      >
        <div className="rounded-lg border border-danger/25 bg-danger/10 p-4 text-sm leading-6 text-danger">
          删除后会移除该站点的状态、模型与操作历史。此操作不可撤销。
        </div>
        {deleteTarget && (
          <div className="mt-4 rounded-lg border border-line bg-elevated/45 p-4 text-sm text-muted">
            <div className="font-medium text-text">{deleteTarget.name}</div>
            <div className="mt-1 break-all">{deleteTarget.baseUrl}</div>
          </div>
        )}
      </Drawer>

      <Drawer
        open={batchDeleteConfirm}
        variant="modal"
        title="批量删除站点"
        subtitle={`已选择 ${selectedIds.length} 个站点`}
        onClose={() => {
          if (!deleting) setBatchDeleteConfirm(false);
        }}
        footer={
          <div className="flex justify-end gap-3">
            <Button onClick={() => setBatchDeleteConfirm(false)} disabled={deleting}>
              取消
            </Button>
            <Button variant="danger" onClick={() => void confirmBatchDelete()} disabled={deleting}>
              {deleting ? '删除中' : '确认删除'}
            </Button>
          </div>
        }
      >
        <div className="rounded-lg border border-danger/25 bg-danger/10 p-4 text-sm leading-6 text-danger">
          删除后会移除这些站点的状态、模型与操作历史。此操作不可撤销。
        </div>
        <div className="mt-4 rounded-lg border border-line bg-elevated/45 p-4 text-sm text-muted">
          将删除 <span className="font-medium text-text">{selectedIds.length}</span> 个站点
        </div>
      </Drawer>

      <Drawer open={Boolean(inspectAccount)} variant="modal" title="站点历史" subtitle={inspectAccount?.name} onClose={() => setInspectAccount(null)}>
        <div className="space-y-3">
          {inspectAccount && (
            <div className="rounded-lg border border-line bg-elevated/45 p-4 text-sm text-muted">
              <div className="font-medium text-text">{inspectAccount.baseUrl}</div>
              <div className="mt-2 grid grid-cols-2 gap-2">
                <span>API：{apiStatusText(inspectAccount.status.apiStatus)}</span>
                <span>站点/API：{formatDetailLatencyPair(inspectAccount.status)}</span>
                <span>模型：{formatDetailModelCount(inspectAccount.status.modelCount, inspectAccount.status.lastModelSyncedAt)}</span>
                <span>账号：{accountStatusText(inspectAccount.status.accountStatus)}</span>
              </div>
            </div>
          )}
          {events.length === 0 ? (
            <div className="rounded-lg border border-line bg-elevated/45 py-10 text-center text-sm text-muted">暂无操作历史</div>
          ) : (
            events.map((event) => (
              <div key={event.id} className="rounded-lg border border-line bg-elevated/45 p-4">
                <div className="flex items-center justify-between gap-3">
                  <span className="text-sm font-medium text-text">{operationText(event.operation)}</span>
                  <span className="text-xs text-muted">{formatEventTime(event.createdAt)}</span>
                </div>
                <div className="mt-2 text-sm text-muted">{event.message || eventStatusText(event.status)}</div>
              </div>
            ))
          )}
        </div>
      </Drawer>

      {testCallAccount && <TestCallModal account={testCallAccount} onClose={() => setTestCallAccount(null)} onStatusUpdate={applyStatusUpdate} />}
    </div>
  );
}

function metricsDetail(
  total: number,
  query: string,
  platform: UpstreamPlatformKind | 'all',
  apiStatus: UpstreamAPIStatus | 'all',
  accountStatus: AccountCredentialStatus | 'all',
  latencyBand: LatencyBand,
): string {
  const hasFilter =
    Boolean(query) || platform !== 'all' || apiStatus !== 'all' || accountStatus !== 'all' || latencyBand !== 'all';
  if (hasFilter) return `匹配 ${formatNumber(total)} 个站点`;
  return '已接入账号';
}

function formatDetailLatencyPair(status: UpstreamAccount['status']): string {
  const site = status.lastModelSyncedAt ? formatLatency(status.latencyMs).replace(' ', '') : '-';
  const api = status.lastApiCheckedAt ? formatLatency(status.apiLatencyMs).replace(' ', '') : '-';
  return `${site}/${api}`;
}

function formatDetailModelCount(modelCount: number, lastModelSyncedAt?: string): string {
  return lastModelSyncedAt ? formatNumber(modelCount) : '-';
}

function applyStatusUpdateToAccounts(accounts: UpstreamAccount[], accountId: string, status?: UpstreamAccount['status']): UpstreamAccount[] {
  if (!status) return accounts;
  return accounts.map((account) => (account.id === accountId ? { ...account, status: { ...account.status, ...status } } : account));
}

function operationText(operation: string): string {
  const map: Record<string, string> = {
    test_api: '测试 API',
    test_account: '检测账号凭据',
    sync_models: '同步模型',
    refresh_quota: '刷新额度',
    checkin: '签到',
    refresh_all: '全量刷新',
  };
  return map[operation] ?? operation;
}

function eventStatusText(status: string): string {
  const map: Record<string, string> = {
    success: '成功',
    failed: '失败',
    not_found: '未找到',
  };
  return map[status] ?? status;
}

function apiStatusText(status: UpstreamAPIStatus): string {
  const map: Record<UpstreamAPIStatus, string> = {
    unknown: '未知',
    healthy: '健康',
    warning: '警告',
    failed: '失败',
    disabled: '禁用',
  };
  return map[status] ?? status;
}

function accountStatusText(status: AccountCredentialStatus): string {
  const map: Record<AccountCredentialStatus, string> = {
    not_configured: '未配置',
    valid: '有效',
    expired: '过期',
    failed: '失败',
    action_required: '需人工处理',
  };
  return map[status] ?? status;
}

function formatEventTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value || '-';
  }
  const pad = (part: number) => String(part).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

function batchSummary(results: UpstreamActionResult[]): string {
  const success = results.filter((item) => item.status === 'success').length;
  const failed = results.length - success;
  const actionRequired = results.filter(
    (item) => item.accountStatus?.accountStatus === 'action_required'
  ).length;
  const expired = results.filter((item) => item.accountStatus?.accountStatus === 'expired').length;
  const details = [`成功 ${success}`, `失败 ${failed}`];
  if (actionRequired > 0) details.push(`需人工处理 ${actionRequired}`);
  if (expired > 0) details.push(`凭据过期 ${expired}`);
  return `批量操作完成：${details.join('，')}`;
}

function validateAccountInput(input: UpstreamAccountInput, editingAccount: UpstreamAccount | null): string {
  if (!input.name.trim()) return '请填写站点名称';
  if (!input.baseUrl.trim()) return '请填写 Base URL';
  if (!editingAccount && !input.apiKey?.trim()) return '请填写 API Key';
  const credential = input.accountCredential?.trim() ?? '';
  const credentialKindChanged = Boolean(editingAccount && input.accountCredentialKind !== editingAccount.accountCredentialKind);
  const mustProvideCredential = !editingAccount || credentialKindChanged || credential !== '';
  if (input.accountCredentialKind === 'none') return '';
  if (!credential && mustProvideCredential) return '请填写账号凭据或选择不配置';
  if (input.accountCredentialKind === 'new_api_access_token' && mustProvideCredential) {
    const parsed = parseJSONCredential(credential);
    if (!stringField(parsed, 'access_token') || !stringField(parsed, 'user_id')) {
      return '请填写 New API Access Token 和 User ID';
    }
  }
  if (input.accountCredentialKind === 'sub2api_refresh_token' && mustProvideCredential) {
    const parsed = parseJSONCredential(credential);
    if (!stringField(parsed, 'refresh_token').startsWith('rt_')) {
      return '请填写 Sub2API Refresh Token';
    }
  }
  return '';
}

function parseJSONCredential(value: string): Record<string, unknown> {
  try {
    return JSON.parse(value) as Record<string, unknown>;
  } catch {
    return {};
  }
}

function stringField(payload: Record<string, unknown>, key: string): string {
  const value = payload[key];
  if (typeof value === 'string') return value.trim();
  if (typeof value === 'number') return String(value);
  return '';
}
