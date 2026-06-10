import type { ReactNode } from 'react';
import type { AdminUser, PageId, ThemeMode } from '../../types';
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

export function AppLayout({ activePage, theme, user, onThemeToggle, onPageChange, onLogout, children }: AppLayoutProps) {
  return (
    <div className="min-h-screen bg-app text-text">
      <Sidebar activePage={activePage} onPageChange={onPageChange} />
      <div className="min-h-screen pl-[232px]">
        <Topbar theme={theme} user={user} onThemeToggle={onThemeToggle} onLogout={onLogout} />
        <main className="thin-scrollbar min-h-[calc(100vh-68px)] overflow-x-hidden px-8 py-6">{children}</main>
      </div>
    </div>
  );
}
