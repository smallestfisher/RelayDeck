import { Bell, ChevronDown, Moon, RefreshCw, Search, Sun } from 'lucide-react';
import type { ThemeMode } from '../../types';
import { Button } from '../ui/Button';

interface TopbarProps {
  theme: ThemeMode;
  onThemeToggle: () => void;
}

export function Topbar({ theme, onThemeToggle }: TopbarProps) {
  return (
    <header className="sticky top-0 z-20 flex h-[68px] items-center justify-between border-b border-line bg-app/82 px-8 backdrop-blur-xl">
      <div className="relative w-full max-w-[380px]">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
        <input
          className="focus-ring h-10 w-full rounded-lg border border-line bg-elevated pl-10 pr-14 text-sm text-text outline-none placeholder:text-muted/70 focus:border-primary/55"
          placeholder="搜索站点 / 模型 / 日志 / 任务"
        />
        <span className="absolute right-3 top-1/2 -translate-y-1/2 rounded-md border border-line bg-panel px-1.5 py-0.5 text-xs text-muted">
          ⌘ K
        </span>
      </div>
      <div className="ml-8 flex shrink-0 items-center gap-5">
        <button type="button" className="flex items-center gap-2 text-sm text-muted hover:text-text">
          系统检测频率：
          <span className="font-semibold text-text">5 分钟</span>
          <ChevronDown className="h-4 w-4" />
        </button>
        <div className="flex h-9 items-center gap-2 rounded-full border border-success/25 bg-success/10 px-4 text-sm font-medium text-success">
          <span className="h-2.5 w-2.5 rounded-full bg-success" />
          系统运行中
        </div>
        <Button variant="icon" aria-label="通知">
          <span className="relative">
            <Bell className="h-5 w-5" />
            <span className="absolute -right-2 -top-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-danger px-1 text-[11px] font-semibold text-white">
              3
            </span>
          </span>
        </Button>
        <div className="flex items-center rounded-full border border-line bg-elevated p-1">
          <button
            type="button"
            className="flex h-7 w-8 items-center justify-center rounded-full text-muted hover:text-text"
            onClick={onThemeToggle}
            aria-label="切换浅色主题"
          >
            <Sun className="h-4 w-4" />
          </button>
          <button
            type="button"
            className="flex h-7 w-8 items-center justify-center rounded-full bg-primary text-white"
            onClick={onThemeToggle}
            aria-label="切换深色主题"
          >
            {theme === 'dark' ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
          </button>
        </div>
        <Button variant="icon" aria-label="刷新">
          <RefreshCw className="h-5 w-5" />
        </Button>
        <button type="button" className="flex items-center gap-3 text-sm font-medium text-text">
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-slate-300 to-slate-600 text-sm text-white">
            管
          </span>
          管理员
          <ChevronDown className="h-4 w-4 text-muted" />
        </button>
      </div>
    </header>
  );
}
