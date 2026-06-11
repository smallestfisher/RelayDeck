import { useEffect, type ReactNode } from 'react';
import { X } from 'lucide-react';
import { cn } from '../../lib/format';
import { Button } from './Button';

interface DrawerProps {
  open: boolean;
  title: string;
  subtitle?: string;
  variant?: 'drawer' | 'modal';
  onClose: () => void;
  footer?: ReactNode;
  children: ReactNode;
}

export function Drawer({ open, title, subtitle, variant = 'drawer', onClose, footer, children }: DrawerProps) {
  useEffect(() => {
    if (!open) return;

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        onClose();
      }
    }

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [open, onClose]);

  if (!open) return null;

  const isModal = variant === 'modal';

  return (
    <div className="fixed inset-0 z-50">
      <div className="fixed inset-0 z-0 bg-slate-950/35" aria-hidden="true" />
      <aside
        className={cn(
          'relative z-10 flex flex-col border bg-panel shadow-2xl',
          isModal
            ? 'left-1/2 top-1/2 h-auto max-h-[90vh] w-full max-w-2xl -translate-x-1/2 -translate-y-1/2 rounded-xl border-line'
            : 'right-0 top-0 h-full w-full max-w-[430px] border-l border-line'
        )}
      >
        <header className="flex items-start justify-between gap-4 border-b border-line px-6 py-5">
          <div>
            <h2 className="text-xl font-semibold text-text">{title}</h2>
            {subtitle && <p className="mt-1 text-sm text-muted">{subtitle}</p>}
          </div>
          <Button variant="ghost" icon={<X className="h-5 w-5" />} onClick={onClose} aria-label="关闭" />
        </header>
        <div className="thin-scrollbar flex-1 overflow-y-auto px-6 py-5">{children}</div>
        {footer && <footer className="border-t border-line px-6 py-4">{footer}</footer>}
      </aside>
    </div>
  );
}
