import { Activity, CheckCircle2, Clock3, Play, Save, WalletCards, XCircle } from 'lucide-react';
import { useState } from 'react';
import { capabilityCoverage, errorSummary, models, sites, testResults, testTemplates } from '../data/mock';
import { formatCurrency, formatLatency } from '../lib/format';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { DataTable, tableCellClass, tableHeadClass } from '../components/ui/DataTable';
import { MetricCard } from '../components/ui/MetricCard';
import { RingChart } from '../components/charts/RingChart';
import { SelectControl, ToggleSwitch } from '../components/ui/Controls';
import { StatusBadge } from '../components/ui/StatusBadge';

export function TestPage() {
  const [running, setRunning] = useState(false);
  const [stream, setStream] = useState(true);
  const [logs, setLogs] = useState(true);
  const [advanced, setAdvanced] = useState(false);

  const successCount = testResults.filter((result) => result.status === 'success').length;
  const failedCount = testResults.filter((result) => result.status !== 'success').length;
  const avgLatency = Math.round(testResults.reduce((sum, result) => sum + result.latencyMs, 0) / testResults.length);

  function startTest() {
    setRunning(true);
    window.setTimeout(() => setRunning(false), 1200);
  }

  return (
    <div className="grid grid-cols-1 gap-5 2xl:grid-cols-[1fr_360px]">
      <div className="space-y-5">
        <div>
          <h1 className="text-2xl font-semibold text-text">调用测试</h1>
          <p className="mt-1 text-sm text-muted">在不同站点上测试模型调用效果，验证可用性与性能表现</p>
        </div>

        <Card title="测试配置">
          <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1fr_1fr_1fr_300px]">
            <SelectControl options={models.map((model) => ({ label: model.name, value: model.id }))} />
            <SelectControl
              options={[
                { label: '全部站点', value: 'all' },
                ...sites.map((site) => ({ label: site.name, value: site.id })),
              ]}
            />
            <SelectControl
              options={[
                { label: '综合能力测试', value: 'general' },
                { label: '代码能力测试', value: 'code' },
                { label: '多模态测试', value: 'vision' },
              ]}
            />
            <div className="rounded-lg border border-line bg-elevated p-3">
              <div className="mb-3 flex items-center justify-between text-sm text-text">
                高级选项
                <span className="text-muted">⌃</span>
              </div>
              <div className="space-y-3">
                <ToggleSwitch checked={advanced} onChange={setAdvanced} label="启用温度、TopP、最大输出等高级参数" />
                <ToggleSwitch checked={stream} onChange={setStream} label="流式输出" />
                <ToggleSwitch checked={logs} onChange={setLogs} label="日志记录" />
              </div>
            </div>
          </div>
          <label className="mt-4 block">
            <span className="text-sm font-medium text-text">测试提示词</span>
            <textarea
              className="mt-2 min-h-28 w-full resize-none rounded-lg border border-line bg-elevated px-3 py-3 text-sm outline-none focus:border-primary/55"
              defaultValue="请对以下城市进行介绍：东京、巴黎、纽约、北京。要求包含地理位置、气候特点、著名景点、特色美食和文化特色。"
            />
          </label>
          <div className="mt-4 flex gap-3">
            <Button variant="primary" icon={<Play className={running ? 'h-4 w-4 animate-pulse' : 'h-4 w-4'} />} onClick={startTest}>
              {running ? '测试中...' : '开始测试'}
            </Button>
            <Button icon={<Activity className="h-4 w-4" />}>并发测试</Button>
            <Button icon={<Save className="h-4 w-4" />}>保存为模板</Button>
          </div>
        </Card>

        <Card title="测试结果概览">
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-6">
            <MetricCard label="总请求数" value={`${testResults.length * 6}`} detail="较昨日" delta="+12" tone="info" icon={<Activity className="h-5 w-5" />} />
            <MetricCard label="成功请求" value={`${successCount * 6 + 1}`} detail="成功率 86.11%" delta="" tone="success" icon={<CheckCircle2 className="h-5 w-5" />} />
            <MetricCard label="失败请求" value={`${failedCount * 2 + 1}`} detail="失败率 13.89%" delta="" tone="danger" icon={<XCircle className="h-5 w-5" />} />
            <MetricCard label="平均延迟" value={formatLatency(avgLatency)} detail="较昨日 -8.3%" delta="" tone="info" icon={<Clock3 className="h-5 w-5" />} />
            <MetricCard label="最慢响应" value="4.23s" detail="API 站点 D" delta="" tone="warning" icon={<Clock3 className="h-5 w-5" />} />
            <MetricCard label="总花费" value={formatCurrency(2.8451)} detail="较昨日 +15.2%" delta="" tone="warning" icon={<WalletCards className="h-5 w-5" />} />
          </div>
        </Card>

        <Card title="测试结果详情">
          <DataTable>
            <thead className={tableHeadClass}>
              <tr>
                <th className={tableCellClass}>站点</th>
                <th className={tableCellClass}>状态</th>
                <th className={tableCellClass}>延迟</th>
                <th className={tableCellClass}>返回 Token</th>
                <th className={tableCellClass}>错误信息</th>
                <th className={tableCellClass}>流式输出</th>
                <th className={tableCellClass}>函数调用</th>
                <th className={tableCellClass}>视觉理解</th>
                <th className={tableCellClass}>测试时间</th>
              </tr>
            </thead>
            <tbody>
              {testResults.map((result) => {
                const site = sites.find((item) => item.id === result.siteId) ?? sites[0];
                return (
                  <tr key={result.id} className="hover:bg-elevated/55">
                    <td className={tableCellClass}>
                      <span className="mr-2">{site.flag}</span>
                      {site.name}
                    </td>
                    <td className={tableCellClass}>
                      <StatusBadge status={result.status} />
                    </td>
                    <td className={tableCellClass}>{formatLatency(result.latencyMs)}</td>
                    <td className={tableCellClass}>{result.tokens}</td>
                    <td className={tableCellClass}>{result.error ?? '-'}</td>
                    <td className={tableCellClass}>{result.supports.stream ? '支持' : '不支持'}</td>
                    <td className={tableCellClass}>{result.supports.function ? '支持' : '不支持'}</td>
                    <td className={tableCellClass}>{result.supports.vision ? '支持' : '不支持'}</td>
                    <td className={tableCellClass}>{result.testedAt}</td>
                  </tr>
                );
              })}
            </tbody>
          </DataTable>
        </Card>
      </div>

      <div className="space-y-5 2xl:pt-[58px]">
        <Card title="最近测试模板" action={<button className="text-sm text-primary">查看全部</button>}>
          <div className="space-y-3">
            {testTemplates.map((template) => (
              <button key={template.id} type="button" className="w-full rounded-lg border border-line bg-elevated/55 p-3 text-left hover:border-primary/45">
                <div className="font-medium text-text">{template.name}</div>
                <div className="mt-1 text-xs text-muted">{template.model}</div>
                <div className="mt-2 text-xs text-muted">{template.time}</div>
              </button>
            ))}
          </div>
        </Card>
        <Card title="错误信息汇总" action={<button className="text-sm text-primary">查看详情</button>}>
          <RingChart data={errorSummary} centerLabel="总错误" centerValue="5" />
        </Card>
        <Card title="能力覆盖情况">
          <div className="space-y-4">
            {capabilityCoverage.map((item) => (
              <div key={item.label} className="grid grid-cols-[80px_1fr_46px] items-center gap-3 text-sm">
                <span className="text-text">{item.label}</span>
                <span className="h-2 overflow-hidden rounded-full bg-line">
                  <span className="block h-full rounded-full bg-primary" style={{ width: `${item.value}%` }} />
                </span>
                <span className="text-muted">{item.count}</span>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </div>
  );
}
