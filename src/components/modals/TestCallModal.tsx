import { RotateCcw, Send, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { adminApi } from '../../lib/adminApi';
import type { UpstreamAccount, UpstreamModel } from '../../types';

interface TestCallModalProps {
  account: UpstreamAccount;
  onClose: () => void;
}

const DEFAULT_MESSAGE = 'Hello, how are you?';
const inputClass =
  'w-full rounded-lg border border-line bg-elevated px-3 py-2 text-sm text-text outline-none focus:border-primary focus:ring-2 focus:ring-primary/20';

export function TestCallModal({ account, onClose }: TestCallModalProps) {
  const [models, setModels] = useState<UpstreamModel[]>([]);
  const [selectedModel, setSelectedModel] = useState('');
  const [protocol, setProtocol] = useState('openai-chat');
  const [streaming, setStreaming] = useState(false);
  const [message, setMessage] = useState(DEFAULT_MESSAGE);
  const [loading, setLoading] = useState(false);
  const [modelError, setModelError] = useState('');
  const [result, setResult] = useState<Record<string, unknown> | null>(null);

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

  function handleReset() {
    setProtocol('openai-chat');
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
      const resp = await fetch(`/api/admin/upstreams/accounts/${account.id}/test-call`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ model_name: selectedModel, protocol, streaming, message }),
      });
      const data = (await resp.json().catch(() => ({}))) as Record<string, unknown>;
      if (!resp.ok) {
        setResult({ error: data.error ?? resp.statusText });
        return;
      }
      setResult(data);
    } catch (error) {
      setResult({ error: error instanceof Error ? error.message : '测试请求失败' });
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
            <p className="mt-1 text-sm text-muted">配置请求参数并发送测试调用，验证站点的接口可用性与响应详情</p>
          </div>
          <button type="button" onClick={onClose} className="shrink-0 text-muted hover:text-text" aria-label="关闭">
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="space-y-4 px-6 py-5">
          <Field label="站点">
            <div className="rounded-lg border border-line bg-elevated px-3 py-2 text-sm text-muted">
              {account.name} ({account.code})
            </div>
          </Field>

          <Field label="模型" hint={modelError || '模型来自已同步的数据库记录；如果为空，请先执行全量刷新'} danger={Boolean(modelError)}>
            <select className={inputClass} value={selectedModel} onChange={(e) => setSelectedModel(e.target.value)}>
              {models.length === 0 && <option value="">暂无可用模型</option>}
              {models.map((m) => (
                <option key={m.id} value={m.upstreamModelName}>
                  {m.displayName || m.upstreamModelName}
                </option>
              ))}
            </select>
          </Field>

          <Field label="协议" hint="选择测试使用的请求协议">
            <select className={inputClass} value={protocol} onChange={(e) => setProtocol(e.target.value)}>
              <option value="openai-chat">OpenAI Chat</option>
              <option value="claude-messages">Claude Messages</option>
              <option value="openai-responses">OpenAI Responses</option>
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

          <Field label="测试消息" hint="输入要发送给模型的测试消息内容">
            <textarea className={inputClass} rows={3} placeholder={DEFAULT_MESSAGE} value={message} onChange={(e) => setMessage(e.target.value)} />
          </Field>

          {result && (
            <div className="rounded-lg border border-line bg-elevated">
              <div className="border-b border-line px-3 py-2 text-sm font-medium text-text">响应结果</div>
              <pre className="max-h-64 overflow-auto p-3 text-xs text-muted">{JSON.stringify(result, null, 2)}</pre>
            </div>
          )}
        </div>

        <div className="flex items-center justify-between gap-3 border-t border-line px-6 py-4">
          <button
            type="button"
            onClick={handleReset}
            className="inline-flex items-center gap-2 rounded-lg border border-line px-4 py-2 text-sm font-medium text-text hover:bg-elevated"
          >
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
              disabled={loading || !selectedModel}
              className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:opacity-50"
            >
              <Send className="h-4 w-4" />
              {loading ? '测试中...' : '发送测试'}
            </button>
          </div>
        </div>

        <div className="border-t border-line px-6 py-3 text-center text-xs text-muted">测试调用不写入 RelayDeck 统计；实际上游计费以平台规则为准</div>
      </div>
    </>
  );
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
