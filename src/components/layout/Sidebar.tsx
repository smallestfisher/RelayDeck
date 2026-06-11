import {
  CalendarCheck,
  ClipboardList,
  FlaskConical,
  KeyRound,
  LayoutDashboard,
  PanelLeftClose,
  PanelLeftOpen,
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
  collapsed: boolean;
  onPageChange: (page: PageId) => void;
  onToggleCollapse: () => void;
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

export function Sidebar({ activePage, collapsed, onPageChange, onToggleCollapse }: SidebarProps) {
  return (
    <aside
      className={cn(
        'fixed inset-y-0 left-0 z-30 flex flex-col border-r border-line bg-panel/86 backdrop-blur-xl transition-[width] duration-200',
        collapsed ? 'w-[76px]' : 'w-[232px]'
      )}
    >
      <div className={cn('py-7', collapsed ? 'px-4' : 'px-6')}>
        <Logo compact={collapsed} className={collapsed ? 'justify-center' : undefined} />
      </div>
      <nav className={cn('thin-scrollbar flex-1 space-y-1 overflow-y-auto', collapsed ? 'px-2' : 'px-3')}>
        {navItems.map((item) => {
          const Icon = item.icon;
          const active = activePage === item.id;

          return (
            <button
              key={item.id}
              type="button"
              title={collapsed ? item.label : undefined}
              onClick={() => onPageChange(item.id)}
              className={cn(
                'focus-ring flex h-12 w-full items-center rounded-lg border text-sm font-medium transition',
                collapsed ? 'justify-center px-0' : 'gap-3 px-4 text-left',
                active
                  ? 'border-primary/45 bg-primary/12 text-primary shadow-[inset_3px_0_0_rgb(var(--color-primary))]'
                  : 'border-transparent text-muted hover:bg-elevated hover:text-text'
              )}
            >
              <Icon className="h-5 w-5 shrink-0" />
              {!collapsed && <span className="truncate">{item.label}</span>}
            </button>
          );
        })}
      </nav>
      <div className={cn('pb-4', collapsed ? 'px-3' : 'px-4')}>
        {!collapsed && (
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
        )}
        <div className={cn('mt-3 flex', collapsed ? 'justify-center' : 'justify-end')}>
          <button
            type="button"
            className="rounded-md p-1.5 text-muted hover:bg-elevated hover:text-text"
            aria-label={collapsed ? '展开侧边栏' : '收起侧边栏'}
            title={collapsed ? '展开侧边栏' : '收起侧边栏'}
            onClick={onToggleCollapse}
          >
            {collapsed ? <PanelLeftOpen className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
          </button>
        </div>
      </div>
    </aside>
  );
}
