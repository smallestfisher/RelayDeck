import { useEffect, useMemo, useState } from 'react';
import { AppLayout } from './components/layout/AppLayout';
import { CheckinQuotaPage } from './pages/CheckinQuotaPage';
import { EmptyPage } from './pages/EmptyPage';
import { LoginPage } from './pages/LoginPage';
import { ModelsPage } from './pages/ModelsPage';
import { OverviewPage } from './pages/OverviewPage';
import { RoutingPage } from './pages/RoutingPage';
import { SitesPage } from './pages/SitesPage';
import { TestPage } from './pages/TestPage';
import { UsersPage } from './pages/UsersPage';
import { ApiKeysPage } from './pages/ApiKeysPage';
import type { AdminUser, PageId, ThemeMode } from './types';

const pageTitles: Record<PageId, string> = {
  overview: '概览',
  sites: '站点管理',
  models: '模型管理',
  routing: '智能路由',
  checkin: '签到中心',
  quota: '额度管理',
  testing: '调用测试',
  users: '用户管理',
  apiKeys: 'API Keys',
  logs: '任务日志',
  settings: '系统设置',
};

function readInitialTheme(): ThemeMode {
  const saved = window.localStorage.getItem('relaydeck-theme');
  return saved === 'light' || saved === 'dark' ? saved : 'dark';
}

export default function App() {
  const [theme, setTheme] = useState<ThemeMode>(readInitialTheme);
  const [activePage, setActivePage] = useState<PageId>('overview');
  const [adminUser, setAdminUser] = useState<AdminUser | null>(null);
  const [authChecked, setAuthChecked] = useState(false);
  const title = useMemo(() => pageTitles[activePage], [activePage]);

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
    window.localStorage.setItem('relaydeck-theme', theme);
  }, [theme]);

  useEffect(() => {
    let cancelled = false;

    async function restoreSession() {
      try {
        const response = await fetch('/api/admin/auth/me', { credentials: 'include' });
        if (!response.ok) return;
        const payload = (await response.json()) as { user?: AdminUser };
        if (!cancelled && payload.user) {
          setAdminUser(payload.user);
        }
      } catch {
        // Staying on the login page is the correct fallback when the backend is unavailable.
      } finally {
        if (!cancelled) {
          setAuthChecked(true);
        }
      }
    }

    restoreSession();
    return () => {
      cancelled = true;
    };
  }, []);

  const toggleTheme = () => setTheme((current) => (current === 'dark' ? 'light' : 'dark'));

  async function handleLogout() {
    try {
      await fetch('/api/admin/auth/logout', { method: 'POST', credentials: 'include' });
    } finally {
      setAdminUser(null);
      setActivePage('overview');
    }
  }

  if (!authChecked) {
    return <main className="flex min-h-screen items-center justify-center bg-app text-sm text-muted">正在检查登录状态...</main>;
  }

  if (!adminUser) {
    return <LoginPage theme={theme} onThemeToggle={toggleTheme} onLogin={setAdminUser} />;
  }

  function renderPage() {
    if (activePage === 'overview') return <OverviewPage />;
    if (activePage === 'sites') return <SitesPage />;
    if (activePage === 'models') return <ModelsPage />;
    if (activePage === 'routing') return <RoutingPage />;
    if (activePage === 'checkin' || activePage === 'quota') return <CheckinQuotaPage />;
    if (activePage === 'testing') return <TestPage />;
    if (activePage === 'users') return <UsersPage />;
    if (activePage === 'apiKeys') return <ApiKeysPage />;

    return (
      <EmptyPage
        title={title}
        description="首版 UI 原型先保留该入口，后续可以继续扩展具体表格、筛选、策略配置和审计视图。"
        onReturn={() => setActivePage('overview')}
      />
    );
  }

  return (
    <AppLayout
      activePage={activePage}
      theme={theme}
      user={adminUser}
      onThemeToggle={toggleTheme}
      onPageChange={setActivePage}
      onLogout={handleLogout}
    >
      {renderPage()}
    </AppLayout>
  );
}
