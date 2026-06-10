import {
  CalendarCheck,
  ClipboardList,
  FlaskConical,
  KeyRound,
  LayoutDashboard,
  PanelLeftClose,
  Route,
  Server,
  Settings,
  Users,
  WalletCards,
  Boxes,
} from 'lucide-react';
import type { PageId } from '../../types';
import { cn } from '../../lib/format';
import { Logo } from '../ui/Logo';

interface SidebarProps {
  activePage: PageId;
  onPageChange: (page: PageId) => void;
}

const navItems: Array<{ id: PageId; label: string; icon: typeof LayoutDashboard }> = [
  { id: 'overview', label: '概览', icon: LayoutDashboard },
  { id: 'sites', label: '站点管理', icon: Server },
  { id: 'models', label: '模型管理', icon: Boxes },
  { id: 'routing', label: '智能路由', icon: Route },
  { id: 'checkin', label: '签到中心', icon: CalendarCheck },
  { id: 'quota', label: '额度管理', icon: WalletCards },
  { id: 'testing', label: '调用测试', icon: FlaskConical },
  { id: 'users', label: '用户管理', icon: Users },
  { id: 'apiKeys', label: 'API Keys', icon: KeyRound },
  { id: 'logs', label: '任务日志', icon: ClipboardList },
  { id: 'settings', label: '系统设置', icon: Settings },
];

export function Sidebar({ activePage, onPageChange }: SidebarProps) {
  return (
    <aside className="fixed inset-y-0 left-0 z-30 flex w-[232px] flex-col border-r border-line bg-panel/86 backdrop-blur-xl">
      <div className="px-6 py-7">
        <Logo />
      </div>
      <nav className="thin-scrollbar flex-1 space-y-1 overflow-y-auto px-3">
        {navItems.map((item) => {
          const Icon = item.icon;
          const active = activePage === item.id;

          return (
            <button
              key={item.id}
              type="button"
              onClick={() => onPageChange(item.id)}
              className={cn(
                'focus-ring flex h-12 w-full items-center gap-3 rounded-lg border px-4 text-left text-sm font-medium transition',
                active
                  ? 'border-primary/45 bg-primary/12 text-primary shadow-[inset_3px_0_0_rgb(var(--color-primary))]'
                  : 'border-transparent text-muted hover:bg-elevated hover:text-text'
              )}
            >
              <Icon className="h-5 w-5 shrink-0" />
              <span className="truncate">{item.label}</span>
            </button>
          );
        })}
      </nav>
      <div className="px-4 pb-4">
        <div className="rounded-lg border border-line bg-elevated/70 p-4">
          <div className="flex items-center gap-2 text-sm text-text">
            <span className="h-2.5 w-2.5 rounded-full bg-success shadow-[0_0_14px_rgb(34_197_94/0.8)]" />
            系统运行状态
          </div>
          <div className="mt-3 text-xs text-muted">所有系统运行正常</div>
          <div className="mt-4 flex items-center justify-between border-t border-line pt-3 text-xs text-muted">
            <span>版本</span>
            <span>v1.4.2</span>
          </div>
        </div>
        <div className="mt-3 flex justify-end">
          <button type="button" className="rounded-md p-1.5 text-muted hover:bg-elevated hover:text-text" aria-label="收起侧边栏">
            <PanelLeftClose className="h-4 w-4" />
          </button>
        </div>
      </div>
    </aside>
  );
}
