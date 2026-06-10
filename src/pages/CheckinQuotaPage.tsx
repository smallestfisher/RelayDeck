import { CalendarCheck, Filter, Gift, RefreshCw, WalletCards } from 'lucide-react';
import { useMemo, useState } from 'react';
import { alerts, quotaDistribution, quotaRecords, quotaTrend, sites } from '../data/mock';
import { formatCurrency } from '../lib/format';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { DataTable, tableCellClass, tableHeadClass } from '../components/ui/DataTable';
import { MetricCard } from '../components/ui/MetricCard';
import { RingChart } from '../components/charts/RingChart';
import { LineChart } from '../components/charts/LineChart';
import { StatusBadge } from '../components/ui/StatusBadge';

export function CheckinQuotaPage() {
  const [checking, setChecking] = useState(false);
  const [completed, setCompleted] = useState(false);
  const checkedCount = quotaRecords.filter((record) => record.status === 'checked').length;
  const uncheckedCount = quotaRecords.filter((record) => record.status === 'unchecked').length;
  const totalQuota = quotaRecords.reduce((sum, record) => sum + record.currentUsd, 0);
  const todayReward = quotaRecords.reduce((sum, record) => sum + (record.rewardUsd ?? 0), 0);

  const progressData = useMemo(
    () => [
      { label: '已签到', value: checkedCount },
      { label: '未签到', value: uncheckedCount },
      { label: '已禁用', value: quotaRecords.filter((record) => record.status === 'disabled').length },
    ],
    [checkedCount, uncheckedCount]
  );

  function runCheckin() {
    setChecking(true);
    setCompleted(false);
    window.setTimeout(() => {
      setChecking(false);
      setCompleted(true);
    }, 1200);
  }

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-text">签到中心 / 额度管理</h1>
          <p className="mt-1 text-sm text-muted">管理站点签到与额度分配，监控额度使用情况</p>
        </div>
        <div className="flex gap-3">
          <Button icon={<RefreshCw className="h-4 w-4" />}>刷新额度</Button>
          <Button icon={<Filter className="h-4 w-4" />}>只看异常</Button>
          <Button variant="primary" icon={checking ? <RefreshCw className="h-4 w-4 animate-spin" /> : <CalendarCheck className="h-4 w-4" />} onClick={runCheckin}>
            {checking ? '签到中...' : completed ? '已完成' : '一键签到'}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="今日已签到" value={`${checkedCount + 23}`} detail="较昨日" delta="+3" tone="info" icon={<CalendarCheck className="h-6 w-6" />} />
        <MetricCard label="未签到站点" value={`${uncheckedCount + 6}`} detail="较昨日" delta="-2" tone="warning" icon={<CalendarCheck className="h-6 w-6" />} />
        <MetricCard label="剩余额度" value={formatCurrency(totalQuota)} detail="较昨日" delta="+115.2%" tone="success" icon={<WalletCards className="h-6 w-6" />} />
        <MetricCard label="今日新增额度" value={formatCurrency(todayReward + 203)} detail="较昨日" delta="+28.00" tone="success" icon={<Gift className="h-6 w-6" />} />
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-[0.95fr_1.6fr_1fr]">
        <Card title="签到进度">
          <RingChart data={progressData} centerLabel="已签到" centerValue={`${checkedCount + 23}`} />
          <div className="mt-4 text-sm text-muted">总站点数 36 个</div>
        </Card>
        <Card title="额度趋势（近 7 天)" action={<Button size="sm">近 7 天</Button>}>
          <LineChart data={quotaTrend} valuePrefix="$" />
        </Card>
        <Card title="任务日志" action={<button className="text-sm text-primary">查看全部</button>}>
          <div className="space-y-3">
            {alerts.map((alert) => (
              <div key={alert.id} className="flex items-start gap-3 text-sm">
                <span className={alert.severity === 'danger' ? 'mt-1 h-5 w-5 rounded-full bg-danger/15 text-center text-danger' : 'mt-1 h-5 w-5 rounded-full bg-success/15 text-center text-success'}>
                  {alert.severity === 'danger' ? '×' : '✓'}
                </span>
                <div className="min-w-0 flex-1">
                  <div className="truncate text-text">{alert.title.replace('连接超时', '签到成功')}</div>
                  <div className="truncate text-xs text-muted">{alert.description}</div>
                </div>
                <span className="text-xs text-muted">{alert.time}</span>
              </div>
            ))}
          </div>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-[1.4fr_0.8fr]">
        <Card title="站点签到记录">
          <DataTable>
            <thead className={tableHeadClass}>
              <tr>
                <th className={tableCellClass}>站点名称</th>
                <th className={tableCellClass}>签到模式</th>
                <th className={tableCellClass}>最后签到</th>
                <th className={tableCellClass}>奖励额度</th>
                <th className={tableCellClass}>当前额度</th>
                <th className={tableCellClass}>状态</th>
              </tr>
            </thead>
            <tbody>
              {quotaRecords.map((record) => {
                const site = sites.find((item) => item.id === record.siteId) ?? sites[0];
                return (
                  <tr key={record.id} className="hover:bg-elevated/55">
                    <td className={tableCellClass}>
                      <span className="mr-2">{site.flag}</span>
                      <span className="font-medium text-text">{site.name}</span>
                    </td>
                    <td className={tableCellClass}>{record.mode}</td>
                    <td className={tableCellClass}>{record.lastCheckin}</td>
                    <td className={tableCellClass}>{record.rewardUsd ? `+${formatCurrency(record.rewardUsd)}` : '-'}</td>
                    <td className={tableCellClass}>{formatCurrency(record.currentUsd)}</td>
                    <td className={tableCellClass}>
                      <StatusBadge status={record.status} />
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </DataTable>
        </Card>
        <Card title="站点额度使用分布">
          <RingChart data={quotaDistribution} centerLabel="总额度 USD" centerValue={totalQuota.toLocaleString('en-US', { maximumFractionDigits: 2 })} />
          <div className="mt-5 text-sm text-muted">共 36 个站点</div>
        </Card>
      </div>
    </div>
  );
}
