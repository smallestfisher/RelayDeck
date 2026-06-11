import { useEffect, useState, type ReactNode } from 'react';
import type { AdminUser, PageId, ThemeMode } from '../../types';
import { cn } from '../../lib/format';
import { Sidebar } from './Sidebar';
import { Topbar } from './Topbar';

interface AppLayoutProps {
  activePage: PageId;
  theme: ThemeMode;
  user: AdminUser;
  onThemeToggle: () => void;
  onPageChange: (page: PageId) => void;
  onLogout: () => void;
  children: ReactNode;
}

const sidebarCollapsedKey = 'relaydeck-sidebar-collapsed';

function readInitialSidebarCollapsed(): boolean {
  return window.localStorage.getItem(sidebarCollapsedKey) === 'true';
}

export function AppLayout({ activePage, theme, user, onThemeToggle, onPageChange, onLogout, children }: AppLayoutProps) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(readInitialSidebarCollapsed);

  useEffect(() => {
    window.localStorage.setItem(sidebarCollapsedKey, String(sidebarCollapsed));
  }, [sidebarCollapsed]);

  return (
    <div className="min-h-screen bg-app text-text">
      <Sidebar
        activePage={activePage}
        collapsed={sidebarCollapsed}
        onPageChange={onPageChange}
        onToggleCollapse={() => setSidebarCollapsed((current) => !current)}
      />
      <div className={cn('min-h-screen transition-[padding-left] duration-200', sidebarCollapsed ? 'pl-[76px]' : 'pl-[232px]')}>
        <Topbar theme={theme} user={user} onThemeToggle={onThemeToggle} onLogout={onLogout} onNavigate={onPageChange} />
        <main className="thin-scrollbar min-h-[calc(100vh-68px)] overflow-x-hidden px-8 py-6">{children}</main>
      </div>
    </div>
  );
}
