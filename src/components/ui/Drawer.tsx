import type { ReactNode } from 'react';
import { X } from 'lucide-react';
import { cn } from '../../lib/format';
import { Button } from './Button';

interface DrawerProps {
  open: boolean;
  title: string;
  subtitle?: string;
  onClose: () => void;
  footer?: ReactNode;
  children: ReactNode;
}

export function Drawer({ open, title, subtitle, onClose, footer, children }: DrawerProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50">
      <button className="absolute inset-0 bg-slate-950/35" type="button" onClick={onClose} aria-label="关闭抽屉遮罩" />
      <aside
        className={cn(
          'absolute right-0 top-0 flex h-full w-full max-w-[430px] flex-col border-l border-line bg-panel shadow-2xl'
        )}
      >
        <header className="flex items-start justify-between gap-4 border-b border-line px-6 py-5">
          <div>
            <h2 className="text-xl font-semibold text-text">{title}</h2>
            {subtitle && <p className="mt-1 text-sm text-muted">{subtitle}</p>}
          </div>
          <Button variant="ghost" className="h-9 w-9 p-0" onClick={onClose} aria-label="关闭">
            <X className="h-5 w-5" />
          </Button>
        </header>
        <div className="thin-scrollbar flex-1 overflow-y-auto px-6 py-5">{children}</div>
        {footer && <footer className="border-t border-line px-6 py-4">{footer}</footer>}
      </aside>
    </div>
  );
}
