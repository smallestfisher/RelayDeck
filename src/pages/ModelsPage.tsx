import { CheckCircle2, Cuboid, ShieldAlert, ShieldCheck, XCircle } from 'lucide-react';
import { useMemo, useState } from 'react';
import { models, sites } from '../data/mock';
import type { Capability } from '../types';
import { formatLatency } from '../lib/format';
import { Card } from '../components/ui/Card';
import { DataTable, tableCellClass, tableHeadClass } from '../components/ui/DataTable';
import { MetricCard } from '../components/ui/MetricCard';
import { StatusBadge } from '../components/ui/StatusBadge';
import { MiniTrend } from '../components/charts/MiniTrend';
import { Button } from '../components/ui/Button';

export function ModelsPage() {
  const [selectedId, setSelectedId] = useState(models[1].id);
  const selected = models.find((model) => model.id === selectedId) ?? models[0];
  const recommendedSite = sites.find((site) => site.id === selected.recommendedSiteId) ?? sites[0];
  const rankedSites = useMemo(
    () => sites.slice(0, 3).map((site, index) => ({ site, score: [96, 93, 91][index] })),
    []
  );

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-text">模型管理</h1>
          <p className="mt-1 text-sm text-muted">查看哪些模型在哪些站点可用及其运行状况</p>
        </div>
        <Button>刷新数据</Button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="模型总数" value="6" detail="较昨日" delta="+1" tone="info" icon={<Cuboid className="h-6 w-6" />} />
        <MetricCard label="正常模型" value="4" detail="较昨日" delta="+1" tone="success" icon={<ShieldCheck className="h-6 w-6" />} />
        <MetricCard label="部分可用模型" value="1" detail="较昨日" delta="-" tone="warning" icon={<ShieldAlert className="h-6 w-6" />} />
        <MetricCard label="不可用模型" value="1" detail="较昨日" delta="-" tone="danger" icon={<XCircle className="h-6 w-6" />} />
      </div>

      <div className="grid grid-cols-1 gap-5 2xl:grid-cols-[1fr_320px]">
        <Card title="模型矩阵 / 库存">
          <DataTable>
            <thead className={tableHeadClass}>
              <tr>
                <th className={tableCellClass}>模型名称</th>
                <th className={tableCellClass}>可用站点数</th>
                <th className={tableCellClass}>推荐站点</th>
                <th className={tableCellClass}>最小延迟</th>
                <th className={tableCellClass}>成功率（7天）</th>
                <th className={tableCellClass}>额度情况</th>
                <th className={tableCellClass}>路由模式</th>
                <th className={tableCellClass}>状态</th>
              </tr>
            </thead>
            <tbody>
              {models.map((model) => {
                const site = sites.find((item) => item.id === model.recommendedSiteId) ?? sites[0];
                return (
                  <tr
                    key={model.id}
                    className={selectedId === model.id ? 'bg-primary/10' : 'hover:bg-elevated/55'}
                    onClick={() => setSelectedId(model.id)}
                  >
                    <td className={tableCellClass}>
                      <div className="flex items-center gap-3">
                        <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-success/15 text-sm font-semibold text-success">
                          {model.iconText}
                        </span>
                        <span className="font-medium text-text">{model.name}</span>
                      </div>
                    </td>
                    <td className={tableCellClass}>
                      {model.availableSites} / {model.totalSites}
                    </td>
                    <td className={tableCellClass}>
                      <span className="mr-2">{site.flag}</span>
                      {site.name}
                    </td>
                    <td className={tableCellClass}>{formatLatency(model.minLatencyMs)}</td>
                    <td className={tableCellClass}>
                      <div className="flex items-center gap-3">
                        <span className="text-success">{model.successRate.toFixed(1)}%</span>
                        <MiniTrend data={model.trend} />
                      </div>
                    </td>
                    <td className={tableCellClass}>
                      <div className="flex items-center gap-2">
                        <span>{model.quotaLabel}</span>
                        <span className="h-2 w-20 overflow-hidden rounded-full bg-line">
                          <span className="block h-full rounded-full bg-success" style={{ width: `${model.quotaUsage}%` }} />
                        </span>
                      </div>
                    </td>
                    <td className={tableCellClass}>
                      <span className="rounded-md border border-line bg-elevated px-2 py-1 text-xs text-muted">{model.routingMode}</span>
                    </td>
                    <td className={tableCellClass}>
                      <StatusBadge status={model.status} />
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </DataTable>
        </Card>

        <Card title="模型详情">
          <div className="flex items-center gap-3">
            <span className="flex h-12 w-12 items-center justify-center rounded-lg bg-success/15 text-sm font-semibold text-success">
              {selected.iconText}
            </span>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="font-semibold text-text">{selected.name}</h2>
                <StatusBadge status={selected.status} />
              </div>
              <div className="mt-1 flex gap-1">
                {selected.kind.map((item) => (
                  <span key={item} className="rounded-md bg-elevated px-2 py-1 text-xs text-muted">
                    {item}
                  </span>
                ))}
              </div>
            </div>
          </div>
          <div className="mt-5 space-y-2 border-t border-line pt-4 text-sm">
            <div className="flex justify-between text-muted">
              模型 ID <span className="font-medium text-text">{selected.id}</span>
            </div>
            <div className="flex justify-between text-muted">
              创建时间 <span className="font-medium text-text">{selected.createdAt}</span>
            </div>
            <div className="flex justify-between text-muted">
              推荐站点 <span className="font-medium text-text">{recommendedSite.name}</span>
            </div>
          </div>
          <div className="mt-5 border-t border-line pt-4">
            <h3 className="text-sm font-semibold text-text">推荐站点（按综合评分）</h3>
            <div className="mt-3 space-y-2">
              {rankedSites.map(({ site, score }, index) => (
                <div key={site.id} className="flex items-center justify-between rounded-lg bg-elevated/55 px-3 py-2 text-sm">
                  <span>
                    {index + 1}. {site.flag} {site.name}
                  </span>
                  <span className="font-semibold text-text">{score}</span>
                </div>
              ))}
            </div>
          </div>
          <div className="mt-5 border-t border-line pt-4">
            <h3 className="text-sm font-semibold text-text">支持能力</h3>
            <div className="mt-3 grid grid-cols-2 gap-2">
              {[
                ['stream', '流式输出'],
                ['function', '函数调用'],
                ['vision', '视觉理解'],
                ['embedding', '向量嵌入'],
              ].map(([key, label]) => (
                <div key={key} className="flex items-center gap-2 rounded-lg border border-line bg-elevated/45 px-3 py-2 text-sm">
                  <CheckCircle2 className={selected.capabilities.includes(key as Capability) ? 'h-4 w-4 text-success' : 'h-4 w-4 text-muted'} />
                  {label}
                </div>
              ))}
            </div>
          </div>
          <div className="mt-5 border-t border-line pt-4">
            <div className="mb-2 flex items-center justify-between text-sm">
              <span className="font-semibold text-text">成功率（7天）</span>
              <span className="text-success">{selected.successRate.toFixed(1)}%</span>
            </div>
            <MiniTrend data={selected.trend} className="h-12 w-full" />
          </div>
        </Card>
      </div>
    </div>
  );
}
