import { Save, TestTube2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Button } from '../../components/ui/Button';
import { SelectControl, ToggleSwitch } from '../../components/ui/Controls';
import { Drawer } from '../../components/ui/Drawer';
import type { UpstreamAccount, UpstreamAccountInput, UpstreamCredentialKind, UpstreamPlatformKind } from '../../types';
import { credentialKindOptions, platformOptions } from './siteOptions';

const emptyForm: UpstreamAccountInput = {
  name: '',
  code: '',
  platformKind: 'new_api',
  baseUrl: '',
  enabled: true,
  includeInRouting: true,
  priority: 50,
  apiKey: '',
  accountCredentialKind: 'none',
  accountCredential: '',
  autoSyncModels: true,
  autoRefreshQuota: false,
  autoCheckin: false,
  note: '',
};

interface SiteDrawerProps {
  open: boolean;
  account?: UpstreamAccount | null;
  saving: boolean;
  testing: boolean;
  error?: string;
  onClose: () => void;
  onSave: (input: UpstreamAccountInput) => void;
  onTestAPI: (input: UpstreamAccountInput) => void;
}

export function SiteDrawer({ open, account, saving, testing, error, onClose, onSave, onTestAPI }: SiteDrawerProps) {
  const [form, setForm] = useState<UpstreamAccountInput>(emptyForm);

  useEffect(() => {
    if (!open) return;
    if (!account) {
      setForm(emptyForm);
      return;
    }
    setForm({
      name: account.name,
      code: account.code,
      platformKind: account.platformKind,
      baseUrl: account.baseUrl,
      enabled: account.enabled,
      includeInRouting: account.includeInRouting,
      priority: account.priority,
      apiKey: '',
      accountCredentialKind: account.accountCredentialKind,
      accountCredential: '',
      autoSyncModels: account.autoSyncModels,
      autoRefreshQuota: account.autoRefreshQuota,
      autoCheckin: account.autoCheckin,
      note: account.note,
    });
  }, [account, open]);

  const isEditing = Boolean(account);
  const platformHint =
    form.platformKind === 'new_api'
      ? '账号凭据可用于用户资料、余额与签到；需要人工验证时会标记为需人工处理。'
      : '账号凭据可用于用户资料和平台额度；当前不假设支持签到。';

  function patch<K extends keyof UpstreamAccountInput>(key: K, value: UpstreamAccountInput[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  return (
    <Drawer
      open={open}
      title={isEditing ? '编辑站点' : '添加站点'}
      subtitle={isEditing ? '更新上游普通用户账号配置' : '接入 New API 或 Sub2API 的普通用户账号'}
      onClose={onClose}
      footer={
        <div className="flex justify-end gap-3">
          <Button onClick={onClose}>取消</Button>
          <Button variant="secondary" icon={<TestTube2 className="h-4 w-4" />} disabled={testing} onClick={() => onTestAPI(form)}>
            {testing ? '测试中' : '测试 API'}
          </Button>
          <Button variant="primary" icon={<Save className="h-4 w-4" />} disabled={saving} onClick={() => onSave(form)}>
            {saving ? '保存中' : '保存'}
          </Button>
        </div>
      }
    >
      <div className="space-y-5">
        {error && <div className="rounded-lg border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger">{error}</div>}

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label="站点名称 *">
            <input value={form.name} onChange={(event) => patch('name', event.target.value)} className={inputClass} placeholder="New API 主账号" />
          </Field>
          <Field label="站点代码 *">
            <input value={form.code} onChange={(event) => patch('code', event.target.value)} className={inputClass} placeholder="newapi-main" />
          </Field>
        </div>

        <Field label="平台 *">
          <SelectControl
            options={platformOptions}
            value={form.platformKind}
            onChange={(event) => patch('platformKind', event.target.value as UpstreamPlatformKind)}
          />
        </Field>

        <Field label="Base URL *">
          <input value={form.baseUrl} onChange={(event) => patch('baseUrl', event.target.value)} className={inputClass} placeholder="https://api.example.com" />
        </Field>

        <Field label={isEditing ? 'API Key（留空则不修改）' : 'API Key *'}>
          <input value={form.apiKey ?? ''} onChange={(event) => patch('apiKey', event.target.value)} className={inputClass} placeholder="sk-..." />
        </Field>

        <div className="rounded-lg border border-line bg-elevated/45 p-4">
          <div className="mb-3 text-sm font-medium text-text">账号凭据（可选）</div>
          <p className="mb-3 text-xs leading-5 text-muted">{platformHint}</p>
          <div className="space-y-3">
            <SelectControl
              options={credentialKindOptions}
              value={form.accountCredentialKind}
              onChange={(event) => patch('accountCredentialKind', event.target.value as UpstreamCredentialKind)}
            />
            {form.accountCredentialKind !== 'none' && (
              <textarea
                value={form.accountCredential ?? ''}
                onChange={(event) => patch('accountCredential', event.target.value)}
                className="min-h-24 w-full resize-none rounded-lg border border-line bg-panel px-3 py-2 text-sm outline-none focus:border-primary/55"
                placeholder="Cookie、Bearer token 或 JSON 凭据"
              />
            )}
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label="优先级">
            <input type="number" value={form.priority} onChange={(event) => patch('priority', Number(event.target.value))} className={inputClass} min={0} max={100} />
          </Field>
        </div>

        <div className="space-y-3 rounded-lg border border-line bg-elevated/45 p-4">
          <ToggleSwitch checked={form.enabled} onChange={(value) => patch('enabled', value)} label="启用站点" />
          <ToggleSwitch checked={form.includeInRouting} onChange={(value) => patch('includeInRouting', value)} label="纳入路由" />
          <ToggleSwitch checked={form.autoSyncModels} onChange={(value) => patch('autoSyncModels', value)} label="自动同步模型" />
          <ToggleSwitch checked={form.autoRefreshQuota} onChange={(value) => patch('autoRefreshQuota', value)} label="自动刷新额度" />
          <ToggleSwitch checked={form.autoCheckin} onChange={(value) => patch('autoCheckin', value)} label="自动签到（支持时）" />
        </div>

        <Field label="备注">
          <textarea
            value={form.note}
            onChange={(event) => patch('note', event.target.value)}
            className="min-h-24 w-full resize-none rounded-lg border border-line bg-elevated px-3 py-2 text-sm outline-none focus:border-primary/55"
            placeholder="用途、限制、维护说明"
          />
        </Field>
      </div>
    </Drawer>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-2 block text-sm font-medium text-text">{label}</span>
      {children}
    </label>
  );
}

const inputClass = 'h-10 w-full rounded-lg border border-line bg-elevated px-3 text-sm outline-none focus:border-primary/55';
