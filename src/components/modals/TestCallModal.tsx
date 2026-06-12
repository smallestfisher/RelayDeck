import { RotateCcw, Send, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { adminApi } from '../../lib/adminApi';
import { formatLatency } from '../../lib/format';
import type { UpstreamAccount, UpstreamModel, UpstreamTestCallResult } from '../../types';

interface TestCallModalProps {
  account: UpstreamAccount;
  onClose: () => void;
  onStatusUpdate?: (accountId: string, status: UpstreamTestCallResult['accountStatus']) => void;
}

const DEFAULT_MESSAGE = 'Hello, how are you?';
const inputClass =
  'w-full rounded-lg border border-line bg-elevated px-3 py-2 text-sm text-text outline-none focus:border-primary focus:ring-2 focus:ring-primary/20';

const protocolLabels: Record<string, string> = {
  'openai-chat': 'OpenAI Chat',
  'openai-responses': 'OpenAI Responses',
  'claude-messages': 'Claude Messages',
};

const wireProtocolToTestProtocol: Record<string, string> = {
  openai_chat_completions: 'openai-chat',
  openai_responses: 'openai-responses',
  anthropic_messages: 'claude-messages',
};

export function TestCallModal({ account, onClose, onStatusUpdate }: TestCallModalProps) {
  const [models, setModels] = useState<UpstreamModel[]>([]);
  const [selectedModel, setSelectedModel] = useState('');
  const [protocol, setProtocol] = useState('auto');
  const [streaming, setStreaming] = useState(false);
  const [message, setMessage] = useState(DEFAULT_MESSAGE);
  const [loading, setLoading] = useState(false);
  const [modelError, setModelError] = useState('');
  const [result, setResult] = useState<UpstreamTestCallResult | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function loadModels() {
      setModels([]);
      setSelectedModel('');
      setModelError('');
      try {
        const items = await adminApi.listUpstreamModels(account.id);
        if (cancelled) return;
        setModels(items);
        if (items.length > 0) {
          setSelectedModel(items[0].upstreamModelName);
        }
      } catch (error) {
        if (!cancelled) {
          setModelError(error instanceof Error ? error.message : '模型加载失败');
        }
      }
    }

    void loadModels();
    return () => {
      cancelled = true;
    };
  }, [account.id]);

  const selectedModelInfo = useMemo(
    () => models.find((item) => item.upstreamModelName === selectedModel || item.normalizedModelName === selectedModel),
    [models, selectedModel]
  );
  const protocolOptions = useMemo(() => supportedTestProtocols(selectedModelInfo), [selectedModelInfo]);
  const canSend = Boolean(selectedModel && protocolOptions.length > 0);
  const protocolHint = protocolOptions.length > 0 ? '自动会按模型支持协议选择第一个可测试协议' : '该模型没有当前支持的测试协议';

  function handleReset() {
    setProtocol('auto');
    setStreaming(false);
    setMessage(DEFAULT_MESSAGE);
    setResult(null);
    if (models.length > 0) {
      setSelectedModel(models[0].upstreamModelName);
    }
  }

  async function handleTest() {
    setLoading(true);
    setResult(null);
    try {
      const data = await adminApi.testUpstreamCall(account.id, { modelName: selectedModel, protocol, streaming, message });
      setResult(data);
      onStatusUpdate?.(account.id, data.accountStatus);
    } catch (error) {
      setResult({
        id: account.id,
        httpStatus: 0,
        protocol,
        ok: false,
        message: error instanceof Error ? error.message : '测试请求失败',
        latencyMs: 0,
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm" onClick={onClose} />
      <div className="fixed left-1/2 top-1/2 z-50 w-full max-w-2xl -translate-x-1/2 -translate-y-1/2 rounded-xl border border-line bg-panel shadow-2xl">
        <div className="flex items-start justify-between gap-4 border-b border-line px-6 py-5">
          <div>
            <h2 className="text-lg font-semibold text-text">测试 API 调用</h2>
            <p className="mt-1 text-sm text-muted">按模型支持协议发送一次真实上游请求，只展示 HTTP 检测摘要</p>
          </div>
          <button type="button" onClick={onClose} className="shrink-0 text-muted hover:text-text" aria-label="关闭">
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="space-y-4 px-6 py-5">
          <Field label="站点">
            <div className="rounded-lg border border-line bg-elevated px-3 py-2 text-sm text-muted">{account.name}</div>
          </Field>

          <Field label="模型" hint={modelError || '模型来自已同步的数据库记录；如果为空，请先执行全量刷新'} danger={Boolean(modelError)}>
            <select className={inputClass} value={selectedModel} onChange={(event) => setSelectedModel(event.target.value)}>
              {models.length === 0 && <option value="">暂无可用模型</option>}
              {models.map((model) => (
                <option key={model.id} value={model.upstreamModelName}>
                  {model.displayName || model.upstreamModelName}
                </option>
              ))}
            </select>
          </Field>

          <Field label="协议" hint={protocolHint} danger={protocolOptions.length === 0 && Boolean(selectedModel)}>
            <select className={inputClass} value={protocol} onChange={(event) => setProtocol(event.target.value)} disabled={protocolOptions.length === 0}>
              <option value="auto">自动</option>
              {protocolOptions.map((item) => (
                <option key={item} value={item}>
                  {protocolLabels[item]}
                </option>
              ))}
            </select>
          </Field>

          <Field label="流式响应">
            <button
              type="button"
              onClick={() => setStreaming(!streaming)}
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition ${streaming ? 'bg-primary' : 'bg-line'}`}
              role="switch"
              aria-checked={streaming}
            >
              <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition ${streaming ? 'translate-x-6' : 'translate-x-1'}`} />
            </button>
          </Field>

          <Field label="测试消息">
            <textarea className={inputClass} rows={3} placeholder={DEFAULT_MESSAGE} value={message} onChange={(event) => setMessage(event.target.value)} />
          </Field>

          {result && (
            <div className={`rounded-lg border p-4 ${result.ok ? 'border-success/30 bg-success/10' : 'border-danger/30 bg-danger/10'}`}>
              <div className={`text-sm font-medium ${result.ok ? 'text-success' : 'text-danger'}`}>{result.ok ? '测试成功' : '测试失败'}</div>
              <div className="mt-3 grid grid-cols-2 gap-2 text-sm text-muted">
                <span>HTTP：{result.httpStatus || '-'}</span>
                <span>协议：{protocolLabels[result.protocol] ?? result.protocol}</span>
                <span>延迟：{formatLatency(result.latencyMs || undefined)}</span>
                <span>状态：{result.ok ? '成功' : '失败'}</span>
              </div>
              {result.message && <div className="mt-3 break-words text-sm text-muted">{result.message}</div>}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between gap-3 border-t border-line px-6 py-4">
          <button type="button" onClick={handleReset} className="inline-flex items-center gap-2 rounded-lg border border-line px-4 py-2 text-sm font-medium text-text hover:bg-elevated">
            <RotateCcw className="h-4 w-4" />
            重置
          </button>
          <div className="flex items-center gap-3">
            <button type="button" onClick={onClose} className="rounded-lg border border-line px-4 py-2 text-sm font-medium text-text hover:bg-elevated">
              取消
            </button>
            <button
              type="button"
              onClick={handleTest}
              disabled={loading || !canSend}
              className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-50"
            >
              <Send className="h-4 w-4" />
              {loading ? '测试中...' : '发送测试'}
            </button>
          </div>
        </div>

        <div className="border-t border-line px-6 py-3 text-center text-xs text-muted">测试调用会真实请求上游，计费以平台规则为准</div>
      </div>
    </>
  );
}

function supportedTestProtocols(model?: UpstreamModel): string[] {
  if (!model) return [];
  const seen = new Set<string>();
  const result: string[] = [];
  for (const protocol of model.supportedWireProtocols) {
    const mapped = wireProtocolToTestProtocol[protocol];
    if (mapped && !seen.has(mapped)) {
      seen.add(mapped);
      result.push(mapped);
    }
  }
  return result;
}

function Field({ label, hint, danger, children }: { label: string; hint?: string; danger?: boolean; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-[80px_1fr] items-start gap-4">
      <label className="pt-2 text-sm font-medium text-text">{label}</label>
      <div>
        {children}
        {hint && <p className={`mt-1 text-xs ${danger ? 'text-danger' : 'text-muted'}`}>{hint}</p>}
      </div>
    </div>
  );
}
