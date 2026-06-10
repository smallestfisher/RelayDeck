import { KeyRound, MailPlus, MoreHorizontal, ShieldCheck, SlidersHorizontal, UserCheck, UserRoundCog, Users, UserX } from 'lucide-react';
import { useMemo, useState } from 'react';
import { userRecords } from '../data/mock';
import type { UserRecord, UserRole, UserStatus } from '../types';
import { cn, formatCurrency, formatPercent } from '../lib/format';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { SearchInput, SelectControl } from '../components/ui/Controls';
import { DataTable, tableCellClass, tableHeadClass } from '../components/ui/DataTable';
import { Drawer } from '../components/ui/Drawer';
import { MetricCard } from '../components/ui/MetricCard';
import { StatusBadge } from '../components/ui/StatusBadge';

const roleLabel: Record<UserRole, string> = {
  super_admin: '超级管理员',
  admin: '管理员',
  developer: '开发者',
  viewer: '只读成员',
};

const roleOptions = [
  { label: '角色：全部', value: 'all' },
  { label: '超级管理员', value: 'super_admin' },
  { label: '管理员', value: 'admin' },
  { label: '开发者', value: 'developer' },
  { label: '只读成员', value: 'viewer' },
];

const statusOptions = [
  { label: '状态：全部', value: 'all' },
  { label: '正常', value: 'active' },
  { label: '待激活', value: 'inactive' },
  { label: '已停用', value: 'blocked' },
];

const modelOptions = [
  { label: '可用模型：全部', value: 'all' },
  { label: 'GPT-4o', value: 'GPT-4o' },
  { label: 'Claude-3.5', value: 'Claude-3.5' },
  { label: 'Gemini Pro', value: 'Gemini Pro' },
  { label: 'Embedding-3-large', value: 'Embedding-3-large' },
];

function quotaPercent(user: UserRecord): number {
  return user.monthlyQuotaUsd === 0 ? 0 : Math.min(100, (user.usedQuotaUsd / user.monthlyQuotaUsd) * 100);
}

function roleClass(role: UserRole): string {
  if (role === 'super_admin') return 'border-violet-400/30 bg-violet-500/12 text-violet-300';
  if (role === 'admin') return 'border-primary/25 bg-primary/10 text-primary';
  if (role === 'developer') return 'border-success/25 bg-success/10 text-success';
  return 'border-line bg-elevated text-muted';
}

function SectionLabel({ children }: { children: string }) {
  return <h3 className="text-sm font-semibold text-text">{children}</h3>;
}

