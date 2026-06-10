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
import type { PageId, ThemeMode } from './types';

const pageTitles: Record<PageId, string> = {
  overview: '概览',
  sites: '站点管理',
  models: '模型管理',
  routing: '智能路由',
  checkin: '签到中心',
  quota: '额度管理',
  testing: '调用测试',
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
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const title = useMemo(() => pageTitles[activePage], [activePage]);

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
    window.localStorage.setItem('relaydeck-theme', theme);
  }, [theme]);

  const toggleTheme = () => setTheme((current) => (current === 'dark' ? 'light' : 'dark'));

  if (!isAuthenticated) {
    return <LoginPage theme={theme} onThemeToggle={toggleTheme} onLogin={() => setIsAuthenticated(true)} />;
  }

  function renderPage() {
    if (activePage === 'overview') return <OverviewPage />;
    if (activePage === 'sites') return <SitesPage />;
    if (activePage === 'models') return <ModelsPage />;
    if (activePage === 'routing') return <RoutingPage />;
    if (activePage === 'checkin' || activePage === 'quota') return <CheckinQuotaPage />;
    if (activePage === 'testing') return <TestPage />;

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
      onThemeToggle={toggleTheme}
      onPageChange={setActivePage}
    >
      {renderPage()}
    </AppLayout>
  );
}
