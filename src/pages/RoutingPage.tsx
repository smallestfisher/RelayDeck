import { Clock, Flame, HeartPulse, Play, RefreshCw, Save, Users, WalletCards } from 'lucide-react';
import { useMemo, useState } from 'react';
import { models, routeHistory, routingCandidates, scoreBreakdown, sites } from '../data/mock';
import { formatLatency } from '../lib/format';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { DataTable, tableCellClass, tableHeadClass } from '../components/ui/DataTable';
import { RangeSlider, SelectControl } from '../components/ui/Controls';
import { StatusBadge } from '../components/ui/StatusBadge';

const tabs = ['全局路由规则', '模型路由规则', '路由日志', '路由分析'];

export function RoutingPage() {
  const [activeTab, setActiveTab] = useState(tabs[0]);
  const [selectedModel, setSelectedModel] = useState('gpt-4o-mini');
  const [weights, setWeights] = useState<Record<string, number>>(
    Object.fromEntries(routingCandidates.map((candidate) => [candidate.id, candidate.manualWeight]))
  );
  const selectedModelLabel = models.find((model) => model.id === selectedModel)?.name ?? 'GPT-4o-mini';
  const bestCandidate = routingCandidates[0];
  const bestSite = sites.find((site) => site.id === bestCandidate.siteId) ?? sites[0];

  const candidateRows = useMemo(
    () =>
      routingCandidates.map((candidate) => ({
        ...candidate,
        site: sites.find((site) => site.id === candidate.siteId) ?? sites[0],
      })),
    []
  );

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-text">智能路由</h1>
          <p className="mt-1 text-sm text-muted">基于多维指标的加权智能路由，自动选择最优站点，保障稳定与高效。</p>
        </div>
        <div className="flex gap-3">
          <Button icon={<Play className="h-4 w-4" />}>测试路由</Button>
          <Button variant="primary" icon={<Save className="h-4 w-4" />}>
            保存规则
          </Button>
        </div>
      </div>

      <div className="flex gap-8 border-b border-line">
        {tabs.map((tab) => (
          <button
            key={tab}
            type="button"
            onClick={() => setActiveTab(tab)}
            className={activeTab === tab ? 'border-b-2 border-primary pb-3 text-sm font-semibold text-primary' : 'pb-3 text-sm text-muted'}
          >
            {tab}
          </button>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-5 2xl:grid-cols-[1fr_380px]">
        <div className="space-y-5">
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-5">
            {[
              ['路由模式', '智能加权', 'Smart Weight', WalletCards],
              ['健康分阈值', '70 分', '低于此分不参与路由', HeartPulse],
              ['熔断阈值', '5 次失败 / 60 秒', '触发熔断', Flame],
              ['熔断冷却时间', '120 秒', '冷却后重新参与路由', Clock],
              ['最少候选站点数', '2 个', '低于此数将降级处理', Users],
            ].map(([label, value, detail, Icon]) => {
              const IconComponent = Icon as typeof WalletCards;
              return (
                <Card key={label as string} className="p-4">
                  <div className="flex items-center gap-3">
                    <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/12 text-primary">
                      <IconComponent className="h-5 w-5" />
                    </span>
                    <div>
                      <div className="text-xs text-muted">{label as string}</div>
                      <div className="mt-1 text-lg font-semibold text-text">{value as string}</div>
                      <div className="mt-1 text-xs text-muted">{detail as string}</div>
                    </div>
                  </div>
                </Card>
              );
            })}
          </div>

          <Card
            title="模型路由配置"
            action={
              <div className="flex items-center gap-3">
                <SelectControl
                  className="w-44"
                  value={selectedModel}
                  onChange={(event) => setSelectedModel(event.target.value)}
                  options={models.map((model) => ({ label: model.name, value: model.id }))}
                />
                <Button size="sm" icon={<RefreshCw className="h-4 w-4" />}>
                  刷新指标
                </Button>
              </div>
            }
          >
            <DataTable>
              <thead className={tableHeadClass}>
                <tr>
                  <th className={tableCellClass}>站点</th>
                  <th className={tableCellClass}>状态</th>
                  <th className={tableCellClass}>手动权重</th>
                  <th className={tableCellClass}>健康分</th>
                  <th className={tableCellClass}>24h 成功率</th>
                  <th className={tableCellClass}>延迟</th>
                  <th className={tableCellClass}>负载均衡</th>
                  <th className={tableCellClass}>熔断状态</th>
                </tr>
              </thead>
              <tbody>
                {candidateRows.map((candidate) => (
                  <tr key={candidate.id} className="hover:bg-elevated/55">
                    <td className={tableCellClass}>
                      <div className="flex items-center gap-3">
                        <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-sm font-semibold text-white">
                          {candidate.site.code}
                        </span>
                        <div>
                          <div className="font-medium text-text">API 站点 {candidate.site.code}</div>
                          <div className="text-xs text-muted">{candidate.site.name}</div>
                        </div>
                      </div>
                    </td>
                    <td className={tableCellClass}>
                      <StatusBadge status={candidate.site.status} />
                    </td>
                    <td className={tableCellClass}>
                      <div className="flex min-w-[130px] items-center gap-3">
                        <RangeSlider value={weights[candidate.id]} onChange={(value) => setWeights((current) => ({ ...current, [candidate.id]: value }))} />
                        <span className="w-8 rounded bg-elevated px-1.5 py-1 text-center text-xs">{weights[candidate.id]}</span>
                      </div>
                    </td>
                    <td className={tableCellClass}>
                      <span className={candidate.healthScore < 50 ? 'text-danger' : 'text-success'}>{candidate.healthScore}</span>
                    </td>
                    <td className={tableCellClass}>{candidate.successRate.toFixed(1)}%</td>
                    <td className={tableCellClass}>{formatLatency(candidate.latencyMs)}</td>
                    <td className={tableCellClass}>
                      <div className="flex items-center gap-2">
                        <span>{candidate.load}%</span>
                        <span className="h-2 w-14 rounded-full bg-line">
                          <span className="block h-full rounded-full bg-success" style={{ width: `${candidate.load}%` }} />
                        </span>
                      </div>
                    </td>
                    <td className={tableCellClass}>
                      <StatusBadge status={candidate.circuitState} label={candidate.circuitState === 'closed' ? '闭合' : '剩余 75s'} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </DataTable>
            <div className="mt-4 text-sm text-muted">
              综合得分 = 手动权重 x30% + 健康分 x30% + 延迟得分 x20% + 负载均衡 x15% - 惩罚项 x5%
            </div>
          </Card>
        </div>

        <Card title={`本次路由为什么选择 ${bestSite.name}`} subtitle={`模型：${selectedModelLabel}`}>
          <div className="flex items-center justify-between border-b border-line pb-5">
            <div>
              <div className="text-5xl font-semibold text-success">{bestCandidate.score.toFixed(1)}</div>
              <div className="mt-1 text-sm text-muted">综合得分（越高越优）</div>
            </div>
            <div className="flex h-24 w-24 items-center justify-center rounded-full border-[10px] border-success/75 bg-success/10 text-3xl">
              🏆
            </div>
          </div>
          <div className="mt-5 space-y-3">
            <h3 className="text-sm font-semibold text-text">得分构成</h3>
            {scoreBreakdown.map((item) => (
              <div key={item.label} className="grid grid-cols-[80px_1fr_58px] items-center gap-3 text-sm">
                <span className="text-text">{item.label}</span>
                <span className="h-2 overflow-hidden rounded-full bg-line">
                  <span
                    className={item.tone === 'danger' ? 'block h-full rounded-full bg-danger' : 'block h-full rounded-full bg-primary'}
                    style={{ width: `${Math.max(0, (item.value / item.max) * 100)}%` }}
                  />
                </span>
                <span className={item.tone === 'danger' ? 'text-danger' : 'text-text'}>
                  {item.value}/{item.max}
                </span>
              </div>
            ))}
          </div>
          <div className="mt-6 border-t border-line pt-5">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="text-sm font-semibold text-text">近期路由历史</h3>
              <button className="text-xs text-primary">更多</button>
            </div>
            <div className="space-y-3">
              {routeHistory.map((item) => (
                <div key={item.id} className="flex items-center justify-between gap-3 text-sm">
                  <span className="flex items-center gap-2">
                    <span className={item.result === 'success' ? 'h-5 w-5 rounded-full bg-success/15 text-center text-success' : 'h-5 w-5 rounded-full bg-danger/15 text-center text-danger'}>
                      {item.result === 'success' ? '✓' : '×'}
                    </span>
                    {item.siteName}
                  </span>
                  <span className="text-muted">{item.score ? `${item.score} 分` : '-'}</span>
                  <span className="text-muted">{item.time}</span>
                </div>
              ))}
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
