import { Bell, ChevronDown, LogOut, Moon, RefreshCw, Search, Settings, Sun } from 'lucide-react';
import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import type { AdminUser, PageId, ThemeMode } from '../../types';
import { Button } from '../ui/Button';

interface TopbarProps {
  theme: ThemeMode;
  user: AdminUser;
  onThemeToggle: () => void;
  onLogout: () => void;
  onNavigate: (page: PageId) => void;
}

export function Topbar({ theme, user, onThemeToggle, onLogout, onNavigate }: TopbarProps) {
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [menuPosition, setMenuPosition] = useState({ top: 68, right: 32 });
  const triggerRef = useRef<HTMLButtonElement>(null);
  const avatarText = user.email.slice(0, 1).toUpperCase();
  const userName = user.email.split('@')[0];

  function updateMenuPosition() {
    const rect = triggerRef.current?.getBoundingClientRect();
    if (!rect) return;

    setMenuPosition({
      top: rect.bottom + 8,
      right: Math.max(16, window.innerWidth - rect.right),
    });
  }

  useEffect(() => {
    if (!dropdownOpen) return;

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setDropdownOpen(false);
      }
    }

    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [dropdownOpen]);

  useLayoutEffect(() => {
    if (!dropdownOpen) return;

    updateMenuPosition();
    window.addEventListener('resize', updateMenuPosition);
    window.addEventListener('scroll', updateMenuPosition, true);

    return () => {
      window.removeEventListener('resize', updateMenuPosition);
      window.removeEventListener('scroll', updateMenuPosition, true);
    };
  }, [dropdownOpen]);

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
            className={`flex h-7 w-8 items-center justify-center rounded-full transition ${theme === 'light' ? 'bg-primary text-white' : 'text-muted hover:text-text'}`}
            onClick={onThemeToggle}
            aria-label="切换浅色主题"
          >
            <Sun className="h-4 w-4" />
          </button>
          <button
            type="button"
            className={`flex h-7 w-8 items-center justify-center rounded-full transition ${theme === 'dark' ? 'bg-primary text-white' : 'text-muted hover:text-text'}`}
            onClick={onThemeToggle}
            aria-label="切换深色主题"
          >
            <Moon className="h-4 w-4" />
          </button>
        </div>
        <Button variant="icon" aria-label="刷新">
          <RefreshCw className="h-5 w-5" />
        </Button>
        <div className="relative">
          <button
            ref={triggerRef}
            type="button"
            className="flex items-center gap-3 text-sm font-medium text-text hover:opacity-80"
            onClick={() => {
              updateMenuPosition();
              setDropdownOpen((open) => !open);
            }}
            aria-expanded={dropdownOpen}
            aria-haspopup="menu"
          >
            <span className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-slate-300 to-slate-600 text-sm text-white">
              {avatarText}
            </span>
            <span>{userName}</span>
            <ChevronDown className="h-4 w-4 text-muted" />
          </button>
          {dropdownOpen &&
            createPortal(
              <>
                <button
                  type="button"
                  className="fixed inset-0 z-40 cursor-default"
                  onClick={() => setDropdownOpen(false)}
                  aria-label="关闭菜单"
                />
                <div
                  className="fixed z-50 w-48 rounded-lg border border-line bg-panel shadow-xl"
                  style={{ top: menuPosition.top, right: menuPosition.right }}
                  role="menu"
                >
                  <button
                    type="button"
                    className="flex w-full items-center gap-3 px-4 py-3 text-sm text-text hover:bg-elevated"
                    onClick={() => {
                      setDropdownOpen(false);
                      onNavigate('settings');
                    }}
                    role="menuitem"
                  >
                    <Settings className="h-4 w-4" />
                    系统设置
                  </button>
                  <button
                    type="button"
                    className="flex w-full items-center gap-3 border-t border-line px-4 py-3 text-sm text-danger hover:bg-elevated"
                    onClick={() => {
                      setDropdownOpen(false);
                      onLogout();
                    }}
                    role="menuitem"
                  >
                    <LogOut className="h-4 w-4" />
                    退出登录
                  </button>
                </div>
              </>,
              document.body
            )}
        </div>
      </div>
    </header>
  );
}