export function UsersPage() {
  const [query, setQuery] = useState('');
  const [role, setRole] = useState('all');
  const [status, setStatus] = useState('all');
  const [model, setModel] = useState('all');
  const [selectedId, setSelectedId] = useState(userRecords[0].id);
  const [drawerOpen, setDrawerOpen] = useState(true);
  const [inviteOpen, setInviteOpen] = useState(false);

  const selectedUser = userRecords.find((user) => user.id === selectedId) ?? userRecords[0];

  const filteredUsers = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return userRecords.filter((user) => {
      const matchesQuery =
        !keyword ||
        user.name.toLowerCase().includes(keyword) ||
        user.email.toLowerCase().includes(keyword) ||
        roleLabel[user.role].toLowerCase().includes(keyword);
      const matchesRole = role === 'all' || user.role === (role as UserRole);
      const matchesStatus = status === 'all' || user.status === (status as UserStatus);
      const matchesModel = model === 'all' || user.availableModels.includes(model);
      return matchesQuery && matchesRole && matchesStatus && matchesModel;
    });
  }, [model, query, role, status]);

  const activeUsers = userRecords.filter((user) => user.status === 'active').length;
  const adminUsers = userRecords.filter((user) => user.role === 'admin' || user.role === 'super_admin').length;
  const monthlyUsage = userRecords.reduce((sum, user) => sum + user.usedQuotaUsd, 0);
  const inactiveUsers = userRecords.filter((user) => user.status !== 'active').length;

  function openUser(user: UserRecord) {
    setSelectedId(user.id);
    setDrawerOpen(true);
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-text">用户管理</h1>
          <p className="mt-1 text-sm text-muted">管理平台用户、角色权限与使用配额</p>
        </div>
        <div className="flex gap-2">
          <Button variant="secondary" icon={<SlidersHorizontal className="h-4 w-4" />}>
            角色策略
          </Button>
          <Button variant="primary" icon={<MailPlus className="h-4 w-4" />} onClick={() => setInviteOpen(true)}>
            邀请用户
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="用户总数" value={String(userRecords.length)} detail="较昨日" delta="+3" tone="info" icon={<Users className="h-6 w-6" />} />
        <MetricCard label="活跃用户" value={String(activeUsers)} detail="较昨日" delta="+5" tone="success" icon={<UserCheck className="h-6 w-6" />} />
        <MetricCard label="管理员" value={String(adminUsers)} detail="较昨日" delta="+0" tone="info" icon={<ShieldCheck className="h-6 w-6" />} />
        <MetricCard label="停用用户" value={String(inactiveUsers)} detail="较昨日" delta="-1" tone="danger" icon={<UserX className="h-6 w-6" />} />
      </div>

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <SearchInput className="w-full sm:w-[280px]" placeholder="搜索用户名、邮箱、角色" value={query} onChange={(event) => setQuery(event.target.value)} />
          <SelectControl className="w-36" options={roleOptions} value={role} onChange={(event) => setRole(event.target.value)} />
          <SelectControl className="w-36" options={statusOptions} value={status} onChange={(event) => setStatus(event.target.value)} />
          <SelectControl className="w-44" options={modelOptions} value={model} onChange={(event) => setModel(event.target.value)} />
          <Button className="ml-auto" variant="secondary" onClick={() => { setQuery(''); setRole('all'); setStatus('all'); setModel('all'); }}>
            重置
          </Button>
        </div>

        <DataTable>
          <thead className={tableHeadClass}>
            <tr>
              <th className={tableCellClass}>用户</th>
              <th className={tableCellClass}>邮箱</th>
              <th className={tableCellClass}>角色</th>
              <th className={tableCellClass}>状态</th>
              <th className={tableCellClass}>可用模型</th>
              <th className={tableCellClass}>月度配额</th>
              <th className={tableCellClass}>已用额度</th>
              <th className={tableCellClass}>API Keys</th>
              <th className={tableCellClass}>最后登录</th>
              <th className={tableCellClass}>操作</th>
            </tr>
          </thead>
          <tbody>
            {filteredUsers.map((user) => {
              const usage = quotaPercent(user);
              return (
                <tr
                  key={user.id}
                  className={cn('cursor-pointer hover:bg-elevated/55', selectedId === user.id && 'bg-primary/10')}
                  onClick={() => openUser(user)}
                >
                  <td className={tableCellClass}>
                    <div className="flex items-center gap-3">
                      <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary/15 text-sm font-semibold text-primary">
                        {user.avatarText}
                      </span>
                      <div>
                        <div className="font-medium text-text">{user.name}</div>
                        <div className="text-xs text-muted">{user.organization}</div>
                      </div>
                    </div>
                  </td>
                  <td className={tableCellClass}>{user.email}</td>
                  <td className={tableCellClass}>
                    <span className={cn('rounded-md border px-2 py-1 text-xs font-medium', roleClass(user.role))}>{roleLabel[user.role]}</span>
                  </td>
                  <td className={tableCellClass}>
                    <StatusBadge status={user.status} />
                  </td>
                  <td className={tableCellClass}>{user.availableModels.length}</td>
                  <td className={tableCellClass}>{formatCurrency(user.monthlyQuotaUsd)}</td>
                  <td className={tableCellClass}>
                    <div className="min-w-32">
                      <div className="flex items-center justify-between gap-3 text-xs">
                        <span>{formatCurrency(user.usedQuotaUsd)}</span>
                        <span className="text-muted">{formatPercent(usage, 1)}</span>
                      </div>
                      <span className="mt-1 block h-1.5 overflow-hidden rounded-full bg-line">
                        <span className="block h-full rounded-full bg-primary" style={{ width: `${usage}%` }} />
                      </span>
                    </div>
                  </td>
                  <td className={tableCellClass}>{user.apiKeyCount}</td>
                  <td className={tableCellClass}>{user.lastLogin}</td>
                  <td className={tableCellClass}>
                    <Button variant="icon" className="h-8 w-8" onClick={(event) => { event.stopPropagation(); openUser(user); }}>
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </DataTable>

        <div className="mt-4 flex items-center justify-between text-sm text-muted">
          <span>共 {filteredUsers.length} 条记录</span>
          <div className="flex items-center gap-2">
            <Button variant="secondary" className="h-9 w-9 p-0">1</Button>
            <Button variant="ghost" className="h-9 w-9 p-0">2</Button>
            <Button variant="ghost" className="h-9 w-9 p-0">3</Button>
          </div>
        </div>
      </Card>

      <Drawer
        open={drawerOpen}
        title="用户详情"
        subtitle={selectedUser.email}
        onClose={() => setDrawerOpen(false)}
        footer={
          <div className="flex justify-between gap-3">
            <Button variant="secondary" icon={<UserRoundCog className="h-4 w-4" />}>编辑用户</Button>
            <Button variant="danger" icon={<KeyRound className="h-4 w-4" />}>重置密钥</Button>
          </div>
        }
      >
        <div className="space-y-6">
          <div className="flex items-center gap-4">
            <span className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/15 text-lg font-semibold text-primary">
              {selectedUser.avatarText}
            </span>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-semibold text-text">{selectedUser.name}</h2>
                <span className={cn('rounded-md border px-2 py-1 text-xs font-medium', roleClass(selectedUser.role))}>{roleLabel[selectedUser.role]}</span>
              </div>
              <div className="mt-1 text-sm text-muted">{selectedUser.note}</div>
              <div className="mt-2"><StatusBadge status={selectedUser.status} /></div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 text-sm">
            {[
              ['用户 ID', selectedUser.id],
              ['创建时间', selectedUser.joinedAt],
              ['所属组织', selectedUser.organization],
              ['速率限制', selectedUser.rateLimit],
            ].map(([label, value]) => (
              <div key={label} className="rounded-lg border border-line bg-elevated/45 p-3">
                <div className="text-xs text-muted">{label}</div>
                <div className="mt-1 font-medium text-text">{value}</div>
              </div>
            ))}
          </div>

          <div className="border-t border-line pt-5">
            <SectionLabel>权限与配额</SectionLabel>
            <div className="mt-3 flex flex-wrap gap-2">
              {selectedUser.availableModels.map((item) => (
                <span key={item} className="rounded-md border border-line bg-elevated px-2.5 py-1 text-xs text-muted">{item}</span>
              ))}
            </div>
            <div className="mt-4 rounded-lg bg-elevated/55 p-4">
              <div className="mb-2 flex items-center justify-between text-sm">
                <span className="text-muted">已用额度</span>
                <span className="font-medium text-text">
                  {formatCurrency(selectedUser.usedQuotaUsd)} / {formatCurrency(selectedUser.monthlyQuotaUsd)}
                </span>
              </div>
              <span className="block h-2 overflow-hidden rounded-full bg-line">
                <span className="block h-full rounded-full bg-primary" style={{ width: `${quotaPercent(selectedUser)}%` }} />
              </span>
            </div>
          </div>

          <div className="border-t border-line pt-5">
            <SectionLabel>最近活动</SectionLabel>
            <div className="mt-3 space-y-3">
              {selectedUser.activity.map((item) => (
                <div key={item.id} className="rounded-lg border border-line bg-elevated/35 p-3 text-sm">
                  <div className="flex items-center justify-between gap-3">
                    <span className="font-medium text-text">{item.title}</span>
                    <span className="text-xs text-muted">{item.time}</span>
                  </div>
                  <div className="mt-1 text-xs text-muted">{item.description}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </Drawer>

      {inviteOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/45 px-4">
          <section className="w-full max-w-lg rounded-xl border border-line bg-panel p-6 shadow-2xl">
            <div className="mb-5 flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold text-text">邀请新用户</h2>
                <p className="mt-1 text-sm text-muted">发送邀请邮件并预设角色、额度和模型权限</p>
              </div>
              <Button variant="ghost" className="h-9 w-9 p-0" onClick={() => setInviteOpen(false)}>×</Button>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="sm:col-span-2">
                <span className="text-sm font-medium text-text">邮箱 *</span>
                <input className="mt-2 h-11 w-full rounded-lg border border-line bg-elevated px-3 text-sm outline-none focus:border-primary/55" placeholder="请输入用户邮箱地址" />
              </label>
              <label>
                <span className="text-sm font-medium text-text">角色 *</span>
                <SelectControl className="mt-2" options={roleOptions.slice(1)} defaultValue="developer" />
              </label>
              <label>
                <span className="text-sm font-medium text-text">初始配额（USD）*</span>
                <input className="mt-2 h-11 w-full rounded-lg border border-line bg-elevated px-3 text-sm outline-none focus:border-primary/55" placeholder="1000" />
              </label>
              <label className="sm:col-span-2">
                <span className="text-sm font-medium text-text">可用模型 *</span>
                <SelectControl className="mt-2" options={modelOptions.slice(1)} defaultValue="GPT-4o" />
              </label>
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <Button onClick={() => setInviteOpen(false)}>取消</Button>
              <Button variant="primary" icon={<MailPlus className="h-4 w-4" />} onClick={() => setInviteOpen(false)}>发送邀请</Button>
            </div>
          </section>
        </div>
      )}
    </div>
  );
}
