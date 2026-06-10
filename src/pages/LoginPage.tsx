import { Building2, Check, Github, Lock, Mail, Moon, Server, Sun, WalletCards, Zap } from 'lucide-react';
import type { ThemeMode } from '../types';
import { Button } from '../components/ui/Button';
import { Logo } from '../components/ui/Logo';
import { cn } from '../lib/format';

interface LoginPageProps {
  theme: ThemeMode;
  onThemeToggle: () => void;
  onLogin: () => void;
}

const features = [
  { label: '多模型接入', icon: Server, x: '18%', y: '24%' },
  { label: '线路管理', icon: Server, x: '78%', y: '30%' },
  { label: '实时监控', icon: Zap, x: '12%', y: '58%' },
  { label: '数据洞察', icon: WalletCards, x: '80%', y: '62%' },
  { label: '智能路由', icon: Check, x: '38%', y: '78%' },
];

export function LoginPage({ theme, onThemeToggle, onLogin }: LoginPageProps) {
  return (
    <main className="min-h-screen bg-app p-5 text-text">
      <div className="relative min-h-[calc(100vh-40px)] overflow-hidden rounded-2xl border border-line bg-panel/55 px-10 py-9 shadow-glow">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_32%_70%,rgb(37_99_235/0.28),transparent_24rem)]" />
        <div className="relative z-10 flex items-start justify-between">
          <Logo />
          <button
            type="button"
            onClick={onThemeToggle}
            className="flex h-11 w-24 items-center justify-between rounded-full border border-line bg-elevated p-1 text-muted"
            aria-label="切换主题"
          >
            <span
              className={cn(
                'flex h-9 w-11 items-center justify-center rounded-full transition',
                theme === 'dark' ? 'bg-primary text-white' : 'text-muted'
              )}
            >
              <Moon className="h-5 w-5" />
            </span>
            <span
              className={cn(
                'flex h-9 w-11 items-center justify-center rounded-full transition',
                theme === 'light' ? 'bg-primary text-white' : 'text-muted'
              )}
            >
              <Sun className="h-5 w-5" />
            </span>
          </button>
        </div>

        <div className="relative z-10 grid min-h-[calc(100vh-180px)] grid-cols-1 items-center gap-12 xl:grid-cols-[1fr_520px]">
          <section className="max-w-3xl">
            <h1 className="max-w-2xl text-5xl font-semibold leading-tight text-text">
              统一管理，智能调度，
              <span className="block text-primary">让每一次调用更高效</span>
            </h1>
            <p className="mt-7 max-w-2xl text-lg leading-9 text-muted">
              集中管理多模型、多线路资源，实时监控调用状态与用量，智能路由与故障切换，助力团队稳定、高效地交付 AI 能力。
            </p>
            <div className="relative mt-12 h-[330px] max-w-[700px]">
              <div className="absolute left-1/2 top-1/2 h-48 w-48 -translate-x-1/2 -translate-y-1/2 rounded-full border border-primary/20 bg-primary/10 blur-sm" />
              <div className="absolute left-1/2 top-1/2 flex h-28 w-28 -translate-x-1/2 -translate-y-1/2 items-center justify-center">
                <Logo compact />
              </div>
              <div className="absolute left-1/2 top-1/2 h-64 w-64 -translate-x-1/2 -translate-y-1/2 rounded-full border border-dashed border-primary/35" />
              <div className="absolute left-1/2 top-1/2 h-80 w-80 -translate-x-1/2 -translate-y-1/2 rounded-full border border-primary/10" />
              {features.map((feature) => {
                const Icon = feature.icon;
                return (
                  <div
                    key={feature.label}
                    className="absolute flex -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-2 text-center text-sm text-muted"
                    style={{ left: feature.x, top: feature.y }}
                  >
                    <span className="flex h-12 w-12 items-center justify-center rounded-full border border-primary/30 bg-primary/10 text-primary">
                      <Icon className="h-5 w-5" />
                    </span>
                    {feature.label}
                  </div>
                );
              })}
            </div>
          </section>

          <section className="surface-card rounded-xl p-8">
            <div className="grid grid-cols-2 border-b border-line text-center text-lg font-semibold">
              <button type="button" className="border-b-2 border-primary pb-5 text-text">
                登录
              </button>
              <button type="button" className="pb-5 text-muted">
                注册
              </button>
            </div>
            <div className="mt-7 space-y-5">
              <label className="block">
                <span className="text-sm font-medium text-text">邮箱地址</span>
                <span className="mt-2 flex h-12 items-center rounded-lg border border-line bg-elevated px-3">
                  <Mail className="h-5 w-5 text-muted" />
                  <input className="ml-3 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-muted/70" placeholder="请输入邮箱地址" />
                </span>
              </label>
              <label className="block">
                <span className="text-sm font-medium text-text">密码</span>
                <span className="mt-2 flex h-12 items-center rounded-lg border border-line bg-elevated px-3">
                  <Lock className="h-5 w-5 text-muted" />
                  <input type="password" className="ml-3 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-muted/70" placeholder="请输入密码" />
                </span>
              </label>
              <div className="flex items-center justify-between text-sm">
                <label className="flex items-center gap-2 text-muted">
                  <span className="flex h-5 w-5 items-center justify-center rounded border border-primary bg-primary text-white">
                    <Check className="h-3.5 w-3.5" />
                  </span>
                  记住我
                </label>
                <button type="button" className="text-primary">
                  忘记密码?
                </button>
              </div>
              <Button variant="primary" size="lg" className="w-full" onClick={onLogin}>
                登录
              </Button>
              <div className="flex items-center gap-4 text-sm text-muted">
                <span className="h-px flex-1 bg-line" />
                或使用以下方式登录
                <span className="h-px flex-1 bg-line" />
              </div>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <Button variant="secondary" icon={<span className="font-semibold text-primary">G</span>}>
                  Google
                </Button>
                <Button variant="secondary" icon={<Github className="h-5 w-5" />}>
                  GitHub
                </Button>
                <Button variant="secondary" icon={<Building2 className="h-5 w-5" />}>
                  企业 SSO
                </Button>
              </div>
              <p className="pt-2 text-center text-sm text-muted">
                登录即表示您同意 <span className="text-primary">服务条款</span> 和 <span className="text-primary">隐私政策</span>
              </p>
            </div>
          </section>
        </div>
        <div className="relative z-10 text-sm text-muted">© 2025 RelayDeck. 保留所有权利。</div>
      </div>
    </main>
  );
}
