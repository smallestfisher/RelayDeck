import { Activity, CalendarCheck, RefreshCw, Server, WalletCards, Boxes } from 'lucide-react';
import { alerts, callTrend, models, overviewMetrics, sites } from '../data/mock';
import { formatCurrency, formatLatency, formatNumber } from '../lib/format';
import { Card } from '../components/ui/Card';
import { MetricCard } from '../components/ui/MetricCard';
import { StatusBadge } from '../components/ui/StatusBadge';
import { DataTable, tableCellClass, tableHeadClass } from '../components/ui/DataTable';
import { RingChart } from '../components/charts/RingChart';
import { LineChart } from '../components/charts/LineChart';
import { MiniTrend } from '../components/charts/MiniTrend';
import { Button } from '../components/ui/Button';

const metricIcons = [Server, Boxes, Activity, WalletCards, CalendarCheck];

export function OverviewPage() {
  const distribution = [
    { label: '正常运行', value: sites.filter((site) => site.status === 'normal').length },
    { label: '部分异常', value: sites.filter((site) => site.status === 'warning').length },
    { label: '连接失败', value: sites.filter((site) => site.status === 'failed').length },
    { label: '维护中', value: sites.filter((site) => site.status === 'maintenance').length },
  ];

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-text">概览</h1>
          <p className="mt-1 text-sm text-muted">全局运行状态与关键指标总览</p>
        </div>
        <div className="flex items-center gap-3 text-sm text-muted">
          <RefreshCw className="h-4 w-4" />
          最后更新：1 分钟前
          <Button variant="secondary" icon={<RefreshCw className="h-4 w-4" />}>
            刷新
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-5">
        {overviewMetrics.map((metric, index) => {
          const Icon = metricIcons[index];
          return (
            <MetricCard
              key={metric.label}
              label={metric.label}
              value={metric.value}
              detail="较昨日"
              delta={metric.delta}
              tone={metric.tone}
              icon={<Icon className="h-6 w-6" />}
            />
          );
        })}
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1.1fr_1.4fr_1fr]">
        <Card title="站点状态分布">
          <RingChart data={distribution} centerLabel="总站点" centerValue="32" />
        </Card>
        <Card title="调用量趋势（近 7 天)" action={<Button size="sm">近7天</Button>}>
          <LineChart data={callTrend} />
        </Card>
        <Card title="异常提醒" action={<button className="text-sm text-primary">查看全部</button>}>
          <div className="space-y-3">
            {alerts.map((alert) => (
              <div key={alert.id} className="flex items-start gap-3 rounded-lg bg-elevated/60 p-3">
                <span className="mt-1 h-2.5 w-2.5 rounded-full bg-danger" />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-medium text-text">{alert.title}</div>
                  <div className="mt-1 truncate text-xs text-muted">{alert.description}</div>
                </div>
                <span className="shrink-0 text-xs text-muted">{alert.time}</span>
              </div>
            ))}
          </div>
        </Card>
      </div>

      <Card title="站点状态概览">
        <DataTable>
          <thead className={tableHeadClass}>
            <tr>
              <th className={tableCellClass}>站点名称</th>
              <th className={tableCellClass}>状态</th>
              <th className={tableCellClass}>可用模型</th>
              <th className={tableCellClass}>延迟（平均）</th>
              <th className={tableCellClass}>余额（USD）</th>
              <th className={tableCellClass}>签到状态</th>
              <th className={tableCellClass}>最后检测时间</th>
            </tr>
          </thead>
          <tbody>
            {sites.slice(0, 5).map((site) => (
              <tr key={site.id} className="hover:bg-elevated/55">
                <td className={tableCellClass}>
                  <span className="mr-2">{site.flag}</span>
                  <span className="font-medium text-text">{site.name}</span>
                </td>
                <td className={tableCellClass}>
                  <StatusBadge status={site.status} />
                </td>
                <td className={tableCellClass}>
                  <div className="flex max-w-[320px] flex-wrap gap-1.5">
                    {site.models.slice(0, 3).map((model) => (
                      <span key={model} className="rounded-md border border-line bg-elevated px-2 py-1 text-xs text-muted">
                        {model}
                      </span>
                    ))}
                  </div>
                </td>
                <td className={tableCellClass}>{formatLatency(site.latencyMs)}</td>
                <td className={tableCellClass}>{formatCurrency(site.balanceUsd)}</td>
                <td className={tableCellClass}>
                  <StatusBadge status={site.checkinStatus} />
                </td>
                <td className={tableCellClass}>{site.lastChecked}</td>
              </tr>
            ))}
          </tbody>
        </DataTable>
      </Card>

      <Card title="模型可用性概览">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-5">
          {models.slice(0, 5).map((model) => (
            <div key={model.id} className="rounded-lg border border-line bg-elevated/50 p-4">
              <div className="flex items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-success/15 text-sm font-semibold text-success">
                    {model.iconText}
                  </span>
                  <div className="font-semibold text-text">{model.name}</div>
                </div>
                <span className="font-semibold text-success">{model.successRate.toFixed(1)}%</span>
              </div>
              <div className="mt-4 grid grid-cols-2 gap-3 text-xs text-muted">
                <span>
                  可用站点
                  <b className="mt-1 block text-sm text-text">
                    {model.availableSites} / {model.totalSites}
                  </b>
                </span>
                <span>
                  今日调用
                  <b className="mt-1 block text-sm text-text">{formatNumber(model.todayCalls)}</b>
                </span>
              </div>
              <MiniTrend data={model.trend} className="mt-3 h-9 w-full" />
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
