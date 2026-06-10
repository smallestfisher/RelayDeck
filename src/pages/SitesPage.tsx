import { BarChart3, Edit, MoreHorizontal, Plus, RefreshCw, ShieldCheck, Signal, TestTube2, XCircle } from 'lucide-react';
import { useMemo, useState } from 'react';
import { sites } from '../data/mock';
import type { SiteStatus, SiteType } from '../types';
import { formatCurrency, formatLatency } from '../lib/format';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { DataTable, tableCellClass, tableHeadClass } from '../components/ui/DataTable';
import { Drawer } from '../components/ui/Drawer';
import { MetricCard } from '../components/ui/MetricCard';
import { SearchInput, SelectControl } from '../components/ui/Controls';
import { StatusBadge } from '../components/ui/StatusBadge';

const statusOptions = [
  { label: '状态：全部', value: 'all' },
  { label: '正常', value: 'normal' },
  { label: '部分异常', value: 'warning' },
  { label: '连接失败', value: 'failed' },
  { label: '维护中', value: 'maintenance' },
];

const typeOptions = [
  { label: '类型：全部', value: 'all' },
  { label: 'OpenAI 兼容', value: 'OpenAI 兼容' },
  { label: 'Claude', value: 'Claude' },
  { label: 'Gemini', value: 'Gemini' },
  { label: 'Azure OpenAI', value: 'Azure OpenAI' },
];

