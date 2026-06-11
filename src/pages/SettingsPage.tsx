import { LogOut, Moon, ShieldCheck, Sun, UserCircle } from 'lucide-react';
import type { AdminUser, ThemeMode } from '../types';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { cn } from '../lib/format';

interface SettingsPageProps {
  user: AdminUser;
  theme: ThemeMode;
  onThemeChange: (theme: ThemeMode) => void;
  onLogout: () => void;
}

const roleLabels: Record<AdminUser['role'], string> = {
  owner: 'Owner',
  admin: 'Admin',
  developer: 'Developer',
  viewer: 'Viewer',
};

const statusLabels: Record<AdminUser['status'], string> = {
  active: '正常',
  inactive: '未启用',
  blocked: '已禁用',
};

export function SettingsPage({ user, theme, onThemeChange, onLogout }: SettingsPageProps) {
  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold text-text">系统设置</h1>
        <p className="mt-1 text-sm text-muted">账号、主题与会话</p>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1.15fr_0.85fr]">
        <Card title="当前账号" subtitle={user.email}>
          <div className="flex items-start gap-4">
            <span className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/12 text-primary">
              <UserCircle className="h-6 w-6" />
            </span>
            <div className="min-w-0 flex-1 space-y-3">
              <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                <div className="rounded-lg border border-line bg-elevated/50 p-4">
                  <div className="text-xs text-muted">邮箱</div>
                  <div className="mt-1 truncate text-sm font-medium text-text">{user.email}</div>
                </div>
                <div className="rounded-lg border border-line bg-elevated/50 p-4">
                  <div className="text-xs text-muted">用户 ID</div>
                  <div className="mt-1 truncate text-sm font-medium text-text">{user.id}</div>
                </div>
                <div className="rounded-lg border border-line bg-elevated/50 p-4">
                  <div className="text-xs text-muted">角色</div>
                  <div className="mt-1 text-sm font-medium text-text">{roleLabels[user.role]}</div>
                </div>
                <div className="rounded-lg border border-line bg-elevated/50 p-4">
                  <div className="text-xs text-muted">状态</div>
                  <div className="mt-1 flex items-center gap-2 text-sm font-medium text-success">
                    <span className="h-2 w-2 rounded-full bg-success" />
                    {statusLabels[user.status]}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Card>

        <Card title="偏好设置">
          <div className="space-y-5">
            <div>
              <div className="mb-3 text-sm font-medium text-text">界面主题</div>
              <div className="grid grid-cols-2 gap-2 rounded-lg border border-line bg-elevated p-1">
                <button
                  type="button"
                  className={cn(
                    'focus-ring flex h-10 items-center justify-center gap-2 rounded-md text-sm transition',
                    theme === 'dark' ? 'bg-primary text-white' : 'text-muted hover:bg-panel hover:text-text'
                  )}
                  onClick={() => onThemeChange('dark')}
                >
                  <Moon className="h-4 w-4" />
                  深色
                </button>
                <button
                  type="button"
                  className={cn(
                    'focus-ring flex h-10 items-center justify-center gap-2 rounded-md text-sm transition',
                    theme === 'light' ? 'bg-primary text-white' : 'text-muted hover:bg-panel hover:text-text'
                  )}
                  onClick={() => onThemeChange('light')}
                >
                  <Sun className="h-4 w-4" />
                  浅色
                </button>
              </div>
            </div>

            <div className="rounded-lg border border-line bg-elevated/50 p-4">
              <div className="flex items-center gap-3">
                <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-success/12 text-success">
                  <ShieldCheck className="h-5 w-5" />
                </span>
                <div>
                  <div className="text-sm font-medium text-text">会话状态</div>
                  <div className="mt-1 text-xs text-muted">已登录</div>
                </div>
              </div>
              <Button className="mt-4 w-full" variant="secondary" icon={<LogOut className="h-4 w-4" />} onClick={onLogout}>
                退出登录
              </Button>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