export function SitesPage() {
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState<string>('all');
  const [type, setType] = useState<string>('all');
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [testing, setTesting] = useState(false);

  const filteredSites = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return sites.filter((site) => {
      const matchesQuery =
        !keyword ||
        site.name.toLowerCase().includes(keyword) ||
        site.url.toLowerCase().includes(keyword) ||
        site.note.toLowerCase().includes(keyword);
      const matchesStatus = status === 'all' || site.status === (status as SiteStatus);
      const matchesType = type === 'all' || site.type === (type as SiteType);
      return matchesQuery && matchesStatus && matchesType;
    });
  }, [query, status, type]);

  function testConnection() {
    setTesting(true);
    window.setTimeout(() => setTesting(false), 1000);
  }

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-text">站点管理</h1>
          <p className="mt-1 text-sm text-muted">管理和维护所有已连接的公益站点，确保路由服务稳定可用</p>
        </div>
        <Button variant="primary" icon={<Plus className="h-4 w-4" />} onClick={() => setDrawerOpen(true)}>
          添加站点
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="站点总数" value="32" detail="较昨日" delta="+2" tone="info" icon={<Signal className="h-6 w-6" />} />
        <MetricCard label="健康站点" value="27" detail="较昨日" delta="+1" tone="success" icon={<ShieldCheck className="h-6 w-6" />} />
        <MetricCard label="异常警告" value="3" detail="较昨日" delta="-1" tone="warning" icon={<XCircle className="h-6 w-6" />} />
        <MetricCard label="未签到站点" value="2" detail="较昨日" delta="-1" tone="danger" icon={<XCircle className="h-6 w-6" />} />
      </div>

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <SearchInput className="w-[270px]" placeholder="搜索站点名称、URL 或备注" value={query} onChange={(event) => setQuery(event.target.value)} />
          <SelectControl className="w-36" options={statusOptions} value={status} onChange={(event) => setStatus(event.target.value)} />
          <SelectControl className="w-44" options={typeOptions} value={type} onChange={(event) => setType(event.target.value)} />
          <SelectControl
            className="w-36"
            options={[
              { label: '延迟：全部', value: 'all' },
              { label: '低延迟', value: 'low' },
              { label: '高延迟', value: 'high' },
            ]}
          />
          <div className="ml-auto flex gap-2">
            <Button variant="secondary">批量检测</Button>
            <Button variant="secondary">批量导入</Button>
          </div>
        </div>
        <DataTable>
          <thead className={tableHeadClass}>
            <tr>
              <th className={tableCellClass}>站点名称</th>
              <th className={tableCellClass}>类型</th>
              <th className={tableCellClass}>状态</th>
              <th className={tableCellClass}>可用模型</th>
              <th className={tableCellClass}>延迟（平均）</th>
              <th className={tableCellClass}>余额（USD）</th>
              <th className={tableCellClass}>今日签到</th>
              <th className={tableCellClass}>操作</th>
            </tr>
          </thead>
          <tbody>
            {filteredSites.map((site) => (
              <tr key={site.id} className="hover:bg-elevated/55">
                <td className={tableCellClass}>
                  <div className="flex items-center gap-3">
                    <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-sm font-semibold text-white">
                      {site.code}
                    </span>
                    <div>
                      <div className="font-medium text-text">{site.name}</div>
                      <div className="text-xs text-muted">{site.region}</div>
                    </div>
                  </div>
                </td>
                <td className={tableCellClass}>
                  <span className="rounded-md border border-line bg-elevated px-2 py-1 text-xs text-muted">{site.type}</span>
                </td>
                <td className={tableCellClass}>
                  <StatusBadge status={site.status} />
                </td>
                <td className={tableCellClass}>{site.models.length}</td>
                <td className={tableCellClass}>{formatLatency(site.latencyMs)}</td>
                <td className={tableCellClass}>{formatCurrency(site.balanceUsd)}</td>
                <td className={tableCellClass}>
                  <StatusBadge status={site.checkinStatus} />
                </td>
                <td className={tableCellClass}>
                  <div className="flex items-center gap-2">
                    <Button variant="icon" className="h-8 w-8">
                      <BarChart3 className="h-4 w-4" />
                    </Button>
                    <Button variant="icon" className="h-8 w-8">
                      <Edit className="h-4 w-4" />
                    </Button>
                    <Button variant="icon" className="h-8 w-8">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </DataTable>
        <div className="mt-4 flex items-center justify-between text-sm text-muted">
          <span>共 {filteredSites.length} 条记录</span>
          <div className="flex items-center gap-2">
            <Button variant="secondary" className="h-9 w-9 p-0">
              1
            </Button>
            <Button variant="ghost" className="h-9 w-9 p-0">
              2
            </Button>
            <Button variant="ghost" className="h-9 w-9 p-0">
              3
            </Button>
          </div>
        </div>
      </Card>

      <Drawer
        open={drawerOpen}
        title="添加站点"
        subtitle="填写站点信息并测试连接以完成添加"
        onClose={() => setDrawerOpen(false)}
        footer={
          <div className="flex justify-end gap-3">
            <Button onClick={() => setDrawerOpen(false)}>取消</Button>
            <Button variant="primary" onClick={() => setDrawerOpen(false)}>
              保存
            </Button>
          </div>
        }
      >
        <div className="space-y-5">
          {[
            ['站点名称 *', '例如：API 站点 A'],
            ['站点 URL *', '例如：https://api.example.com/v1'],
            ['API Key *', 'sk-...'],
            ['Cookie（可选）', '输入 Cookie 字符串'],
          ].map(([label, placeholder]) => (
            <label key={label} className="block">
              <span className="text-sm font-medium text-text">{label}</span>
              <input className="mt-2 h-11 w-full rounded-lg border border-line bg-elevated px-3 text-sm outline-none focus:border-primary/55" placeholder={placeholder} />
            </label>
          ))}
          <label className="block">
            <span className="text-sm font-medium text-text">类型 *</span>
            <SelectControl className="mt-2" options={typeOptions.slice(1)} />
          </label>
          <label className="block">
            <span className="text-sm font-medium text-text">备注（可选）</span>
            <textarea className="mt-2 min-h-24 w-full resize-none rounded-lg border border-line bg-elevated px-3 py-2 text-sm outline-none focus:border-primary/55" placeholder="请输入备注信息" />
          </label>
          <Button variant="secondary" icon={testing ? <RefreshCw className="h-4 w-4 animate-spin" /> : <TestTube2 className="h-4 w-4" />} onClick={testConnection}>
            {testing ? '测试中...' : '测试连接'}
          </Button>
        </div>
      </Drawer>
    </div>
  );
}
